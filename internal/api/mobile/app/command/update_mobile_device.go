package command

import (
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/mobile/domain"
)

type UpdateMobileDevice struct {
	Id       string `uri:"device_id" validate:"required,uuid4"`
	Name     string `json:"name" validate:"max=128"`
	Platform string `json:"platform" validate:"omitempty,oneof=ios android huawei"`
	FcmToken string `json:"fcm_token" validate:"max=256"`
}

type UpdateMobileDeviceHandler struct {
	Repository domain.MobileDeviceRepository
}

func (h *UpdateMobileDeviceHandler) Handle(cmd *UpdateMobileDevice) error {
	deviceId, _ := uuid.Parse(cmd.Id)

	mobileDevice, err := h.Repository.FindById(deviceId)

	if err != nil {
		return err
	}

	if cmd.Name != "" {
		mobileDevice.Name = cmd.Name
	}

	if cmd.FcmToken != "" {
		mobileDevice.FcmToken = cmd.FcmToken
	}

	return h.Repository.Update(mobileDevice)
}
