package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/websocket"

	app_http "github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type WebsocketApiClient struct {
	wsAddr string
}

func NewWebsocketApiClient(websocketApiUrl string) *WebsocketApiClient {
	return &WebsocketApiClient{
		wsAddr: websocketApiUrl,
	}
}

func (ws *WebsocketApiClient) SendMessage(uri string, message interface{}) error {
	u, err := url.Parse(ws.wsAddr)
	if err != nil {
		return fmt.Errorf("failed to parse %q: %w", ws.wsAddr, err)
	}
	u.Path = path.Join(u.Path, uri)

	msg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	logging.WithFields(logging.Fields{
		"message": string(msg),
		"ws_url":  u.String(),
	}).Info("Start command `SendWebsocketMessage`")

	requestHeaders := http.Header{
		app_http.CorrelationIdHeader: []string{app_http.CorrelationId},
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), requestHeaders)
	if err != nil {
		return fmt.Errorf("failed to dial: %q: %w", u.String(), err)
	}

	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Cannot send websocket message")
		return fmt.Errorf("failed to write message to the conection: %w", err)
	}

	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		logging.WithField("error", err.Error()).Error("Cannot close websocket connection")
		return fmt.Errorf("failed to write close message to the conection: %w", err)
	}

	return nil
}
