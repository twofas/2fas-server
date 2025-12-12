package tests

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
	query "github.com/twofas/2fas-server/internal/api/support/app/queries"
)

func TestDebugLogsAuditTestSuite(t *testing.T) {
	suite.Run(t, new(DebugLogsAuditTestSuite))
}

type DebugLogsAuditTestSuite struct {
	suite.Suite
}

func (s *DebugLogsAuditTestSuite) SetupTest() {
	e2e_tests.DoAdminSuccessDelete(s.T(), "mobile/support/debug_logs/audit")
}

func (s *DebugLogsAuditTestSuite) TestCreateDebugLogsAuditClaim() {
	payload := []byte(`{"username": "app-user", "description": "some description"}`)

	auditClaim := new(query.DebugLogsAuditPresenter)

	e2e_tests.DoAdminAPISuccessPost(s.T(), "mobile/support/debug_logs/audit/claim", payload, auditClaim)

	s.Equal("app-user", auditClaim.Username)
	s.Equal("some description", auditClaim.Description)
}

func (s *DebugLogsAuditTestSuite) TestUpdateDebugLogsAuditClaim() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	var updatedAuditClaim *query.DebugLogsAuditPresenter
	updatePayload := []byte(`{"username": "app-user-1", "description": "another description"}`)
	e2e_tests.DoAdminSuccessPut(s.T(), "mobile/support/debug_logs/audit/claim/"+auditClaim.Id, updatePayload, &updatedAuditClaim)

	s.Equal("app-user-1", updatedAuditClaim.Username)
	s.Equal("another description", updatedAuditClaim.Description)
}

func (s *DebugLogsAuditTestSuite) TestFulfillDebugLogsAuditClaim() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "logs.json")

	_, _ = part.Write([]byte(`{"log":"data"}`))

	writer.Close()

	request, _ := http.NewRequest(http.MethodPost, "http://localhost/mobile/support/debug_logs/audit/"+auditClaim.Id, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	response, err := http.DefaultClient.Do(request)
	s.Require().NoError(err)
	s.Equal(200, response.StatusCode)

	reqB, _ := ioutil.ReadAll(body)
	s.T().Log(string(reqB))

	rawBody, _ := ioutil.ReadAll(response.Body)

	s.T().Log(string(rawBody))
}

func mkFormFileBody() (*bytes.Buffer, string, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "logs.json")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = part.Write([]byte(`{"log":"data"}`))
	if err != nil {
		return nil, "", fmt.Errorf("failed to write to form file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close writer: %w", err)
	}
	return body, writer.FormDataContentType(), nil
}

func (s *DebugLogsAuditTestSuite) TestTryToFulfillDebugLogsAuditClaimTwice() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	body, formDataContentType, err := mkFormFileBody()
	s.Require().NoError(err)

	request, _ := http.NewRequest(http.MethodPost, "http://localhost/mobile/support/debug_logs/audit/"+auditClaim.Id, body)
	request.Header.Add("Content-Type", formDataContentType)
	response, err := http.DefaultClient.Do(request)
	s.Require().NoError(err)
	s.Equal(200, response.StatusCode)

	body, formDataContentType, err = mkFormFileBody()
	s.Require().NoError(err)
	secondRequest, _ := http.NewRequest(http.MethodPost, "http://localhost/mobile/support/debug_logs/audit/"+auditClaim.Id, body)
	secondRequest.Header.Add("Content-Type", formDataContentType)
	secondResponse, err := http.DefaultClient.Do(secondRequest)
	s.Require().NoError(err)

	responseBody, err := io.ReadAll(secondResponse.Body)
	s.Require().NoError(err)

	s.Equal(410, secondResponse.StatusCode, "Response body: %s", string(responseBody))
}

func (s *DebugLogsAuditTestSuite) TestTryToFulfillNotExistingDebugLogsAuditClaim() {
	notExistingAuditClaimId := uuid.New().String()

	body, formDataContentType, err := mkFormFileBody()
	s.Require().NoError(err)

	request, _ := http.NewRequest(http.MethodPost, "http://localhost/mobile/support/debug_logs/audit/"+notExistingAuditClaimId, body)
	request.Header.Add("Content-Type", formDataContentType)
	response, err := http.DefaultClient.Do(request)
	s.Require().NoError(err)

	responseBody, err := io.ReadAll(response.Body)
	s.Require().NoError(err)

	s.Equal(404, response.StatusCode, "Response body: %s", string(responseBody))
}

func (s *DebugLogsAuditTestSuite) TestTryToFulfillDebugLogsAuditClaimUsingInvalidId() {
	auditClaimId := uuid.New().String()
	invalidId := strings.ToUpper(auditClaimId)

	body, formDataContentType, err := mkFormFileBody()
	s.Require().NoError(err)

	request, _ := http.NewRequest(http.MethodPost, "http://localhost/mobile/support/debug_logs/audit/"+invalidId, body)
	request.Header.Add("Content-Type", formDataContentType)
	response, err := http.DefaultClient.Do(request)
	s.Require().NoError(err)

	s.Equal(404, response.StatusCode)
}

func (s *DebugLogsAuditTestSuite) TestGetDebugLogsAudit() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	audit := new(query.DebugLogsAuditPresenter)
	e2e_tests.DoAdminSuccessGet(s.T(), "mobile/support/debug_logs/audit/"+auditClaim.Id, audit)

	s.Equal(auditClaim.Id, audit.Id)
	s.Equal("user1", audit.Username)
	s.Equal("desc1", audit.Description)
}

func (s *DebugLogsAuditTestSuite) TestDeleteDebugLogsAudit() {
	auditClaim := createDebugLogsAuditClaim(s.T(), "user1", "desc1")

	e2e_tests.DoAdminSuccessDelete(s.T(), "mobile/support/debug_logs/audit/"+auditClaim.Id)

	response := e2e_tests.DoAPIGet(s.T(), "mobile/support/debug_logs/audit/"+auditClaim.Id, nil)
	s.Equal(404, response.StatusCode)
}

func (s *DebugLogsAuditTestSuite) TestFindAllDebugLogsAudit() {
	createDebugLogsAuditClaim(s.T(), "user1", "desc1")
	createDebugLogsAuditClaim(s.T(), "user2", "desc2")

	var audits []*query.DebugLogsAuditPresenter
	e2e_tests.DoAdminSuccessGet(s.T(), "mobile/support/debug_logs/audit", &audits)

	s.Len(audits, 2)
}

func createDebugLogsAuditClaim(t *testing.T, username, description string) *query.DebugLogsAuditPresenter {
	t.Helper()
	payload := []byte(`{"username": "` + username + `", "description": "` + description + `"}`)

	auditClaim := new(query.DebugLogsAuditPresenter)
	e2e_tests.DoAdminAPISuccessPost(t, "mobile/support/debug_logs/audit/claim", payload, auditClaim)

	return auditClaim
}
