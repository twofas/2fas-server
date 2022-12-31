package command

import (
	"github.com/2fas/api/internal/api/browser_extension/domain"
	"github.com/google/uuid"
)

type RemoveExtensionPairedDevice struct {
	ExtensionId string `uri:"extension_id" validate:"required,uuid4"`
	DeviceId    string `uri:"device_id" validate:"required,uuid4"`
}

type RemoveExtensionPairedDeviceHandler struct {
	BrowserExtensionRepository              domain.BrowserExtensionRepository
	BrowserExtensionPairedDevicesRepository domain.BrowserExtensionDevicesRepository
}

func (h *RemoveExtensionPairedDeviceHandler) Handle(cmd *RemoveExtensionPairedDevice) error {
	extId, _ := uuid.Parse(cmd.ExtensionId)
	deviceId, _ := uuid.Parse(cmd.DeviceId)

	extension, err := h.BrowserExtensionRepository.FindById(extId)

	if err != nil {
		return err
	}

	pairedDevice, err := h.BrowserExtensionPairedDevicesRepository.GetById(extension.Id, deviceId)

	if err != nil {
		return err
	}

	return h.BrowserExtensionPairedDevicesRepository.Delete(pairedDevice)
}
