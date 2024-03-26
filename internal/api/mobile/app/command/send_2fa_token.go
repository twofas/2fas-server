package command

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
	"github.com/twofas/2fas-server/internal/api/mobile/adapters"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/websocket"
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

func (h *Send2FaTokenHandler) Handle(ctx context.Context, cmd *Send2FaToken) error {
	extId, _ := uuid.Parse(cmd.ExtensionId)
	log := logging.FromContext(ctx).WithFields(logging.Fields{
		"browser_extension_id": cmd.ExtensionId,
		"device_id":            cmd.DeviceId,
		"token_request_id":     cmd.TokenRequestId,
	})
	log.Info("Start command `Send2FaToken`")

	browserExtension, err := h.BrowserExtensionsRepository.FindById(extId)

	if err != nil {
		return err
	}

	message := NewSend2FaTokenWebsocketMessage(cmd.ExtensionId, cmd.DeviceId, cmd.Token, cmd.TokenRequestId)

	uri := fmt.Sprintf("browser_extensions/%s/2fa_requests/%s", browserExtension.Id.String(), cmd.TokenRequestId)

	err = retry.Do(
		func() error {
			return h.WebsocketClient.SendMessage(uri, message)
		},
		retry.Attempts(5),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		log.WithFields(logging.Fields{
			"error":             err.Error(),
			"websocket_message": message,
		}).Error("Cannot send websocket message")
	}

	return err
}
