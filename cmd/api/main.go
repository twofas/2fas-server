package main

import (
	"github.com/gin-gonic/gin"

	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/api"
	"github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/logging"
)

func main() {
	logging.Init(logging.Fields{"service_name": "api"})

	config.LoadConfiguration()

	application, err := api.NewApplication("api", config.Config)
	if err != nil {
		logging.Fatalf("Failed to initialize application: %v", err)
	}

	logging.Info("Initialize application ", config.Config.App.ListenAddr)
	logging.Info("Environment is: ", config.Config.Env)

	http.RunHttpServer(config.Config.App.ListenAddr, func(engine *gin.Engine) {
		application.RegisterRoutes(engine)
	})
}
