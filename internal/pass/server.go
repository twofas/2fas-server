package pass

import (
	"github.com/gin-gonic/gin"
	httphelpers "github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/recovery"
	"github.com/twofas/2fas-server/internal/pass/pairing"
)

type Server struct {
	router *gin.Engine
	addr   string
}

func NewServer(addr string) *Server {
	pairingApp := pairing.NewPairingApp()
	proxyApp := pairing.NewProxy()

	router := gin.New()
	router.Use(recovery.RecoveryMiddleware())
	router.Use(httphelpers.RequestIdMiddleware())
	router.Use(httphelpers.CorrelationIdMiddleware())
	// TODO: don't log auth headers.
	router.Use(httphelpers.RequestJsonLogger())

	router.GET("/health", func(context *gin.Context) {
		context.Status(200)
	})

	router.POST("/browser_extension/configure", pairing.ExtensionConfigureHandler(pairingApp))
	router.GET("/browser_extension/wait_for_connection", pairing.ExtensionWaitForConnWSHandler(pairingApp))
	router.GET("/browser_extension/proxy_to_mobile", pairing.ExtensionProxyWSHandler(pairingApp, proxyApp))
	router.POST("/mobile/confirm", pairing.MobileConfirmHandler(pairingApp))
	router.GET("/mobile/proxy_to_browser_extension", pairing.MobileProxyWSHandler(pairingApp, proxyApp))

	return &Server{
		router: router,
		addr:   addr,
	}
}

func (s *Server) Run() error {
	return s.router.Run(s.addr)
}
