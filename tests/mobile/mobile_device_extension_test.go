package tests

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/tests"
)

func TestMobileDeviceExtensionTestSuite(t *testing.T) {
	suite.Run(t, new(MobileDeviceExtensionTestSuite))
}

type MobileDeviceExtensionTestSuite struct {
	suite.Suite
}

func (s *MobileDeviceExtensionTestSuite) SetupTest() {
	tests.RemoveAllMobileDevices(s.T())
	tests.RemoveAllBrowserExtensions(s.T())
	tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *MobileDeviceExtensionTestSuite) TestDoNotFindExtensionsForNotExistingDevice() {
	notExistingDeviceId := uuid.New()

	response := tests.DoGet("/mobile/devices/"+notExistingDeviceId.String()+"/browser_extensions", nil)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *MobileDeviceExtensionTestSuite) TestDoNotFindNotExistingMobileDeviceExtension() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")
	device, devicePubKey := tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)

	notExistingExtensionId := uuid.New()
	response := tests.DoGet("/mobile/devices/"+device.Id+"/browser_extensions/"+notExistingExtensionId.String(), nil)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *MobileDeviceExtensionTestSuite) Test_FindExtensionForDevice() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")
	device, devicePubKey := tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)

	var deviceBrowserExtension *tests.BrowserExtensionResponse
	tests.DoSuccessGet(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/"+browserExt.Id, &deviceBrowserExtension)

	assert.Equal(s.T(), browserExt.Id, deviceBrowserExtension.Id)
}

func (s *MobileDeviceExtensionTestSuite) Test_FindAllDeviceExtensions() {
	browserExt1 := tests.CreateBrowserExtension(s.T(), "go-test-1")
	browserExt2 := tests.CreateBrowserExtension(s.T(), "go-test-2")
	device, devicePubKey := tests.CreateDevice(s.T(), "go-test-device", "some-device-id")

	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt1, device)
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt2, device)

	var deviceBrowserExtensions []*tests.BrowserExtensionResponse
	tests.DoSuccessGet(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/", &deviceBrowserExtensions)

	assert.Len(s.T(), deviceBrowserExtensions, 2)
}

func (s *MobileDeviceExtensionTestSuite) Test_DisconnectExtensionFromDevice() {
	browserExt1 := tests.CreateBrowserExtension(s.T(), "go-test")
	browserExt2 := tests.CreateBrowserExtension(s.T(), "go-test")
	device, devicePubKey := tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt1, device)
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt2, device)

	tests.DoSuccessDelete(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/"+browserExt1.Id)

	var deviceBrowserExtension1 *tests.BrowserExtensionResponse
	response := tests.DoGet("/mobile/devices/"+device.Id+"/browser_extensions/"+browserExt1.Id, &deviceBrowserExtension1)
	assert.Equal(s.T(), 404, response.StatusCode)

	var deviceBrowserExtension2 *tests.BrowserExtensionResponse
	tests.DoSuccessGet(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/"+browserExt2.Id, &deviceBrowserExtension2)
	assert.Equal(s.T(), browserExt2.Id, deviceBrowserExtension2.Id)
}

func (s *MobileDeviceExtensionTestSuite) TestExtensionHasAlreadyBeenConnected() {
	extension := tests.CreateBrowserExtension(s.T(), "go-test")
	device, devicePubKey := tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, extension, device)

	payload := []byte(fmt.Sprintf(`{"extension_id":"%s","device_name":"%s","device_public_key":"%s"}`, extension.Id, device.Name, devicePubKey))

	response := tests.DoPost("/mobile/devices/"+device.Id+"/browser_extensions", payload, nil)
	assert.Equal(s.T(), 409, response.StatusCode)
}
