package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ConvertKeyPairToStringAndBackward(t *testing.T) {
	keyPair := GenerateKeyPair(2048)

	privateKeyAsPemStr := ExportRsaPrivateKeyAsPemStr(keyPair.PrivateKey)
	publicKeyAsPemStr := ExportRsaPublicKeyAsPemStr(keyPair.PublicKey)

	assert.NotEmpty(t, publicKeyAsPemStr)

	_, err := ParseRsaPrivateKeyFromPemStr(privateKeyAsPemStr)

	assert.NoError(t, err, "Cannot convert PEM string to private key")

	_, err = ParseRsaPublicKeyFromPemStr(publicKeyAsPemStr)

	assert.NoError(t, err)
}
