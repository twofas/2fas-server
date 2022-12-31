package common

import (
	"github.com/2fas/api/internal/common/logging"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

func (h *ConnectionHandler) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Request.URL.Path

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

	go client.writePump()
	go client.readPump()

	go func() {
		<-time.After(time.Duration(3) * time.Minute)

		defer func() {
			logging.Debug("Disconnect websocket client")
			client.hub.unregister <- client
			client.conn.Close()
		}()
	}()
}
