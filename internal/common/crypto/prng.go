package crypto

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateNonce() (string, error) {
	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
	//return base64.URLEncoding.EncodeToString(nonceBytes), nil
}
