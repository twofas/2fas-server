package sign

import (
	"crypto"
	"encoding/asn1"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/golang-jwt/jwt/v5"
)

type Message struct {
	ConnectionID   string
	ExpiresAt      time.Time
	ConnectionType ConnectionType
}

// SignAndEncode information in the message. The result
// is second and third part of jwt token. Since the first
// part is constant it is omitted.
func (s Service) SignAndEncode(m Message) (string, error) {
	token := jwt.NewWithClaims(s.signingMethod, jwt.MapClaims{
		"exp":  m.ExpiresAt.Unix(),
		"aud":  []string{string(m.ConnectionType)},
		"c_id": m.ConnectionID,
	})

	// no key is needed, as we use custom signing method.
	signed, err := token.SignedString(nil)
	if err != nil {
		return "", fmt.Errorf("failed to sign jwt: %w", err)
	}

	if !strings.HasPrefix(signed, jwtHeader) {
		return "", fmt.Errorf("unpexpected signed string format")
	}

	return strings.TrimPrefix(signed, jwtHeader), nil
}

type kmsSigningMethod struct {
	client *kms.KMS
	keyID  string
	hash   crypto.Hash
}

// Verify implements jwt.SigningMethod#Method. Because we
// provide key to jwt library, this is never called.
func (s kmsSigningMethod) Verify(signingString string, sig []byte, key interface{}) error {
	panic("should never be called")
}

// Sign implements jwt.SigningMethod#Sign method.
func (s kmsSigningMethod) Sign(signingString string, key interface{}) ([]byte, error) {
	messageType := "DIGEST"

	hasher := s.hash.New()
	if _, err := hasher.Write([]byte(signingString)); err != nil {
		return nil, fmt.Errorf("failed to hash input")
	}
	hashedSigningString := hasher.Sum(nil)

	resp, err := s.client.Sign(&kms.SignInput{
		KeyId:            &s.keyID,
		Message:          hashedSigningString,
		MessageType:      &messageType,
		SigningAlgorithm: aws.String(awsSigningAlgorithm),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign the message: %w", err)
	}
	// We are using encryption method with SHA_256 digest. Hence, key has 256/8=32 bytes.
	keySizeInBytes := 256 / 8
	return formatKMSSignatureForJWT(keySizeInBytes, resp.Signature)
}

// Alg implements jwt.SigningMethod#Method.
func (s kmsSigningMethod) Alg() string {
	return jwtSigningAlgorithm
}

// formatKMSSignatureForJWT translates asn1 encoded signature (returned by AWS)
// to format expected by JWT standard.
// It is an algorithm I found on the internet
// (here: https://github.com/twofas/2fas-server/pull/24/files/4f68cc2e611dca18b9787942e5cf12fc16518dd4#r1452702669 )
// It should be tested using e2e tests.
func formatKMSSignatureForJWT(keyBytes int, sig []byte) ([]byte, error) {
	p := struct {
		R *big.Int
		S *big.Int
	}{}

	_, err := asn1.Unmarshal(sig, &p)
	if err != nil {
		return nil, err
	}
	rBytes := p.R.Bytes()
	rBytesPadded := make([]byte, keyBytes)
	copy(rBytesPadded[keyBytes-len(rBytes):], rBytes)

	sBytes := p.S.Bytes()
	sBytesPadded := make([]byte, keyBytes)
	copy(sBytesPadded[keyBytes-len(sBytes):], sBytes)

	out := append(rBytesPadded, sBytesPadded...)
	return out, nil
}
