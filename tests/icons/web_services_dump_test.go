package tests

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/tests"
)

func TestWebServicesDumpTestSuite(t *testing.T) {
	suite.Run(t, new(WebServicesDumpTestSuite))
}

type WebServicesDumpTestSuite struct {
	suite.Suite
}

func (s *WebServicesDumpTestSuite) SetupTest() {
	tests.RemoveAllMobileIcons(s.T())
	tests.RemoveAllMobileIconsCollections(s.T())
	tests.RemoveAllMobileWebServices(s.T())
}

func (s *WebServicesDumpTestSuite) TestWebServicesDump() {
	createWebService(s.T())
	createWebService(s.T())

	response := tests.DoAPIGet(s.T(), "mobile/web_services/dump", nil)

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

	tests.DoAdminAPISuccessPost(t, "mobile/web_services", payload, &webService)

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

	tests.DoAdminAPISuccessPost(t, "mobile/icons/collections", payload, &createdIconsCollection)

	return createdIconsCollection
}
