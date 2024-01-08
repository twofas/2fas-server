package pairing

import (
	"context"
	"errors"
)

// VerifyPairingToken verifies pairing token and returns extension_id
func (p *Pairing) VerifyPairingToken(ctx context.Context, pairingToken string) (string, error) {
	// TODO verify pairing token and take extension from token, this is for debug only.
	extensionID := pairingToken
	ok := p.store.ExtensionExists(ctx, extensionID)
	if !ok {
		return "", errors.New("extension is not configured")
	}
	return extensionID, nil
}

// VerifyProxyToken verifies proxy token and returns extension_id
func (p *Pairing) VerifyProxyToken(ctx context.Context, proxyToken string) (string, error) {
	// TODO verify proxy token and take extension from token, this is for debug only.
	extensionID := proxyToken
	ok := p.store.ExtensionExists(ctx, extensionID)
	if !ok {
		return "", errors.New("extension is not configured")
	}
	return extensionID, nil
}

// VerifyConnectionToken verifies connection token and returns extension_id
func (p *Pairing) VerifyConnectionToken(ctx context.Context, connectionToken string) (string, error) {
	// TODO verify proxy token and take extension from token, this is for debug only.
	extensionID := connectionToken
	ok := p.store.ExtensionExists(ctx, extensionID)
	if !ok {
		return "", errors.New("extension is not configured")
	}
	return extensionID, nil
}
