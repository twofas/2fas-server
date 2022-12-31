package command

import (
	"fmt"
	"github.com/2fas/api/internal/api/browser_extension/domain"
	"github.com/2fas/api/internal/api/mobile/adapters"
	"github.com/2fas/api/internal/common/logging"
	"github.com/2fas/api/internal/common/websocket"
	"github.com/google/uuid"
)

type Send2FaTokenWebsocketMessage struct {
	Event          string `json:"event"`
	ExtensionId    string `json:"extension_id"`
	TokenRequestId string `json:"token_request_id"`
	DeviceId       string `json:"device_id"`
	Token          string `json:"token"`
}

func NewSend2FaTokenWebsocketMessage(extensionId, deviceId, token, tokenRequestId string) Send2FaTokenWebsocketMessage {
	return Send2FaTokenWebsocketMessage{
		Event:          "browser_extensions.device.2fa_response",
		ExtensionId:    extensionId,
		DeviceId:       deviceId,
		TokenRequestId: tokenRequestId,
		Token:          token,
	}
}

type Send2FaToken struct {
	DeviceId       string `uri:"device_id" validate:"required,uuid4"`
	ExtensionId    string `json:"extension_id" validate:"required,uuid4"`
	TokenRequestId string `json:"token_request_id" validate:"required,uuid4"`
	Token          string `json:"token" validate:"required,lte=768"`
}

type Send2FaTokenHandler struct {
	BrowserExtensionsRepository        domain.BrowserExtensionRepository
	MobileApplicationExtensionsService *adapters.DeviceExtensionsService
	WebsocketClient                    *websocket.WebsocketApiClient
}

func (h *Send2FaTokenHandler) Handle(cmd *Send2FaToken) error {
	extId, _ := uuid.Parse(cmd.ExtensionId)

	logging.WithFields(logging.Fields{
		"browser_extension_id": cmd.ExtensionId,
		"device_id":            cmd.DeviceId,
		"token":                cmd.Token,
		"token_request_id":     cmd.TokenRequestId,
	}).Info("Start command `Send2FaToken`")

	browserExtension, err := h.BrowserExtensionsRepository.FindById(extId)

	if err != nil {
		return err
	}

	message := NewSend2FaTokenWebsocketMessage(cmd.ExtensionId, cmd.DeviceId, cmd.Token, cmd.TokenRequestId)

	uri := fmt.Sprintf("browser_extensions/%s/2fa_requests/%s", browserExtension.Id.String(), cmd.TokenRequestId)
	err = h.WebsocketClient.SendMessage(uri, message)

	if err != nil {
		logging.WithFields(logging.Fields{
			"error":   err.Error(),
			"message": message,
		}).Error("Cannot send websocket message")
	}

	return err
}
