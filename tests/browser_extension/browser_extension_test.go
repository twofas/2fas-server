package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/tests"
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
