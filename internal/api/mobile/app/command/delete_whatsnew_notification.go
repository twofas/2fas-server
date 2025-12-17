package command

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/mobile/domain"
)

type DeleteNotification struct {
	Id string `uri:"notification_id" validate:"required,uuid4"`
}

type DeleteNotificationHandler struct {
	Repository domain.MobileNotificationsRepository
}

func (h *DeleteNotificationHandler) Handle(cmd *DeleteNotification) error {
	id, _ := uuid.Parse(cmd.Id)

	mobileNotification, err := h.Repository.FindById(id)

	if err != nil {
		return err
	}

	return h.Repository.Delete(mobileNotification)
}

type DeleteAllNotifications struct{}

type DeleteAllNotificationsHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *DeleteAllNotificationsHandler) Handle(cmd *DeleteAllNotifications) {
	sql, _, _ := h.Qb.Truncate("mobile_notifications").ToSQL()

	h.Database.Exec(sql)
}
