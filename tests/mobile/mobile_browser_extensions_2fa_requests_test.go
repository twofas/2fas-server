package tests

import (
	"github.com/2fas/api/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GetPending2FaRequests(t *testing.T) {
	device, devicePubKey := tests.CreateDevice(t, "SM-955F", "fcm-token")
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")
	tests.PairDeviceWithBrowserExtension(t, devicePubKey, browserExtension, device)

	var tokenRequest *tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"domain.com"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	var tokenRequestsCollection []*tests.AuthTokenRequestResponse
	tests.DoSuccessGet(t, "mobile/devices/"+device.Id+"/browser_extensions/2fa_requests", &tokenRequestsCollection)
	assert.Len(t, tokenRequestsCollection, 1)
}

func Test_DoNotReturnCompleted2FaRequests(t *testing.T) {
	device, devicePubKey := tests.CreateDevice(t, "SM-955F", "fcm-token")
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")
	tests.PairDeviceWithBrowserExtension(t, devicePubKey, browserExtension, device)

	var tokenRequest *tests.AuthTokenRequestResponse
	request2FaTokenPayload := []byte(`{"domain":"domain.com"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/commands/request_2fa_token", request2FaTokenPayload, &tokenRequest)

	closeTokenRequestPayload := []byte(`{"status":"completed"}`)
	tests.DoSuccessPost(t, "browser_extensions/"+browserExtension.Id+"/2fa_requests/"+tokenRequest.Id+"/commands/close_2fa_request", closeTokenRequestPayload, nil)

	var tokenRequestsCollection []*tests.AuthTokenRequestResponse
	tests.DoSuccessGet(t, "mobile/devices/"+device.Id+"/browser_extensions/2fa_requests", &tokenRequestsCollection)
	assert.Len(t, tokenRequestsCollection, 0)
}
