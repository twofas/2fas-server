package pass

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/pass/sign"
)

func TestSignAndVerifyHappyPath(t *testing.T) {
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
	srv, err := sign.NewService("alias/pass_service", kmsClient)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()

	token, err := srv.SignAndEncode(sign.Message{
		ConnectionID:   uuid.New().String(),
		ExpiresAt:      now.Add(time.Hour),
		ConnectionType: sign.ConnectionTypeBrowserExtensionProxy,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)
	t.Log("Length of the token is", len(token))

	extensionID, err := srv.CanI(token, sign.ConnectionTypeBrowserExtensionProxy)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(extensionID)
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
	srv, err := sign.NewService("alias/pass_service", kmsClient)
	if err != nil {
		t.Fatal(err)
	}
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
				token, err := srv.SignAndEncode(sign.Message{
					ConnectionID:   uuid.New().String(),
					ExpiresAt:      now.Add(-time.Hour),
					ConnectionType: sign.ConnectionTypeBrowserExtensionProxy,
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
				token, err := srv.SignAndEncode(sign.Message{
					ConnectionID:   uuid.New().String(),
					ExpiresAt:      now.Add(time.Hour),
					ConnectionType: sign.ConnectionTypeBrowserExtensionWait,
				})
				if err != nil {
					t.Fatal(err)
				}
				return token
			},
			expectedError: sign.ErrInvalidClaims,
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
				serviceWithAnotherKey, err := sign.NewService(*resp.KeyMetadata.KeyId, kmsClient)
				if err != nil {
					t.Fatal(err)
				}

				token, err := serviceWithAnotherKey.SignAndEncode(sign.Message{
					ConnectionID:   uuid.New().String(),
					ExpiresAt:      now.Add(-time.Hour),
					ConnectionType: sign.ConnectionTypeBrowserExtensionProxy,
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
			_, err := srv.CanI(token, sign.ConnectionTypeBrowserExtensionProxy)
			if err == nil {
				t.Fatalf("Expected error %v, got nil", tc.expectedError)
			}
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("Expected error %v, got %v", tc.expectedError, err)
			}
		})
	}

}
