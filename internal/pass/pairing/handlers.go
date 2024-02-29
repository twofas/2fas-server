package pairing

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/pass/connection"
)

func ExtensionConfigureHandler(pairingApp *Pairing) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		var req ConfigureBrowserExtensionRequest
		if err := gCtx.BindJSON(&req); err != nil {
			gCtx.String(http.StatusBadRequest, err.Error())
			return
		}
		if _, err := uuid.Parse(req.ExtensionID); err != nil {
			gCtx.String(http.StatusBadRequest, "extension_id is not valid uuid")
			return
		}

		resp, err := pairingApp.ConfigureBrowserExtension(gCtx, req)
		if err != nil {
			logging.Errorf("Failed to configure: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
		gCtx.JSON(http.StatusCreated, resp)
	}
}

func ExtensionWaitForConnWSHandler(pairingApp *Pairing) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := connection.TokenFromWSProtocol(gCtx.Request)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}

		extensionID, err := pairingApp.VerifyPairingToken(gCtx, token)
		if err != nil {
			logging.Errorf("Failed to verify pairing token: %v", err)
			gCtx.Status(http.StatusUnauthorized)
			return
		}

		if err := pairingApp.ServePairingWS(gCtx.Writer, gCtx.Request, extensionID); err != nil {
			logging.Errorf("Failed serve ws: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
	}
}

func ExtensionProxyWSHandler(pairingApp *Pairing, proxyApp *connection.Proxy) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := connection.TokenFromWSProtocol(gCtx.Request)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}
		extensionID, err := pairingApp.VerifyExtProxyToken(gCtx, token)
		if err != nil {
			logging.Errorf("Failed to verify proxy token: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
		pairingInfo, err := pairingApp.GetPairingInfo(gCtx, extensionID)
		if err != nil {
			logging.Errorf("Failed to get pairing info: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
		if !pairingInfo.IsPaired() {
			gCtx.String(http.StatusForbidden, "Pairing is not yet done")
			return
		}
		if err := proxyApp.ServeExtensionProxyToMobileWS(gCtx.Writer, gCtx.Request, pairingInfo.Device.DeviceID); err != nil {
			logging.Errorf("Failed to serve ws: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
	}
}

func MobileConfirmHandler(pairingApp *Pairing) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := tokenFromRequest(gCtx)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}
		extensionID, err := pairingApp.VerifyConnectionToken(gCtx, token)
		if err != nil {
			logging.Errorf("Failed to verify connection token: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
		var req ConfirmPairingRequest
		if err := gCtx.BindJSON(&req); err != nil {
			gCtx.String(http.StatusBadRequest, err.Error())
			return
		}

		if _, err := uuid.Parse(req.DeviceID); err != nil {
			gCtx.String(http.StatusBadRequest, "extension_id is not valid uuid")
			return
		}

		resp, err := pairingApp.ConfirmPairing(gCtx, req, extensionID)
		if err != nil {
			logging.Errorf("Failed to ConfirmPairing: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
		gCtx.JSON(http.StatusOK, resp)
	}
}

func MobileProxyWSHandler(pairingApp *Pairing, proxy *connection.Proxy) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := connection.TokenFromWSProtocol(gCtx.Request)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}
		extensionID, err := pairingApp.VerifyMobileProxyToken(gCtx, token)
		if err != nil {
			logging.Errorf("Failed to verify connection token: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
		log := logging.WithField("extension_id", extensionID)
		pairingInfo, err := pairingApp.GetPairingInfo(gCtx, extensionID)
		if err != nil {
			log.Errorf("Failed to get pairing info: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
		if !pairingInfo.IsPaired() {
			gCtx.String(http.StatusForbidden, "Pairing is not yet done")
			return
		}
		if err := proxy.ServeMobileProxyToExtensionWS(gCtx.Writer, gCtx.Request, pairingInfo.Device.DeviceID); err != nil {
			log.Errorf("Failed to serve ws: %w", err)
			gCtx.Status(http.StatusInternalServerError)
			return
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
