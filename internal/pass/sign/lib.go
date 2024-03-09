package sign

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/golang-jwt/jwt/v5"
)

const (
	awsKeySpec          = "ECC_NIST_P256"
	awsSigningAlgorithm = "ECDSA_SHA_256"
	jwtSigningAlgorithm = "ES256"

	// since we control both signature and verification, and we always use the same
	// algorithm, jwt header part (first segment) is always the same.
	// we can skip it (as in not send it) to save bytes in QR code.
	// note: header has only key type, not key id.
	jwtHeader = "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9."
)

type Service struct {
	publicKey     *ecdsa.PublicKey
	signingMethod jwt.SigningMethod
}

func NewService(keyID string, client *kms.KMS) (*Service, error) {
	resp, err := client.GetPublicKey(&kms.GetPublicKeyInput{
		KeyId: &keyID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch key for %q: %w", keyID, err)
	}
	if *resp.KeySpec != awsKeySpec {
		return nil, fmt.Errorf("the only supported key spec is %q, received: %q", awsKeySpec, *resp.KeySpec)
	}

	key, err := x509.ParsePKIXPublicKey(resp.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response from KMSas public key: %w", err)
	}

	return &Service{
		publicKey: key.(*ecdsa.PublicKey),
		signingMethod: kmsSigningMethod{
			client: client,
			keyID:  keyID,
			hash:   crypto.SHA256,
		},
	}, nil
}

type ConnectionType string

const (
	ConnectionTypeBrowserExtensionWait        ConnectionType = "be/wait"
	ConnectionTypeBrowserExtensionProxy       ConnectionType = "be/proxy"
	ConnectionTypeBrowserExtensionSyncRequest ConnectionType = "be/sync/request"
	ConnectionTypeBrowserExtensionSync        ConnectionType = "be/sync/proxy"
	ConnectionTypeMobileProxy                 ConnectionType = "mobile/proxy"
	ConnectionTypeMobileConfirm               ConnectionType = "mobile/confirm"
	ConnectionTypeMobileSyncConfirm           ConnectionType = "mobile/sync/confirm"
	ConnectionTypeMobileSyncProxy             ConnectionType = "mobile/sync/proxy"
)
