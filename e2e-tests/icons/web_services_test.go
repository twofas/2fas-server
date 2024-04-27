package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/e2e-tests"
	"github.com/twofas/2fas-server/internal/api/icons/app/command"
)

type webServiceResponse struct {
	Id               string               `json:"id"`
	Name             string               `json:"name"`
	Description      string               `json:"description"`
	Issuers          []string             `json:"issuers"`
	Tags             []string             `json:"tags"`
	IconsCollections []string             `json:"icons_collections"`
	MatchRules       []*command.MatchRule `json:"match_rules"`
	CreatedAt        string               `json:"created_at"`
	UpdatedAt        string               `json:"updated_at"`
}

func TestWebServicesTestSuite(t *testing.T) {
	suite.Run(t, new(WebServicesTestSuite))
}

type WebServicesTestSuite struct {
	suite.Suite
}

func (s *WebServicesTestSuite) SetupTest() {
	e2e_tests.RemoveAllMobileWebServices(s.T())
}

func (s *WebServicesTestSuite) TestCreateWebService() {
	payload := []byte(`
		{
			"name":"facebook",
			"description":"desc",
			"issuers":["facebook", "meta"],
			"tags":["shitbook"],
			"icons_collections":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)

	var webService *webServiceResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/web_services", payload, &webService)

	assert.Equal(s.T(), "facebook", webService.Name)
	assert.Equal(s.T(), "desc", webService.Description)
	assert.Equal(s.T(), []string{"facebook", "meta"}, webService.Issuers)
	assert.Equal(s.T(), []string{"shitbook"}, webService.Tags)
	assert.Equal(s.T(), []string{"123e4567-e89b-12d3-a456-426614174000"}, webService.IconsCollections)
}

func (s *WebServicesTestSuite) TestCreateWebServiceWithAlreadyExistingName() {
	payload := []byte(`
		{
			"name":"facebook",
			"description":"desc",
			"issuers":["facebook", "meta"],
			"tags":["shitbook"],
			"icons_collections":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)

	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/web_services", payload, nil)
	e2e_tests.DoAdminPostAndAssertCode(s.T(), 409, "mobile/web_services", payload, nil)
}

func (s *WebServicesTestSuite) TestCreateWebServiceWithMatchRules() {
	payload := []byte(`
		{
			"name":"facebook",
			"issuers":["facebook", "meta"],
			"icons_collections":["123e4567-e89b-12d3-a456-426614174000"],
			"match_rules":[{"field":"label","text":"facebook.com","matcher":"contains","ignore_case":true}]
		}
	`)

	var webService *webServiceResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/web_services", payload, &webService)

	assert.Equal(s.T(), []*command.MatchRule{{
		Field:      "label",
		Text:       "facebook.com",
		Matcher:    "contains",
		IgnoreCase: true,
	}}, webService.MatchRules)
}

func (s *WebServicesTestSuite) TestUpdateWebService() {
	payload := []byte(`
		{
			"name":"facebook",
			"description":"another",
			"issuers":["facebook", "m.facebook"],
			"tags":["shitbook"],
			"icons_collections":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)
	var webService *webServiceResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/web_services", payload, &webService)

	updatePayload := []byte(`{
			"name":"meta",
			"issuers":["meta", "facebook"],
			"tags":["tag1", "tag2"],
			"icons_collections":["set1", "set2"]
		}
	`)

	var updatedWebService *webServiceResponse
	e2e_tests.DoAdminSuccessPut(s.T(), "mobile/web_services/"+webService.Id, updatePayload, &updatedWebService)

	assert.Equal(s.T(), "meta", updatedWebService.Name)
	assert.Equal(s.T(), []string{"meta", "facebook"}, updatedWebService.Issuers)
	assert.Equal(s.T(), []string{"tag1", "tag2"}, updatedWebService.Tags)
	assert.Equal(s.T(), []string{"set1", "set2"}, updatedWebService.IconsCollections)
}

func (s *WebServicesTestSuite) TestUpdateWebServiceMatchRule() {
	payload := []byte(`
		{
			"name":"facebook",
			"issuers":["facebook", "m.facebook"],
			"icons_collections":["123e4567-e89b-12d3-a456-426614174000"],
			"match_rules":[{"field":"label","text":"facebook.com","matcher":"contains","ignore_case":true}]
		}
	`)
	var webService *webServiceResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/web_services", payload, &webService)

	updatePayload := []byte(`{
			"name":"meta",
			"issuers":["meta", "facebook"],
			"icons_collections":["set1", "set2"],
			"match_rules":[{"field":"issuer","text":"facebook.pl","matcher":"starts_with","ignore_case":false}]
		}
	`)

	var updatedWebService *webServiceResponse
	e2e_tests.DoAdminSuccessPut(s.T(), "mobile/web_services/"+webService.Id, updatePayload, &updatedWebService)

	assert.Equal(s.T(), "issuer", updatedWebService.MatchRules[0].Field)
	assert.Equal(s.T(), "facebook.pl", updatedWebService.MatchRules[0].Text)
	assert.Equal(s.T(), "starts_with", updatedWebService.MatchRules[0].Matcher)
	assert.Equal(s.T(), false, updatedWebService.MatchRules[0].IgnoreCase)
}

func (s *WebServicesTestSuite) TestDeleteWebService() {
	payload := []byte(`
		{
			"name":"facebook",
			"description":"another",
			"issuers":["facebook", "m.facebook"],
			"tags":["shitbook"],
			"icons_collections":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)
	var webService *webServiceResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/web_services", payload, &webService)

	e2e_tests.DoAdminSuccessDelete(s.T(), "mobile/web_services/"+webService.Id)

	response := e2e_tests.DoAPIGet(s.T(), "mobile/web_services/"+webService.Id, nil)
	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *WebServicesTestSuite) TestFindAllWebServices() {
	payload := []byte(`
		{
			"name":"facebook",
			"description":"another",
			"issuers":["facebook", "m.facebook"],
			"tags":["shitbook"],
			"icons_collections":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/web_services", payload, nil)

	payload2 := []byte(`
		{
			"name":"google",
			"description":"google google",
			"issuers":["gmail", "google"],
			"tags":["google"],
			"icons_collections":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/web_services", payload2, nil)

	var webServices []*webServiceResponse
	e2e_tests.DoAPISuccessGet(s.T(), "mobile/web_services", &webServices)
	assert.Len(s.T(), webServices, 2)
}

func (s *WebServicesTestSuite) TestFindWebService() {
	payload := []byte(`
		{
			"name":"just-one",
			"description":"another",
			"issuers":["facebook", "m.facebook"],
			"tags":["shitbook"],
			"icons_collections":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)
	var createdWebService *webServiceResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/web_services", payload, &createdWebService)

	var webService *webServiceResponse
	e2e_tests.DoAPISuccessGet(s.T(), "mobile/web_services/"+createdWebService.Id, &webService)

	assert.Equal(s.T(), "just-one", webService.Name)
}
