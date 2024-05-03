package main

import (
	"github.com/kelseyhightower/envconfig"

	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/pass"
)

func main() {
	logging.Init(logging.Fields{"service_name": "pass"})

	var cfg config.PassConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		logging.Fatal(err.Error())
	}

	server := pass.NewServer(cfg)

	if err := server.Run(); err != nil {
		logging.Fatal(err.Error())
	}
}
