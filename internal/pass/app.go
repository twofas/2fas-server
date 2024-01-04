package pass

import (
	"github.com/gin-gonic/gin"

	"github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/recovery"
)

type Server struct {
	router *gin.Engine
	addr   string
}

func NewServer(addr string) *Server {
	router := gin.New()

	router.Use(recovery.RecoveryMiddleware())
	router.Use(http.RequestIdMiddleware())
	router.Use(http.CorrelationIdMiddleware())
	router.Use(http.RequestJsonLogger())

	router.GET("/health", func(context *gin.Context) {
		context.Status(200)
	})

	return &Server{
		router: router,
		addr:   addr,
	}
}

func (s *Server) Run() error {
	return s.router.Run(s.addr)
}
