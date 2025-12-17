package tests

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
)

func TestWebServicesDumpTestSuite(t *testing.T) {
	suite.Run(t, new(WebServicesDumpTestSuite))
}

type WebServicesDumpTestSuite struct {
	suite.Suite
}

func (s *WebServicesDumpTestSuite) SetupTest() {
	e2e_tests.RemoveAllMobileIcons(s.T())
	e2e_tests.RemoveAllMobileIconsCollections(s.T())
	e2e_tests.RemoveAllMobileWebServices(s.T())
}

func (s *WebServicesDumpTestSuite) TestWebServicesDump() {
	createWebService(s.T())
	createWebService(s.T())

	response := e2e_tests.DoAPIGet(s.T(), "mobile/web_services/dump", nil)

	s.Equal(200, response.StatusCode)
}

func createWebService(t *testing.T) *webServiceResponse {
	t.Helper()
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

	e2e_tests.DoAdminAPISuccessPost(t, "mobile/web_services", payload, &webService)

	return webService
}

func createIconsCollection(t *testing.T) *iconsCollectionResponse {
	t.Helper()
	icon := createIcon(t)

	payload := []byte(`
		{
			"name":"just-one",
			"description":"another",
			"icons":["` + icon.Id + `"]
		}
	`)

	var createdIconsCollection *iconsCollectionResponse

	e2e_tests.DoAdminAPISuccessPost(t, "mobile/icons/collections", payload, &createdIconsCollection)

	return createdIconsCollection
}
