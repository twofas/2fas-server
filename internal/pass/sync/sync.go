package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/pass/connection"
	"github.com/twofas/2fas-server/internal/pass/sign"
)

type Syncing struct {
	store   store
	signSvc *sign.Service
}

type store interface {
	RequestSync(fmtToken string)
	ConfirmSync(fmtToken string) bool
	IsSyncCofirmed(fmtToken string) bool
}

func NewPairingApp(signService *sign.Service, fakeMobilePush bool) *Syncing {
	return &Syncing{
		store:   NewMemoryStore(),
		signSvc: signService,
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

type ExtensionWaitForConnectionInput struct {
	ResponseWriter http.ResponseWriter
	HttpReq        *http.Request
}

type RequestSyncResponse struct {
	BrowserExtensionProxyToken string `json:"browser_extension_proxy_token"`
	Status                     string `json:"status"`
}

type MobileSyncPayload struct {
	MobileSyncToken string `json:"mobile_sync_token"`
}

func (p *Syncing) ServeSyncingRequestWS(w http.ResponseWriter, r *http.Request, fcmToken string) error {
	log := logging.WithField("fcm_token", fcmToken)
	conn, err := connection.Upgrade(w, r)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %w", err)
	}
	defer conn.Close()

	log.Infof("Starting sync request WS for %q", fcmToken)
	p.requestSync(r.Context(), fcmToken)

	if syncDone := p.isSyncConfirmed(r.Context(), fcmToken); syncDone {
		if err := p.sendTokenAndCloseConn(fcmToken, conn); err != nil {
			log.Errorf("Failed to send token: %v", err)
		}
		log.Infof("Paring ws finished")
		return nil
	}

	const (
		maxWaitTime              = 3 * time.Minute
		checkIfConnectedInterval = time.Second
	)
	maxWaitC := time.After(maxWaitTime)
	connectedCheckTicker := time.NewTicker(checkIfConnectedInterval)
	defer connectedCheckTicker.Stop()
	for {
		select {
		case <-maxWaitC:
			log.Info("Closing paring ws after timeout")
			return nil
		case <-connectedCheckTicker.C:
			if syncConfirmed := p.isSyncConfirmed(r.Context(), fcmToken); syncConfirmed {
				if err := p.sendTokenAndCloseConn(fcmToken, conn); err != nil {
					log.Errorf("Failed to send token: %v", err)
					return nil
				}
				log.Infof("Paring ws finished")
				return nil
			}
		}
	}
}

func (p *Syncing) isSyncConfirmed(ctx context.Context, fcmToken string) bool {
	return p.store.IsSyncCofirmed(fcmToken)
}

func (p *Syncing) requestSync(ctx context.Context, fcmToken string) {
	p.store.RequestSync(fcmToken)
}

func (p *Syncing) sendTokenAndCloseConn(fcmToken string, conn *websocket.Conn) error {
	extProxyToken, err := p.signSvc.SignAndEncode(sign.Message{
		ConnectionID:   fcmToken,
		ExpiresAt:      time.Now().Add(pairingTokenValidityDuration),
		ConnectionType: sign.ConnectionTypeBrowserExtensionSync,
	})
	if err != nil {
		return fmt.Errorf("failed to generate ext proxy token: %v", err)
	}

	if err := conn.WriteJSON(RequestSyncResponse{
		BrowserExtensionProxyToken: extProxyToken,
		Status:                     "ok",
	}); err != nil {
		return fmt.Errorf("failed to write to extension: %v", err)
	}
	return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (p *Syncing) sendMobileToken(fcmToken string, resp http.ResponseWriter) error {
	extProxyToken, err := p.signSvc.SignAndEncode(sign.Message{
		ConnectionID:   fcmToken,
		ExpiresAt:      time.Now().Add(pairingTokenValidityDuration),
		ConnectionType: sign.ConnectionTypeMobileSyncConfirm,
	})
	if err != nil {
		return fmt.Errorf("failed to generate ext proxy token: %v", err)
	}

	bb, err := json.Marshal(struct {
		MobileSyncConfirmToken string `json:"mobile_sync_confirm_token"`
	}{
		MobileSyncConfirmToken: extProxyToken,
	})
	if err != nil {
		return fmt.Errorf("failed to write to extension: %v", err)
	}
	resp.Write(bb)
	return nil
}

type ConfirmSyncResponse struct {
	ProxyToken string `json:"proxy_token"`
}

var noSyncRequestErr = errors.New("sync request was not created")

func (p *Syncing) confirmSync(ctx context.Context, fcmToken string) (ConfirmSyncResponse, error) {
	logging.Infof("Starting sync confirm for %q", fcmToken)

	mobileProxyToken, err := p.signSvc.SignAndEncode(sign.Message{
		ConnectionID:   fcmToken,
		ExpiresAt:      time.Now().Add(pairingTokenValidityDuration),
		ConnectionType: sign.ConnectionTypeMobileSyncConfirm,
	})
	if err != nil {
		return ConfirmSyncResponse{}, fmt.Errorf("failed to generate ext proxy token: %v", err)
	}
	if ok := p.store.ConfirmSync(fcmToken); !ok {
		return ConfirmSyncResponse{}, noSyncRequestErr
	}

	return ConfirmSyncResponse{ProxyToken: mobileProxyToken}, nil
}
