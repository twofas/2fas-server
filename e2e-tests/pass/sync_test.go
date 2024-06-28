package pass

import (
	"testing"

	"github.com/google/uuid"
)

func TestSyncHappyFlow(t *testing.T) {
	resp, err := configureBrowserExtension()
	if err != nil {
		t.Fatalf("Failed to configure browser extension: %v", err)
	}

	browserExtensionDone := make(chan struct{})
	mobileParingDone := make(chan struct{})
	confirmMobileChannel := make(chan string)

	fcm := uuid.NewString()
	deviceID := getDeviceID()

	go func() {
		defer close(browserExtensionDone)
		_, syncToken, err := browserExtensionWaitForConfirm(resp.BrowserExtensionPairingToken)
		if err != nil {
			t.Errorf("Error when Browser Extension waited for pairing confirm: %v", err)
			return
		}

		requestSyncResp, err := browserExtensionRequestSync(syncToken)
		if err != nil {
			t.Errorf("Error when Browser Extension requested sync confirm: %v", err)
			return
		}

		pushResp, err := browserExtensionPush(requestSyncResp.BrowserExtensionWaitToken, map[string]string{"hello": "world!"})
		if err != nil {
			t.Errorf("Error when Browser Extension tried to send push notification: %v", err)
			return
		}
		t.Logf("Push response: %v", pushResp)

		confirmMobileChannel <- requestSyncResp.MobileConfirmToken

		proxyToken, err := browserExtensionWaitForSyncConfirm(requestSyncResp.BrowserExtensionWaitToken)
		if err != nil {
			t.Errorf("Error when Browser Extension waited for sync confirm: %v", err)
			return
		}

		err = proxyWebSocket(
			getWsURL()+"/browser_extension/sync/proxy",
			proxyToken,
			"sent from browser extension",
			"sent from mobile")
		if err != nil {
			t.Errorf("Browser Extension: proxy failed: %v", err)
			return
		}
	}()
	go func() {
		defer close(mobileParingDone)

		_, err := confirmMobile(resp.ConnectionToken, deviceID, fcm)
		if err != nil {
			t.Errorf("Mobile: confirm failed: %v", err)
			return
		}

		confirmToken := <-confirmMobileChannel

		proxyToken, err := confirmSyncMobile(confirmToken)
		if err != nil {
			t.Errorf("Failed to confirm mobile: %v", err)
			return
		}
		if proxyToken == "" {
			t.Errorf("Mobile: proxy token is empty")
			return
		}

		err = proxyWebSocket(
			getWsURL()+"/mobile/sync/proxy",
			proxyToken,
			"sent from mobile",
			"sent from browser extension",
		)
		if err != nil {
			t.Errorf("Mobile: proxy failed: %v", err)
			return
		}
	}()

	<-browserExtensionDone
	<-mobileParingDone
}
