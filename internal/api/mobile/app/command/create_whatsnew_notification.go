package command

import (
	"github.com/2fas/api/internal/api/mobile/domain"
	"github.com/google/uuid"
)

type CreateNotification struct {
	Id       uuid.UUID
	Icon     string `json:"icon" validate:"required,oneof=updates news features youtube"`
	Link     string `json:"link" validate:"required,max=128"`
	Message  string `json:"message" validate:"required,max=256"`
	Platform string `json:"platform" validate:"required,oneof=ios android huawei"`
	Version  string `json:"version" validate:"omitempty,max=12"`
}

type CreateNotificationHandler struct {
	Repository domain.MobileNotificationsRepository
}

func (h *CreateNotificationHandler) Handle(cmd *CreateNotification) error {
	mobileNotification := &domain.MobileNotification{
		Id:       cmd.Id,
		Icon:     cmd.Icon,
		Link:     cmd.Link,
		Message:  cmd.Message,
		Platform: cmd.Platform,
		Version:  cmd.Version,
	}

	return h.Repository.Save(mobileNotification)
}
