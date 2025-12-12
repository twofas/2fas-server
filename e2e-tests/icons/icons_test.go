package tests

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/suite"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
	query "github.com/twofas/2fas-server/internal/api/icons/app/queries"
)

func createIcon(t *testing.T) *query.IconPresenter {
	t.Helper()
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

	e2e_tests.DoAdminAPISuccessPost(t, "mobile/icons", payload, &icon)

	return icon
}

func TestIconsTestSuite(t *testing.T) {
	suite.Run(t, new(IconsTestSuite))
}

type IconsTestSuite struct {
	suite.Suite
}

func (s *IconsTestSuite) SetupTest() {
	e2e_tests.DoAdminSuccessDelete(s.T(), "mobile/icons")
}

func (s *IconsTestSuite) TestCreateIcon() {
	icon := createIcon(s.T())

	s.Equal("facebook", icon.Name)
}

func (s *IconsTestSuite) TestUpdateIcon() {
	icon := createIcon(s.T())

	updatePayload := []byte(`
		{
			"name":"meta"
		}
	`)

	var updatedIcon *query.IconPresenter
	e2e_tests.DoAdminSuccessPut(s.T(), "mobile/icons/"+icon.Id, updatePayload, &updatedIcon)

	s.Equal("meta", updatedIcon.Name)
}

func (s *IconsTestSuite) TestDeleteIcon() {
	icon := createIcon(s.T())

	e2e_tests.DoAdminSuccessDelete(s.T(), "mobile/icons/"+icon.Id)

	response := e2e_tests.DoAPIGet(s.T(), "mobile/icons/"+icon.Id, nil)
	s.Equal(404, response.StatusCode)
}

func (s *IconsTestSuite) TestFindAllIcons() {
	createIcon(s.T())
	createIcon(s.T())

	var Icons []*query.IconPresenter
	e2e_tests.DoAPISuccessGet(s.T(), "mobile/icons", &Icons)

	s.Len(Icons, 2)
}

func (s *IconsTestSuite) TestFindIcon() {
	icon := createIcon(s.T())

	var searchResult *query.IconPresenter
	e2e_tests.DoAPISuccessGet(s.T(), "mobile/icons/"+icon.Id, &searchResult)

	s.Equal("facebook", searchResult.Name)
}
