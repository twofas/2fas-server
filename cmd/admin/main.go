package main

import (
	"github.com/gin-gonic/gin"

	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/api"
	"github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/logging"
)

func main() {
	logging.Init(logging.Fields{"service_name": "admin_api"})

	config.LoadConfiguration()

	application, err := api.NewApplication("admin-api", config.Config)
	if err != nil {
		logging.Fatalf("Failed to initialize application: %v", err)
	}

	logging.Infof("Initialize admin-api application: %q", config.Config.App.ListenAddr)
	logging.Infof("Environment is: %q", config.Config.Env)

	http.RunHttpServer(config.Config.App.ListenAddr, func(engine *gin.Engine) {
		application.RegisterAdminRoutes(engine)
	})
}
