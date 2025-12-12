package command

import (
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/api/mobile/domain"
)

type RegisterMobileDevice struct {
	Id       uuid.UUID
	Name     string `json:"name" validate:"not_blank,max=128"`
	Platform string `json:"platform" validate:"required,oneof=ios android huawei"`
	FcmToken string `json:"fcm_token" validate:"max=256"`
}

type RegisterMobileDeviceHandler struct {
	Repository domain.MobileDeviceRepository
}

func (h *RegisterMobileDeviceHandler) Handle(cmd *RegisterMobileDevice) error {
	mobileDevice := domain.NewMobileDevice(cmd.Id, cmd.Name, cmd.Platform, cmd.FcmToken)

	return h.Repository.Save(mobileDevice)
}
