package pass

import (
	"testing"

	"github.com/google/uuid"
)

func msgOfSize(size int, c byte) string {
	msg := make([]byte, size)

	for i := range msg {
		msg[i] = c
	}

	return string(msg)
}

func TestPairHappyFlow(t *testing.T) {
	resp, err := configureBrowserExtension()
	if err != nil {
		t.Fatalf("Failed to configure browser extension: %v", err)
	}
	testPairing(t, resp)
}

func TestPairMultipleTimes(t *testing.T) {
	resp, err := configureBrowserExtension()
	if err != nil {
		t.Fatalf("Failed to configure browser extension: %v", err)
	}
	const messageSize = 1024 * 1024

	for i := 0; i < 10; i++ {
		testPairing(t, resp)
	}
}

func testPairing(t *testing.T, resp ConfigureBrowserExtensionResponse) {
	t.Helper()

	browserExtensionDone := make(chan struct{})
	mobileDone := make(chan struct{})

	const messageSize = 1024 * 1024

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
			msgOfSize(messageSize, 'b'),
			msgOfSize(messageSize, 'm'),
		)
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
			msgOfSize(messageSize, 'm'),
			msgOfSize(messageSize, 'b'),
		)
		if err != nil {
			t.Errorf("Mobile: proxy failed: %v", err)
			return
		}
	}()
	<-browserExtensionDone
	<-mobileDone
}
