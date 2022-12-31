package query

import (
	"github.com/doug-martin/goqu/v9"
	"gorm.io/gorm"
)

type BrowserExtensionPairedDevicesQuery struct {
	ExtensionId string `uri:"extension_id" validate:"required,uuid4"`
}

type BrowserExtensionPairedDeviceQuery struct {
	ExtensionId string `uri:"extension_id" validate:"required,uuid4"`
	DeviceId    string `uri:"device_id" validate:"required,uuid4"`
}

type BrowserPairedDevicePresenter struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	CreatedAt string `json:"paired_at"`
}

type BrowserExtensionPairedMobileDevicesQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *BrowserExtensionPairedMobileDevicesQueryHandler) Handle(query *BrowserExtensionPairedDevicesQuery) []BrowserPairedDevicePresenter {
	var presenter []BrowserPairedDevicePresenter

	relationTable := goqu.T("mobile_device_browser_extension")
	devicesTable := goqu.T("mobile_devices")

	sql, _, _ := h.Qb.From(relationTable).
		Select(
			devicesTable.Col("id").As("id"),
			devicesTable.Col("name").As("name"),
			devicesTable.Col("platform").As("platform"),
			relationTable.Col("created_at").As("created_at")).
		LeftJoin(devicesTable, goqu.On(relationTable.Col("device_id").Eq(devicesTable.Col("id")))).
		Where(relationTable.Col("extension_id").Eq(query.ExtensionId)).
		ToSQL()

	h.Database.Raw(sql).Find(&presenter)

	return presenter
}

type BrowserExtensionPairedMobileDeviceQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *BrowserExtensionPairedMobileDeviceQueryHandler) Handle(query *BrowserExtensionPairedDeviceQuery) (*BrowserPairedDevicePresenter, error) {
	var presenter *BrowserPairedDevicePresenter

	relationTable := goqu.T("mobile_device_browser_extension")
	devicesTable := goqu.T("mobile_devices")

	sql, _, _ := h.Qb.From(relationTable).
		Select(
			devicesTable.Col("id").As("id"),
			devicesTable.Col("name").As("name"),
			devicesTable.Col("platform").As("platform"),
			relationTable.Col("created_at").As("created_at")).
		LeftJoin(devicesTable, goqu.On(relationTable.Col("device_id").Eq(devicesTable.Col("id")))).
		Where(
			relationTable.Col("extension_id").Eq(query.ExtensionId),
			relationTable.Col("device_id").Eq(query.DeviceId),
		).
		ToSQL()

	result := h.Database.Raw(sql).First(&presenter)

	if result.Error != nil {
		return nil, result.Error
	}

	return presenter, nil
}
