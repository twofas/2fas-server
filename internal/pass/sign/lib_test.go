package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ecdsaSigningMethodWithStaticKey struct {
	privateKey *ecdsa.PrivateKey
}

func (e ecdsaSigningMethodWithStaticKey) Verify(signingString string, sig []byte, key interface{}) error {
	panic("not needed")
}

func (e ecdsaSigningMethodWithStaticKey) Sign(signingString string, key interface{}) ([]byte, error) {
	return jwt.SigningMethodES256.Sign(signingString, e.privateKey)
}

func (e ecdsaSigningMethodWithStaticKey) Alg() string {
	return jwt.SigningMethodES256.Alg()
}

func TestSignAndVerifyHappyPath(t *testing.T) {
	srv := createTestService(t)

	now := time.Now()

	token, err := srv.SignAndEncode(Message{
		ConnectionID:   uuid.New().String(),
		ExpiresAt:      now.Add(time.Hour),
		ConnectionType: ConnectionTypeBrowserExtensionProxy,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := srv.CanI(token, ConnectionTypeBrowserExtensionProxy); err != nil {
		t.Fatal(err)
	}
}

func createTestService(t *testing.T) Service {
	t.Helper()

	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	srv := Service{
		publicKey: &pk.PublicKey,
		signingMethod: ecdsaSigningMethodWithStaticKey{
			privateKey: pk,
		},
	}
	return srv
}

func TestSignAndVerify(t *testing.T) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String("http://localhost:4566"),
	})
	if err != nil {
		t.Fatal(err)
	}
	kmsClient := kms.New(sess)
	srv := createTestService(t)
	now := time.Now()

	tests := []struct {
		name          string
		tokenFn       func() string
		expectedError error
	}{
		{
			name: "not even jwt token",
			tokenFn: func() string {
				return "xxx"
			},
			expectedError: jwt.ErrTokenMalformed,
		},
		{
			name: "token is expired",
			tokenFn: func() string {
				token, err := srv.SignAndEncode(Message{
					ConnectionID:   uuid.New().String(),
					ExpiresAt:      now.Add(-time.Hour),
					ConnectionType: ConnectionTypeBrowserExtensionProxy,
				})
				if err != nil {
					t.Fatal(err)
				}
				return token
			},
			expectedError: jwt.ErrTokenExpired,
		},
		{
			name: "invalid claim",
			tokenFn: func() string {
				token, err := srv.SignAndEncode(Message{
					ConnectionID:   uuid.New().String(),
					ExpiresAt:      now.Add(time.Hour),
					ConnectionType: ConnectionTypeBrowserExtensionWait,
				})
				if err != nil {
					t.Fatal(err)
				}
				return token
			},
			expectedError: ErrInvalidClaims,
		},
		{
			name: "invalid signature",
			tokenFn: func() string {
				resp, err := kmsClient.CreateKey(&kms.CreateKeyInput{
					KeySpec:  aws.String("ECC_NIST_P256"),
					KeyUsage: aws.String("SIGN_VERIFY"),
				})
				if err != nil {
					t.Fatal(err)
				}
				serviceWithAnotherKey, err := NewService(*resp.KeyMetadata.KeyId, kmsClient)
				if err != nil {
					t.Fatal(err)
				}

				token, err := serviceWithAnotherKey.SignAndEncode(Message{
					ConnectionID:   uuid.New().String(),
					ExpiresAt:      now.Add(-time.Hour),
					ConnectionType: ConnectionTypeBrowserExtensionProxy,
				})
				if err != nil {
					t.Fatal(err)
				}
				return token
			},
			expectedError: jwt.ErrTokenSignatureInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.tokenFn()
			err := srv.CanI(token, ConnectionTypeBrowserExtensionProxy)
			if err == nil {
				t.Fatalf("Expected error %v, got nil", tc.expectedError)
			}
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("Expected error %v, got %v", tc.expectedError, err)
			}
		})
	}

}
