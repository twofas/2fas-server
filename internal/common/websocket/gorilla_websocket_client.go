package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	app_http "github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/logging"
	"net/http"
	"net/url"
	"path"
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
	u, _ := url.Parse(ws.wsAddr)
	u.Path = path.Join(u.Path, uri)

	msg, err := json.Marshal(message)

	if err != nil {
		return err
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
		return err
	}

	err = c.WriteMessage(websocket.TextMessage, msg)

	if err != nil {
		logging.WithField("error", err.Error()).Error("Cannot send websocket message")
		return err
	}

	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	if err != nil {
		logging.WithField("error", err.Error()).Error("Cannot close websocket connection")
		return err
	}

	return nil
}
