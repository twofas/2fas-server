package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/twofas/2fas-server/tests"
)

func Test_MobileApiBandwidthAbuse(t *testing.T) {
	someId := uuid.New()

	for i := 0; i <= 100; i++ {
		tests.DoAPIGet(t, "/mobile/devices/"+someId.String()+"/browser_extensions", nil)
	}
}

func Test_BrowserExtensionApiBandwidthAbuse(t *testing.T) {
	someId := uuid.New()

	for i := 0; i <= 100; i++ {
		tests.DoAPIGet(t, "/browser_extensions/"+someId.String(), nil)
	}
}
