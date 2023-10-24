package tests

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/tests"
)

func TestTwoFactorAuthTestSuite(t *testing.T) {
	suite.Run(t, new(TwoFactorAuthTestSuite))
}

type TwoFactorAuthTestSuite struct {
	suite.Suite
}

func (s *TwoFactorAuthTestSuite) SetupTest() {
	tests.RemoveAllMobileDevices(s.T())
	tests.RemoveAllBrowserExtensions(s.T())
	tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *TwoFactorAuthTestSuite) TestBrowserExtensionAuthFullFlow() {
	device, devicePubKey := tests.CreateDevice(s.T(), "SM-955F", "some-token")
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")

	websocketTestListener := tests.NewWebsocketTestListener("browser_extensions/" + browserExtension.Id)
	websocketConnection := websocketTestListener.StartListening()
	defer websocketConnection.Close()

	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	assertDeviceHasPairedExtension(s.T(), device, browserExtension)
	assertBrowserExtensionHasPairedDevice(s.T(), browserExtension, device)
	expectedPairingSuccessWebsocket := createPairingSuccessWebsocketMessage(browserExtension, device, devicePubKey)
	websocketTestListener.AssertMessageHasBeenReceived(s.T(), expectedPairingSuccessWebsocket)

	tokenRequest := tests.Request2FaToken(s.T(), "facebook.com", browserExtension.Id)

	extensionTokenRequestWebsocketListener := tests.NewWebsocketTestListener("browser_extensions/" + browserExtension.Id + "/2fa_requests/" + tokenRequest.Id)
	extensionTokenRequestWebsocketConnection := extensionTokenRequestWebsocketListener.StartListening()
	defer extensionTokenRequestWebsocketConnection.Close()

	tests.Send2FaTokenToExtension(s.T(), browserExtension.Id, device.Id, tokenRequest.Id, "2fa-token")

	expected2FaTokenWebsocket := createBrowserExtensionReceived2FaTokenMessage(browserExtension.Id, device.Id, tokenRequest.Id)
	extensionTokenRequestWebsocketListener.AssertMessageHasBeenReceived(s.T(), expected2FaTokenWebsocket)
}

func createBrowserExtensionReceived2FaTokenMessage(extensionId, deviceId, requestId string) string {
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

	message, _ := json.Marshal(expected2FaTokenWebsocketMessageRaw)

	return string(message)
}

func createPairingSuccessWebsocketMessage(browserExtension *tests.BrowserExtensionResponse, device *tests.DeviceResponse, devicePubKey string) string {
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

	message, _ := json.Marshal(expectedPairingWebsocketMessageRaw)

	return string(message)
}

func assertBrowserExtensionHasPairedDevice(t *testing.T, browserExtension *tests.BrowserExtensionResponse, device *tests.DeviceResponse) {
	var browserExtensionDevices []*tests.DeviceResponse
	tests.DoAPISuccessGet(t, "browser_extensions/"+browserExtension.Id+"/devices", &browserExtensionDevices)

	assert.Len(t, browserExtensionDevices, 1)
	assert.Equal(t, device.Id, browserExtensionDevices[0].Id)
}

func assertDeviceHasPairedExtension(t *testing.T, device *tests.DeviceResponse, browserExtension *tests.BrowserExtensionResponse) {
	var deviceBrowserExtensions []*tests.BrowserExtensionResponse
	tests.DoAPISuccessGet(t, "mobile/devices/"+device.Id+"/browser_extensions", &deviceBrowserExtensions)

	assert.Len(t, deviceBrowserExtensions, 1)
	assert.Equal(t, browserExtension.Id, deviceBrowserExtensions[0].Id)
}
