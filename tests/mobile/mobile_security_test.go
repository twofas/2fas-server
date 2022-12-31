package tests

import (
	"github.com/2fas/api/tests"
	"github.com/google/uuid"
	"testing"
)

func Test_MobileApiBandwidthAbuse(t *testing.T) {
	someId := uuid.New()

	for i := 0; i <= 100; i++ {
		tests.DoGet("/mobile/devices/"+someId.String()+"/browser_extensions", nil)
	}
}

func Test_BrowserExtensionApiBandwidthAbuse(t *testing.T) {
	someId := uuid.New()

	for i := 0; i <= 100; i++ {
		tests.DoGet("/browser_extensions/"+someId.String(), nil)
	}
}
