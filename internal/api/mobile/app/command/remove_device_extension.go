package command

import (
	"github.com/2fas/api/internal/api/mobile/domain"
	"github.com/google/uuid"
)

type RemoveDevicePairedExtension struct {
	DeviceId    string `uri:"device_id" validate:"required,uuid4"`
	ExtensionId string `uri:"extension_id" validate:"required,uuid4"`
}

type RemoveDeviceExtensionHandler struct {
	MobileDeviceExtensionsRepository domain.MobileDeviceExtensionsRepository
}

func (h *RemoveDeviceExtensionHandler) Handle(cmd *RemoveDevicePairedExtension) error {
	deviceId, _ := uuid.Parse(cmd.DeviceId)
	extId, _ := uuid.Parse(cmd.ExtensionId)

	extension, err := h.MobileDeviceExtensionsRepository.FindById(deviceId, extId)

	if err != nil {
		return err
	}

	return h.MobileDeviceExtensionsRepository.Delete(extension)
}
