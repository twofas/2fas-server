package pass

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	wsDialer = websocket.DefaultDialer
)

func getWsURL() string {
	wsURL := os.Getenv("WS_URL")
	if wsURL != "" {
		return wsURL
	}
	return "ws://" + getPassAddr()
}

func browserExtensionWaitForSyncConfirm(token string) (string, error) {
	url := getWsURL() + "/browser_extension/sync/wait"

	var resp struct {
		BrowserExtensionSyncToken string `json:"browser_extension_proxy_token"`
		Status                    string `json:"status"`
	}

	conn, err := dialWS(url, token)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
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
	return resp.BrowserExtensionSyncToken, nil
}

func browserExtensionWaitForConfirm(token string) (string, string, error) {
	url := getWsURL() + "/browser_extension/wait_for_connection"

	var resp struct {
		BrowserExtensionProxyToken string `json:"browser_extension_proxy_token"`
		BrowserExtensionSyncToken  string `json:"browser_extension_sync_token"`
		Status                     string `json:"status"`
		DeviceID                   string `json:"device_id"`
	}

	conn, err := dialWS(url, token)
	if err != nil {
		return "", "", err
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		return "", "", fmt.Errorf("error reading from connection: %w", err)
	}
	if err := json.Unmarshal(message, &resp); err != nil {
		return "", "", fmt.Errorf("failed to decode message: %w", err)
	}
	const expectedStatus = "ok"
	if resp.Status != expectedStatus {
		return "", "", fmt.Errorf("received status %q, expected %q", resp.Status, expectedStatus)
	}
	return resp.BrowserExtensionProxyToken, resp.BrowserExtensionSyncToken, nil
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

// proxyWebSocket will dial `endpoint`, using `token` for auth. It will then write exactly one message and
// read exactly one message (and then check it is `expectedReadMsg`).
func proxyWebSocket(url, token string, writeMsg, expectedReadMsg string, sleepBeforeSend time.Duration) error {
	const wsPingFrequency = 10 * time.Second // how often server send pings

	conn, err := dialWS(url, token)
	if err != nil {
		return err
	}
	defer conn.Close()

	doneReading := make(chan error)
	doneWriting := atomic.Bool{}
	doneWriting.Store(false)

	go func() {
		defer close(doneReading)
		_, message, err := conn.ReadMessage()
		if err != nil {
			doneReading <- fmt.Errorf("failed to read message: %w", err)
		}
		if string(message) != expectedReadMsg {
			doneReading <- fmt.Errorf("expected to read %q, read %q", expectedReadMsg, string(message))
		}
		for !doneWriting.Load() {
			conn.SetReadDeadline(time.Now().Add(wsPingFrequency + time.Second))
			_, _, err = conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	time.Sleep(sleepBeforeSend)

	defer doneWriting.Store(true)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(writeMsg)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	err, _ = <-doneReading
	if err != nil {
		return fmt.Errorf("error when reading: %w", err)
	}
	return nil
}
