package pass

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

const api = "localhost:8082"

func TestPassHappyFlow(t *testing.T) {
	resp, err := configureBrowserExtension()
	if err != nil {
		t.Fatalf("Failed to configure browser extension: %v", err)
	}

	browserExtensionDone := make(chan struct{})
	mobileDone := make(chan struct{})

	go func() {
		defer close(browserExtensionDone)

		err := browserExtensionWaitForConfirm(resp.BrowserExtensionPairingToken)
		if err != nil {
			t.Errorf("Error when Browser Extension waited for confirm: %v", err)
			return
		}

		err = proxyWebSocket(
			"ws://"+api+"/browser_extension/proxy_to_mobile",
			resp.BrowserExtensionPairingToken,
			"sent from browser extension",
			"sent from mobile")
		if err != nil {
			t.Errorf("Browser Extension: proxy failed: %v", err)
			return
		}

	}()
	go func() {
		defer close(mobileDone)

		err := confirmMobile(resp.ConnectionToken)
		if err != nil {
			t.Errorf("Mobile: confirm failed: %v", err)
			return
		}

		err = proxyWebSocket(
			"ws://"+api+"/mobile/proxy_to_browser_extension",
			resp.BrowserExtensionPairingToken,
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

func browserExtensionWaitForConfirm(token string) error {
	url := "ws://" + api + "/browser_extension/wait_for_connection"

	var resp struct {
		Status string `json:"status"`
	}

	conn, err := dialWS(url, token)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("error reading from connection: %w", err)
	}
	if err := json.Unmarshal(message, &resp); err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}
	const expectedStatus = "ok"
	if resp.Status != expectedStatus {
		return fmt.Errorf("received status %q, expected %q", resp.Status, expectedStatus)
	}
	return nil
}

func configureBrowserExtension() (ConfigureBrowserExtensionResponse, error) {
	url := "http://" + api + "/browser_extension/configure"

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

func confirmMobile(connectionToken string) error {
	url := "http://" + api + "/mobile/confirm"

	req, err := http.NewRequest("POST", url, bytesPrintf(`{"device_id":"%s"}`, uuid.New().String()))
	if err != nil {
		return fmt.Errorf("failed to prepare the reqest: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", connectionToken))

	httpResp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform the reqest: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode > 299 {
		return fmt.Errorf("unexpected response: %s", httpResp.Status)
	}

	return nil
}

// proxyWebSocket will dial `endpoint`, using `token` for auth. It will then write exactly one message and
// read exactly one message (and then check it is `expectedReadMsg`).
func proxyWebSocket(url, token string, writeMsg, expectedReadMsg string) error {
	conn, err := dialWS(url, token)
	if err != nil {
		return nil
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
		return nil, fmt.Errorf("failed to dial ws %q: %v", url, err)
	}
	return conn, nil
}

func bytesPrintf(format string, ii ...interface{}) io.Reader {
	s := fmt.Sprintf(format, ii...)
	return bytes.NewBufferString(s)
}
