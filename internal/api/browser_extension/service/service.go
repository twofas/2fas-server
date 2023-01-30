package service

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/api/browser_extension/adapters"
	"github.com/twofas/2fas-server/internal/api/browser_extension/app"
	"github.com/twofas/2fas-server/internal/api/browser_extension/app/command"
	"github.com/twofas/2fas-server/internal/api/browser_extension/app/query"
	apisec "github.com/twofas/2fas-server/internal/api/browser_extension/app/security"
	"github.com/twofas/2fas-server/internal/api/browser_extension/ports"
	"github.com/twofas/2fas-server/internal/api/mobile/domain"
	"github.com/twofas/2fas-server/internal/common/aws"
	"github.com/twofas/2fas-server/internal/common/db"
	mobile "github.com/twofas/2fas-server/internal/common/push"
	"github.com/twofas/2fas-server/internal/common/rate_limit"
	"github.com/twofas/2fas-server/internal/common/security"
	"gorm.io/gorm"
)

type BrowserExtensionModule struct {
	Cqrs          *app.Cqrs
	RoutesHandler *ports.RoutesHandler
	Redis         *redis.Client
	Config        config.Configuration
}

func NewBrowserExtensionModule(config config.Configuration, gorm *gorm.DB, database *sql.DB, redisClient *redis.Client) *BrowserExtensionModule {
	queryBuilder := db.NewQueryBuilder(database)

	browserExtensionsMysqlRepository := adapters.NewBrowserExtensionsMysqlRepository(gorm)
	browserExtension2FaRequestRepository := adapters.NewBrowserExtension2FaRequestsMysqlRepository(gorm)
	pairedDevicesRepository := adapters.NewBrowserExtensionDevicesMysqlRepository(gorm, queryBuilder)

	var pushClient mobile.Pusher

	if config.IsTestingEnv() {
		pushClient = mobile.NewFakePushClient()
	} else {
		s3 := aws.NewAwsS3(config.Aws.Region, config.Aws.S3AccessKeyId, config.Aws.S3AccessSecretKey)
		pushConfig := domain.NewFcmPushConfig(s3)
		pushClient = mobile.NewFcmPushClient(pushConfig)
	}

	cqrs := &app.Cqrs{
		Commands: app.Commands{
			RegisterBrowserExtension: command.RegisterBrowserExtensionHandler{
				Repository: browserExtensionsMysqlRepository,
			},
			RemoveAllBrowserExtensions: command.RemoveAllBrowserExtensionsHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			RemoveAllBrowserExtensionsDevices: command.RemoveAllBrowserExtensionsDevicesHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			UpdateBrowserExtension: command.UpdateBrowserExtensionHandler{
				Repository: browserExtensionsMysqlRepository,
			},
			Request2FaToken: command.Request2FaTokenHandler{
				BrowserExtensionsRepository:          browserExtensionsMysqlRepository,
				BrowserExtension2FaRequestRepository: browserExtension2FaRequestRepository,
				PairedDevicesRepository:              pairedDevicesRepository,
				Pusher:                               pushClient,
			},
			Close2FaRequest: command.Close2FaRequestHandler{
				BrowserExtensionsRepository:          browserExtensionsMysqlRepository,
				BrowserExtension2FaRequestRepository: browserExtension2FaRequestRepository,
			},
			RemoveExtensionPairedDevice: command.RemoveExtensionPairedDeviceHandler{
				BrowserExtensionRepository:              browserExtensionsMysqlRepository,
				BrowserExtensionPairedDevicesRepository: pairedDevicesRepository,
			},
			RemoveAllExtensionPairedDevices: command.RemoveALlExtensionPairedDevicesHandler{
				BrowserExtensionRepository:              browserExtensionsMysqlRepository,
				BrowserExtensionPairedDevicesRepository: pairedDevicesRepository,
			},
			StoreLogEvent: command.StoreLogEventHandler{
				BrowserExtensionsRepository: browserExtensionsMysqlRepository,
			},
		},

		Queries: app.Queries{
			BrowserExtensionQuery: query.BrowserExtensionQueryHandler{
				Database: gorm,
			},
			BrowserExtensionPairedDevicesQuery: query.BrowserExtensionPairedMobileDevicesQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			BrowserExtensionPairedDeviceQuery: query.BrowserExtensionPairedMobileDeviceQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			BrowserExtension2FaRequestQuery: query.BrowserExtension2FaRequestQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
		},
	}

	routesHandler := ports.NewRoutesHandler(cqrs)

	module := &BrowserExtensionModule{
		Cqrs:          cqrs,
		RoutesHandler: routesHandler,
		Redis:         redisClient,
		Config:        config,
	}

	return module
}

func (m *BrowserExtensionModule) RegisterRoutes(router *gin.Engine) {
	rateLimiter := rate_limit.New(m.Redis)

	bandwidthAuditMiddleware := apisec.BrowserExtensionBandwidthAuditMiddleware(rateLimiter)
	iPAbuseAuditMiddleware := security.IPAbuseAuditMiddleware(rateLimiter)

	publicRouter := router.Group("/")
	publicRouter.Use(iPAbuseAuditMiddleware)
	publicRouter.Use(bandwidthAuditMiddleware)

	publicRouter.POST("/browser_extensions", m.RoutesHandler.RegisterBrowserExtension)
	publicRouter.GET("/browser_extensions/:extension_id", m.RoutesHandler.FindBrowserExtension)
	publicRouter.PUT("/browser_extensions/:extension_id", m.RoutesHandler.UpdateBrowserExtension)

	publicRouter.GET("/browser_extensions/:extension_id/devices", m.RoutesHandler.FindBrowserExtensionPairedMobileDevices)
	publicRouter.GET("/browser_extensions/:extension_id/devices/:device_id", m.RoutesHandler.GetBrowserExtensionPairedMobileDevice)
	publicRouter.DELETE("/browser_extensions/:extension_id/devices", m.RoutesHandler.RemoveAllExtensionPairedDevices)
	publicRouter.DELETE("/browser_extensions/:extension_id/devices/:device_id", m.RoutesHandler.RemovePairedDeviceFromExtension)

	publicRouter.POST("/browser_extensions/:extension_id/commands/request_2fa_token", m.RoutesHandler.Request2FaToken)
	publicRouter.POST("/browser_extensions/:extension_id/commands/store_log", m.RoutesHandler.Log)

	publicRouter.GET("/browser_extensions/:extension_id/2fa_requests", m.RoutesHandler.GetAllBrowserExtension2FaTokenRequests)
	publicRouter.GET("/browser_extensions/:extension_id/2fa_requests/:token_request_id", m.RoutesHandler.GetBrowserExtension2FaTokenRequest)
	publicRouter.POST("/browser_extensions/:extension_id/2fa_requests/:token_request_id/commands/close_2fa_request", m.RoutesHandler.Close2FaRequest)

	if m.Config.IsTestingEnv() {
		publicRouter.DELETE("/browser_extensions", m.RoutesHandler.RemoveAllBrowserExtensions)
		publicRouter.DELETE("/browser_extensions/devices", m.RoutesHandler.RemoveAllBrowserExtensionsDevices)
	}

}
