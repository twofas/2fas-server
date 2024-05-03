package tests

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/e2e-tests"
)

func TestMobileDeviceExtensionTestSuite(t *testing.T) {
	suite.Run(t, new(MobileDeviceExtensionTestSuite))
}

type MobileDeviceExtensionTestSuite struct {
	suite.Suite
}

func (s *MobileDeviceExtensionTestSuite) SetupTest() {
	e2e_tests.RemoveAllMobileDevices(s.T())
	e2e_tests.RemoveAllBrowserExtensions(s.T())
	e2e_tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *MobileDeviceExtensionTestSuite) TestDoNotFindExtensionsForNotExistingDevice() {
	notExistingDeviceId := uuid.New()

	response := e2e_tests.DoAPIGet(s.T(), "/mobile/devices/"+notExistingDeviceId.String()+"/browser_extensions", nil)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *MobileDeviceExtensionTestSuite) TestDoNotFindNotExistingMobileDeviceExtension() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)

	notExistingExtensionId := uuid.New()
	response := e2e_tests.DoAPIGet(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/"+notExistingExtensionId.String(), nil)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *MobileDeviceExtensionTestSuite) Test_FindExtensionForDevice() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt, device)

	var deviceBrowserExtension *e2e_tests.BrowserExtensionResponse
	e2e_tests.DoAPISuccessGet(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/"+browserExt.Id, &deviceBrowserExtension)

	assert.Equal(s.T(), browserExt.Id, deviceBrowserExtension.Id)
}

func (s *MobileDeviceExtensionTestSuite) Test_FindAllDeviceExtensions() {
	browserExt1 := e2e_tests.CreateBrowserExtension(s.T(), "go-test-1")
	browserExt2 := e2e_tests.CreateBrowserExtension(s.T(), "go-test-2")
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "go-test-device", "some-device-id")

	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt1, device)
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt2, device)

	var deviceBrowserExtensions []*e2e_tests.BrowserExtensionResponse
	e2e_tests.DoAPISuccessGet(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/", &deviceBrowserExtensions)

	assert.Len(s.T(), deviceBrowserExtensions, 2)
}

func (s *MobileDeviceExtensionTestSuite) Test_DisconnectExtensionFromDevice() {
	browserExt1 := e2e_tests.CreateBrowserExtension(s.T(), "go-test")
	browserExt2 := e2e_tests.CreateBrowserExtension(s.T(), "go-test")
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt1, device)
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExt2, device)

	e2e_tests.DoAPISuccessDelete(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/"+browserExt1.Id)

	var deviceBrowserExtension1 *e2e_tests.BrowserExtensionResponse
	response := e2e_tests.DoAPIGet(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/"+browserExt1.Id, &deviceBrowserExtension1)
	assert.Equal(s.T(), 404, response.StatusCode)

	var deviceBrowserExtension2 *e2e_tests.BrowserExtensionResponse
	e2e_tests.DoAPISuccessGet(s.T(), "/mobile/devices/"+device.Id+"/browser_extensions/"+browserExt2.Id, &deviceBrowserExtension2)
	assert.Equal(s.T(), browserExt2.Id, deviceBrowserExtension2.Id)
}

func (s *MobileDeviceExtensionTestSuite) TestExtensionHasAlreadyBeenConnected() {
	extension := e2e_tests.CreateBrowserExtension(s.T(), "go-test")
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "go-test-device", "some-device-id")
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, extension, device)

	payload := []byte(fmt.Sprintf(`{"extension_id":"%s","device_name":"%s","device_public_key":"%s"}`, extension.Id, device.Name, devicePubKey))

	e2e_tests.DoAPIPostAndAssertCode(s.T(), 409, "/mobile/devices/"+device.Id+"/browser_extensions", payload, nil)
}
