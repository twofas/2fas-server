package service

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/api/support/adapters"
	"github.com/twofas/2fas-server/internal/api/support/app"
	"github.com/twofas/2fas-server/internal/api/support/app/command"
	"github.com/twofas/2fas-server/internal/api/support/app/queries"
	"github.com/twofas/2fas-server/internal/api/support/domain"
	"github.com/twofas/2fas-server/internal/api/support/ports"
	"github.com/twofas/2fas-server/internal/common/aws"
	"github.com/twofas/2fas-server/internal/common/clock"
	"github.com/twofas/2fas-server/internal/common/db"
	"github.com/twofas/2fas-server/internal/common/storage"
	"gorm.io/gorm"
)

type SupportModule struct {
	Cqrs          *app.Cqrs
	RoutesHandler *ports.RoutesHandler
	Config        config.Configuration
}

func NewSupportModule(config config.Configuration, gorm *gorm.DB, database *sql.DB, validate *validator.Validate) *SupportModule {
	queryBuilder := db.NewQueryBuilder(database)

	debugLogsConfig := domain.LoadDebugLogsConfig()

	var debugLogsStorage storage.FileSystemStorage

	if config.IsTestingEnv() {
		debugLogsStorage = storage.NewTmpFileSystem()
	} else {
		debugLogsStorage = aws.NewAwsS3(debugLogsConfig.AwsRegion, debugLogsConfig.AwsAccessKeyId, debugLogsConfig.AwsSecretAccessKey)
	}

	debugLogsAuditRepository := adapters.NewDebugLogsAuditMysqlRepository(gorm)

	cqrs := &app.Cqrs{
		Commands: app.Commands{
			CreateDebugLogsAudit: &command.CreateDebugLogsAuditHandler{
				DebugLogsAuditRepository: debugLogsAuditRepository,
				FileSystem:               debugLogsStorage,
				Config:                   debugLogsConfig,
				Clock:                    clock.New(),
			},
			CreateDebugLogsAuditClain: &command.CreateDebugLogsAuditClaimHandler{
				DebugLogsAuditRepository: debugLogsAuditRepository,
				DebugLogsAuditConfig:     debugLogsConfig,
				Clock:                    clock.New(),
			},
			UpdateDebugLogsAudit: &command.UpdateDebugLogsAuditHandler{
				DebugLogsAuditRepository: debugLogsAuditRepository,
			},
			DeleteDebugLogsAudit: &command.DeleteDebugLogsAuditHandler{
				DebugLogsAuditRepository: debugLogsAuditRepository,
			},
			DeleteAllDebugLogsAudit: &command.DeleteAllDebugLogsAuditHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
		},
		Queries: app.Queries{
			DebugLogsAuditQuery: &queries.DebugLogsAuditQueryHandler{
				Database: gorm,
				Qb:       queryBuilder,
			},
		},
	}

	routesHandler := ports.NewRoutesHandler(cqrs, validate)

	module := &SupportModule{
		Cqrs:          cqrs,
		RoutesHandler: routesHandler,
		Config:        config,
	}

	return module
}

func (m *SupportModule) RegisterPublicRoutes(router *gin.Engine) {
	publicRouter := router.Group("/")

	publicRouter.POST("/mobile/support/debug_logs/audit/:audit_id", m.RoutesHandler.CreateDebugLogsAudit)
}

func (m *SupportModule) RegisterAdminRoutes(g *gin.RouterGroup) {
	g.POST("/mobile/support/debug_logs/audit/claim", m.RoutesHandler.CreateDebugLogsAuditClaim)
	g.PUT("/mobile/support/debug_logs/audit/claim/:audit_id", m.RoutesHandler.UpdateDebugLogsAuditClaim)
	g.DELETE("/mobile/support/debug_logs/audit/:audit_id", m.RoutesHandler.DeleteDebugLogsAudit)
	g.GET("/mobile/support/debug_logs/audit/:audit_id", m.RoutesHandler.GetDebugLogsAudit)
	g.GET("/mobile/support/debug_logs/audit", m.RoutesHandler.GetDebugAllLogsAudit)

	if m.Config.IsTestingEnv() {
		g.DELETE("/mobile/support/debug_logs/audit", m.RoutesHandler.DeleteAllDebugLogsAudit)
	}
}
