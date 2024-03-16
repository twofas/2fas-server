package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/e2e-tests"
)

func TestBrowserExtensionTwoFactorAuthTestSuite(t *testing.T) {
	suite.Run(t, new(BrowserExtensionTwoFactorAuthTestSuite))
}

type BrowserExtensionTwoFactorAuthTestSuite struct {
	suite.Suite
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) SetupTest() {
	e2e_tests.RemoveAllMobileDevices(s.T())
	e2e_tests.RemoveAllBrowserExtensions(s.T())
	e2e_tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestRequest2FaToken() {
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")

	var tokenRequest *e2e_tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"https://facebook.com/path/nested"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	assert.Equal(s.T(), browserExtension.Id, tokenRequest.ExtensionId)

	var tokenRequestById *e2e_tests.AuthTokenRequestResponse
	e2e_tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &tokenRequestById)
	assert.Equal(s.T(), tokenRequest.Id, tokenRequestById.Id)
	assert.Equal(s.T(), "https://facebook.com", tokenRequestById.Domain)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestFindAll2FaRequestsForBrowserExtension() {
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")

	facebook2FaTokenRequest := []byte(`{"domain":"facebook.com"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", facebook2FaTokenRequest, nil)

	google2FaTokenRequest := []byte(`{"domain":"google.com"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", google2FaTokenRequest, nil)

	var tokenRequestsCollection []*e2e_tests.AuthTokenRequestResponse
	e2e_tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests", &tokenRequestsCollection)

	assert.Len(s.T(), tokenRequestsCollection, 2)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestClose2FaTokenRequest() {
	var tokenRequest *e2e_tests.AuthTokenRequestResponse
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")
	tokenRequestPayload := []byte(`{"domain":"facebook.com"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", tokenRequestPayload, &tokenRequest)
	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var closedTokenRequest *e2e_tests.AuthTokenRequestResponse
	e2e_tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &closedTokenRequest)
	assert.Equal(s.T(), "completed", closedTokenRequest.Status)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestCloseNotExisting2FaTokenRequest() {
	notExistingTokenRequestId := uuid.New()
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	e2e_tests.DoAPIPostAndAssertCode(s.T(), 404, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+notExistingTokenRequestId.String()+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestDoNotReturnClosed2FaRequests() {
	var tokenRequest *e2e_tests.AuthTokenRequestResponse
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")
	tokenRequestPayload := []byte(`{"domain":"facebook.com"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", tokenRequestPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var response []*e2e_tests.AuthTokenRequestResponse
	e2e_tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests", &response)
	assert.Len(s.T(), response, 0)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestTerminate2FaRequest() {
	var tokenRequest *e2e_tests.AuthTokenRequestResponse
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")
	tokenRequestPayload := []byte(`{"domain":"facebook.com"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", tokenRequestPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"terminated"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var response *e2e_tests.AuthTokenRequestResponse
	e2e_tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &response)
	assert.Equal(s.T(), "terminated", response.Status)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestClose2FaRequest() {
	device, devicePubKey := e2e_tests.CreateDevice(s.T(), "SM-955F", "fcm-token")
	browserExtension := e2e_tests.CreateBrowserExtension(s.T(), "go-ext")
	e2e_tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	var tokenRequest *e2e_tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"domain.com"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	e2e_tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var closedTokenRequest *e2e_tests.AuthTokenRequestResponse
	e2e_tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &closedTokenRequest)
	assert.Equal(s.T(), "completed", closedTokenRequest.Status)
}
