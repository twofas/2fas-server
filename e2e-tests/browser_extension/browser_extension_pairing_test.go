package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
)

func TestBrowserExtensionPairingTestSuite(t *testing.T) {
	suite.Run(t, new(BrowserExtensionPairingTestSuite))
}

type BrowserExtensionPairingTestSuite struct {
	suite.Suite
}

func (s *BrowserExtensionPairingTestSuite) SetupTest() {
	e2e_tests.RemoveAllBrowserExtensions(s.T())
	e2e_tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *BrowserExtensionPairingTestSuite) TestPairBrowserExtensionWithMobileDevice() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")
	_, err := uuid.Parse(browserExt.Id)
	require.NoError(s.T(), err)

	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	_, err = uuid.Parse(device.Id)
	require.NoError(s.T(), err)

	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)

	var extensionDevice *e2e_tests.DevicePairedBrowserExtensionResponse
	e2e_tests.DoAPISuccessGet(s.T(), "/browser_extensions/"+browserExt.Id+"/devices/"+device.Id, &extensionDevice)

	assert.Equal(s.T(), extensionDevice.Id, device.Id)
}

func (s *BrowserExtensionPairingTestSuite) TestDoNotFindNotPairedBrowserExtensionMobileDevice() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")
	_, err := uuid.Parse(browserExt.Id)
	require.NoError(s.T(), err)

	device, _ := e2e_tests.CreateDevice(s.T(), "go-test-device", "some-device-id")

	response := e2e_tests.DoAPIGet(s.T(), "/browser_extensions/"+browserExt.Id+"/devices/"+device.Id, nil)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *BrowserExtensionPairingTestSuite) TestPairBrowserExtensionWithMultipleDevices() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")
	_, err := uuid.Parse(browserExt.Id)
	require.NoError(s.T(), err)

	device1, devicePubKey1 := e2e_tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := e2e_tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt, device1)
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt, device2)

	extensionDevices := e2e_tests.GetExtensionDevices(s.T(), browserExt.Id)

	assert.Len(s.T(), extensionDevices, 2)
}

func (s *BrowserExtensionPairingTestSuite) TestRemoveBrowserExtensionPairedDevice() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")

	device1, devicePubKey1 := e2e_tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := e2e_tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt, device1)
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt, device2)

	extensionDevices := getExtensionPairedDevices(s.T(), browserExt)
	assert.Len(s.T(), extensionDevices, 2)

	e2e_tests.DoAPISuccessDelete(s.T(), "/browser_extensions/"+browserExt.Id+"/devices/"+device1.Id)

	extensionDevices = getExtensionPairedDevices(s.T(), browserExt)
	assert.Len(s.T(), extensionDevices, 1)
	assert.Equal(s.T(), device2.Id, extensionDevices[0].Id)
}

func (s *BrowserExtensionPairingTestSuite) TestRemoveBrowserExtensionPairedDeviceTwice() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")

	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)

	e2e_tests.DoAPISuccessDelete(s.T(), "/browser_extensions/"+browserExt.Id+"/devices/"+device.Id)
	response := e2e_tests.DoAPIRequest(s.T(), "/browser_extensions/"+browserExt.Id+"/devices/"+device.Id, http.MethodDelete, nil /*payload*/, nil /*resp*/)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *BrowserExtensionPairingTestSuite) TestRemoveAllBrowserExtensionPairedDevices() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")
	device1, devicePubKey1 := e2e_tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id1")
	device2, devicePubKey2 := e2e_tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id2")
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt, device1)
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt, device2)

	e2e_tests.DoAPISuccessDelete(s.T(), "/browser_extensions/"+browserExt.Id+"/devices")

	extensionDevices := e2e_tests.GetExtensionDevices(s.T(), browserExt.Id)
	assert.Len(s.T(), extensionDevices, 0)
}

