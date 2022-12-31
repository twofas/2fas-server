package main

import (
	"github.com/2fas/api/config"
	"github.com/2fas/api/internal/api"
	"github.com/2fas/api/internal/common/http"
	"github.com/2fas/api/internal/common/logging"
	"github.com/gin-gonic/gin"
)

func main() {
	logging.WithDefaultField("service_name", "api")

	config.LoadConfiguration()

	application := api.NewApplication(config.Config)

	logging.Info("Initialize application ", config.Config.App.ListenAddr)
	logging.Info("Environment is: ", config.Config.Env)

	http.RunHttpServer(config.Config.App.ListenAddr, func(engine *gin.Engine) {
		application.RegisterRoutes(engine)
	})
}
