package pass

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func msgOfSize(size int, c byte) string {
	msg := make([]byte, size)

	for i := range msg {
		msg[i] = c
	}

	return string(msg)
}

func TestDelayedCommunication(t *testing.T) {
	resp, err := configureBrowserExtension()
	if err != nil {
		t.Fatalf("Failed to configure browser extension: %v", err)
	}

	t.Run("BE sleeps before sending a message", func(t *testing.T) {
		deviceID := getDeviceID()
		testPairing(t, deviceID, resp, time.Minute, 0)
	})
	t.Run("Mobile sleeps before sending a message", func(t *testing.T) {
		deviceID := getDeviceID()
		testPairing(t, deviceID, resp, 0, time.Minute)
	})
	t.Run("Both sleep before sending a message", func(t *testing.T) {
		deviceID := getDeviceID()
		testPairing(t, deviceID, resp, time.Minute, time.Minute)
	})
}

func TestPairHappyFlow(t *testing.T) {
	resp, err := configureBrowserExtension()
	if err != nil {
		t.Fatalf("Failed to configure browser extension: %v", err)
	}

	deviceID := getDeviceID()
	testPairing(t, deviceID, resp, 0, 0)
}

func TestPairMultipleTimes(t *testing.T) {
	resp, err := configureBrowserExtension()
	if err != nil {
		t.Fatalf("Failed to configure browser extension: %v", err)
	}

	deviceID := getDeviceID()
	for i := 0; i < 10; i++ {
		testPairing(t, deviceID, resp, 0, 0)
		if t.Failed() {
			break
		}
	}
}

func testPairing(t *testing.T, deviceID string, resp ConfigureBrowserExtensionResponse, sleepBeforeSendBE, sleepBeforeSendMobile time.Duration) {
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
			sleepBeforeSendBE,
		)
		if err != nil {
			t.Errorf("Browser Extension: proxy failed: %v", err)
			return
		}

	}()
	go func() {
		defer close(mobileDone)

		mobileProxyToken, err := confirmMobile(resp.ConnectionToken, deviceID, uuid.NewString())
		if err != nil {
			t.Errorf("Mobile: confirm failed: %v", err)
			return
		}

		err = proxyWebSocket(
			getWsURL()+"/mobile/proxy_to_browser_extension",
			mobileProxyToken,
			msgOfSize(messageSize, 'm'),
			msgOfSize(messageSize, 'b'),
			sleepBeforeSendMobile,
		)
		if err != nil {
			t.Errorf("Mobile: proxy failed: %v", err)
			return
		}
	}()
	<-browserExtensionDone
	<-mobileDone
}
