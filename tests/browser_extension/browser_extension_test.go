package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/internal/common/crypto"
	"github.com/twofas/2fas-server/tests"
)

func TestBrowserExtensionTestSuite(t *testing.T) {
	suite.Run(t, new(BrowserExtensionTestSuite))
}

type BrowserExtensionTestSuite struct {
	suite.Suite
}

func (s *BrowserExtensionTestSuite) SetupTest() {
	tests.RemoveAllBrowserExtensions(s.T())
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
		response := createBrowserExtension(s.T(), tc.extensionName)

		assert.Equal(s.T(), tc.expectedHttpCode, response.StatusCode)
	}
}

func (s *BrowserExtensionTestSuite) TestUpdateBrowserExtension() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")

	payload := []byte(`{"name": "updated-extension-name"}`)
	tests.DoAPISuccessPut(s.T(), "/browser_extensions/"+browserExt.Id, payload, nil)

	var browserExtension *tests.BrowserExtensionResponse
	tests.DoAPISuccessGet(s.T(), "/browser_extensions/"+browserExt.Id, &browserExtension)

	assert.Equal(s.T(), "updated-extension-name", browserExtension.Name)
}

func (s *BrowserExtensionTestSuite) TestUpdateNotExistingBrowserExtension() {
	id := uuid.New()

	payload := []byte(`{"name": "updated-extension-name"}`)
	response := tests.DoAPIRequest(s.T(), "/browser_extensions/"+id.String(), http.MethodPut, payload, nil)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *BrowserExtensionTestSuite) TestUpdateBrowserExtensionSetEmptyName() {
	browserExt := tests.CreateBrowserExtension(s.T(), "go-test")

	payload := []byte(`{"name": ""}`)
	response := tests.DoAPIRequest(s.T(), "/browser_extensions/"+browserExt.Id, http.MethodPut, payload, nil)

	assert.Equal(s.T(), 400, response.StatusCode)
}

func (s *BrowserExtensionTestSuite) TestDoNotFindNotExistingExtension() {
	notExistingId := uuid.New()

	var browserExtension *tests.BrowserExtensionResponse
	response := tests.DoAPIGet(s.T(), "/browser_extensions/"+notExistingId.String(), &browserExtension)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func createBrowserExtension(t *testing.T, name string) *http.Response {
	keyPair := crypto.GenerateKeyPair(2048)

	pubKey := crypto.PublicKeyToBase64(keyPair.PublicKey)

	payload := []byte(fmt.Sprintf(`{"name":"%s","browser_name":"go-browser","browser_version":"0.1","public_key":"%s"}`, name, pubKey))

	return tests.DoAPIRequest(t, "/browser_extensions", http.MethodPost, payload, nil)

}
