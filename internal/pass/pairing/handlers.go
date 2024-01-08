package pairing

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/common/logging"
)

func BrowserExtensionConfigureHandler(pairingApp *Pairing) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		var req ConfigureBrowserExtensionRequest
		if err := gCtx.BindJSON(&req); err != nil {
			gCtx.String(http.StatusBadRequest, "invalid request")
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

func BrowserExtensionWaitForConnHandler(pairingApp *Pairing) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		// TODO: consider moving auth to middleware.
		token, err := tokenFromRequest(gCtx)
		if err != nil {
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

func BrowserExtensionProxyHandler(pairingApp *Pairing, proxyApp *Proxy) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		// TODO: consider moving auth to middleware.
		token, err := tokenFromRequest(gCtx)
		if err != nil {
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
			logging.Info("Pairing is not yet done")
			gCtx.Status(http.StatusForbidden)
			return
		}
		proxyApp.ServeExtensionProxyToMobileWS(gCtx.Writer, gCtx.Request, extensionID, pairingInfo.Device.DeviceID)
	}
}

func MobileConfirmHandler(pairingApp *Pairing) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		// TODO: consider moving auth to middleware.
		token, err := tokenFromRequest(gCtx)
		if err != nil {
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
			gCtx.String(http.StatusBadRequest, "invalid request")
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

func MobileProxyHandler(pairingApp *Pairing, proxyApp *Proxy) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		// TODO: consider moving auth to middleware.
		token, err := tokenFromRequest(gCtx)
		if err != nil {
			gCtx.Status(http.StatusForbidden)
			return
		}
		extensionID, err := pairingApp.VerifyConnectionToken(gCtx, token)
		if err != nil {
			logging.Errorf("Failed to verify connection token: %v", err)
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
			logging.Info("Pairing is not yet done")
			gCtx.Status(http.StatusForbidden)
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
