package sync

import (
	"context"
	"fmt"

	"github.com/twofas/2fas-server/internal/pass/sign"
)

// VerifyExtRequestSyncToken verifies sync request token and returns fcm_token.
func (s *Syncing) VerifyExtRequestSyncToken(ctx context.Context, proxyToken string) (string, error) {
	fcmToken, err := s.signSvc.CanI(proxyToken, sign.ConnectionTypeBrowserExtensionSyncRequest)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	return fcmToken, nil
}

// VerifyExtWaitForSyncToken verifies wait for sync request token and returns fcm_token.
func (s *Syncing) VerifyExtWaitForSyncToken(ctx context.Context, proxyToken string) (string, error) {
	fcmToken, err := s.signSvc.CanI(proxyToken, sign.ConnectionTypeBrowserExtensionSyncWait)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	return fcmToken, nil
}

// VerifyExtSyncToken verifies sync token and returns fcm_token.
func (s *Syncing) VerifyExtSyncToken(ctx context.Context, proxyToken string) (string, error) {
	fcmToken, err := s.signSvc.CanI(proxyToken, sign.ConnectionTypeBrowserExtensionSync)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	return fcmToken, nil
}

// VerifyMobileSyncConfirmToken verifies mobile token and returns connection id.
func (s *Syncing) VerifyMobileSyncConfirmToken(ctx context.Context, proxyToken string) (string, error) {
	extensionID, err := s.signSvc.CanI(proxyToken, sign.ConnectionTypeMobileSyncConfirm)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	return extensionID, nil
}

// VerifyMobileSyncProxyToken verifies mobile token and returns connection id.
func (s *Syncing) VerifyMobileSyncProxyToken(ctx context.Context, proxyToken string) (string, error) {
	extensionID, err := s.signSvc.CanI(proxyToken, sign.ConnectionTypeMobileSyncProxy)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	return extensionID, nil
}
