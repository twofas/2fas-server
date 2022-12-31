package http

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/2fas/api/internal/common/logging"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

var tunedHttpTransport = &http.Transport{
	MaxIdleConns:        1024,
	MaxIdleConnsPerHost: 1024,
	TLSHandshakeTimeout: 10 * time.Second,
	DialContext: (&net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}).DialContext,
}

type HttpClient struct {
	client              *http.Client
	baseUrl             *url.URL
	credentialsCallback func(r *http.Request)
}

func (w *HttpClient) CredentialsProvider(credentialsCallback func(r *http.Request)) {
	w.credentialsCallback = credentialsCallback
}

func (w *HttpClient) Post(ctx context.Context, path string, result interface{}, data interface{}) error {
	req, err := w.newJsonRequest("POST", path, data)

	if err != nil {
		return err
	}

	return w.executeRequest(ctx, req, result)
}

func (w *HttpClient) newJsonRequest(method, path string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter

	logging.WithFields(logging.Fields{
		"method": method,
		"body":   body,
	}).Debug("HTTP Request")

	if body != nil {
		buf = new(bytes.Buffer)

		encoder := json.NewEncoder(buf)
		err := encoder.Encode(body)

		if err != nil {
			return nil, err
		}
	}

	return w.newRequest(method, path, buf, "application/json")
}

func (w *HttpClient) newRequest(method, path string, buf io.Reader, contentType string) (*http.Request, error) {
	u, err := w.baseUrl.Parse(path)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), buf)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return req, nil
}

func (w *HttpClient) executeRequest(ctx context.Context, req *http.Request, v interface{}) error {
	req = req.WithContext(ctx)

	if w.credentialsCallback != nil {
		w.credentialsCallback(req)
	}

	resp, err := w.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	responseData, err := w.checkError(resp)

	if err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	responseDataReader := bytes.NewReader(responseData)

	err = json.NewDecoder(responseDataReader).Decode(v)

	return err
}

func (w *HttpClient) checkError(r *http.Response) ([]byte, error) {
	errorResponse := &ErrorResponse{}

	responseData, err := ioutil.ReadAll(r.Body)

	if err == nil && responseData != nil {
		json.Unmarshal(responseData, errorResponse)
	}

	if httpCode := r.StatusCode; 200 <= httpCode && httpCode <= 300 {
		return responseData, nil
	}

	errorResponse.Status = r.StatusCode

	return responseData, errorResponse
}

func NewHttpClient(baseUrl string) *HttpClient {
	clientBaseUrl, err := url.Parse(baseUrl)

	if err != nil {
		panic(err)
	}

	return &HttpClient{
		client:  &http.Client{Transport: tunedHttpTransport},
		baseUrl: clientBaseUrl,
	}
}
