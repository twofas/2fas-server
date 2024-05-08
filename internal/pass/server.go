package pass

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/gin-gonic/gin"

	"github.com/twofas/2fas-server/config"
	httphelpers "github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/recovery"
	"github.com/twofas/2fas-server/internal/pass/connection"
	"github.com/twofas/2fas-server/internal/pass/fcm"
	"github.com/twofas/2fas-server/internal/pass/pairing"
	"github.com/twofas/2fas-server/internal/pass/sign"
	"github.com/twofas/2fas-server/internal/pass/sync"
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
	sess, err := session.NewSessionWithOptions(
		session.Options{
			Config: aws.Config{
				Region:           aws.String(region),
				S3ForcePathStyle: aws.Bool(true),
				Endpoint:         awsEndpoint,
			},
			SharedConfigState: session.SharedConfigEnable,
		})
	if err != nil {
		log.Fatal(err)
	}
	kmsClient := kms.New(sess)

	signSvc, err := sign.NewService(cfg.KMSKeyID, kmsClient)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	var fcmClient fcm.Client
	if cfg.FirebaseServiceAccount != "" {
		fcmClient, err = fcm.NewClient(ctx, cfg.FirebaseServiceAccount)
		if err != nil {
			log.Fatalf("Error initializing Messaging Client: %v", err)
		}
	} else {
		fcmClient = fcm.NewFakePushClient()
	}

	pairingApp := pairing.NewApp(signSvc, cfg.PairingRequestTokenValidityDuration)
	proxyPairingApp := connection.NewProxyServer("device_id")

	syncApp := sync.NewApp(signSvc, fcmClient)
	proxySyncApp := connection.NewProxyServer("fcm_token")

	router := gin.New()
	router.Use(recovery.RecoveryMiddleware())
	router.Use(httphelpers.LoggingMiddleware())
	router.Use(httphelpers.RequestJsonLogger())

	router.GET("/health", func(context *gin.Context) {
		context.Status(200)
	})

	// Deprecated paths start here.
	router.GET("/browser_extension/wait_for_connection", pairing.ExtensionWaitForConnWSHandler(pairingApp))
	router.GET("/browser_extension/proxy_to_mobile", pairing.ExtensionProxyWSHandler(pairingApp, proxyPairingApp))
	router.POST("/mobile/confirm", pairing.MobileConfirmHandler(pairingApp))
	router.GET("/mobile/proxy_to_browser_extension", pairing.MobileProxyWSHandler(pairingApp, proxyPairingApp))
	// Deprecated paths end here.

	router.POST("/browser_extension/configure", pairing.ExtensionConfigureHandler(pairingApp))

	router.GET("/browser_extension/pairing/wait", pairing.ExtensionWaitForConnWSHandler(pairingApp))
	router.GET("/browser_extension/pairing/proxy", pairing.ExtensionProxyWSHandler(pairingApp, proxyPairingApp))
	router.POST("/mobile/pairing/confirm", pairing.MobileConfirmHandler(pairingApp))
	router.GET("/mobile/pairing/proxy", pairing.MobileProxyWSHandler(pairingApp, proxyPairingApp))

	router.POST("/browser_extension/sync/request", sync.ExtensionRequestSync(syncApp))
	router.POST("/browser_extension/sync/push", sync.ExtensionRequestPush(syncApp))
	router.GET("/browser_extension/sync/wait", sync.ExtensionRequestWait(syncApp))

	router.GET("/browser_extension/sync/proxy", sync.ExtensionProxyWSHandler(syncApp, proxySyncApp))
	router.POST("/mobile/sync/confirm", sync.MobileConfirmHandler(syncApp))
	router.GET("/mobile/sync/proxy", sync.MobileProxyWSHandler(syncApp, proxySyncApp))

	return &Server{
		router: router,
		addr:   cfg.Addr,
	}
}

func (s *Server) Run() error {
	return s.router.Run(s.addr)
}
