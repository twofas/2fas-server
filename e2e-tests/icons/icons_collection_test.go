package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/twofas/2fas-server/e2e-tests"
)

type iconsCollectionResponse struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icons       []string `json:"icons"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func TestIconsCollectionsTestSuite(t *testing.T) {
	suite.Run(t, new(IconsCollectionsTestSuite))
}

type IconsCollectionsTestSuite struct {
	suite.Suite
}

func (s *IconsCollectionsTestSuite) SetupTest() {
	e2e_tests.RemoveAllMobileIconsCollections(s.T())
}

func (s *IconsCollectionsTestSuite) TestCreateIconsCollection() {
	payload := []byte(`
		{
			"name":"facebook",
			"description":"desc",
			"icons":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)

	var IconsCollection *iconsCollectionResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/collections", payload, &IconsCollection)

	assert.Equal(s.T(), "facebook", IconsCollection.Name)
	assert.Equal(s.T(), "desc", IconsCollection.Description)
	assert.Equal(s.T(), []string{"123e4567-e89b-12d3-a456-426614174000"}, IconsCollection.Icons)
}

func (s *IconsCollectionsTestSuite) TestUpdateIconsCollection() {
	payload := []byte(`
		{
			"name":"facebook",
			"description":"another",
			"icons":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)
	var iconsCollection *iconsCollectionResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/collections", payload, &iconsCollection)

	updatePayload := []byte(`
		{
			"name":"meta",
			"icons":["icon-1", "icon-2"]
		}
	`)

	var updatedIconsCollection *iconsCollectionResponse
	e2e_tests.DoAdminSuccessPut(s.T(), "mobile/icons/collections/"+iconsCollection.Id, updatePayload, &updatedIconsCollection)

	assert.Equal(s.T(), "meta", updatedIconsCollection.Name)
	assert.Equal(s.T(), []string{"icon-1", "icon-2"}, updatedIconsCollection.Icons)
}

func (s *IconsCollectionsTestSuite) TestDeleteIconsCollection() {
	payload := []byte(`
		{
			"name":"facebook icons",
			"icons":["icon-1", "icon-2"]
		}
	`)
	var iconsCollection *iconsCollectionResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/collections", payload, &iconsCollection)

	e2e_tests.DoAdminSuccessDelete(s.T(), "mobile/icons/collections/"+iconsCollection.Id)

	response := e2e_tests.DoAPIGet(s.T(), "mobile/icons/collections/"+iconsCollection.Id, nil)
	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *IconsCollectionsTestSuite) TestFindAllIconsCollections() {
	payload := []byte(`
		{
			"name":"facebook",
			"icons":["icon-1", "icon-2"]
		}
	`)
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/collections", payload, nil)

	payload2 := []byte(`
		{
			"name":"google",
			"description":"google google",
			"icons":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/collections", payload2, nil)

	var IconsCollections []*iconsCollectionResponse
	e2e_tests.DoAPISuccessGet(s.T(), "mobile/icons/collections", &IconsCollections)
	assert.Len(s.T(), IconsCollections, 2)
}

func (s *IconsCollectionsTestSuite) TestFindIconsCollection() {
	payload := []byte(`
		{
			"name":"just-one",
			"description":"another",
			"icons":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)
	var createdIconsCollection *iconsCollectionResponse
	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/icons/collections", payload, &createdIconsCollection)

	var IconsCollection *iconsCollectionResponse
	e2e_tests.DoAPISuccessGet(s.T(), "mobile/icons/collections/"+createdIconsCollection.Id, &IconsCollection)

	assert.Equal(s.T(), "just-one", IconsCollection.Name)
}
