package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
)

func TestMobileDeviceExtensionIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(MobileDeviceExtensionIntegrationTestSuite))
}

type MobileDeviceExtensionIntegrationTestSuite struct {
	suite.Suite
}

func (s *MobileDeviceExtensionIntegrationTestSuite) SetupTest() {
	e2e_tests.RemoveAllMobileDevices(s.T())
	e2e_tests.RemoveAllBrowserExtensions(s.T())
	e2e_tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *MobileDeviceExtensionIntegrationTestSuite) TestGetPending2FaRequests() {
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "SM-955F", "fcm-token")
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	var tokenRequest *e2e_tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"domain.com"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	var tokenRequestsCollection []*e2e_tests.AuthTokenRequestResponse
	e2e_tests.DoAPISuccessGet(s.T(), "mobile/devices/"+device.Id+"/browser_extensions/2fa_requests", &tokenRequestsCollection)
	s.Len(tokenRequestsCollection, 1)
}

func (s *MobileDeviceExtensionIntegrationTestSuite) TestDoNotReturnCompleted2FaRequests() {
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "SM-955F", "fcm-token")
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	var tokenRequest *e2e_tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"domain.com"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var tokenRequestsCollection []*e2e_tests.AuthTokenRequestResponse
	e2e_tests.DoAPISuccessGet(s.T(), "mobile/devices/"+device.Id+"/browser_extensions/2fa_requests", &tokenRequestsCollection)
	s.Empty(tokenRequestsCollection)
}
