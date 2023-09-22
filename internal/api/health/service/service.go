package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/api/health/ports"
	"github.com/twofas/2fas-server/internal/common/http"
)

type HealthModule struct {
	RoutesHandler *ports.RoutesHandler
	Config        config.Configuration
}

func NewHealthModule(applicationName string, config config.Configuration, redis *redis.Client) *HealthModule {
	routesHandler := ports.NewRoutesHandler(applicationName, redis)

	return &HealthModule{
		RoutesHandler: routesHandler,
		Config:        config,
	}
}

func (m *HealthModule) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", m.RoutesHandler.CheckApplicationHealth)

	internalFor2FasUsersOnly := router.Group("/")
	internalFor2FasUsersOnly.Use(http.IPWhitelistMiddleware(m.Config.Security))

	internalFor2FasUsersOnly.GET("/system/redis/info", m.RoutesHandler.RedisInfo)
	internalFor2FasUsersOnly.GET("/system/info", m.RoutesHandler.GetApplicationConfiguration)
	internalFor2FasUsersOnly.GET("/system/fake_error", m.RoutesHandler.FakeError)
	internalFor2FasUsersOnly.GET("/system/fake_warning", m.RoutesHandler.FakeWarning)
	internalFor2FasUsersOnly.GET("/system/fake_security_warning", m.RoutesHandler.FakeSecurityWarning)
}

func (m *HealthModule) RegisterHealth(router *gin.Engine) {
	router.GET("/health", m.RoutesHandler.CheckApplicationHealth)
}

func (m *HealthModule) RegisterAdminRoutes(g *gin.RouterGroup) {
	g.GET("/health", m.RoutesHandler.CheckApplicationHealth)
	g.GET("/system/redis/info", m.RoutesHandler.RedisInfo)
	g.GET("/system/info", m.RoutesHandler.GetApplicationConfiguration)
	g.GET("/system/fake_error", m.RoutesHandler.FakeError)
	g.GET("/system/fake_warning", m.RoutesHandler.FakeWarning)
	g.GET("/system/fake_security_warning", m.RoutesHandler.FakeSecurityWarning)
}
