package proxy

import (
	"bytes"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/recovery"
)

const (
	DefaultWriteTimeout    = 10 * time.Second
	DefaultReadTimeout     = 20 * time.Second
	DefaultPingFrequency   = DefaultReadTimeout / 4
	DefaultDisconnectAfter = 3 * time.Minute

	// Maximum message size allowed from peer.
	maxMessageSize = 10 * (2 << 20)
)

type Config struct {
	WriteTimeout    time.Duration
	ReadTimeout     time.Duration
	PingFrequency   time.Duration
	DisconnectAfter time.Duration
}

func DefaultConfig() Config {
	return Config{
		WriteTimeout:    DefaultWriteTimeout,
		ReadTimeout:     DefaultReadTimeout,
		PingFrequency:   DefaultPingFrequency,
		DisconnectAfter: DefaultDisconnectAfter,
	}
}

type WriterCloser interface {
	Write(msg []byte)
	Close()
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

// proxy is a responsible for reading from reader chan and sending it over conn
// and reading fom conn and sending it over writer.
type proxy struct {
	writer WriterCloser
	reader chan []byte
	conn   *websocket.Conn
	cfg    Config
}

func Start(wsConn *websocket.Conn, writer WriterCloser, reader chan []byte, cfg Config) {
	proxy := &proxy{
		writer: writer,
		reader: reader,
		conn:   wsConn,
		cfg:    cfg,
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go recovery.DoNotPanic(func() {
		defer wg.Done()
		proxy.writePump()
	})

	go recovery.DoNotPanic(func() {
		defer wg.Done()
		proxy.readPump()
	})

	go recovery.DoNotPanic(func() {
		timeout := time.After(cfg.DisconnectAfter)

		<-timeout
		logging.Info("Connection closed after", cfg.DisconnectAfter)

		proxy.conn.Close()
	})

	wg.Wait()
}

// readPump pumps messages from the websocket proxy to writer.
//
// The application runs readPump in a per-proxy goroutine. The application
// ensures that there is at most one reader on a proxy by executing all
// reads from this goroutine.
func (p *proxy) readPump() {
	defer func() {
		p.conn.Close()
		p.writer.Close()
	}()

	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(p.cfg.ReadTimeout))
	p.conn.SetPongHandler(func(string) error {
		p.conn.SetReadDeadline(time.Now().Add(p.cfg.ReadTimeout))
		return nil
	})

	for {
		_, message, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, acceptedCloseStatus...) {
				logging.WithFields(logging.Fields{
					"reason": err.Error(),
				}).Error("Websocket proxy closed unexpected")
			} else {
				logging.WithFields(logging.Fields{
					"reason": err.Error(),
				}).Info("Connection closed")
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		p.writer.Write(message)
	}
}

// writePump pumps messages from the reader chan to the websocket proxy.
//
// A goroutine running writePump is started for each proxy. The
// application ensures that there is at most one writer to a proxy by
// executing all writes from this goroutine.
func (p *proxy) writePump() {
	ticker := time.NewTicker(p.cfg.PingFrequency)
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()

	for {
		select {
		case message, ok := <-p.reader:
			p.conn.SetWriteDeadline(time.Now().Add(p.cfg.WriteTimeout))
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
			p.conn.SetWriteDeadline(time.Now().Add(p.cfg.WriteTimeout))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
