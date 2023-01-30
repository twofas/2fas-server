package command

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/mobile/domain"
	"time"
)

type PublishNotification struct {
	Id string `uri:"notification_id" validate:"required,uuid4"`
}

type PublishNotificationHandler struct {
	Repository domain.MobileNotificationsRepository
}

func (h *PublishNotificationHandler) Handle(cmd *PublishNotification) error {
	id, _ := uuid.Parse(cmd.Id)

	mobileNotification, err := h.Repository.FindById(id)

	if err != nil {
		return err
	}

	mobileNotification.PublishedAt = sql.NullTime{Time: time.Now(), Valid: true}

	return h.Repository.Update(mobileNotification)
}
