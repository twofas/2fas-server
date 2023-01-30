package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/twofas/2fas-server/config"
	extension "github.com/twofas/2fas-server/internal/api/browser_extension/service"
	health "github.com/twofas/2fas-server/internal/api/health/service"
	icons "github.com/twofas/2fas-server/internal/api/icons/service"
	mobile "github.com/twofas/2fas-server/internal/api/mobile/service"
	support "github.com/twofas/2fas-server/internal/api/support/service"
	"github.com/twofas/2fas-server/internal/common/api"
	"github.com/twofas/2fas-server/internal/common/db"
	"github.com/twofas/2fas-server/internal/common/redis"
)

var validate *validator.Validate

type Module interface {
	RegisterRoutes(router *gin.Engine)
}

type Application struct {
	Addr string

	Router *gin.Engine

	Config config.Configuration

	Modules []Module
}

func NewApplication(config config.Configuration) *Application {
	validate = validator.New()

	gorm := db.NewGormConnection(config)
	database := db.NewDbConnection(config)
	redisClient := redis.New(config.Redis.ServiceUrl, config.Redis.Port)

	modules := []Module{
		health.NewHealthModule(config, redisClient),
		support.NewSupportModule(config, gorm, database, validate),
		icons.NewIconsModule(config, gorm, database, validate),
		extension.NewBrowserExtensionModule(config, gorm, database, redisClient),
		mobile.NewMobileModule(config, gorm, database, validate, redisClient),
	}

	app := &Application{
		Addr:    config.App.ListenAddr,
		Config:  config,
		Modules: modules,
	}

	return app
}

func (a *Application) RegisterRoutes(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, api.NotFoundError(errors.New("URI not found")))
	})

	for _, module := range a.Modules {
		module.RegisterRoutes(router)
	}
}
