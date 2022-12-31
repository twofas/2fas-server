package tests

import (
	query "github.com/2fas/api/internal/api/mobile/app/queries"
	"github.com/2fas/api/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestMobileNotificationsTestSuite(t *testing.T) {
	suite.Run(t, new(MobileNotificationsTestSuite))
}

type MobileNotificationsTestSuite struct {
	suite.Suite
}

func (s *MobileNotificationsTestSuite) SetupTest() {
	tests.DoSuccessDelete(s.T(), "mobile/notifications")
}

func (s *MobileNotificationsTestSuite) TestCreateMobileNotification() {
	payload := []byte(`{"icon":"features","platform":"android","link":"2fas.com","message":"demo","version":"0.1"}`)

	var notification *query.MobileNotificationPresenter

	tests.DoSuccessPost(s.T(), "mobile/notifications", payload, &notification)

	assert.Equal(s.T(), "android", notification.Platform)
	assert.Equal(s.T(), "0.1", notification.Version)
	assert.Equal(s.T(), "2fas.com", notification.Link)
	assert.Equal(s.T(), "demo", notification.Message)
	assert.Equal(s.T(), "features", notification.Icon)
}

func (s *MobileNotificationsTestSuite) TestUpdateMobileNotification() {
	payload := []byte(`{"icon":"features","platform":"android","link":"2fas.com","message":"demo","version":"0.1"}`)
	var notification *query.MobileNotificationPresenter
	tests.DoSuccessPost(s.T(), "mobile/notifications", payload, &notification)

	payload = []byte(`{"icon":"youtube","platform":"ios","link":"new-2fas.com","message":"new-demo","version":"1.1"}`)
	var updatedNotification *query.MobileNotificationPresenter
	tests.DoSuccessPut(s.T(), "mobile/notifications/"+notification.Id, payload, &updatedNotification)

	assert.Equal(s.T(), "ios", updatedNotification.Platform)
	assert.Equal(s.T(), "1.1", updatedNotification.Version)
	assert.Equal(s.T(), "new-2fas.com", updatedNotification.Link)
	assert.Equal(s.T(), "new-demo", updatedNotification.Message)
	assert.Equal(s.T(), "youtube", updatedNotification.Icon)
}

func (s *MobileNotificationsTestSuite) TestDeleteMobileNotification() {
	payload := []byte(`{"icon":"features","platform":"android","link":"2fas.com","message":"demo","version":"0.1"}`)
	var notification *query.MobileNotificationPresenter
	tests.DoSuccessPost(s.T(), "mobile/notifications", payload, &notification)

	tests.DoSuccessDelete(s.T(), "mobile/notifications/"+notification.Id)

	response := tests.DoGet("mobile/notifications/"+notification.Id, nil)
	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *MobileNotificationsTestSuite) TestDeleteNotExistingMobileNotification() {
	id := uuid.New()

	response := tests.DoDelete("mobile/notifications/" + id.String())

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *MobileNotificationsTestSuite) TestFindAllNotifications() {
	payload1 := []byte(`{"icon":"features","platform":"android","link":"2fas.com","message":"demo","version":"0.1"}`)
	var notification1 *query.MobileNotificationPresenter
	tests.DoSuccessPost(s.T(), "mobile/notifications", payload1, &notification1)

	payload2 := []byte(`{"icon":"youtube","platform":"android","link":"2fas.com","message":"demo2","version":"1.1"}`)
	var notification2 *query.MobileNotificationPresenter
	tests.DoSuccessPost(s.T(), "mobile/notifications", payload2, &notification2)

	var collection []*query.MobileNotificationPresenter
	tests.DoSuccessGet(s.T(), "mobile/notifications", &collection)

	assert.Len(s.T(), collection, 2)
}

func (s *MobileNotificationsTestSuite) TestDoNotFindNotifications() {
	var collection []*query.MobileNotificationPresenter

	tests.DoSuccessGet(s.T(), "mobile/notifications", &collection)

	assert.Len(s.T(), collection, 0)
}

func (s *MobileNotificationsTestSuite) TestPublishNotification() {
	payload := []byte(`{"icon":"features","platform":"android","link":"2fas.com","message":"demo","version":"0.1"}`)
	var notification *query.MobileNotificationPresenter
	tests.DoSuccessPost(s.T(), "mobile/notifications", payload, &notification)

	var publishedNotification *query.MobileNotificationPresenter
	tests.DoSuccessPost(s.T(), "mobile/notifications/"+notification.Id+"/commands/publish", payload, &publishedNotification)

	assert.NotEmpty(s.T(), "published_at", notification.PublishedAt)
}
