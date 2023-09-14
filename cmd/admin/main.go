package main

import (
	"github.com/gin-gonic/gin"

	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/api"
	"github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/logging"
)

func main() {
	logging.WithDefaultField("service_name", "admin_api")

	config.LoadConfiguration()

	application := api.NewApplication(config.Config)

	logging.Infof("Initialize admin application: %q", config.Config.App.ListenAddr)
	logging.Infof("Environment is: %q", config.Config.Env)

	http.RunHttpServer(config.Config.App.ListenAddr, func(engine *gin.Engine) {
		application.RegisterAdminRoutes(engine)
	})
}
