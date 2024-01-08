package pairing

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type Proxy struct {
	wsStore wsStore
}

func NewProxy() *Proxy {
	return &Proxy{wsStore: NewWSMemoryStore()}
}

type wsStore interface {
	SetMobileConn(ctx context.Context, deviceID string, conn *websocket.Conn)
	GetMobileConn(ctx context.Context, deviceID string) (*websocket.Conn, bool)
}

func (p *Proxy) ServeExtensionProxyToMobileWS(w http.ResponseWriter, r *http.Request, extID, deviceID string) {
	extConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Errorf("Failed to upgrade on ServeExtensionProxyToMobileWS: %v", err)
		return
	}
	logging.Infof("Starting ServeExtensionProxyToMobileWS for extension: %v", extID)
	const (
		maxWaitTime = 3 * time.Minute
	)
	mobileConn, ok := p.wsStore.GetMobileConn(r.Context(), deviceID)
	if !ok {
		logging.Errorf("Could not found ws mobile connection: %v - %v", extID, deviceID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO: pong handlers

	defer extConn.Close()
	defer mobileConn.Close()

	go copyLoop(extConn, mobileConn)
	copyLoop(mobileConn, extConn)
}

func copyLoop(dst, src *websocket.Conn) {
	for {
		msgT, msg, err := src.ReadMessage()
		if err != nil {
			logging.Errorf("read:", err)
			break
		}
		if err := dst.WriteMessage(msgT, msg); err != nil {
			logging.Errorf("wrote:", err)
			break
		}
	}
}

func (p *Proxy) ServeMobileProxyToExtensionWS(w http.ResponseWriter, r *http.Request, deviceID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Errorf("Failed to upgrade on ServeMobileProxyToExtensionWS: %v", err)
		return
	}
	logging.Infof("Starting ServeMobileProxyToExtensionWS for dev: %v", deviceID)
	// TODO: when we need to set timeouts? and close conn.
	p.wsStore.SetMobileConn(r.Context(), deviceID, conn)
	// ServeExtensionProxyToMobileWS is responsible for proxing, here we only store ws conn.

	// TODO: what to do if there is 2nd connection to proxy endpoint (on retry)?
}
