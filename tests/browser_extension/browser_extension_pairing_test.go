package tests

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/tests"
)

func TestBrowserExtensionPairingTestSuite(t *testing.T) {
	suite.Run(t, new(BrowserExtensionPairingTestSuite))
}

type BrowserExtensionPairingTestSuite struct {
	suite.Suite
}

func (s *BrowserExtensionPairingTestSuite) SetupTest() {
	tests.RemoveAllBrowserExtensions(s.T())
	tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *BrowserExtensionPairingTestSuite) TestPairBrowserExtensionWithMobileDevice() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")
	_, err := uuid.Parse(browserExt.Id)
	require.NoError(s.T(), err)

	device, devicePubKey := tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	_, err = uuid.Parse(device.Id)
	require.NoError(s.T(), err)

	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)

	var extensionDevice *tests.DevicePairedBrowserExtensionResponse
	tests.DoSuccessGet(s.T(), "/browser_extensions/"+browserExt.Id+"/devices/"+device.Id, &extensionDevice)

	assert.Equal(s.T(), extensionDevice.Id, device.Id)
}

func (s *BrowserExtensionPairingTestSuite) TestDoNotFindNotPairedBrowserExtensionMobileDevice() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")
	_, err := uuid.Parse(browserExt.Id)
	require.NoError(s.T(), err)

	device, _ := tests.CreateDevice(s.T(), "go-test-device", "some-device-id")

	response := tests.DoGet("/browser_extensions/"+browserExt.Id+"/devices/"+device.Id, nil)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *BrowserExtensionPairingTestSuite) TestPairBrowserExtensionWithMultipleDevices() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")
	_, err := uuid.Parse(browserExt.Id)
	require.NoError(s.T(), err)

	device1, devicePubKey1 := tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt, device1)
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt, device2)

	extensionDevices := tests.GetExtensionDevices(s.T(), browserExt.Id)

	assert.Len(s.T(), extensionDevices, 2)
}

func (s *BrowserExtensionPairingTestSuite) TestRemoveBrowserExtensionPairedDevice() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")

	device1, devicePubKey1 := tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt, device1)
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt, device2)

	extensionDevices := getExtensionPairedDevices(s.T(), browserExt)
	assert.Len(s.T(), extensionDevices, 2)

	tests.DoSuccessDelete(s.T(), "/browser_extensions/"+browserExt.Id+"/devices/"+device1.Id)

	extensionDevices = getExtensionPairedDevices(s.T(), browserExt)
	assert.Len(s.T(), extensionDevices, 1)
	assert.Equal(s.T(), device2.Id, extensionDevices[0].Id)
}

func (s *BrowserExtensionPairingTestSuite) TestRemoveBrowserExtensionPairedDeviceTwice() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")

	device, devicePubKey := tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)

	tests.DoSuccessDelete(s.T(), "/browser_extensions/"+browserExt.Id+"/devices/"+device.Id)
	response := tests.DoDelete("/browser_extensions/" + browserExt.Id + "/devices/" + device.Id)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *BrowserExtensionPairingTestSuite) TestRemoveAllBrowserExtensionPairedDevices() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")
	device1, devicePubKey1 := tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id1")
	device2, devicePubKey2 := tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id2")
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt, device1)
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt, device2)

	tests.DoSuccessDelete(s.T(), "/browser_extensions/"+browserExt.Id+"/devices")

	extensionDevices := tests.GetExtensionDevices(s.T(), browserExt.Id)
	assert.Len(s.T(), extensionDevices, 0)
}

func (s *BrowserExtensionPairingTestSuite) TestGetPairedDevicesWhichIDoNotOwn() {
	browserExt1 := tests.CreateBrowserExtension(s.T(), "go-test-1")
	browserExt2 := tests.CreateBrowserExtension(s.T(), "go-test-2")

	device1, devicePubKey1 := tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt1, device1)
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt2, device2)

	firstExtensionDevices := getExtensionPairedDevices(s.T(), browserExt1)
	assert.Len(s.T(), firstExtensionDevices, 1)
	assert.Equal(s.T(), device1.Id, firstExtensionDevices[0].Id)

	secondExtensionDevices := getExtensionPairedDevices(s.T(), browserExt2)
	assert.Len(s.T(), secondExtensionDevices, 1)
	assert.Equal(s.T(), device2.Id, secondExtensionDevices[0].Id)
}

func (s *BrowserExtensionPairingTestSuite) TestGetPairedDevicesByInvalidExtensionId() {
	browserExt1 := tests.CreateBrowserExtension(s.T(), "go-test-1")
	browserExt2 := tests.CreateBrowserExtension(s.T(), "go-test-2")

	device1, devicePubKey1 := tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt1, device1)
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt2, device2)

	var firstExtensionDevices []*tests.ExtensionPairedDeviceResponse
	response := tests.DoGet("/browser_extensions/some-invalid-id/devices/", &firstExtensionDevices)
	assert.Len(s.T(), firstExtensionDevices, 0)
	assert.Equal(s.T(), 400, response.StatusCode)
}

func (s *BrowserExtensionPairingTestSuite) TestGetPairedDevicesByNotExistingExtensionId() {
	browserExt1 := tests.CreateBrowserExtension(s.T(), "go-test-1")
	browserExt2 := tests.CreateBrowserExtension(s.T(), "go-test-2")

	device1, devicePubKey1 := tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt1, device1)
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt2, device2)

	notExistingExtensionId := uuid.New()
	var firstExtensionDevices []*tests.ExtensionPairedDeviceResponse
	tests.DoSuccessGet(s.T(), "/browser_extensions/"+notExistingExtensionId.String()+"/devices/", &firstExtensionDevices)
	assert.Len(s.T(), firstExtensionDevices, 0)
}

func (s *BrowserExtensionPairingTestSuite) TestShareExtensionPublicKeyWithMobileDevice() {
	browserExt := tests.CreateBrowserExtensionWithPublicKey(s.T(), "go-test", "b64-rsa-pub-key")
	_, err := uuid.Parse(browserExt.Id)
	require.NoError(s.T(), err)

	device, devicePubKey := tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	_, err = uuid.Parse(device.Id)
	require.NoError(s.T(), err)

	result := tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)
	assert.Equal(s.T(), "b64-rsa-pub-key", result.ExtensionPublicKey)
}

func (s *BrowserExtensionPairingTestSuite) TestCannotPairSameDeviceAndExtensionTwice() {
	browserExtension := tests.CreateBrowserExtensionWithPublicKey(s.T(), "go-test", "b64-rsa-pub-key")
	device, devicePubKey := tests.CreateDevice(s.T(), "go-test-device", "some-device-id")

	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	payload := struct {
		ExtensionId     string `json:"extension_id"`
		DeviceName      string `json:"device_name"`
		DevicePublicKey string `json:"device_public_key"`
	}{
		ExtensionId:     browserExtension.Id,
		DeviceName:      device.Name,
		DevicePublicKey: "device-pub-key",
	}

	pairingResult := new(tests.PairingResultResponse)
	payloadJson, _ := json.Marshal(payload)

	response := tests.DoPost("/mobile/devices/"+device.Id+"/browser_extensions", payloadJson, pairingResult)
	assert.Equal(s.T(), 409, response.StatusCode)
}

func getExtensionPairedDevices(t *testing.T, browserExt *tests.BrowserExtensionResponse) []*tests.ExtensionPairedDeviceResponse {
	var extensionDevices []*tests.ExtensionPairedDeviceResponse
	tests.DoSuccessGet(t, "/browser_extensions/"+browserExt.Id+"/devices/", &extensionDevices)
	return extensionDevices
}
