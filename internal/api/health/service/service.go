package service

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/api/health/ports"
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

func (m *HealthModule) RegisterPublicRoutes(router *gin.Engine) {
	router.GET("/health", m.RoutesHandler.CheckApplicationHealth)
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
