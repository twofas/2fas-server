package tests

import (
	"fmt"
	"github.com/2fas/api/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
)

func TestWebServicesDumpTestSuite(t *testing.T) {
	suite.Run(t, new(WebServicesDumpTestSuite))
}

type WebServicesDumpTestSuite struct {
	suite.Suite
}

func (s *WebServicesDumpTestSuite) SetupTest() {
	tests.DoSuccessDelete(s.T(), "mobile/icons")
	tests.DoSuccessDelete(s.T(), "mobile/icons/collections")
	tests.DoSuccessDelete(s.T(), "mobile/web_services")
}

func (s *WebServicesDumpTestSuite) TestWebServicesDump() {
	createWebService(s.T())
	createWebService(s.T())

	response := tests.DoGet("mobile/web_services/dump", nil)

	assert.Equal(s.T(), 200, response.StatusCode)
}

func createWebService(t *testing.T) *webServiceResponse {
	iconsCollection := createIconsCollection(t)

	payload := []byte(`
		{
			"name":"` + fmt.Sprintf("service-%d", rand.Int()) + `",
			"description":"another",
			"issuers":["facebook", "m.facebook"],
			"tags":["shitbook"],
			"icons_collections":["` + iconsCollection.Id + `"]
		}
	`)

	var webService *webServiceResponse

	tests.DoSuccessPost(t, "mobile/web_services", payload, &webService)

	return webService
}

func createIconsCollection(t *testing.T) *iconsCollectionResponse {
	icon := createIcon(t)

	payload := []byte(`
		{
			"name":"just-one",
			"description":"another",
			"icons":["` + icon.Id + `"]
		}
	`)

	var createdIconsCollection *iconsCollectionResponse

	tests.DoSuccessPost(t, "mobile/icons/collections", payload, &createdIconsCollection)

	return createdIconsCollection
}
