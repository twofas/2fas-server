package pairing

import (
	"context"
	"errors"
	"fmt"

	"github.com/twofas/2fas-server/internal/pass/sign"
)

// VerifyPairingToken verifies pairing token and returns extension_id
func (p *Pairing) VerifyPairingToken(ctx context.Context, pairingToken string) (string, error) {
	extensionID, err := p.signSvc.CanI(pairingToken, sign.ConnectionTypeBrowserExtensionWait)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	ok := p.store.ExtensionExists(ctx, extensionID)
	if !ok {
		return "", errors.New("extension is not configured")
	}
	return extensionID, nil
}

// VerifyExtProxyToken verifies proxy token and returns extension_id
func (p *Pairing) VerifyExtProxyToken(ctx context.Context, proxyToken string) (string, error) {
	extensionID, err := p.signSvc.CanI(proxyToken, sign.ConnectionTypeBrowserExtensionProxy)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	ok := p.store.ExtensionExists(ctx, extensionID)
	if !ok {
		return "", errors.New("extension is not configured")
	}
	return extensionID, nil
}

// VerifyMobileProxyToken verifies mobile token and returns extension_id
func (p *Pairing) VerifyMobileProxyToken(ctx context.Context, proxyToken string) (string, error) {
	extensionID, err := p.signSvc.CanI(proxyToken, sign.ConnectionTypeMobileProxy)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	ok := p.store.ExtensionExists(ctx, extensionID)
	if !ok {
		return "", errors.New("extension is not configured")
	}
	return extensionID, nil
}

// VerifyConnectionToken verifies connection token and returns extension_id
func (p *Pairing) VerifyConnectionToken(ctx context.Context, connectionToken string) (string, error) {
	extensionID, err := p.signSvc.CanI(connectionToken, sign.ConnectionTypeMobileConfirm)
	if err != nil {
		return "", fmt.Errorf("failed to check token signature: %w", err)
	}
	ok := p.store.ExtensionExists(ctx, extensionID)
	if !ok {
		return "", errors.New("extension is not configured")
	}
	return extensionID, nil
}
