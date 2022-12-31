package query

import (
	"github.com/2fas/api/internal/api/mobile/adapters"
	"github.com/doug-martin/goqu/v9"
	"gorm.io/gorm"
)

type MobileNotificationPresenter struct {
	Id          string `json:"id"`
	Icon        string `json:"icon"`
	Link        string `json:"link"`
	Message     string `json:"message"`
	PublishedAt string `json:"published_at"`
	Push        bool   `json:"push"`
	Platform    string `json:"platform"`
	Version     string `json:"version"`
	CreatedAt   string `json:"created_at"`
}

type MobileNotificationsQuery struct {
	Id             string `uri:"notification_id" validate:"omitempty,uuid4"`
	Platform       string `form:"platform" validate:"omitempty,oneof=android ios huawei"`
	Version        string `form:"version" validate:"omitempty"`
	PublishedAfter string `form:"published_after" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type MobileNotificationsQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *MobileNotificationsQueryHandler) FindOne(query *MobileNotificationsQuery) (*MobileNotificationPresenter, error) {
	ds := h.Qb.From("mobile_notifications").Where(goqu.And(
		goqu.C("id").Eq(query.Id),
		goqu.C("deleted_at").IsNull(),
	))

	sql, _, _ := ds.ToSQL()

	presenter := &MobileNotificationPresenter{}

	result := h.Database.Raw(sql).First(&presenter)

	if result.Error != nil {
		return nil, adapters.MobileNotificationCouldNotBeFound{NotificationId: query.Id}
	}

	return presenter, nil
}

func (h *MobileNotificationsQueryHandler) FindAll(query *MobileNotificationsQuery) ([]*MobileNotificationPresenter, error) {
	var presenter []*MobileNotificationPresenter

	ds := h.Qb.From("mobile_notifications").Where(goqu.And(
		goqu.C("deleted_at").IsNull(),
	))

	if query.Platform != "" {
		ds = ds.Where(goqu.C("platform").Eq(query.Platform))
	}

	if query.Version != "" {
		ds = ds.Where(goqu.C("version").Eq(query.Version))
	}

	if query.PublishedAfter != "" {
		ds = ds.Where(goqu.C("published_at").Gte(query.PublishedAfter))
	}

	sql, _, _ := ds.ToSQL()

	h.Database.Raw(sql).Find(&presenter)

	return presenter, nil
}
