package query

import (
	"time"

	"github.com/doug-martin/goqu/v9"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/common/clock"
)

type DeviceBrowserExtension2FaRequestPresenter struct {
	Id          string `json:"token_request_id"`
	ExtensionId string `json:"extension_id"`
	Domain      string `json:"domain"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type DeviceBrowserExtension2FaRequestQuery struct {
	DeviceId string `uri:"device_id" validate:"required,uuid4"`
}

type DeviceBrowserExtension2FaRequestQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
	Clock    clock.Clock
}

func (h *DeviceBrowserExtension2FaRequestQueryHandler) Handle(query *DeviceBrowserExtension2FaRequestQuery) ([]*DeviceBrowserExtension2FaRequestPresenter, error) {
	var presenter []*DeviceBrowserExtension2FaRequestPresenter

	sourceT := goqu.T("browser_extensions_2fa_requests")
	joinT := goqu.T("mobile_device_browser_extension")

	ds := h.Qb.From(sourceT)
	ds = ds.LeftJoin(joinT, goqu.On(joinT.Col("extension_id").Eq(sourceT.Col("extension_id"))))
	ds = ds.Where(
		joinT.Col("device_id").Eq(query.DeviceId),
		sourceT.Col("status").Eq("pending"),
		sourceT.Col("created_at").Gte(h.Clock.Now().Add(-time.Minute+2)))

	sql, _, _ := ds.ToSQL()

	result := h.Database.Raw(sql).Find(&presenter)

	if result.Error != nil {
		return nil, result.Error
	}

	return presenter, nil
}
