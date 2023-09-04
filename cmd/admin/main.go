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

	logging.Info("Initialize admin application ", config.Config.App.ListenAddr)
	logging.Info("Environment is: ", config.Config.Env)

	http.RunHttpServer(config.Config.App.ListenAddr, func(engine *gin.Engine) {
		application.RegisterAdminRoutes(engine)
	})
}
