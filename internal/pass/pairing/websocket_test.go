package pairing

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_tokenFromWSProtocol(t *testing.T) {
	tests := []struct {
		name           string
		protocolHeader string
		assertFn       func(t *testing.T, token string, err error)
	}{
		{
			name:           "valid token with additional subprotocol",
			protocolHeader: "base64url.bearer.authorization.2pass.io.dGVzdDE,ws",
			assertFn: func(t *testing.T, token string, err error) {
				require.NoError(t, err)
				require.Equal(t, "test1", token)
			},
		},
		{
			name: "missing header",
			assertFn: func(t *testing.T, token string, err error) {
				require.ErrorContains(t, err, "empty token")
			},
		},
		{
			name:           "invalid encoding",
			protocolHeader: "base64url.bearer.authorization.2pass.io.dGVzdA==",
			assertFn: func(t *testing.T, token string, err error) {
				require.ErrorContains(t, err, "invalid base64.bearer.authorization token encoding")
			},
		},
		{
			name:           "missing other protocol",
			protocolHeader: "base64url.bearer.authorization.2pass.io.dGVzdDE",
			assertFn: func(t *testing.T, token string, err error) {
				require.ErrorContains(t, err, "missing additional subprotocol")
			},
		},
		{
			name:           "double authorization header",
			protocolHeader: "base64url.bearer.authorization.2pass.io.dGVzdDE,base64url.bearer.authorization.2pass.io.dGVzdDE",
			assertFn: func(t *testing.T, token string, err error) {
				require.ErrorContains(t, err, "multiple base64.bearer.authorization tokens specified")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://localhost", nil)
			require.NoError(t, err)
			if protocolHeader != "" {
				req.Header.Set(protocolHeader, tt.protocolHeader)
			}
			got, err := tokenFromWSProtocol(req)
			tt.assertFn(t, got, err)
		})
	}
}
