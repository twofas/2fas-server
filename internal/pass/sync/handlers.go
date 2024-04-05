package sync

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/pass/connection"
)

func ExtensionRequestSync(syncingApp *Syncing) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := connection.TokenFromWSProtocol(gCtx.Request)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}
		fcmToken, err := syncingApp.VerifyExtRequestSyncToken(gCtx, token)
		if err != nil {
			logging.Errorf("Failed to verify proxy token: %v", err)
			gCtx.String(http.StatusUnauthorized, "Invalid auth token")
			return
		}

		if err := syncingApp.ServeSyncingRequestWS(gCtx.Writer, gCtx.Request, fcmToken); err != nil {
			logging.Errorf("Failed to verify proxy token: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
	}
}

func ExtensionProxyWSHandler(syncingApp *Syncing, proxy *connection.ProxyServer) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := connection.TokenFromWSProtocol(gCtx.Request)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}
		fcmToken, err := syncingApp.VerifyExtSyncToken(gCtx, token)
		if err != nil {
			logging.Errorf("Failed to verify proxy token: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
		ok := syncingApp.isSyncConfirmed(gCtx, fcmToken)
		if !ok {
			gCtx.String(http.StatusForbidden, "Syncing is not yet done")
			return
		}
		if err := proxy.ServeExtensionProxyToMobileWS(gCtx.Writer, gCtx.Request, fcmToken); err != nil {
			logging.Errorf("Failed to serve ws: %v", err)
			gCtx.Status(http.StatusInternalServerError)
		}
	}
}

func MobileConfirmHandler(syncApp *Syncing) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := tokenFromRequest(gCtx)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}
		fcmToken, err := syncApp.VerifyMobileSyncConfirmToken(gCtx, token)
		if err != nil {
			logging.Errorf("Failed to verify connection token: %v", err)
			gCtx.Status(http.StatusUnauthorized)
			return
		}
		resp, err := syncApp.confirmSync(gCtx, fcmToken)
		if err != nil {
			if errors.Is(err, noSyncRequestErr) {
				gCtx.String(http.StatusBadRequest, "no sync request was created for this token")
				return
			}
			logging.Errorf("Failed to ConfirmPairing: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
		gCtx.JSON(http.StatusOK, resp)
	}
}

func MobileProxyWSHandler(syncingApp *Syncing, proxy *connection.ProxyServer) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := connection.TokenFromWSProtocol(gCtx.Request)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}
		fcmToken, err := syncingApp.VerifyMobileSyncConfirmToken(gCtx, token)
		if err != nil {
			logging.Errorf("Invalid connection token: %v", err)
			gCtx.Status(http.StatusUnauthorized)
			return
		}
		ok := syncingApp.isSyncConfirmed(gCtx, fcmToken)
		if !ok {
			gCtx.String(http.StatusForbidden, "Syncing is not yet done")
			return
		}
		if err := proxy.ServeMobileProxyToExtensionWS(gCtx.Writer, gCtx.Request, fcmToken); err != nil {
			logging.Errorf("Failed to serve ws: %v", err)
			gCtx.Status(http.StatusInternalServerError)
		}
	}
}

func MobileGenerateSyncToken(syncingApp *Syncing) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		fcm := gCtx.Param("fcm")
		if fcm == "" {
			gCtx.Status(http.StatusBadRequest)
			return
		}
		if err := syncingApp.sendMobileToken(fcm, gCtx.Writer); err != nil {
			logging.Errorf("Failed to send mobile token: %v", err)
			gCtx.Status(http.StatusInternalServerError)
		}
	}
}

func tokenFromRequest(gCtx *gin.Context) (string, error) {
	tokenHeader := gCtx.GetHeader("Authorization")
	if tokenHeader == "" {
		return "", errors.New("missing Authorization header")
	}
	splitToken := strings.Split(tokenHeader, "Bearer ")
	if len(splitToken) != 2 {
		gCtx.Status(http.StatusForbidden)
		return "", errors.New("missing 'Bearer: value'")
	}
	return splitToken[1], nil
}