func (s *BrowserExtensionPairingTestSuite) TestGetPairedDevicesWhichIDoNotOwn() {
	browserExt1 := e2e_tests.CreateBrowserExtension(s.T(), "go-test-1")
	browserExt2 := e2e_tests.CreateBrowserExtension(s.T(), "go-test-2")

	device1, devicePubKey1 := e2e_tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := e2e_tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt1, device1)
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt2, device2)

	firstExtensionDevices := getExtensionPairedDevices(s.T(), browserExt1)
	assert.Len(s.T(), firstExtensionDevices, 1)
	assert.Equal(s.T(), device1.Id, firstExtensionDevices[0].Id)

	secondExtensionDevices := getExtensionPairedDevices(s.T(), browserExt2)
	assert.Len(s.T(), secondExtensionDevices, 1)
	assert.Equal(s.T(), device2.Id, secondExtensionDevices[0].Id)
}

func (s *BrowserExtensionPairingTestSuite) TestGetPairedDevicesByInvalidExtensionId() {
	browserExt1 := e2e_tests.CreateBrowserExtension(s.T(), "go-test-1")
	browserExt2 := e2e_tests.CreateBrowserExtension(s.T(), "go-test-2")

	device1, devicePubKey1 := e2e_tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := e2e_tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt1, device1)
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt2, device2)

	invalidResp := map[string]any{}
	response := e2e_tests.DoAPIGet(s.T(), "/browser_extensions/some-invalid-id/devices/", &invalidResp)
	assert.Equal(s.T(), 400, response.StatusCode)
	assert.Contains(s.T(), invalidResp["Reason"], `Field validation for 'ExtensionId' failed on the 'uuid4'`)
}

func (s *BrowserExtensionPairingTestSuite) TestGetPairedDevicesByNotExistingExtensionId() {
	browserExt1 := e2e_tests.CreateBrowserExtension(s.T(), "go-test-1")
	browserExt2 := e2e_tests.CreateBrowserExtension(s.T(), "go-test-2")

	device1, devicePubKey1 := e2e_tests.CreateDevice(s.T(), "go-test-device-1", "some-device-id-1")
	device2, devicePubKey2 := e2e_tests.CreateDevice(s.T(), "go-test-device-2", "some-device-id-2")

	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey1, browserExt1, device1)
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey2, browserExt2, device2)

	notExistingExtensionId := uuid.New()
	var firstExtensionDevices []*e2e_tests.ExtensionPairedDeviceResponse
	e2e_tests.DoAPISuccessGet(s.T(), "/browser_extensions/"+notExistingExtensionId.String()+"/devices/", &firstExtensionDevices)
	assert.Len(s.T(), firstExtensionDevices, 0)
}

func (s *BrowserExtensionPairingTestSuite) TestShareExtensionPublicKeyWithMobileDevice() {
	browserExt := e2e_tests.CreateBrowserExtensionWithPublicKey(s.T(), "go-test", "b64-rsa-pub-key")
	_, err := uuid.Parse(browserExt.Id)
	require.NoError(s.T(), err)

	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	_, err = uuid.Parse(device.Id)
	require.NoError(s.T(), err)

	result := e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)
	assert.Equal(s.T(), "b64-rsa-pub-key", result.ExtensionPublicKey)
}

func (s *BrowserExtensionPairingTestSuite) TestCannotPairSameDeviceAndExtensionTwice(t *testing.T) {
	browserExtension := e2e_tests.CreateBrowserExtensionWithPublicKey(s.T(), "go-test", "b64-rsa-pub-key")
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "go-test-device", "some-device-id")

	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	payload := struct {
		ExtensionId     string `json:"extension_id"`
		DeviceName      string `json:"device_name"`
		DevicePublicKey string `json:"device_public_key"`
	}{
		ExtensionId:     browserExtension.Id,
		DeviceName:      device.Name,
		DevicePublicKey: "device-pub-key",
	}

	pairingResult := new(e2e_tests.PairingResultResponse)
	payloadJson, err := json.Marshal(payload)
	require.NoError(t, err)

	e2e_tests.DoAPIPostAndAssertCode(s.T(), 409, "/mobile/devices/"+device.Id+"/browser_extensions", payloadJson, pairingResult)
}

func getExtensionPairedDevices(t *testing.T, browserExt *e2e_tests.BrowserExtensionResponse) []*e2e_tests.ExtensionPairedDeviceResponse {
	var extensionDevices []*e2e_tests.ExtensionPairedDeviceResponse
	e2e_tests.DoAPISuccessGet(t, "/browser_extensions/"+browserExt.Id+"/devices/", &extensionDevices)
	return extensionDevices
}
