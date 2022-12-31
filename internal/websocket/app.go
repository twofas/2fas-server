package websocket

import (
	"github.com/2fas/api/internal/common/http"
	"github.com/2fas/api/internal/websocket/browser_extension"
	"github.com/2fas/api/internal/websocket/common"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router            *gin.Engine
	addr              string
	connectionHandler *common.ConnectionHandler
}

func NewServer(addr string) *Server {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(http.RequestIdMiddleware())
	router.Use(http.CorrelationIdMiddleware())
	router.Use(http.RequestJsonLogger())

	connectionHandler := common.NewConnectionHandler()

	routesHandler := browser_extension.NewRoutesHandler(connectionHandler)

	browser_extension.GinRoutesHandler(routesHandler, router)

	return &Server{
		router:            router,
		addr:              addr,
		connectionHandler: connectionHandler,
	}
}

func (s *Server) RunWebsocketServer() {
	err := s.router.Run(s.addr)

	if err != nil {
		panic(err)
	}
}
