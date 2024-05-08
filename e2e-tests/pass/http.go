package pass

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/avast/retry-go/v4"
	"github.com/google/uuid"
)

var (
	httpClient = http.DefaultClient
)

func getApiURL() string {
	apiURL := os.Getenv("API_URL")
	if apiURL != "" {
		return apiURL
	}
	return "http://" + getPassAddr()
}

func getPassAddr() string {
	addr := os.Getenv("PASS_ADDR")
	if addr != "" {
		return addr
	}
	return "localhost:8082"
}

type ConfigureBrowserExtensionResponse struct {
	BrowserExtensionPairingToken string `json:"browser_extension_pairing_token"`
	ConnectionToken              string `json:"connection_token"`
}

func configureBrowserExtension() (ConfigureBrowserExtensionResponse, error) {
	extensionID := uuid.NewString()
	if extensionIDFromEnv := os.Getenv("TEST_EXTENSION_ID"); extensionIDFromEnv != "" {
		extensionID = extensionIDFromEnv
	}
	req := struct {
		ExtensionID string `json:"extension_id"`
	}{
		ExtensionID: extensionID,
	}
	var resp ConfigureBrowserExtensionResponse

	if err := request("POST", "/browser_extension/configure", "", req, &resp); err != nil {
		return resp, fmt.Errorf("failed to configure browser: %w", err)
	}

	return resp, nil
}

// confirmMobile confirms pairing and returns mobile proxy token.
func confirmMobile(connectionToken, deviceID, fcm string) (string, error) {
	req := struct {
		DeviceID string `json:"device_id"`
		FCMToken string `json:"fcm_token"`
	}{
		DeviceID: deviceID,
		FCMToken: fcm,
	}
	resp := struct {
		ProxyToken string `json:"proxy_token"`
	}{}

	if err := request("POST", "/mobile/confirm", connectionToken, req, &resp); err != nil {
		return "", fmt.Errorf("failed to configure browser: %w", err)
	}

	return resp.ProxyToken, nil
}

// confirmSyncMobile confirms pairing and returns mobile proxy token.
func confirmSyncMobile(connectionToken string) (string, error) {
	var result string

	err := retry.Do(func() error {
		var err error
		result, err = confirmSyncMobileRequest(connectionToken)
		return err
	})

	return result, err
}

func confirmSyncMobileRequest(connectionToken string) (string, error) {
	var resp struct {
		ProxyToken string `json:"proxy_token"`
	}

	if err := request("POST", "/mobile/sync/confirm", connectionToken, nil, &resp); err != nil {
		return "", fmt.Errorf("failed to confirm mobile: %w", err)
	}

	return resp.ProxyToken, nil
}

func request(method, path, auth string, req, resp interface{}) error {
	url := getApiURL() + path
	var body io.Reader
	if req != nil {
		bb, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to request marshal: %w", err)
		}
		body = bytes.NewBuffer(bb)
	}
	httpReq, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	if auth != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
	}

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {

		return fmt.Errorf("failed perform the request: %w", err)
	}
	defer httpResp.Body.Close()

	bb, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body from response: %w", err)
	}

	if httpResp.StatusCode >= 300 {
		return fmt.Errorf("received status %s and body %q", httpResp.Status, string(bb))
	}
	if err := json.Unmarshal(bb, &resp); err != nil {
		return fmt.Errorf("failed to decode the response: %w", err)
	}

	return nil
}

type RequestSyncResponse struct {
	BrowserExtensionWaitToken string `json:"browser_extension_wait_token"`
	MobileConfirmToken        string `json:"mobile_confirm_token"`
}

func browserExtensionRequestSync(token string) (RequestSyncResponse, error) {
	var resp RequestSyncResponse

	if err := request("POST", "/browser_extension/sync/request", token, nil, &resp); err != nil {
		return resp, fmt.Errorf("failed to configure browser: %w", err)
	}

	return resp, nil
}

func browserExtensionPush(token, body string) (string, error) {
	var resp struct {
		Response string `json:"response"`
	}
	req := struct {
		Body string `json:"push_body"`
	}{
		Body: body,
	}

	if err := request("POST", "/browser_extension/sync/push", token, req, &resp); err != nil {
		return "", fmt.Errorf("failed to send push notification: %w", err)
	}

	return resp.Response, nil
}
