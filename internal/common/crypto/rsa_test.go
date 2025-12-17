package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ConvertKeyPairToStringAndBackward(t *testing.T) {
	t.Helper()

	keyPair := GenerateKeyPair(2048)

	privateKeyAsPemStr := ExportRsaPrivateKeyAsPemStr(keyPair.PrivateKey)
	publicKeyAsPemStr := ExportRsaPublicKeyAsPemStr(keyPair.PublicKey)

	require.NotEmpty(t, publicKeyAsPemStr)

	_, err := ParseRsaPrivateKeyFromPemStr(privateKeyAsPemStr)

	require.NoError(t, err, "Cannot convert PEM string to private key")

	_, err = ParseRsaPublicKeyFromPemStr(publicKeyAsPemStr)

	require.NoError(t, err)
}
