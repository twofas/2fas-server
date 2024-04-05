package connection

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

const bearerProtocolPrefix = "base64url.bearer.authorization.2pass.io."

var protocolHeader = textproto.CanonicalMIMEHeaderKey("Sec-WebSocket-Protocol")

// TokenFromWSProtocol returns authorization token from 'Sec-WebSocket-Protocol' request header.
// It is used because websocket API in browser does not allow to pass headers.
// https://github.com/kubernetes/kubernetes/commit/714f97d7baf4975ad3aa47735a868a81a984d1f0
//
// Client provides a bearer token as a subprotocol in the format
// "base64url.bearer.authorization.2pass.io.<base64url-without-padding(bearer-token)>".
// This function also modified request header by removing authorization header.
// Server according to spec must return at least 1 protocol to proxy, so this fn checks
// if at least 1 protocol is sent, beside authorization header.
func TokenFromWSProtocol(req *http.Request) (string, error) {
	token := ""
	sawTokenProtocol := false
	filteredProtocols := []string{}
	for _, protocolHeader := range req.Header[protocolHeader] {
		for _, protocol := range strings.Split(protocolHeader, ",") {
			protocol = strings.TrimSpace(protocol)

			if !strings.HasPrefix(protocol, bearerProtocolPrefix) {
				filteredProtocols = append(filteredProtocols, protocol)
				continue
			}

			if sawTokenProtocol {
				return "", errors.New("multiple base64.bearer.authorization tokens specified")
			}
			sawTokenProtocol = true

			encodedToken := strings.TrimPrefix(protocol, bearerProtocolPrefix)
			decodedToken, err := base64.RawURLEncoding.DecodeString(encodedToken)
			if err != nil {
				return "", errors.New("invalid base64.bearer.authorization token encoding")
			}
			if !utf8.Valid(decodedToken) {
				return "", errors.New("invalid base64.bearer.authorization token")
			}
			token = string(decodedToken)
		}
	}

	if len(token) == 0 {
		return "", errors.New("empty token")
	}

	// Must pass at least one other subprotocol so that we can remove the one containing the bearer token,
	// and there is at least one to echo back to the proxy
	if len(filteredProtocols) == 0 {
		return "", errors.New("missing additional subprotocol")
	}

	// https://tools.ietf.org/html/rfc6455#section-11.3.4 indicates the Sec-WebSocket-Protocol header may appear multiple times
	// in a request, and is logically the same as a single Sec-WebSocket-Protocol header field that contains all values
	req.Header.Set(protocolHeader, strings.Join(filteredProtocols, ","))
	return token, nil
}

// supportedProtocol2pass is Protocol that will be sent back on upgrade mechanism.
const supportedProtocol2pass = "2pass.io"

var upgrader2pass = websocket.Upgrader{
	ReadBufferSize:  4 * 1024,
	WriteBufferSize: 4 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigin := os.Getenv("WEBSOCKET_ALLOWED_ORIGIN")

		if allowedOrigin != "" {
			return r.Header.Get("Origin") == allowedOrigin
		}

		return true
	},
	Subprotocols: []string{supportedProtocol2pass},
}

func Upgrade(w http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
	protocols := strings.Split(req.Header.Get(protocolHeader), ",")
	if !slices.Contains(protocols, supportedProtocol2pass) {
		return nil, fmt.Errorf("upgrader not available for protocols: %v", protocols)
	}
	conn, err := upgrader2pass.Upgrade(w, req, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade proxy: %w", err)
	}
	return conn, nil
}
