package tests

import (
	"github.com/2fas/api/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Request2FaToken(t *testing.T) {
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")

	var tokenRequest *tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"https://facebook.com/path/nested"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	assert.Equal(t, browserExtension.Id, tokenRequest.ExtensionId)

	var tokenRequestById *tests.AuthTokenRequestResponse
	tests.DoSuccessGet(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &tokenRequestById)
	assert.Equal(t, tokenRequest.Id, tokenRequestById.Id)
	assert.Equal(t, "https://facebook.com", tokenRequestById.Domain)
}

func Test_FindAll2FaRequestsForBrowserExtension(t *testing.T) {
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")

	facebook2FaTokenRequest := []byte(`{"domain":"facebook.com"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", facebook2FaTokenRequest, nil)

	google2FaTokenRequest := []byte(`{"domain":"google.com"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", google2FaTokenRequest, nil)

	var tokenRequestsCollection []*tests.AuthTokenRequestResponse
	tests.DoSuccessGet(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests", &tokenRequestsCollection)

	assert.Len(t, tokenRequestsCollection, 2)
}

func Test_Close2FaTokenRequest(t *testing.T) {
	var tokenRequest *tests.AuthTokenRequestResponse
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")
	tokenRequestPayload := []byte(`{"domain":"facebook.com"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", tokenRequestPayload, &tokenRequest)
	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var closedTokenRequest *tests.AuthTokenRequestResponse
	tests.DoSuccessGet(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &closedTokenRequest)
	assert.Equal(t, "completed", closedTokenRequest.Status)
}

func Test_CloseNotExisting2FaTokenRequest(t *testing.T) {
	notExistingTokenRequestId := uuid.New()
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	response := tests.DoPost("browser_extensions/"+browserExtension.Id+"/2fa_requests/"+notExistingTokenRequestId.String()+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	assert.Equal(t, 404, response.StatusCode)
}

func Test_DoNotReturnClosed2FaRequests(t *testing.T) {
	var tokenRequest *tests.AuthTokenRequestResponse
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")
	tokenRequestPayload := []byte(`{"domain":"facebook.com"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", tokenRequestPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var response []*tests.AuthTokenRequestResponse
	tests.DoSuccessGet(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests", &response)
	assert.Len(t, response, 0)
}

func Test_Terminate2FaRequest(t *testing.T) {
	var tokenRequest *tests.AuthTokenRequestResponse
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")
	tokenRequestPayload := []byte(`{"domain":"facebook.com"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", tokenRequestPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"terminated"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var response *tests.AuthTokenRequestResponse
	tests.DoSuccessGet(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &response)
	assert.Equal(t, "terminated", response.Status)
}

func Test_Close2FaRequest(t *testing.T) {
	device, devicePubKey := tests.CreateDevice(t, "SM-955F", "fcm-token")
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")
	tests.PairDeviceWithBrowserExtension(t, devicePubKey, browserExtension, device)

	var tokenRequest *tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"domain.com"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var closedTokenRequest *tests.AuthTokenRequestResponse
	tests.DoSuccessGet(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id, &closedTokenRequest)
	assert.Equal(t, "completed", closedTokenRequest.Status)
}
