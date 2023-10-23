package tests

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/internal/api/icons/app/queries"
	"github.com/twofas/2fas-server/tests"
)

func TestIconsRequestsTestSuite(t *testing.T) {
	suite.Run(t, new(IconsRequestsTestSuite))
}

type IconsRequestsTestSuite struct {
	suite.Suite
}

func (s *IconsRequestsTestSuite) SetupTest() {
	tests.RemoveAllMobileWebServices(s.T())
	tests.RemoveAllMobileIcons(s.T())
	tests.RemoveAllMobileIconsCollections(s.T())
	tests.RemoveAllMobileIconsRequests(s.T())
}

func (s *IconsRequestsTestSuite) TestCreateIconRequest() {
	iconRequest := createIconRequest(s.T(), "service")

	assert.Equal(s.T(), "service", iconRequest.ServiceName)
	assert.Equal(s.T(), "desc", iconRequest.Description)
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

	tests.DoAPIPostAndAssertCode(s.T(), 400, "mobile/icons/requests", payload, &iconRequest)
}

func (s *IconsRequestsTestSuite) TestDeleteIconRequest() {
	iconRequest := createIconRequest(s.T(), "service")

	tests.DoAdminSuccessDelete(s.T(), "mobile/icons/requests/"+iconRequest.Id)

	response := tests.DoAPIGet(s.T(), "mobile/icons/requests/"+iconRequest.Id, nil)
	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *IconsRequestsTestSuite) TestFindAllIconsRequests() {
	createIconRequest(s.T(), "service1")
	createIconRequest(s.T(), "service2")

	var iconsRequests []*queries.IconRequestPresenter
	tests.DoAPISuccessGet(s.T(), "mobile/icons/requests", &iconsRequests)

	assert.Len(s.T(), iconsRequests, 2)
}

func (s *IconsRequestsTestSuite) TestFindIconRequest() {
	iconRequest := createIconRequest(s.T(), "service")

	var searchResult *queries.IconPresenter
	tests.DoAdminSuccessGet(s.T(), "mobile/icons/requests/"+iconRequest.Id, &searchResult)

	assert.Equal(s.T(), "service", searchResult.Name)
}

func (s *IconsRequestsTestSuite) TestTransformIconRequestIntoWebService() {
	iconRequest := createIconRequest(s.T(), "service")

	var result *queries.WebServicePresenter
	tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/requests/"+iconRequest.Id+"/commands/transform_to_web_service", nil, &result)

	assert.Equal(s.T(), "service", result.Name)
}

func (s *IconsRequestsTestSuite) TestTransformSingleIconRequestsIntoWebServiceFromManyRequestsWithSameServiceName() {
	iconRequest := createIconRequest(s.T(), "service")
	createIconRequest(s.T(), "service")

	var result *queries.WebServicePresenter
	tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/requests/"+iconRequest.Id+"/commands/transform_to_web_service", nil, &result)

	var icons []*queries.IconPresenter
	tests.DoAPIGet(s.T(), "mobile/icons", &icons)

	assert.Len(s.T(), icons, 1)
}

func (s *IconsRequestsTestSuite) TestTransformIconRequestWithAlreadyExistingWebService() {
	webService := createWebService(s.T())
	iconRequest := createIconRequest(s.T(), webService.Name)

	var result *queries.WebServicePresenter
	tests.DoAdminPostAndAssertCode(s.T(), 409, "mobile/icons/requests/"+iconRequest.Id+"/commands/transform_to_web_service", nil, &result)
}

func (s *IconsRequestsTestSuite) TestUpdateWebServiceFromIconRequest() {
	iconRequest := createIconRequest(s.T(), "service")
	webService := createWebService(s.T())

	var result *queries.WebServicePresenter
	payload := []byte(`{"web_service_id":"` + webService.Id + `"}`)
	tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/requests/"+iconRequest.Id+"/commands/update_web_service", payload, &result)

	assert.Equal(s.T(), webService.Name, result.Name)
}

func createIconRequest(t *testing.T, serviceName string) *queries.IconRequestPresenter {
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

	tests.DoAPISuccessPost(t, "mobile/icons/requests", payload, &iconRequest)

	return iconRequest
}
