package tests

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
)

func Test_Default404Response(t *testing.T) {
	response := e2e_tests.DoAPIGet(t, "some/not/existing/endpoint", nil)

	rawBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	expected := `{"Code":404,"Type":"NotFound","Description":"Requested resource can not be found","Reason":"URI not found"}`

	assert.Equal(t, 404, response.StatusCode)
	assert.JSONEq(t, expected, string(rawBody))
}
