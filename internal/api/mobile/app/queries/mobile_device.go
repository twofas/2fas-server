package query

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrMobileDeviceNotFound = errors.New("Mobile device can not be found")
)

type MobileDevicePresenter struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type MobileDeviceQuery struct {
	Id string `uri:"id"`
}

type MobileDeviceQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *MobileDeviceQueryHandler) Handle(query *MobileDeviceQuery) (*MobileDevicePresenter, error) {
	sql, _, _ := h.Qb.From("mobile_devices").Where(goqu.Ex{
		"id": query.Id,
	}).ToSQL()

	presenter := &MobileDevicePresenter{}

	result := h.Database.Raw(sql).Scan(&presenter)

	if result.Error != nil {
		return nil, ErrMobileDeviceNotFound
	}

	return presenter, nil
}
