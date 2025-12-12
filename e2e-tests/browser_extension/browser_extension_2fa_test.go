package tests

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
)

func TestTwoFactorAuthTestSuite(t *testing.T) {
	suite.Run(t, new(TwoFactorAuthTestSuite))
}

type TwoFactorAuthTestSuite struct {
	suite.Suite
}

func (s *TwoFactorAuthTestSuite) SetupTest() {
	e2e_tests.RemoveAllMobileDevices(s.T())
	e2e_tests.RemoveAllBrowserExtensions(s.T())
	e2e_tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *TwoFactorAuthTestSuite) TestBrowserExtensionAuthFullFlow() {
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "SM-955F", "some-token")
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")

	websocketTestListener := e2e_tests.NewWebsocketTestListener("browser_extensions/" + browserExtension.Id)
	websocketConnection := websocketTestListener.StartListening()
	defer websocketConnection.Close()

	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	assertDeviceHasPairedExtension(s.T(), device, browserExtension)
	assertBrowserExtensionHasPairedDevice(s.T(), browserExtension, device)
	expectedPairingSuccessWebsocket := createPairingSuccessWebsocketMessage(s.T(), browserExtension, device, devicePubKey)
	websocketTestListener.AssertMessageHasBeenReceived(s.T(), expectedPairingSuccessWebsocket)

	tokenRequest := e2e_tests.Request2FaToken(s.T(), "facebook.com", browserExtension.Id)

	extensionTokenRequestWebsocketListener := e2e_tests.NewWebsocketTestListener("browser_extensions/" + browserExtension.Id + "/2fa_requests/" + tokenRequest.Id)
	extensionTokenRequestWebsocketConnection := extensionTokenRequestWebsocketListener.StartListening()
	defer extensionTokenRequestWebsocketConnection.Close()

	e2e_tests.Send2FaTokenToExtension(s.T(), browserExtension.Id, device.Id, tokenRequest.Id, "2fa-token")

	expected2FaTokenWebsocket := createBrowserExtensionReceived2FaTokenMessage(s.T(), browserExtension.Id, device.Id, tokenRequest.Id)
	extensionTokenRequestWebsocketListener.AssertMessageHasBeenReceived(s.T(), expected2FaTokenWebsocket)
}

func createBrowserExtensionReceived2FaTokenMessage(t *testing.T, extensionId, deviceId, requestId string) string {
	t.Helper()

	expected2FaTokenWebsocketMessageRaw := struct {
		Event          string `json:"event"`
		ExtensionId    string `json:"extension_id"`
		DeviceId       string `json:"device_id"`
		TokenRequestId string `json:"token_request_id"`
		Token          string `json:"token"`
	}{
		Event:          "browser_extensions.device.2fa_response",
		ExtensionId:    extensionId,
		DeviceId:       deviceId,
		TokenRequestId: requestId,
		Token:          "2fa-token",
	}

	message, err := json.Marshal(expected2FaTokenWebsocketMessageRaw)
	if err != nil {
		t.Fatalf("failed to marshal expected 2FA token websocket message: %v", err)
	}

	return string(message)
}

func createPairingSuccessWebsocketMessage(t *testing.T, browserExtension *e2e_tests.BrowserExtensionResponse, device *e2e_tests.DeviceResponse, devicePubKey string) string {
	t.Helper()

	expectedPairingWebsocketMessageRaw := &struct {
		Event              string `json:"event"`
		BrowserExtensionId string `json:"browser_extension_id"`
		DeviceId           string `json:"device_id"`
		DevicePublicKey    string `json:"device_public_key"`
	}{
		Event:              "browser_extensions.pairing.success",
		BrowserExtensionId: browserExtension.Id,
		DeviceId:           device.Id,
		DevicePublicKey:    devicePubKey,
	}

	message, err := json.Marshal(expectedPairingWebsocketMessageRaw)
	if err != nil {
		t.Fatalf("failed to marshal expected pairing websocket message: %v", err)
	}

	return string(message)
}

func assertBrowserExtensionHasPairedDevice(t *testing.T, browserExtension *e2e_tests.BrowserExtensionResponse, device *e2e_tests.DeviceResponse) {
	var browserExtensionDevices []*e2e_tests.DeviceResponse
	e2e_tests.DoAPISuccessGet(t, "browser_extensions/"+browserExtension.Id+"/devices", &browserExtensionDevices)

	assert.Len(t, browserExtensionDevices, 1)
	assert.Equal(t, device.Id, browserExtensionDevices[0].Id)
}

func assertDeviceHasPairedExtension(t *testing.T, device *e2e_tests.DeviceResponse, browserExtension *e2e_tests.BrowserExtensionResponse) {
	var deviceBrowserExtensions []*e2e_tests.BrowserExtensionResponse
	e2e_tests.DoAPISuccessGet(t, "mobile/devices/"+device.Id+"/browser_extensions", &deviceBrowserExtensions)

	assert.Len(t, deviceBrowserExtensions, 1)
	assert.Equal(t, browserExtension.Id, deviceBrowserExtensions[0].Id)
}
