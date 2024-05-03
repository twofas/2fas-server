package pairing

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/pass/connection"
	"github.com/twofas/2fas-server/internal/pass/sign"
)

type Pairing struct {
	store                               store
	signSvc                             *sign.Service
	pairingRequestTokenValidityDuration time.Duration
}

type store interface {
	AddExtension(ctx context.Context, extensionID string)
	ExtensionExists(ctx context.Context, extensionID string) bool
	GetPairingInfo(ctx context.Context, extensionID string) (PairingInfo, error)
	SetPairingInfo(ctx context.Context, extensionID string, pi PairingInfo) error
}

func NewApp(signService *sign.Service, pairingRequestTokenValidityDuration time.Duration) *Pairing {
	return &Pairing{
		store:                               NewMemoryStore(),
		signSvc:                             signService,
		pairingRequestTokenValidityDuration: pairingRequestTokenValidityDuration,
	}
}

const (
	pairingTokenValidityDuration = 3 * time.Minute
)

type ConfigureBrowserExtensionRequest struct {
	ExtensionID string `json:"extension_id"`
}

type ConfigureBrowserExtensionResponse struct {
	BrowserExtensionPairingToken string `json:"browser_extension_pairing_token"`
	ConnectionToken              string `json:"connection_token"`
}

func (p *Pairing) ConfigureBrowserExtension(ctx context.Context, req ConfigureBrowserExtensionRequest) (ConfigureBrowserExtensionResponse, error) {
	p.store.AddExtension(ctx, req.ExtensionID)

	pairingToken, err := p.signSvc.SignAndEncode(sign.Message{
		ConnectionID:   req.ExtensionID,
		ExpiresAt:      time.Now().Add(pairingTokenValidityDuration),
		ConnectionType: sign.ConnectionTypeBrowserExtensionWait,
	})
	if err != nil {
		return ConfigureBrowserExtensionResponse{}, fmt.Errorf("failed to generate pairing token: %v", err)
	}

	mobileToken, err := p.signSvc.SignAndEncode(sign.Message{
		ConnectionID:   req.ExtensionID,
		ExpiresAt:      time.Now().Add(pairingTokenValidityDuration),
		ConnectionType: sign.ConnectionTypeMobileConfirm,
	})
	if err != nil {
		return ConfigureBrowserExtensionResponse{}, fmt.Errorf("Failed to generate mobile confirm token: %v", err)
	}
	return ConfigureBrowserExtensionResponse{
		ConnectionToken:              mobileToken,
		BrowserExtensionPairingToken: pairingToken,
	}, nil
}

type ExtensionWaitForConnectionInput struct {
	ResponseWriter http.ResponseWriter
	HttpReq        *http.Request
}

type WaitForConnectionResponse struct {
	BrowserExtensionProxyToken string `json:"browser_extension_proxy_token"`
	BrowserExtensionSyncToken  string `json:"browser_extension_sync_token"`
	Status                     string `json:"status"`
	DeviceID                   string `json:"device_id"`
}

func (p *Pairing) ServePairingWS(w http.ResponseWriter, r *http.Request, extID string) error {
	log := logging.WithField("extension_id", extID)
	conn, err := connection.Upgrade(w, r)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %w", err)
	}

	log.Info("Starting pairing WS")

	if pairing, pairingDone := p.isExtensionPaired(r.Context(), extID, log); pairingDone {
		if err := p.sendTokenAndCloseConn(extID, pairing, conn); err != nil {
			log.Errorf("Failed to send token: %v", err)
		}
		return nil
	}

	const (
		maxWaitTime              = 3 * time.Minute
		checkIfConnectedInterval = time.Second
	)
	maxWaitC := time.After(maxWaitTime)
	// TODO: consider returning event from store on change.
	connectedCheckTicker := time.NewTicker(checkIfConnectedInterval)
	defer connectedCheckTicker.Stop()
	for {
		select {
		case <-maxWaitC:
			log.Info("Closing paring ws after timeout")
			return nil
		case <-connectedCheckTicker.C:
			if pairing, pairingDone := p.isExtensionPaired(r.Context(), extID, log); pairingDone {
				if err := p.sendTokenAndCloseConn(extID, pairing, conn); err != nil {
					log.Errorf("Failed to send token: %v", err)
					return nil
				}
				log.WithField("device_id", pairing.Device.DeviceID).Infof("Paring ws finished")
				return nil
			}
		}
	}
}

func (p *Pairing) isExtensionPaired(ctx context.Context, extID string, log logging.FieldLogger) (PairingInfo, bool) {
	pairingInfo, err := p.store.GetPairingInfo(ctx, extID)
	if err != nil {
		log.Warn("Failed to get pairing info")
		return PairingInfo{}, false
	}
	return pairingInfo, pairingInfo.IsPaired()
}

func (p *Pairing) sendTokenAndCloseConn(extID string, pairingInfo PairingInfo, conn *websocket.Conn) error {
	extProxyToken, err := p.signSvc.SignAndEncode(sign.Message{
		ConnectionID:   extID,
		ExpiresAt:      time.Now().Add(pairingTokenValidityDuration),
		ConnectionType: sign.ConnectionTypeBrowserExtensionProxy,
	})
	if err != nil {
		return fmt.Errorf("failed to generate ext proxy token: %v", err)
	}
	var syncToken string
	if pairingInfo.Device.FCMToken != "" {
		syncToken, err = p.signSvc.SignAndEncode(sign.Message{
			ConnectionID:   pairingInfo.Device.FCMToken,
			ExpiresAt:      time.Now().Add(p.pairingRequestTokenValidityDuration),
			ConnectionType: sign.ConnectionTypeBrowserExtensionSyncRequest,
		})
		if err != nil {
			return fmt.Errorf("failed to generate proxy sync request token: %v", err)
		}

	}

	if err := conn.WriteJSON(WaitForConnectionResponse{
		BrowserExtensionProxyToken: extProxyToken,
		BrowserExtensionSyncToken:  syncToken,
		Status:                     "ok",
		DeviceID:                   pairingInfo.Device.DeviceID,
	}); err != nil {
		return fmt.Errorf("failed to write to extension: %v", err)
	}
	return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

// GetPairingInfo returns paired device and information if pairing was done.
func (p *Pairing) GetPairingInfo(ctx context.Context, extensionID string) (PairingInfo, error) {
	return p.store.GetPairingInfo(ctx, extensionID)
}

type ConfirmPairingRequest struct {
	FCMToken string `json:"fcm_token"`
	DeviceID string `json:"device_id"`
}

type ConfirmPairingResponse struct {
	ProxyToken string `json:"proxy_token"`
}

func (p *Pairing) ConfirmPairing(ctx context.Context, req ConfirmPairingRequest, extensionID string) (ConfirmPairingResponse, error) {
	mobileProxyToken, err := p.signSvc.SignAndEncode(sign.Message{
		ConnectionID:   extensionID,
		ExpiresAt:      time.Now().Add(pairingTokenValidityDuration),
		ConnectionType: sign.ConnectionTypeMobileProxy,
	})
	if err != nil {
		return ConfirmPairingResponse{}, fmt.Errorf("Failed to generate ext proxy token: %v", err)
	}
	if err := p.store.SetPairingInfo(ctx, extensionID, PairingInfo{
		Device: MobileDevice{
			DeviceID: req.DeviceID,
			FCMToken: req.FCMToken,
		},
		PairedAt: time.Now().UTC(),
	}); err != nil {
		return ConfirmPairingResponse{}, err
	}

	return ConfirmPairingResponse{ProxyToken: mobileProxyToken}, nil
}
