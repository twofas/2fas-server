package browser_extension

import (
	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/internal/websocket/common"
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
	router.GET("/browser_extensions/:extension_id", routes.connectionHandler.Handler())
	router.GET("/browser_extensions/:extension_id/2fa_requests/:token_request_id", routes.connectionHandler.Handler())

	router.GET("/health", func(c *gin.Context) {
		c.String(200, "")
	})
}
