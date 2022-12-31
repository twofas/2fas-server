package main

import (
	"github.com/2fas/api/config"
	"github.com/2fas/api/internal/common/logging"
	"github.com/2fas/api/internal/websocket"
)

func main() {
	logging.WithDefaultField("service_name", "websocket_api")

	config.LoadConfiguration()

	server := websocket.NewServer(config.Config.Websocket.ListenAddr)

	server.RunWebsocketServer()
}
