package tests

import (
	"github.com/2fas/api/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
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
	tests.DoSuccessDelete(s.T(), "mobile/icons/collections")
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
	tests.DoSuccessPost(s.T(), "mobile/icons/collections", payload, &IconsCollection)

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
	tests.DoSuccessPost(s.T(), "mobile/icons/collections", payload, &iconsCollection)

	updatePayload := []byte(`
		{
			"name":"meta",
			"icons":["icon-1", "icon-2"]
		}
	`)

	var updatedIconsCollection *iconsCollectionResponse
	tests.DoSuccessPut(s.T(), "mobile/icons/collections/"+iconsCollection.Id, updatePayload, &updatedIconsCollection)

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
	tests.DoSuccessPost(s.T(), "mobile/icons/collections", payload, &iconsCollection)

	tests.DoSuccessDelete(s.T(), "mobile/icons/collections/"+iconsCollection.Id)

	response := tests.DoGet("mobile/icons/collections/"+iconsCollection.Id, nil)
	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *IconsCollectionsTestSuite) TestFindAllIconsCollections() {
	payload := []byte(`
		{
			"name":"facebook",
			"icons":["icon-1", "icon-2"]
		}
	`)
	tests.DoSuccessPost(s.T(), "mobile/icons/collections", payload, nil)

	payload2 := []byte(`
		{
			"name":"google",
			"description":"google google",
			"icons":["123e4567-e89b-12d3-a456-426614174000"]
		}
	`)
	tests.DoSuccessPost(s.T(), "mobile/icons/collections", payload2, nil)

	var IconsCollections []*iconsCollectionResponse
	tests.DoSuccessGet(s.T(), "mobile/icons/collections", &IconsCollections)
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
	tests.DoSuccessPost(s.T(), "mobile/icons/collections", payload, &createdIconsCollection)

	var IconsCollection *iconsCollectionResponse
	tests.DoSuccessGet(s.T(), "mobile/icons/collections/"+createdIconsCollection.Id, &IconsCollection)

	assert.Equal(s.T(), "just-one", IconsCollection.Name)
}
