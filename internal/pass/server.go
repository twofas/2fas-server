package pass

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/internal/pass/sign"

	"github.com/twofas/2fas-server/config"
	httphelpers "github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/recovery"
	"github.com/twofas/2fas-server/internal/pass/pairing"
)

type Server struct {
	router *gin.Engine
	addr   string
}

func NewServer(cfg config.PassConfig) *Server {
	var awsEndpoint *string
	if cfg.AWSEndpoint != "" {
		awsEndpoint = aws.String(cfg.AWSEndpoint)
	}
	region := cfg.AWSRegion
	if region == "" {
		region = "us-east-1"
	}
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         awsEndpoint,
	})
	if err != nil {
		log.Fatal(err)
	}
	kmsClient := kms.New(sess)

	signSvc, err := sign.NewService(cfg.KMSKeyID, kmsClient)
	if err != nil {
		log.Fatal(err)
	}

	pairingApp := pairing.NewPairingApp(signSvc)
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
		addr:   cfg.Addr,
	}
}

func (s *Server) Run() error {
	return s.router.Run(s.addr)
}
