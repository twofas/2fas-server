package command

import (
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/mobile/domain"
)

type UpdateNotification struct {
	Id       string `uri:"notification_id" validate:"required,uuid4"`
	Icon     string `json:"icon" validate:"required,max=128"`
	Link     string `json:"link" validate:"required,max=128"`
	Message  string `json:"message" validate:"required,max=256"`
	Platform string `json:"platform" validate:"required,oneof=ios android huawei"`
	Version  string `json:"version" validate:"omitempty,max=12"`
}

type UpdateNotificationHandler struct {
	Repository domain.MobileNotificationsRepository
}

func (h *UpdateNotificationHandler) Handle(cmd *UpdateNotification) error {
	id, _ := uuid.Parse(cmd.Id)

	mobileNotification, err := h.Repository.FindById(id)

	if err != nil {
		return err
	}

	if cmd.Icon != "" {
		mobileNotification.Icon = cmd.Icon
	}

	if cmd.Link != "" {
		mobileNotification.Link = cmd.Link
	}

	if cmd.Message != "" {
		mobileNotification.Message = cmd.Message
	}

	if cmd.Platform != "" {
		mobileNotification.Platform = cmd.Platform
	}

	if cmd.Version != "" {
		mobileNotification.Version = cmd.Version
	}

	return h.Repository.Update(mobileNotification)
}
