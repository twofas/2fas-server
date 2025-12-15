package tests

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	e2e_tests "github.com/twofas/2fas-server/e2e-tests"
)

func Test_BrowserExtensionLogging(t *testing.T) {
	browserExtension := e2e_tests.CreateBrowserExtension(t, "go-ext")

	log := &struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}{
		Level:   "info",
		Message: "test log",
	}

	payload, err := json.Marshal(log)
	require.NoError(t, err)

	e2e_tests.DoAPISuccessPost(t, "/browser_extensions/"+browserExtension.Id+"/commands/store_log", payload, nil)
}

func Test_NotExistingBrowserExtensionLogging(t *testing.T) {
	someId := uuid.New()

	log := &struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}{
		Level:   "info",
		Message: "test log",
	}

	payload, err := json.Marshal(log)
	require.NoError(t, err)

	e2e_tests.DoAPISuccessPost(t, "/browser_extensions/"+someId.String()+"/commands/store_log", payload, nil)
}
