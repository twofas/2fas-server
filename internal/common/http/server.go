package http

import (
	"github.com/gin-gonic/gin"
)

func RunHttpServer(addr string, init func(engine *gin.Engine)) {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(LoggingMiddleware())
	router.Use(RequestJsonLogger())

	init(router)

	err := router.Run(addr)

	if err != nil {
		panic(err)
	}
}
