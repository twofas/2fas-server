package tests

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/twofas/2fas-server/tests"
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
	tests.DoAPISuccessPost(t, "/browser_extensions/"+browserExtension.Id+"/commands/store_log", payload, nil)
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
	tests.DoAPISuccessPost(t, "/browser_extensions/"+someId.String()+"/commands/store_log", payload, nil)
}
