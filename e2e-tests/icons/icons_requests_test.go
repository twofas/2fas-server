package tests

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/suite"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
	"github.com/twofas/2fas-server/internal/api/icons/app/queries"
)

func TestIconsRequestsTestSuite(t *testing.T) {
	suite.Run(t, new(IconsRequestsTestSuite))
}

type IconsRequestsTestSuite struct {
	suite.Suite
}

func (s *IconsRequestsTestSuite) SetupTest() {
	e2e_tests.RemoveAllMobileWebServices(s.T())
	e2e_tests.RemoveAllMobileIcons(s.T())
	e2e_tests.RemoveAllMobileIconsCollections(s.T())
	e2e_tests.RemoveAllMobileIconsRequests(s.T())
}

func (s *IconsRequestsTestSuite) TestCreateIconRequest() {
	iconRequest := createIconRequest(s.T(), "service")

	s.Equal("service", iconRequest.ServiceName)
	s.Equal("desc", iconRequest.Description)
}

func (s *IconsRequestsTestSuite) TestCreateIconRequestWithNotAllowedIconDimensions() {
	img := faker.New().Image().Image(120, 60)

	pngImg, err := ioutil.ReadFile(img.Name())

	if err != nil {
		s.T().Error(err)
	}

	iconBase64Encoded := base64.StdEncoding.EncodeToString(pngImg)

	payload := []byte(`
		{
			"caller_id":"some-caller-uniq-name",
			"service_name":"some-service",
			"issuers": ["fb"],
			"description":"desc",
			"light_icon":"` + iconBase64Encoded + `"
		}
	`)

	var iconRequest *queries.IconRequestPresenter

	e2e_tests.DoAPIPostAndAssertCode(s.T(), 400, "mobile/icons/requests", payload, &iconRequest)
}

func (s *IconsRequestsTestSuite) TestDeleteIconRequest() {
	iconRequest := createIconRequest(s.T(), "service")

	e2e_tests.DoAdminSuccessDelete(s.T(), "mobile/icons/requests/"+iconRequest.Id)

	response := e2e_tests.DoAPIGet(s.T(), "mobile/icons/requests/"+iconRequest.Id, nil)
	s.Equal(404, response.StatusCode)
}

func (s *IconsRequestsTestSuite) TestFindAllIconsRequests() {
	createIconRequest(s.T(), "service1")
	createIconRequest(s.T(), "service2")

	var iconsRequests []*queries.IconRequestPresenter
	e2e_tests.DoAPISuccessGet(s.T(), "mobile/icons/requests", &iconsRequests)

	s.Len(iconsRequests, 2)
}

func (s *IconsRequestsTestSuite) TestFindIconRequest() {
	iconRequest := createIconRequest(s.T(), "service")

	var searchResult *queries.IconPresenter
	e2e_tests.DoAdminSuccessGet(s.T(), "mobile/icons/requests/"+iconRequest.Id, &searchResult)

	s.Equal("service", searchResult.Name)
}

func (s *IconsRequestsTestSuite) TestTransformIconRequestIntoWebService() {
	iconRequest := createIconRequest(s.T(), "service")

	var result *queries.WebServicePresenter
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/requests/"+iconRequest.Id+"/commands/transform_to_web_service", nil, &result)

	s.Equal("service", result.Name)
}

func (s *IconsRequestsTestSuite) TestTransformSingleIconRequestsIntoWebServiceFromManyRequestsWithSameServiceName() {
	iconRequest := createIconRequest(s.T(), "service")
	createIconRequest(s.T(), "service")

	var result *queries.WebServicePresenter
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/requests/"+iconRequest.Id+"/commands/transform_to_web_service", nil, &result)

	var icons []*queries.IconPresenter
	e2e_tests.DoAPIGet(s.T(), "mobile/icons", &icons)

	s.Len(icons, 1)
}

func (s *IconsRequestsTestSuite) TestTransformIconRequestWithAlreadyExistingWebService() {
	webService := createWebService(s.T())
	iconRequest := createIconRequest(s.T(), webService.Name)

	var result *queries.WebServicePresenter
	e2e_tests.DoAdminPostAndAssertCode(s.T(), 409, "mobile/icons/requests/"+iconRequest.Id+"/commands/transform_to_web_service", nil, &result)
}

func (s *IconsRequestsTestSuite) TestUpdateWebServiceFromIconRequest() {
	iconRequest := createIconRequest(s.T(), "service")
	webService := createWebService(s.T())

	var result *queries.WebServicePresenter
	payload := []byte(`{"web_service_id":"` + webService.Id + `"}`)
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/requests/"+iconRequest.Id+"/commands/update_web_service", payload, &result)

	s.Equal(webService.Name, result.Name)

	var iconsCollections []string
	if err := json.Unmarshal(result.IconsCollections, &iconsCollections); err != nil {
		s.NoError(err)
	}
	s.Equal(webService.IconsCollections, iconsCollections, "icons collections id should not change")
}

func createIconRequest(t *testing.T, serviceName string) *queries.IconRequestPresenter {
	t.Helper()
	if serviceName == "" {
		serviceName = "some-service"
	}

	img := faker.New().Image().Image(120, 120)

	pngImg, err := ioutil.ReadFile(img.Name())

	if err != nil {
		t.Error(err)
	}

	iconBase64Encoded := base64.StdEncoding.EncodeToString(pngImg)

	payload := []byte(`
		{
			"caller_id":"some-caller-uniq-name",
			"service_name":"` + serviceName + `",
			"issuers": ["fb"],
			"description":"desc",
			"light_icon":"` + iconBase64Encoded + `"
		}
	`)

	var iconRequest *queries.IconRequestPresenter

	e2e_tests.DoAPISuccessPost(t, "mobile/icons/requests", payload, &iconRequest)

	return iconRequest
}
