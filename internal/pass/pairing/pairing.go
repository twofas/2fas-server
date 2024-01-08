package pairing

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type Pairing struct {
	store store
}

type store interface {
	AddExtension(ctx context.Context, extensionID string)
	ExtensionExists(ctx context.Context, extensionID string) bool
	GetPairingInfo(ctx context.Context, extensionID string) (PairingInfo, error)
	SetPairingInfo(ctx context.Context, extensionID string, pi PairingInfo) error
}

func NewPairingApp() *Pairing {
	return &Pairing{
		store: NewMemoryStore(),
	}
}

type ConfigureBrowserExtensionRequest struct {
	ExtensionID string `json:"extension_id"`
}

type ConfigureBrowserExtensionResponse struct {
	BrowserExtensionPairingToken string `json:"browser_extension_pairing_token"`
	ConnectionToken              string `json:"connection_token"`
}

func (p *Pairing) ConfigureBrowserExtension(ctx context.Context, req ConfigureBrowserExtensionRequest) (ConfigureBrowserExtensionResponse, error) {
	p.store.AddExtension(ctx, req.ExtensionID)
	// TODO: generate connection token and pairing token.
	connectionToken := uuid.NewString()
	pairingToken := uuid.NewString()

	return ConfigureBrowserExtensionResponse{
		ConnectionToken:              connectionToken,
		BrowserExtensionPairingToken: pairingToken,
	}, nil
}

type ExtensionWaitForConnectionInput struct {
	ResponseWriter http.ResponseWriter
	HttpReq        *http.Request
}

type WaitForConnectionResponse struct {
	BrowserExtensionProxyToken string `json:"browser_extension_proxy_token"`
	Status                     string `json:"status"`
	DeviceID                   string `json:"device_id"`
}

func (p *Pairing) ServePairingWS(w http.ResponseWriter, r *http.Request, extID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Errorf("Failed to upgrade on ServePairingWS: %v", err)
		return
	}
	logging.Infof("Starting paring WS for extension: %v", extID)
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
			logging.Infof("Closing paring ws after timeout: %v", extID)
			// todo: check if this is graceful
			conn.Close()
			return
		case <-connectedCheckTicker.C:
			pairingInfo, error := p.store.GetPairingInfo(r.Context(), extID)
			if err != nil {
				logging.Errorf("Failed to get pairing info: %v, retrying", error)
				continue
			}
			if !pairingInfo.IsPaired() {
				logging.Errorf("Paring ws device not connected: %v, retrying", extID)
				continue
			}
			if err := conn.WriteJSON(WaitForConnectionResponse{
				BrowserExtensionProxyToken: "fill",
				Status:                     "ok",
				DeviceID:                   pairingInfo.Device.DeviceID,
			}); err != nil {
				logging.Errorf("Failed to write to extension: %v, %v", extID, err)
				continue
			}
			logging.Infof("Paring ws finished for ext %v and device %v", extID, pairingInfo.Device.DeviceID)
			// TODO: write close message.
			conn.Close()
			return
		}
	}
}

// GetPairedDevice returns paired device and information if pairing was done.
func (p *Pairing) GetPairingInfo(ctx context.Context, extensionID string) (PairingInfo, error) {
	return p.store.GetPairingInfo(ctx, extensionID)
}

type ConfirmPairingRequest struct {
	FCMToken string `json:"fcm_token"`
	DeviceID string `json:"device_id"`
}

func (p *Pairing) ConfirmPairing(ctx context.Context, req ConfirmPairingRequest, extensionID string) error {
	return p.store.SetPairingInfo(ctx, extensionID, PairingInfo{
		Device: MobileDevice{
			DeviceID: req.DeviceID,
			FCMToken: req.FCMToken,
		},
		PairedAt: time.Now().UTC(),
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4 * 1024,
	WriteBufferSize: 4 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigin := os.Getenv("WEBSOCKET_ALLOWED_ORIGIN")

		if allowedOrigin != "" {
			return r.Header.Get("Origin") == allowedOrigin
		}

		return true
	},
}
