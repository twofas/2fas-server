package tests

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/twofas/2fas-server/tests"
	"testing"
)

func Test_BrowserExtensionLogging(t *testing.T) {
	browserExtension := tests.CreateBrowserExtension(t, "go-ext")

	log := &struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}{
		Level:   "info",
		Message: "test log",
	}

	payload, _ := json.Marshal(log)
	response := tests.DoPost("/browser_extensions/"+browserExtension.Id+"/commands/store_log", payload, nil)

	assert.Equal(t, 200, response.StatusCode)
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

	payload, _ := json.Marshal(log)
	response := tests.DoPost("/browser_extensions/"+someId.String()+"/commands/store_log", payload, nil)

	assert.Equal(t, 200, response.StatusCode)
}
