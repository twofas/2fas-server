package tests

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	query "github.com/twofas/2fas-server/internal/api/icons/app/queries"
	"github.com/twofas/2fas-server/tests"
)

func createIcon(t *testing.T) *query.IconPresenter {
	img := faker.New().Image().Image(120, 120)

	pngImg, err := ioutil.ReadFile(img.Name())

	if err != nil {
		t.Error(err)
	}

	iconBase64Encoded := base64.StdEncoding.EncodeToString(pngImg)

	payload := []byte(`
		{
			"name":"facebook",
			"description":"desc",
			"type":"light",
			"icon":"` + iconBase64Encoded + `"
		}
	`)

	var icon *query.IconPresenter

	tests.DoAdminAPISuccessPost(t, "mobile/icons", payload, &icon)

	return icon
}

func TestIconsTestSuite(t *testing.T) {
	suite.Run(t, new(IconsTestSuite))
}

type IconsTestSuite struct {
	suite.Suite
}

func (s *IconsTestSuite) SetupTest() {
	tests.DoAdminSuccessDelete(s.T(), "mobile/icons")
}

func (s *IconsTestSuite) TestCreateIcon() {
	icon := createIcon(s.T())

	assert.Equal(s.T(), "facebook", icon.Name)
}

func (s *IconsTestSuite) TestUpdateIcon() {
	icon := createIcon(s.T())

	updatePayload := []byte(`
		{
			"name":"meta"
		}
	`)

	var updatedIcon *query.IconPresenter
	tests.DoAdminSuccessPut(s.T(), "mobile/icons/"+icon.Id, updatePayload, &updatedIcon)

	assert.Equal(s.T(), "meta", updatedIcon.Name)
}

func (s *IconsTestSuite) TestDeleteIcon() {
	icon := createIcon(s.T())

	tests.DoAdminSuccessDelete(s.T(), "mobile/icons/"+icon.Id)

	response := tests.DoAPIGet(s.T(), "mobile/icons/"+icon.Id, nil)
	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *IconsTestSuite) TestFindAllIcons() {
	createIcon(s.T())
	createIcon(s.T())

	var Icons []*query.IconPresenter
	tests.DoAPISuccessGet(s.T(), "mobile/icons", &Icons)

	assert.Len(s.T(), Icons, 2)
}

func (s *IconsTestSuite) TestFindIcon() {
	icon := createIcon(s.T())

	var searchResult *query.IconPresenter
	tests.DoAPISuccessGet(s.T(), "mobile/icons/"+icon.Id, &searchResult)

	assert.Equal(s.T(), "facebook", searchResult.Name)
}
