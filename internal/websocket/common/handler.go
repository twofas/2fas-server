package common

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/recovery"
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
		logging.Errorf("Starting new hub, there are %d hubs in total", len(h.channels))
	}

	return hub
}

func (h *ConnectionHandler) serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Errorf("Failed to upgrade connection: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

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

		client.hub.unregister <- client
		client.conn.Close()
	})
}
