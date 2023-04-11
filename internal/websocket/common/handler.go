package common

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/recovery"
	"net/http"
	"os"
	"time"
)

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

type ConnectionHandler struct {
	channels map[string]*Hub
}

func NewConnectionHandler() *ConnectionHandler {
	channels := make(map[string]*Hub)

	return &ConnectionHandler{
		channels: channels,
	}
}

func (h *ConnectionHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Request.URL.Path

		logging.WithDefaultField("channel", channel)
		logging.WithDefaultField("ip", c.ClientIP())

		logging.Info("New channel subscriber")

		hub := h.getHub(channel)

		h.serveWs(hub, c.Writer, c.Request)
	}
}

func (h *ConnectionHandler) getHub(channel string) *Hub {
	var hub *Hub

	hub, ok := h.channels[channel]

	if !ok {
		hub = NewHub()

		go hub.Run()

		h.channels[channel] = hub
	}

	return hub
}

func (h *ConnectionHandler) serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go recovery.DoNotPanic(func() {
		client.writePump()
	})

	go recovery.DoNotPanic(func() {
		client.readPump()
	})

	go func() {
		disconnectAfter := 3 * time.Minute
		timeout := time.After(disconnectAfter)

		select {
		case <-timeout:
			logging.Info("Connection closed after", disconnectAfter)

			client.hub.unregister <- client
			client.conn.Close()
		}
	}()
}
