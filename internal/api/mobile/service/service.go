package service

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/twofas/2fas-server/config"
	browser_extension_adapters "github.com/twofas/2fas-server/internal/api/browser_extension/adapters"
	"github.com/twofas/2fas-server/internal/api/mobile/adapters"
	"github.com/twofas/2fas-server/internal/api/mobile/app"
	"github.com/twofas/2fas-server/internal/api/mobile/app/command"
	query "github.com/twofas/2fas-server/internal/api/mobile/app/queries"
	apisec "github.com/twofas/2fas-server/internal/api/mobile/app/security"
	"github.com/twofas/2fas-server/internal/api/mobile/ports"
	"github.com/twofas/2fas-server/internal/common/clock"
	"github.com/twofas/2fas-server/internal/common/db"
	"github.com/twofas/2fas-server/internal/common/rate_limit"
	"github.com/twofas/2fas-server/internal/common/security"
	"github.com/twofas/2fas-server/internal/common/websocket"
	"gorm.io/gorm"
)

type MobileModule struct {
	Cqrs          *app.Cqrs
	RoutesHandler *ports.RoutesHandler
	Config        config.Configuration
	Redis         *redis.Client
}

func NewMobileModule(config config.Configuration, gorm *gorm.DB, database *sql.DB, validate *validator.Validate, redisClient *redis.Client) *MobileModule {
	queryBuilder := db.NewQueryBuilder(database)

	mobileDeviceRepository := adapters.NewMobileDeviceMysqlRepository(gorm)
	notificationsRepository := adapters.NewMobileNotificationMysqlRepository(gorm)

	mobileApplicationExtensionsService := adapters.NewDeviceExtensionsService(gorm, queryBuilder, clock.New())

	websocketClient := websocket.NewWebsocketApiClient(config.Websocket.ApiUrl)

	browserExtensionsMysqlRepository := browser_extension_adapters.NewBrowserExtensionsMysqlRepository(gorm)
	mobileDeviceExtensionsRepository := adapters.NewMobileDeviceExtensionsGormRepository(gorm, queryBuilder)

	validate.RegisterValidation("is-device-id", DeviceIdExistsValidator(mobileDeviceRepository))

	cqrs := &app.Cqrs{
		Commands: app.Commands{
			RegisterMobileDevice: &command.RegisterMobileDeviceHandler{
				Repository: mobileDeviceRepository,
			},
			RemoveAllMobileDevices: &command.RemoveAllMobileDevicesHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			UpdateMobileDevice: &command.UpdateMobileDeviceHandler{Repository: mobileDeviceRepository},
			CreateNotification: &command.CreateNotificationHandler{Repository: notificationsRepository},
			UpdateNotification: &command.UpdateNotificationHandler{Repository: notificationsRepository},
			DeleteNotification: &command.DeleteNotificationHandler{Repository: notificationsRepository},
			RemoveAllMobileNotifications: &command.DeleteAllNotificationsHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			PublishNotification: &command.PublishNotificationHandler{Repository: notificationsRepository},
			PairMobileWithExtension: &command.PairMobileWithExtensionHandler{
				BrowserExtensionsRepository:        browserExtensionsMysqlRepository,
				MobileApplicationExtensionsService: mobileApplicationExtensionsService,
				MobileDeviceExtensionsRepository:   mobileDeviceExtensionsRepository,
				WebsocketClient:                    websocketClient,
			},
			RemovePairingWithExtension: &command.RemoveDeviceExtensionHandler{
				MobileDeviceExtensionsRepository: mobileDeviceExtensionsRepository,
			},
			Send2FaToken: &command.Send2FaTokenHandler{
				BrowserExtensionsRepository:        browserExtensionsMysqlRepository,
				MobileApplicationExtensionsService: mobileApplicationExtensionsService,
				WebsocketClient:                    websocketClient,
			},
		},
		Queries: app.Queries{
			MobileDeviceQuery: &query.MobileDeviceQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			DeviceBrowserExtensionsQuery: &query.DeviceBrowserExtensionsQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			DeviceBrowserExtension2FaRequestQuery: &query.DeviceBrowserExtension2FaRequestQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
				Clock:    clock.New(),
			},
			PairedBrowserExtensionQuery: &query.PairedBrowserExtensionQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			MobileNotificationsQuery: &query.MobileNotificationsQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
		},
	}

	routesHandler := ports.NewRoutesHandler(cqrs, validate, mobileDeviceRepository)

	module := &MobileModule{
		Cqrs:          cqrs,
		RoutesHandler: routesHandler,
		Config:        config,
		Redis:         redisClient,
	}

	return module
}

func (m *MobileModule) RegisterPublicRoutes(router *gin.Engine) {
	rateLimiter := rate_limit.New(m.Redis)

	bandwidthMobileApiMiddleware := apisec.MobileIpAbuseAuditMiddleware(rateLimiter, m.Config.Security.RateLimitMobile)
	iPAbuseAuditMiddleware := security.IPAbuseAuditMiddleware(rateLimiter, m.Config.Security.RateLimitIP)

	publicRouter := router.Group("/")
	publicRouter.Use(iPAbuseAuditMiddleware)
	publicRouter.Use(bandwidthMobileApiMiddleware)

	publicRouter.POST("/mobile/devices", m.RoutesHandler.RegisterMobileDevice)

	publicRouter.PUT("/mobile/devices/:device_id", m.RoutesHandler.UpdateMobileDevice)

	publicRouter.GET("/mobile/notifications", m.RoutesHandler.FindAllMobileNotifications)
	publicRouter.GET("/mobile/notifications/:notification_id", m.RoutesHandler.FindMobileNotification)

	publicRouter.POST("/mobile/devices/:device_id/commands/send_2fa_token", m.RoutesHandler.Send2FaToken)
	publicRouter.GET("/mobile/devices/:device_id/browser_extensions/2fa_requests", m.RoutesHandler.GetAll2FaTokenRequests)
	publicRouter.POST("/mobile/devices/:device_id/browser_extensions", m.RoutesHandler.PairMobileWithExtension)
	publicRouter.DELETE("/mobile/devices/:device_id/browser_extensions/:extension_id", m.RoutesHandler.RemovePairingWithExtension)
	publicRouter.GET("/mobile/devices/:device_id/browser_extensions", m.RoutesHandler.FindAllMobileAppExtensions)
	publicRouter.GET("/mobile/devices/:device_id/browser_extensions/:extension_id", m.RoutesHandler.FindMobileAppExtensionById)
}

func (m *MobileModule) RegisterAdminRoutes(g *gin.RouterGroup) {
	g.POST("/mobile/notifications", m.RoutesHandler.CreateMobileNotification)
	g.PUT("/mobile/notifications/:notification_id", m.RoutesHandler.UpdateMobileNotification)
	g.DELETE("/mobile/notifications/:notification_id", m.RoutesHandler.RemoveMobileNotification)
	g.POST("/mobile/notifications/:notification_id/commands/publish", m.RoutesHandler.PublishMobileNotification)

	if m.Config.IsTestingEnv() {
		g.DELETE("/mobile/notifications", m.RoutesHandler.RemoveAllMobileNotifications)
		g.DELETE("/mobile/devices", m.RoutesHandler.RemoveAllMobileDevices)
	}
}
