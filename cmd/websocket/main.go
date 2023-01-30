package main

import (
	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/websocket"
)

func main() {
	logging.WithDefaultField("service_name", "websocket_api")

	config.LoadConfiguration()

	server := websocket.NewServer(config.Config.Websocket.ListenAddr)

	server.RunWebsocketServer()
}
