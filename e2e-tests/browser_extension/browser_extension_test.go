package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
	"github.com/twofas/2fas-server/internal/common/crypto"
)

func TestBrowserExtensionTestSuite(t *testing.T) {
	suite.Run(t, new(BrowserExtensionTestSuite))
}

type BrowserExtensionTestSuite struct {
	suite.Suite
}

func (s *BrowserExtensionTestSuite) SetupTest() {
	e2e_tests.RemoveAllBrowserExtensions(s.T())
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

		s.Equal(tc.expectedHttpCode, response.StatusCode)
	}
}

func (s *BrowserExtensionTestSuite) TestUpdateBrowserExtension() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")

	payload := []byte(`{"name": "updated-extension-name"}`)
	e2e_tests.DoAPISuccessPut(s.T(), "/browser_extensions/"+browserExt.Id, payload, nil)

	var browserExtension *e2e_tests.BrowserExtensionResponse
	e2e_tests.DoAPISuccessGet(s.T(), "/browser_extensions/"+browserExt.Id, &browserExtension)

	s.Equal("updated-extension-name", browserExtension.Name)
}

func (s *BrowserExtensionTestSuite) TestUpdateNotExistingBrowserExtension() {
	id := uuid.New()

	payload := []byte(`{"name": "updated-extension-name"}`)
	response := e2e_tests.DoAPIRequest(s.T(), "/browser_extensions/"+id.String(), http.MethodPut, payload, nil)

	s.Equal(404, response.StatusCode)
}

func (s *BrowserExtensionTestSuite) TestUpdateBrowserExtensionSetEmptyName() {
	browserExt := e2e_tests.CreateBrowserExtension(s.T(), "go-test")

	payload := []byte(`{"name": ""}`)
	response := e2e_tests.DoAPIRequest(s.T(), "/browser_extensions/"+browserExt.Id, http.MethodPut, payload, nil)

	s.Equal(400, response.StatusCode)
}

func (s *BrowserExtensionTestSuite) TestDoNotFindNotExistingExtension() {
	notExistingId := uuid.New()

	var browserExtension *e2e_tests.BrowserExtensionResponse
	response := e2e_tests.DoAPIGet(s.T(), "/browser_extensions/"+notExistingId.String(), &browserExtension)

	s.Equal(404, response.StatusCode)
}

func createBrowserExtension(t *testing.T, name string) *http.Response {
	t.Helper()
	keyPair := crypto.GenerateKeyPair(2048)

	pubKey := crypto.PublicKeyToBase64(keyPair.PublicKey)

	payload := []byte(fmt.Sprintf(`{"name":"%s","browser_name":"go-browser","browser_version":"0.1","public_key":"%s"}`, name, pubKey))

	return e2e_tests.DoAPIRequest(t, "/browser_extensions", http.MethodPost, payload, nil)
}
