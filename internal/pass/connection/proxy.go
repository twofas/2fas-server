package connection

import (
	"bytes"
	"time"

	"github.com/gorilla/websocket"

	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/recovery"
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

// proxy is a responsible for reading from read chan and sending it over wsConn
// and reading fom wsChan and sending it over send chan
type proxy struct {
	send chan []byte
	read chan []byte

	conn *websocket.Conn
}

func StartProxy(wsConn *websocket.Conn, send, read chan []byte) {
	proxy := &proxy{
		send: send,
		read: read,
		conn: wsConn,
	}

	go recovery.DoNotPanic(func() {
		proxy.writePump()
	})

	go recovery.DoNotPanic(func() {
		proxy.readPump()
	})

	go recovery.DoNotPanic(func() {
		disconnectAfter := 3 * time.Minute
		timeout := time.After(disconnectAfter)

		<-timeout
		logging.Info("Connection closed after", disconnectAfter)

		proxy.conn.Close()
	})
}

// readPump pumps messages from the websocket connection to send.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (p *proxy) readPump() {
	defer func() {
		p.conn.Close()
		close(p.send)
	}()

	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(func(string) error {
		p.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := p.conn.ReadMessage()
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
		p.send <- message
	}
}

// writePump pumps messages from the read chan to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (p *proxy) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()

	for {
		select {
		case message, ok := <-p.read:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := p.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
