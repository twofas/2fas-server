package api

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/twofas/2fas-server/config"
	extension "github.com/twofas/2fas-server/internal/api/browser_extension/service"
	health "github.com/twofas/2fas-server/internal/api/health/service"
	icons "github.com/twofas/2fas-server/internal/api/icons/service"
	"github.com/twofas/2fas-server/internal/api/mobile/domain"
	mobile "github.com/twofas/2fas-server/internal/api/mobile/service"
	support "github.com/twofas/2fas-server/internal/api/support/service"
	"github.com/twofas/2fas-server/internal/common/api"
	"github.com/twofas/2fas-server/internal/common/db"
	"github.com/twofas/2fas-server/internal/common/push"
	"github.com/twofas/2fas-server/internal/common/redis"
	"github.com/twofas/2fas-server/internal/common/storage"
	"github.com/twofas/2fas-server/internal/common/validation"
)

var validate *validator.Validate

type Module interface {
	RegisterPublicRoutes(router *gin.Engine)
	RegisterAdminRoutes(g *gin.RouterGroup)
}

type Application struct {
	Addr         string
	Router       *gin.Engine
	Config       config.Configuration
	Modules      []Module
	HealthModule *health.HealthModule
}

func NewApplication(applicationName string, config config.Configuration) (*Application, error) {
	validate = validator.New()

	gorm := db.NewGormConnection(config)
	database := db.NewDbConnection(config)
	redisClient := redis.New(config.Redis.ServiceUrl, config.Redis.Port)

	validate.RegisterValidation("not_blank", validation.NotBlank)

	h := health.NewHealthModule(applicationName, config, redisClient)

	var strg storage.FileSystemStorage
	var pushClient push.Pusher

	if config.IsTestingEnv() {
		strg = storage.NewTmpFileSystem()
		pushClient = push.NewFakePushClient()
	} else {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(config.Aws.Region),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create aws session: %w", err)
		}
		strg = storage.NewS3Storage(sess)
		pushConfig, err := domain.NewFcmPushConfig(strg)
		if err != nil {
			return nil, fmt.Errorf("failed to create fcm push config: %w", err)
		}
		pushClient = push.NewFcmPushClient(pushConfig)
	}

	modules := []Module{
		h,
		support.NewSupportModule(config, gorm, database, validate, strg),
		icons.NewIconsModule(config, gorm, database, validate, strg),
		extension.NewBrowserExtensionModule(config, gorm, database, redisClient, validate, pushClient),
		mobile.NewMobileModule(config, gorm, database, validate, redisClient),
	}

	app := &Application{
		Addr:         config.App.ListenAddr,
		Config:       config,
		Modules:      modules,
		HealthModule: h,
	}

	return app, nil
}

func (a *Application) RegisterRoutes(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, api.NotFoundError(errors.New("URI not found")))
	})

	for _, module := range a.Modules {
		module.RegisterPublicRoutes(router)
	}
}

func (a *Application) RegisterAdminRoutes(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, api.NotFoundError(errors.New("URI not found")))
	})

	// The only route method is /health. Everything else
	// is nested under /admin so that oAuth proxy can route to it.
	a.HealthModule.RegisterHealth(router)

	g := router.Group("/admin")

	for _, module := range a.Modules {
		module.RegisterAdminRoutes(g)
	}
}
