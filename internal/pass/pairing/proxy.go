package pairing

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/recovery"
)

type Proxy struct {
	proxyPool *proxyPool
}

func NewProxy() *Proxy {
	proxyPool := &proxyPool{proxies: map[string]*proxyPair{}}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for {
			<-ticker.C
			proxyPool.deleteExpiresPairs()
		}
	}()
	return &Proxy{
		proxyPool: proxyPool,
	}
}

type proxyPool struct {
	mu      sync.Mutex
	proxies map[string]*proxyPair
}

// registerMobileConn register proxyPair if not existing in pool and returns it.
func (pp *proxyPool) getOrCreateProxyPair(deviceID string) *proxyPair {
	// TODO: handle delete.
	// TODO: right now two connections to the same WS results in race for messages/ decide if we want multiple conn or not.
	pp.mu.Lock()
	defer pp.mu.Unlock()
	v, ok := pp.proxies[deviceID]
	if !ok {
		v = initProxyPair()
	}
	pp.proxies[deviceID] = v
	return v
}

func (pp *proxyPool) deleteExpiresPairs() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	for key, pair := range pp.proxies {
		if time.Now().After(pair.expiresAt) {
			delete(pp.proxies, key)
		}
	}
}

type proxyPair struct {
	toMobileDataCh    chan []byte
	toExtensionDataCh chan []byte
	expiresAt         time.Time
}

// initProxyPair returns proxyPair and runs loop responsible for proxing data.
func initProxyPair() *proxyPair {
	const proxyTimeout = 3 * time.Minute
	return &proxyPair{
		toMobileDataCh:    make(chan []byte),
		toExtensionDataCh: make(chan []byte),
		expiresAt:         time.Now().Add(proxyTimeout),
	}
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}

	acceptedCloseStatus = []int{
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
		websocket.CloseAbnormalClosure,
	}
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4 * 1048
)

// client is a responsible for reading from read chan and sending it over wsConn
// and reading fom wsChan and sending it over send chan
type client struct {
	send chan []byte
	read chan []byte

	conn *websocket.Conn
}

func newClient(wsConn *websocket.Conn, send, read chan []byte) *client {
	return &client{
		send: send,
		read: read,
		conn: wsConn,
	}
}

// readPump pumps messages from the websocket connection to send.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *client) readPump() {
	defer func() {
		c.conn.Close()
		close(c.send)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, acceptedCloseStatus...) {
				logging.WithFields(logging.Fields{
					"reason": err.Error(),
				}).Error("Websocket connection closed unexpected")
			} else {
				logging.WithFields(logging.Fields{
					"reason": err.Error(),
				}).Info("Connection closed")
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.send <- message
	}
}

// writePump pumps messages from the read chan to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.read:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (p *Proxy) ServeExtensionProxyToMobileWS(w http.ResponseWriter, r *http.Request, extID, deviceID string) {
	log := logging.WithField("extension_id", extID).WithField("device_id", deviceID)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Failed to upgrade on ServeExtensionProxyToMobileWS: %v", err)
		return
	}

	log.Infof("Starting ServeExtensionProxyToMobileWS")

	proxyPair := p.proxyPool.getOrCreateProxyPair(deviceID)
	client := newClient(conn, proxyPair.toMobileDataCh, proxyPair.toExtensionDataCh)

	go recovery.DoNotPanic(func() {
		client.writePump()
	})

	go recovery.DoNotPanic(func() {
		client.readPump()
	})

	go recovery.DoNotPanic(func() {
		disconnectAfter := 3 * time.Minute
		timeout := time.After(disconnectAfter)

		<-timeout
		logging.Info("Connection closed after", disconnectAfter)

		client.conn.Close()
	})
}

func (p *Proxy) ServeMobileProxyToExtensionWS(w http.ResponseWriter, r *http.Request, deviceID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Errorf("Failed to upgrade on ServeMobileProxyToExtensionWS: %v", err)
		return
	}

	logging.Infof("Starting ServeMobileProxyToExtensionWS for dev: %v", deviceID)
	proxyPair := p.proxyPool.getOrCreateProxyPair(deviceID)

	client := newClient(conn, proxyPair.toExtensionDataCh, proxyPair.toMobileDataCh)

	go recovery.DoNotPanic(func() {
		client.writePump()
	})

	go recovery.DoNotPanic(func() {
		client.readPump()
	})

	go recovery.DoNotPanic(func() {
		disconnectAfter := 3 * time.Minute
		timeout := time.After(disconnectAfter)

		<-timeout
		logging.Info("Connection closed after", disconnectAfter)

		client.conn.Close()
	})
}
