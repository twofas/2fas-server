package browser_extension

import (
	"github.com/2fas/api/internal/websocket/common"
	"github.com/gin-gonic/gin"
)

type RoutesHandler struct {
	connectionHandler *common.ConnectionHandler
}

func NewRoutesHandler(handler *common.ConnectionHandler) *RoutesHandler {
	return &RoutesHandler{
		connectionHandler: handler,
	}
}

func GinRoutesHandler(routes *RoutesHandler, router *gin.Engine) {
	connectionHandler := routes.connectionHandler.Handle()

	router.GET("/browser_extensions/:extension_id", connectionHandler)
	router.GET("/browser_extensions/:extension_id/2fa_requests/:token_request_id", connectionHandler)

	router.GET("/health", func(c *gin.Context) {
		c.String(200, "")
	})
}
