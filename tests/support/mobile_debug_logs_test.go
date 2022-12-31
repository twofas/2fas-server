package tests

import (
	"bytes"
	query "github.com/2fas/api/internal/api/support/app/queries"
	"github.com/2fas/api/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func TestDebugLogsAuditTestSuite(t *testing.T) {
	suite.Run(t, new(DebugLogsAuditTestSuite))
}

type DebugLogsAuditTestSuite struct {
	suite.Suite
}

func (s *DebugLogsAuditTestSuite) SetupTest() {
	tests.DoSuccessDelete(s.T(), "mobile/support/debug_logs/audit")
}

func (s *DebugLogsAuditTestSuite) TestCreateDebugLogsAuditClaim() {
	payload := []byte(`{"username": "app-user", "description": "some description"}`)

	auditClaim := new(query.DebugLogsAuditPresenter)

	tests.DoSuccessPost(s.T(), "mobile/support/debug_logs/audit/claim", payload, auditClaim)

	assert.Equal(s.T(), "app-user", auditClaim.Username)
	assert.Equal(s.T(), "some description", auditClaim.Description)
}

func (s *DebugLogsAuditTestSuite) TestUpdateDebugLogsAuditClaim() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	var updatedAuditClaim *query.DebugLogsAuditPresenter
	updatePayload := []byte(`{"username": "app-user-1", "description": "another description"}`)
	tests.DoSuccessPut(s.T(), "mobile/support/debug_logs/audit/claim/"+auditClaim.Id, updatePayload, &updatedAuditClaim)

	assert.Equal(s.T(), "app-user-1", updatedAuditClaim.Username)
	assert.Equal(s.T(), "another description", updatedAuditClaim.Description)
}

func (s *DebugLogsAuditTestSuite) TestFulfillDebugLogsAuditClaim() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "logs.json")

	part.Write([]byte(`{"log":"data"}`))

	writer.Close()

	request, _ := http.NewRequest("POST", "http://localhost/mobile/support/debug_logs/audit/"+auditClaim.Id, body)
	request.Header.Add("Content-type", writer.FormDataContentType())

	response, err := http.DefaultClient.Do(request)
	require.NoError(s.T(), err)

	reqB, _ := ioutil.ReadAll(body)
	s.T().Log(string(reqB))

	rawBody, _ := ioutil.ReadAll(response.Body)

	s.T().Log(string(rawBody))
}

func (s *DebugLogsAuditTestSuite) TestTryToFulfillDebugLogsAuditClaimTwice() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "logs.json")

	part.Write([]byte(`{"log":"data"}`))

	writer.Close()

	request, _ := http.NewRequest("POST", "http://localhost/mobile/support/debug_logs/audit/"+auditClaim.Id, body)
	request.Header.Add("Content-type", writer.FormDataContentType())
	_, err := http.DefaultClient.Do(request)
	require.NoError(s.T(), err)

	secondRequest, _ := http.NewRequest("POST", "http://localhost/mobile/support/debug_logs/audit/"+auditClaim.Id, body)
	request.Header.Add("Content-type", writer.FormDataContentType())
	secondResponse, err := http.DefaultClient.Do(secondRequest)

	assert.Equal(s.T(), 410, secondResponse.StatusCode)
}

func (s *DebugLogsAuditTestSuite) TestTryToFulfillNotExistingDebugLogsAuditClaim() {
	notExistingAuditClaimId := uuid.New().String()

	request, _ := http.NewRequest("POST", "http://localhost/mobile/support/debug_logs/audit/"+notExistingAuditClaimId, nil)
	response, _ := http.DefaultClient.Do(request)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *DebugLogsAuditTestSuite) TestTryToFulfillDebugLogsAuditClaimUsingInvalidId() {
	auditClaimId := uuid.New().String()
	invalidId := strings.ToUpper(auditClaimId)

	request, _ := http.NewRequest("POST", "http://localhost/mobile/support/debug_logs/audit/"+invalidId, nil)
	response, _ := http.DefaultClient.Do(request)

	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *DebugLogsAuditTestSuite) TestGetDebugLogsAudit() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	audit := new(query.DebugLogsAuditPresenter)
	tests.DoSuccessGet(s.T(), "mobile/support/debug_logs/audit/"+auditClaim.Id, audit)

	assert.Equal(s.T(), auditClaim.Id, audit.Id)
	assert.Equal(s.T(), "user1", audit.Username)
	assert.Equal(s.T(), "desc1", audit.Description)
}

func (s *DebugLogsAuditTestSuite) TestDeleteDebugLogsAudit() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	tests.DoSuccessDelete(s.T(), "mobile/support/debug_logs/audit/"+auditClaim.Id)

	response := tests.DoGet("mobile/support/debug_logs/audit/"+auditClaim.Id, nil)
	assert.Equal(s.T(), 404, response.StatusCode)
}

func (s *DebugLogsAuditTestSuite) TestFindAllDebugLogsAudit() {
	createDebugLogsAuditClaim(s.T(), "user1", "desc1")
	createDebugLogsAuditClaim(s.T(), "user2", "desc2")

	var audits []*query.DebugLogsAuditPresenter
	tests.DoSuccessGet(s.T(), "mobile/support/debug_logs/audit", &audits)

	assert.Len(s.T(), audits, 2)
}

func createDebugLogsAuditClaim(t *testing.T, username, description string) *query.DebugLogsAuditPresenter {
	payload := []byte(`{"username": "` + username + `", "description": "` + description + `"}`)

	auditClaim := new(query.DebugLogsAuditPresenter)
	tests.DoSuccessPost(t, "mobile/support/debug_logs/audit/claim", payload, auditClaim)

	return auditClaim
}
