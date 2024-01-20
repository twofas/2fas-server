package pairing

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/common/logging"
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
		token, err := tokenFromWSProtocol(gCtx.Request)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}

		extensionID, err := pairingApp.VerifyPairingToken(gCtx, token)
		if err != nil {
			logging.Errorf("Failed to verify pairing token: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}

		pairingApp.ServePairingWS(gCtx.Writer, gCtx.Request, extensionID)
	}
}

func ExtensionProxyWSHandler(pairingApp *Pairing, proxyApp *Proxy) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := tokenFromWSProtocol(gCtx.Request)
		if err != nil {
			logging.Errorf("Failed to get token from request: %v", err)
			gCtx.Status(http.StatusForbidden)
			return
		}
		extensionID, err := pairingApp.VerifyProxyToken(gCtx, token)
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
		proxyApp.ServeExtensionProxyToMobileWS(gCtx.Writer, gCtx.Request, extensionID, pairingInfo.Device.DeviceID)
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

		if err := pairingApp.ConfirmPairing(gCtx, req, extensionID); err != nil {
			logging.Errorf("Failed to ConfirmPairing: %v", err)
			gCtx.Status(http.StatusInternalServerError)
			return
		}
	}
}

func MobileProxyWSHandler(pairingApp *Pairing, proxyApp *Proxy) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		token, err := tokenFromWSProtocol(gCtx.Request)
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
		proxyApp.ServeMobileProxyToExtensionWS(gCtx.Writer, gCtx.Request, pairingInfo.Device.DeviceID)
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
