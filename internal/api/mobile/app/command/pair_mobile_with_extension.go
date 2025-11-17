package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
	"github.com/twofas/2fas-server/internal/api/mobile/adapters"
	mobile_domain "github.com/twofas/2fas-server/internal/api/mobile/domain"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/websocket"
)

type BrowserExtensionHasBeenPairedWithDevice struct {
	Event              string `json:"event"`
	BrowserExtensionId string `json:"browser_extension_id"`
	DeviceId           string `json:"device_id"`
	DevicePublicKey    string `json:"device_public_key"`
}

type BrowserExtensionHasNotBeenPairedWithDevice struct {
	Event              string `json:"event"`
	BrowserExtensionId string `json:"browser_extension_id"`
	DeviceId           string `json:"device_id"`
	Reason             string `json:"reason"`
}

func NewBrowserExtensionHasBeenPairedWithDevice(deviceId, devicePublicKey string, extId uuid.UUID) *BrowserExtensionHasBeenPairedWithDevice {
	return &BrowserExtensionHasBeenPairedWithDevice{
		Event:              "browser_extensions.pairing.success",
		BrowserExtensionId: extId.String(),
		DeviceId:           deviceId,
		DevicePublicKey:    devicePublicKey,
	}
}

func NewBrowserExtensionHasNotBeenPairedWithDevice(err error, deviceId string, extId uuid.UUID) *BrowserExtensionHasNotBeenPairedWithDevice {
	return &BrowserExtensionHasNotBeenPairedWithDevice{
		Event:              "browser_extensions.pairing.failure",
		BrowserExtensionId: extId.String(),
		DeviceId:           deviceId,
		Reason:             err.Error(),
	}
}

type PairMobileWithBrowserExtension struct {
	DeviceId        string `uri:"device_id" validate:"required"`
	ExtensionId     string `json:"extension_id" validate:"required,uuid4"`
	DevicePublicKey string `json:"device_public_key" validate:"required,lte=768"`
}

type PairMobileWithExtensionHandler struct {
	BrowserExtensionsRepository        domain.BrowserExtensionRepository
	MobileDeviceExtensionsRepository   mobile_domain.MobileDeviceExtensionsRepository
	MobileApplicationExtensionsService *adapters.DeviceExtensionsService
	WebsocketClient                    *websocket.WebsocketApiClient
}

func (h *PairMobileWithExtensionHandler) Handle(ctx context.Context, cmd *PairMobileWithBrowserExtension) error {
	log := logging.FromContext(ctx)

	deviceId, _ := uuid.Parse(cmd.DeviceId)
	extensionId, _ := uuid.Parse(cmd.ExtensionId)

	browserExtension, err := h.BrowserExtensionsRepository.FindById(extensionId)

	if err != nil {
		return err
	}

	mobileDeviceExtension, _ := h.MobileDeviceExtensionsRepository.FindById(deviceId, extensionId)

	if mobileDeviceExtension != nil {
		return mobile_domain.ExtensionHasAlreadyBeenPairedError{ExtensionId: extensionId.String()}
	}

	err = h.MobileApplicationExtensionsService.PairDeviceWithBrowserExtension(cmd.DeviceId, extensionId)

	websocketUri := fmt.Sprintf("browser_extensions/%s", browserExtension.Id.String())

	if err != nil {
		message := NewBrowserExtensionHasNotBeenPairedWithDevice(err, cmd.DeviceId, extensionId)

		sendErrr := h.WebsocketClient.SendMessage(websocketUri, message)
		if sendErrr != nil {
			log.Error("Failed to send browser extension hasn't been paired with the device: %v", sendErrr)
		}

		return err
	}

	message := NewBrowserExtensionHasBeenPairedWithDevice(cmd.DeviceId, cmd.DevicePublicKey, extensionId)

	return h.WebsocketClient.SendMessage(websocketUri, message)
}
