package pass

import (
	"testing"

	"github.com/google/uuid"
)

func TestPairHappyFlow(t *testing.T) {
	resp, err := configureBrowserExtension()
	if err != nil {
		t.Fatalf("Failed to configure browser extension: %v", err)
	}

	browserExtensionDone := make(chan struct{})
	mobileDone := make(chan struct{})

	go func() {
		defer close(browserExtensionDone)

		extProxyToken, _, err := browserExtensionWaitForConfirm(resp.BrowserExtensionPairingToken)
		if err != nil {
			t.Errorf("Error when Browser Extension waited for confirm: %v", err)
			return
		}

		err = proxyWebSocket(
			getWsURL()+"/browser_extension/proxy_to_mobile",
			extProxyToken,
			"sent from browser extension",
			"sent from mobile")
		if err != nil {
			t.Errorf("Browser Extension: proxy failed: %v", err)
			return
		}

	}()
	go func() {
		defer close(mobileDone)

		mobileProxyToken, err := confirmMobile(resp.ConnectionToken, uuid.NewString())
		if err != nil {
			t.Errorf("Mobile: confirm failed: %v", err)
			return
		}

		err = proxyWebSocket(
			getWsURL()+"/mobile/proxy_to_browser_extension",
			mobileProxyToken,
			"sent from mobile",
			"sent from browser extension",
		)
		if err != nil {
			t.Errorf("Mobile: proxy failed: %v", err)
			return
		}
	}()
	<-browserExtensionDone
	<-mobileDone
}
