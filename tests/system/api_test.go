package tests

import (
	"github.com/2fas/api/tests"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func Test_Default404Response(t *testing.T) {
	response := tests.DoGet("some/not/existing/endpoint", nil)

	rawBody, _ := ioutil.ReadAll(response.Body)

	expected := `{"Code":404,"Type":"NotFound","Description":"Requested resource can not be found","Reason":"URI not found"}`

	assert.Equal(t, 404, response.StatusCode)
	assert.JSONEq(t, expected, string(rawBody))
}
