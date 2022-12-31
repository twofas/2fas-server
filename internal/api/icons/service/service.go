package service

import (
	"database/sql"
	"github.com/2fas/api/config"
	"github.com/2fas/api/internal/api/icons/adapters"
	"github.com/2fas/api/internal/api/icons/app"
	"github.com/2fas/api/internal/api/icons/app/command"
	"github.com/2fas/api/internal/api/icons/app/queries"
	"github.com/2fas/api/internal/api/icons/ports"
	"github.com/2fas/api/internal/common/aws"
	"github.com/2fas/api/internal/common/db"
	httpsec "github.com/2fas/api/internal/common/http"
	"github.com/2fas/api/internal/common/storage"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type IconsModule struct {
	Cqrs          *app.Cqrs
	RoutesHandler *ports.RoutesHandler
	Config        config.Configuration
}

func NewIconsModule(config config.Configuration, gorm *gorm.DB, database *sql.DB, validate *validator.Validate) *IconsModule {
	queryBuilder := db.NewQueryBuilder(database)

	var iconsStorage storage.FileSystemStorage

	if config.IsTestingEnv() {
		iconsStorage = storage.NewTmpFileSystem()
	} else {
		iconsStorage = aws.NewAwsS3(config.Aws.Region, config.Aws.S3AccessKeyId, config.Aws.S3AccessSecretKey)
	}

	webServicesRepository := adapters.NewWebServiceMysqlRepository(gorm)
	iconsRepository := adapters.NewIconMysqlRepository(gorm)
	iconsRelationsRepository := adapters.NewIconsRelationsMysqlRepository(gorm)
	iconsRequestRepository := adapters.NewIconRequestMysqlRepository(gorm)
	iconsCollectionRepository := adapters.NewIconsCollectionMysqlRepository(gorm)
	iconsCollectionRelationsRepository := adapters.NewIconsCollectionsRelationsMysqlRepository(gorm)

	cqrs := &app.Cqrs{
		Commands: app.Commands{
			CreateWebService: &command.CreateWebServiceHandler{Repository: webServicesRepository},
			UpdateWebService: &command.UpdateWebServiceHandler{Repository: webServicesRepository},
			RemoveWebService: &command.DeleteWebServiceHandler{Repository: webServicesRepository},
			RemoveAllWebServices: &command.DeleteAllWebServicesHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			CreateIcon: &command.CreateIconHandler{Repository: iconsRepository, Storage: iconsStorage},
			UpdateIcon: &command.UpdateIconHandler{Repository: iconsRepository, Storage: iconsStorage},
			RemoveIcon: &command.DeleteIconHandler{
				Repository:              iconsRepository,
				IconsRelationRepository: iconsRelationsRepository,
			},
			RemoveAllIcons: &command.DeleteAllIconsHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			CreateIconRequest: &command.CreateIconRequestHandler{
				Storage:    iconsStorage,
				Repository: iconsRequestRepository,
			},
			RemoveIconRequest: &command.DeleteIconRequestHandler{
				Repository: iconsRequestRepository,
			},
			RemoveAllIconsRequests: &command.DeleteAllIconsRequestsHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			UpdateWebServiceFromIconRequest: &command.UpdateWebServiceFromIconRequestHandler{
				IconsStorage:               iconsStorage,
				WebServiceRepository:       webServicesRepository,
				IconsCollectionsRepository: iconsCollectionRepository,
				IconsRequestsRepository:    iconsRequestRepository,
				IconsRepository:            iconsRepository,
			},
			TransformIconRequestToWebService: &command.TransformIconRequestToWebServiceHandler{
				IconsStorage:               iconsStorage,
				WebServiceRepository:       webServicesRepository,
				IconsRepository:            iconsRepository,
				IconsCollectionsRepository: iconsCollectionRepository,
				IconsRequestsRepository:    iconsRequestRepository,
			},
			CreateIconsCollection: &command.CreateIconsCollectionHandler{Repository: iconsCollectionRepository},
			UpdateIconsCollection: &command.UpdateIconsCollectionHandler{Repository: iconsCollectionRepository},
			RemoveIconsCollection: &command.DeleteIconsCollectionHandler{
				Repository:                          iconsCollectionRepository,
				IconsCollectionsRelationsRepository: iconsCollectionRelationsRepository,
			},
			RemoveAllIconsCollections: &command.DeleteAllIconsCollectionsHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
		},
		Queries: app.Queries{
			WebServiceQuery: &queries.WebServiceQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			WebServicesDumpQuery: &queries.WebServicesDumpQueryHandler{
				Database: database,
			},
			IconQuery: &queries.IconQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			IconsCollectionQuery: &queries.IconsCollectionQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
			IconRequestQuery: &queries.IconRequestQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
		},
	}

	routesHandler := ports.NewRoutesHandler(cqrs, validate)

	module := &IconsModule{
		Cqrs:          cqrs,
		RoutesHandler: routesHandler,
		Config:        config,
	}

	return module
}

func (m *IconsModule) RegisterRoutes(router *gin.Engine) {
	// internal/admin
	adminRouter := router.Group("/")
	adminRouter.Use(httpsec.IPWhitelistMiddleware(m.Config.Security))

	adminRouter.POST("/mobile/web_services", m.RoutesHandler.CreateWebService)
	adminRouter.PUT("/mobile/web_services/:service_id", m.RoutesHandler.UpdateWebService)
	adminRouter.DELETE("/mobile/web_services/:service_id", m.RoutesHandler.RemoveWebService)

	if m.Config.IsTestingEnv() {
		adminRouter.DELETE("/mobile/web_services", m.RoutesHandler.RemoveAllWebServices)
		adminRouter.DELETE("/mobile/icons", m.RoutesHandler.RemoveAllIcons)
		adminRouter.DELETE("/mobile/icons/collections", m.RoutesHandler.RemoveAllIconsCollections)
		adminRouter.DELETE("/mobile/icons/requests", m.RoutesHandler.RemoveAllIconsRequests)
	}

	adminRouter.POST("/mobile/icons/collections", m.RoutesHandler.CreateIconsCollection)
	adminRouter.PUT("/mobile/icons/collections/:collection_id", m.RoutesHandler.UpdateIconsCollection)
	adminRouter.DELETE("/mobile/icons/collections/:collection_id", m.RoutesHandler.RemoveIconsCollection)

	adminRouter.POST("/mobile/icons", m.RoutesHandler.CreateIcon)
	adminRouter.PUT("/mobile/icons/:icon_id", m.RoutesHandler.UpdateIcon)
	adminRouter.DELETE("/mobile/icons/:icon_id", m.RoutesHandler.RemoveIcon)

	adminRouter.DELETE("/mobile/icons/requests/:icon_request_id", m.RoutesHandler.RemoveIconRequest)
	adminRouter.POST("/mobile/icons/requests/:icon_request_id/commands/update_web_service", m.RoutesHandler.UpdateWebServiceFromIconRequest)
	adminRouter.POST("/mobile/icons/requests/:icon_request_id/commands/transform_to_web_service", m.RoutesHandler.TransformToWebService)
	adminRouter.GET("/mobile/icons/requests/:icon_request_id", m.RoutesHandler.FindIconRequest)
	adminRouter.GET("/mobile/icons/requests", m.RoutesHandler.FindAllIconsRequests)

	// public
	publicRouter := router.Group("/")

	publicRouter.GET("/mobile/web_services/:service_id", m.RoutesHandler.FindWebService)
	publicRouter.GET("/mobile/web_services", m.RoutesHandler.FindAllWebServices)

	publicRouter.GET("/mobile/web_services/dump", m.RoutesHandler.DumpWebServices)

	publicRouter.GET("/mobile/icons/:icon_id", m.RoutesHandler.FindIcon)
	publicRouter.GET("/mobile/icons", m.RoutesHandler.FindAllIcons)

	publicRouter.GET("/mobile/icons/collections/:collection_id", m.RoutesHandler.FindIconsCollection)
	publicRouter.GET("/mobile/icons/collections", m.RoutesHandler.FindAllIconsCollection)

	publicRouter.POST("/mobile/icons/requests", m.RoutesHandler.CreateIconRequest)
}
