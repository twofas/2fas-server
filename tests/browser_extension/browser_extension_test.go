package tests

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/internal/common/crypto"
	"github.com/twofas/2fas-server/tests"
	"net/http"
	"testing"
)

func TestBrowserExtensionTestSuite(t *testing.T) {
	suite.Run(t, new(BrowserExtensionTestSuite))
}

type BrowserExtensionTestSuite struct {
	suite.Suite
}

func (s *BrowserExtensionTestSuite) SetupTest() {
	tests.DoSuccessDelete(s.T(), "/browser_extensions")
}

func (s *BrowserExtensionTestSuite) TestCreateBrowserExtension() {
	type testCase struct {
		extensionName    string
		expectedHttpCode int
	}

	testsCases := []testCase{
		{extensionName: "", expectedHttpCode: 400},
		{extensionName: " ", expectedHttpCode: 400},
		{extensionName: "   ", expectedHttpCode: 400},
		{extensionName: "abc", expectedHttpCode: 200},
		{extensionName: "efg ", expectedHttpCode: 200},
		{extensionName: " ab123 ", expectedHttpCode: 200},
	}

	for _, tc := range testsCases {
		response := createBrowserExtension(tc.extensionName)

		assert.Equal(s.T(), tc.expectedHttpCode, response.StatusCode)
	}
}

func (s *BrowserExtensionTestSuite) TestUpdateBrowserExtension() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")

	payload := []byte(`{"name": "updated-extension-name"}`)
	tests.DoSuccessPut(s.T(), "/browser_extensions/"+browserExt.Id, payload, nil)

	var browserExtension *tests.BrowserExtensionResponse
	tests.DoSuccessGet(s.T(), "/browser_extensions/"+browserExt.Id, &browserExtension)

	assert.Equal(s.T(), "updated-extension-name", browserExtension.Name)
}

func (s *BrowserExtensionTestSuite) TestUpdateNotExistingBrowserExtension() {
	id := uuid.New()

	payload := []byte(`{"name": "updated-extension-name"}`)
	response := tests.DoPut("/browser_extensions/"+id.String(), payload, nil)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *BrowserExtensionTestSuite) TestDoNotFindNotExistingExtension() {
	notExistingId := uuid.New()

	var browserExtension *tests.BrowserExtensionResponse
	response := tests.DoGet("/browser_extensions/"+notExistingId.String(), &browserExtension)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func createBrowserExtension(name string) *http.Response {
	keyPair := crypto.GenerateKeyPair(2048)

	pubKey := crypto.PublicKeyToBase64(keyPair.PublicKey)

	payload := []byte(fmt.Sprintf(`{"name":"%s","browser_name":"go-browser","browser_version":"0.1","public_key":"%s"}`, name, pubKey))

	return tests.DoPost("/browser_extensions", payload, nil)

}
