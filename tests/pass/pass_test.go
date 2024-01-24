package pass

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ConfigureBrowserExtensionResponse struct {
	BrowserExtensionPairingToken string `json:"browser_extension_pairing_token"`
	ConnectionToken              string `json:"connection_token"`
}

var (
	httpClient = http.DefaultClient
	wsDialer   = websocket.DefaultDialer
)

func getAPIURL() string {
	addr := os.Getenv("PASS_ADDR")
	if addr != "" {
		return addr
	}
	return "localhost:8082"
}

func TestPassHappyFlow(t *testing.T) {
	resp, err := configureBrowserExtension()
	if err != nil {
		t.Fatalf("Failed to configure browser extension: %v", err)
	}

	browserExtensionDone := make(chan struct{})
	mobileDone := make(chan struct{})

	go func() {
		defer close(browserExtensionDone)

		extProxyToken, err := browserExtensionWaitForConfirm(resp.BrowserExtensionPairingToken)
		if err != nil {
			t.Errorf("Error when Browser Extension waited for confirm: %v", err)
			return
		}

		err = proxyWebSocket(
			"ws://"+getAPIURL()+"/browser_extension/proxy_to_mobile",
			extProxyToken,
			"sent from browser extension",
			"sent from mobile")
		if err != nil {
			t.Errorf("Browser Extension: proxy failed: %v", err)
			return
		}

	}()
	go func() {
		defer close(mobileDone)

		mobileProxyToken, err := confirmMobile(resp.ConnectionToken)
		if err != nil {
			t.Errorf("Mobile: confirm failed: %v", err)
			return
		}

		err = proxyWebSocket(
			"ws://"+getAPIURL()+"/mobile/proxy_to_browser_extension",
			mobileProxyToken,
			"sent from mobile",
			"sent from browser extension",
		)
		if err != nil {
			t.Errorf("Mobile: proxy failed: %v", err)
			return
		}
	}()
	<-browserExtensionDone
	<-mobileDone
}

func browserExtensionWaitForConfirm(token string) (string, error) {
	url := "ws://" + getAPIURL() + "/browser_extension/wait_for_connection"

	var resp struct {
		BrowserExtensionProxyToken string `json:"browser_extension_proxy_token"`
		Status                     string `json:"status"`
		DeviceID                   string `json:"device_id"`
	}

	conn, err := dialWS(url, token)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("error reading from connection: %w", err)
	}
	if err := json.Unmarshal(message, &resp); err != nil {
		return "", fmt.Errorf("failed to decode message: %w", err)
	}
	const expectedStatus = "ok"
	if resp.Status != expectedStatus {
		return "", fmt.Errorf("received status %q, expected %q", resp.Status, expectedStatus)
	}
	return resp.BrowserExtensionProxyToken, nil
}

func configureBrowserExtension() (ConfigureBrowserExtensionResponse, error) {
	url := "http://" + getAPIURL() + "/browser_extension/configure"

	req, err := http.NewRequest("POST", url, bytesPrintf(`{"extension_id":"%s"}`, uuid.New().String()))
	if err != nil {
		return ConfigureBrowserExtensionResponse{}, fmt.Errorf("failed to create http request: %w", err)
	}
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return ConfigureBrowserExtensionResponse{}, fmt.Errorf("failed perform the request: %w", err)
	}
	defer httpResp.Body.Close()

	bb, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return ConfigureBrowserExtensionResponse{}, fmt.Errorf("failed to read body from response: %w", err)
	}

	if httpResp.StatusCode >= 300 {
		return ConfigureBrowserExtensionResponse{}, fmt.Errorf("received status %s and body %q", httpResp.Status, string(bb))
	}

	var resp ConfigureBrowserExtensionResponse
	if err := json.Unmarshal(bb, &resp); err != nil {
		return resp, fmt.Errorf("failed to decode the response: %w", err)
	}

	return resp, nil
}

// confirmMobile confirms pairing and returns mobile proxy token.
func confirmMobile(connectionToken string) (string, error) {
	url := "http://" + getAPIURL() + "/mobile/confirm"

	req, err := http.NewRequest("POST", url, bytesPrintf(`{"device_id":"%s"}`, uuid.New().String()))
	if err != nil {
		return "", fmt.Errorf("failed to prepare the reqest: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", connectionToken))

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform the reqest: %w", err)
	}
	defer httpResp.Body.Close()

	bb, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body from response: %w", err)
	}

	if httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("received status %s and body %q", httpResp.Status, string(bb))
	}

	var resp struct {
		ProxyToken string `json:"proxy_token"`
	}
	if err := json.Unmarshal(bb, &resp); err != nil {
		return "", fmt.Errorf("failed to decode the response: %w", err)
	}

	return resp.ProxyToken, nil
}

// proxyWebSocket will dial `endpoint`, using `token` for auth. It will then write exactly one message and
// read exactly one message (and then check it is `expectedReadMsg`).
func proxyWebSocket(url, token string, writeMsg, expectedReadMsg string) error {
	conn, err := dialWS(url, token)
	if err != nil {
		return err
	}
	defer conn.Close()

	doneReading := make(chan error)

	go func() {
		defer close(doneReading)
		_, message, err := conn.ReadMessage()
		if err != nil {
			doneReading <- fmt.Errorf("faile to read message: %w", err)
			return
		}
		if string(message) != expectedReadMsg {
			doneReading <- fmt.Errorf("expected to read %q, read %q", expectedReadMsg, string(message))
			return
		}
	}()

	if err := conn.WriteMessage(websocket.TextMessage, []byte(writeMsg)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	err, _ = <-doneReading
	if err != nil {
		return fmt.Errorf("error when reading: %w", err)
	}
	return nil
}

func dialWS(url, auth string) (*websocket.Conn, error) {
	authEncodedAsProtocol := fmt.Sprintf("base64url.bearer.authorization.2pass.io.%s", base64.RawURLEncoding.EncodeToString([]byte(auth)))

	conn, _, err := wsDialer.Dial(url, http.Header{
		"Sec-WebSocket-Protocol": []string{
			"2pass.io",
			authEncodedAsProtocol,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial ws %q: %w", url, err)
	}
	return conn, nil
}

func bytesPrintf(format string, ii ...interface{}) io.Reader {
	s := fmt.Sprintf(format, ii...)
	return bytes.NewBufferString(s)
}
