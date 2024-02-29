package sync

import (
	"context"
	"fmt"

	"github.com/twofas/2fas-server/internal/pass/sign"
)

// VerifyExtRequestSyncToken verifies sync request token and returns fcm_token.
func (p *Syncing) VerifyExtRequestSyncToken(ctx context.Context, proxyToken string) (string, error) {
	fcmToken, err := p.signSvc.CanI(proxyToken, sign.ConnectionTypeBrowserExtensionSyncRequest)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	return fcmToken, nil
}

// VerifyExtSyncToken verifies sync token and returns fcm_token.
func (p *Syncing) VerifyExtSyncToken(ctx context.Context, proxyToken string) (string, error) {
	fcmToken, err := p.signSvc.CanI(proxyToken, sign.ConnectionTypeBrowserExtensionSync)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	return fcmToken, nil
}

// VerifyMobileSyncConfirmToken verifies mobile token and returns connection id.
func (p *Syncing) VerifyMobileSyncConfirmToken(ctx context.Context, proxyToken string) (string, error) {
	extensionID, err := p.signSvc.CanI(proxyToken, sign.ConnectionTypeMobileSyncConfirm)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	return extensionID, nil
}

// VerifyMobileSyncProxyToken verifies mobile token and returns connection id.
func (p *Syncing) VerifyMobileSyncProxyToken(ctx context.Context, proxyToken string) (string, error) {
	extensionID, err := p.signSvc.CanI(proxyToken, sign.ConnectionTypeMobileSyncProxy)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	return extensionID, nil
}
