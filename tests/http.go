package tests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

const DebugHttpRequests = false

var baseUrl *url.URL
var Auth *BasicAuth

type BasicAuth struct {
	Username string
	Password string
}

func (a *BasicAuth) Header() string {
	base := a.Username + ":" + a.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}

func init() {
	baseUrl, _ = url.Parse("http://localhost")
}

func DoSuccessPost(t *testing.T, uri string, payload []byte, resp interface{}) {
	u, _ := baseUrl.Parse(uri)

	request := createRequest(http.MethodPost, u.String(), payload)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)

	logRequest(request, response)

	rawBody, _ := ioutil.ReadAll(response.Body)

	if resp != nil {
		responseDataReader := bytes.NewReader(rawBody)
		err = json.NewDecoder(responseDataReader).Decode(resp)
	}

	require.Equal(t, 200, response.StatusCode)
}

func DoPost(uri string, payload []byte, resp interface{}) *http.Response {
	u, _ := baseUrl.Parse(uri)

	request := createRequest(http.MethodPost, u.String(), payload)
	response, _ := http.DefaultClient.Do(request)

	logRequest(request, response)

	rawBody, _ := ioutil.ReadAll(response.Body)

	if resp != nil {
		responseDataReader := bytes.NewReader(rawBody)
		_ = json.NewDecoder(responseDataReader).Decode(resp)
	}

	response.Body.Close()
	response.Body = ioutil.NopCloser(bytes.NewBuffer(rawBody))

	return response
}

func DoSuccessPut(t *testing.T, uri string, payload []byte, resp interface{}) {
	u, _ := baseUrl.Parse(uri)

	request := createRequest(http.MethodPut, u.String(), payload)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)

	logRequest(request, response)

	rawBody, _ := ioutil.ReadAll(response.Body)

	if resp != nil {
		responseDataReader := bytes.NewReader(rawBody)
		err = json.NewDecoder(responseDataReader).Decode(resp)
	}

	assert.Equal(t, 200, response.StatusCode)
}

func DoPut(uri string, payload []byte, resp interface{}) *http.Response {
	u, _ := baseUrl.Parse(uri)

	request := createRequest(http.MethodPut, u.String(), payload)
	response, _ := http.DefaultClient.Do(request)

	logRequest(request, response)

	rawBody, _ := ioutil.ReadAll(response.Body)

	if resp != nil {
		responseDataReader := bytes.NewReader(rawBody)
		json.NewDecoder(responseDataReader).Decode(resp)
	}

	return response
}

func DoSuccessGet(t *testing.T, uri string, resp interface{}) {
	u, _ := baseUrl.Parse(uri)

	request := createRequest(http.MethodGet, u.String(), nil)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)

	rawBody, _ := ioutil.ReadAll(response.Body)

	logRequest(request, response)

	require.Equal(t, 200, response.StatusCode)

	err = json.Unmarshal(rawBody, resp)

	require.NoError(t, err)
}

func DoGet(uri string, resp interface{}) *http.Response {
	u, _ := baseUrl.Parse(uri)

	request := createRequest(http.MethodGet, u.String(), nil)
	response, _ := http.DefaultClient.Do(request)

	rawBody, _ := ioutil.ReadAll(response.Body)

	response.Body.Close()
	response.Body = ioutil.NopCloser(bytes.NewBuffer(rawBody))

	logRequest(request, response)

	json.Unmarshal(rawBody, resp)

	return response
}

func DoSuccessDelete(t *testing.T, uri string) *http.Response {
	u, _ := baseUrl.Parse(uri)

	request := createRequest(http.MethodDelete, u.String(), nil)
	response, err := http.DefaultClient.Do(request)

	logRequest(request, response)

	require.NoError(t, err)
	require.Equal(t, 200, response.StatusCode)

	return response
}

func DoDelete(uri string) *http.Response {
	u, _ := baseUrl.Parse(uri)

	request := createRequest(http.MethodDelete, u.String(), nil)
	response, _ := http.DefaultClient.Do(request)

	logRequest(request, response)

	return response
}

func createRequest(method, uri string, payload []byte) *http.Request {
	request, _ := http.NewRequest(method, uri, bytes.NewBuffer(payload))

	request.Header.Add("Content-type", "application/json")

	if Auth != nil {
		request.Header.Add("Authorization", Auth.Header())
	}

	return request
}

func logRequest(req *http.Request, resp *http.Response) {
	if DebugHttpRequests {
		rawBody, _ := ioutil.ReadAll(resp.Body)

		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(rawBody))

		fmt.Printf("Request: %s: %s %v \n", req.Method, req.URL.RequestURI(), req.Body)
		fmt.Println("Response: ", req.URL.RequestURI(), resp.StatusCode, string(rawBody))
	}
}
