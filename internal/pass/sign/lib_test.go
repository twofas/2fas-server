package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"
	"time"

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

	_, err = srv.CanI(token, ConnectionTypeBrowserExtensionProxy)
	if err != nil {
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
				serviceWithAnotherKey := createTestService(t)
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
			_, err := srv.CanI(token, ConnectionTypeBrowserExtensionProxy)
			if err == nil {
				t.Fatalf("Expected error %v, got nil", tc.expectedError)
			}
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("Expected error %v, got %v", tc.expectedError, err)
			}
		})
	}

}
