package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/tests"
)

func TestBrowserExtensionTwoFactorAuthTestSuite(t *testing.T) {
	suite.Run(t, new(BrowserExtensionTwoFactorAuthTestSuite))
}

type BrowserExtensionTwoFactorAuthTestSuite struct {
	suite.Suite
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) SetupTest() {
	tests.RemoveAllMobileDevices(s.T())
	tests.RemoveAllBrowserExtensions(s.T())
	tests.RemoveAllBrowserExtensionsDevices(s.T())
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestRequest2FaToken() {
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")

	var tokenRequest *tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"https://facebook.com/path/nested"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	assert.Equal(s.T(), browserExtension.Id, tokenRequest.ExtensionId)

	var tokenRequestById *tests.AuthTokenRequestResponse
	tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &tokenRequestById)
	assert.Equal(s.T(), tokenRequest.Id, tokenRequestById.Id)
	assert.Equal(s.T(), "https://facebook.com", tokenRequestById.Domain)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestFindAll2FaRequestsForBrowserExtension() {
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")

	facebook2FaTokenRequest := []byte(`{"domain":"facebook.com"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", facebook2FaTokenRequest, nil)

	google2FaTokenRequest := []byte(`{"domain":"google.com"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", google2FaTokenRequest, nil)

	var tokenRequestsCollection []*tests.AuthTokenRequestResponse
	tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests", &tokenRequestsCollection)

	assert.Len(s.T(), tokenRequestsCollection, 2)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestClose2FaTokenRequest() {
	var tokenRequest *tests.AuthTokenRequestResponse
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")
	tokenRequestPayload := []byte(`{"domain":"facebook.com"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", tokenRequestPayload, &tokenRequest)
	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var closedTokenRequest *tests.AuthTokenRequestResponse
	tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &closedTokenRequest)
	assert.Equal(s.T(), "completed", closedTokenRequest.Status)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestCloseNotExisting2FaTokenRequest() {
	notExistingTokenRequestId := uuid.New()
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	tests.DoAPIPostAndAssertCode(s.T(), 404, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+notExistingTokenRequestId.String()+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestDoNotReturnClosed2FaRequests() {
	var tokenRequest *tests.AuthTokenRequestResponse
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")
	tokenRequestPayload := []byte(`{"domain":"facebook.com"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", tokenRequestPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var response []*tests.AuthTokenRequestResponse
	tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests", &response)
	assert.Len(s.T(), response, 0)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestTerminate2FaRequest() {
	var tokenRequest *tests.AuthTokenRequestResponse
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")
	tokenRequestPayload := []byte(`{"domain":"facebook.com"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", tokenRequestPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"terminated"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var response *tests.AuthTokenRequestResponse
	tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &response)
	assert.Equal(s.T(), "terminated", response.Status)
}

func (s *BrowserExtensionTwoFactorAuthTestSuite) TestClose2FaRequest() {
	device, devicePubKey := tests.CreateDevice(s.T(), "SM-955F", "fcm-token")
	browserExtension := tests.CreateBrowserExtension(s.T(), "go-ext")
	tests.PairDeviceWithBrowserExtension(s.T(), devicePubKey, browserExtension, device)

	var tokenRequest *tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"domain.com"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	tests.DoAPISuccessPost(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var closedTokenRequest *tests.AuthTokenRequestResponse
	tests.DoAPISuccessGet(s.T(), "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &closedTokenRequest)
	assert.Equal(s.T(), "completed", closedTokenRequest.Status)
}
