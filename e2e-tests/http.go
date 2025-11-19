package e2e_tests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	DebugHttpRequests = false
	adminRawURL       = "http://localhost:8082/admin"
	apiRawURL         = "http://localhost:80"
)

var Auth *BasicAuth

type BasicAuth struct {
	Username string
	Password string
}

func (a *BasicAuth) Header() string {
	base := a.Username + ":" + a.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}

func DoAPISuccessPost(t *testing.T, uri string, payload []byte, resp interface{}) {
	response := doRequest(t, apiRawURL, uri, http.MethodPost, payload, resp)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func DoAdminAPISuccessPost(t *testing.T, uri string, payload []byte, resp interface{}) {
	response := doRequest(t, adminRawURL, uri, http.MethodPost, payload, resp)
	bb, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, response.StatusCode, "invalid status code, response payload is: %q", string(bb))
}

func DoAdminPostAndAssertCode(t *testing.T, expCode int, uri string, payload []byte, resp interface{}) {
	response := doRequest(t, adminRawURL, uri, http.MethodPost, payload, resp)
	require.Equal(t, expCode, response.StatusCode)
}

func DoAPIPostAndAssertCode(t *testing.T, expCode int, uri string, payload []byte, resp interface{}) {
	response := doRequest(t, apiRawURL, uri, http.MethodPost, payload, resp)
	require.Equal(t, expCode, response.StatusCode)
}

func DoAPIRequest(t *testing.T, uri, method string, payload []byte, resp interface{}) *http.Response {
	return doRequest(t, apiRawURL, uri, method, payload, resp)
}

func DoAdminRequest(t *testing.T, uri, method string, payload []byte, resp interface{}) *http.Response {
	return doRequest(t, apiRawURL, uri, method, payload, resp)
}

func DoAdminSuccessPut(t *testing.T, uri string, payload []byte, resp interface{}) {
	response := doRequest(t, adminRawURL, uri, http.MethodPut, payload, resp)
	bb, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, response.StatusCode, fmt.Sprintf("invalid status code, response payload is: %q", string(bb)))
}

func DoAPISuccessPut(t *testing.T, uri string, payload []byte, resp interface{}) {
	response := doRequest(t, apiRawURL, uri, http.MethodPut, payload, resp)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func DoAPISuccessGet(t *testing.T, uri string, resp interface{}) {
	response := doRequest(t, apiRawURL, uri, http.MethodGet, nil /*payload*/, resp)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func DoAPIGet(t *testing.T, uri string, resp interface{}) *http.Response {
	return doRequest(t, apiRawURL, uri, http.MethodGet, nil /*payload*/, resp)
}

func DoAdminSuccessGet(t *testing.T, uri string, resp interface{}) {
	response := doRequest(t, adminRawURL, uri, http.MethodGet, nil /*payload*/, resp)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func DoAdminSuccessDelete(t *testing.T, uri string) {
	response := doRequest(t, adminRawURL, uri, http.MethodDelete, nil /*payload*/, nil /*response*/)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func DoAPISuccessDelete(t *testing.T, uri string) {
	response := doRequest(t, apiRawURL, uri, http.MethodDelete, nil /*payload*/, nil /*response*/)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func doRequest(t *testing.T, base, uri, method string, payload []byte, resp interface{}) *http.Response {
	t.Helper()
	baseURL, err := url.Parse(base)
	require.NoError(t, err)

	request := createRequest(method, baseURL.JoinPath(uri).String(), payload)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)

	logRequest(request, response)

	rawBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	if resp != nil {
		responseDataReader := bytes.NewReader(rawBody)
		err = json.NewDecoder(responseDataReader).Decode(resp)
		require.NoError(t, err)
	}

	response.Body.Close()
	response.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	return response
}

func createRequest(method, uri string, payload []byte) *http.Request {
	request, _ := http.NewRequest(method, uri, bytes.NewBuffer(payload))

	request.Header.Add("Content-Type", "application/json")

	if Auth != nil {
		request.Header.Add("Authorization", Auth.Header())
	}

	return request
}

func logRequest(req *http.Request, resp *http.Response) {
	if DebugHttpRequests {
		rawBody, _ := io.ReadAll(resp.Body)

		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(rawBody))

		fmt.Printf("Request: %s: %s %v \n", req.Method, req.URL.RequestURI(), req.Body)
		fmt.Println("Response: ", req.URL.RequestURI(), resp.StatusCode, string(rawBody))
	}
}
