package query

import (
	"github.com/doug-martin/goqu/v9"
	"gorm.io/gorm"
)

type DeviceBrowserExtensionsQuery struct {
	DeviceId    string `uri:"device_id" validate:"required,uuid4"`
	ExtensionId string `uri:"extension_id" validate:"omitempty,uuid4"`
}

type DeviceBrowserExtensionPresenter struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	BrowserName    string `json:"browser_name"`
	BrowserVersion string `json:"browser_version"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	PairedAt       string `json:"paired_at"`
}

type DeviceBrowserExtensionsQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *DeviceBrowserExtensionsQueryHandler) Handle(query *DeviceBrowserExtensionsQuery) ([]*DeviceBrowserExtensionPresenter, error) {
	var presenter []*DeviceBrowserExtensionPresenter

	relationTable := goqu.T("mobile_device_browser_extension")
	extensionsTable := goqu.T("browser_extensions")

	ds := h.Qb.From(relationTable)

	ds = ds.Select(
		extensionsTable.Col("id").As("id"),
		extensionsTable.Col("name").As("name"),
		extensionsTable.Col("browser_name").As("browser_name"),
		extensionsTable.Col("browser_version").As("browser_version"),
		extensionsTable.Col("created_at").As("created_at"),
		extensionsTable.Col("updated_at").As("updated_at"),
		relationTable.Col("created_at").As("paired_at")).
		LeftJoin(extensionsTable, goqu.On(relationTable.Col("extension_id").Eq(extensionsTable.Col("id")))).
		Where(relationTable.Col("device_id").Eq(query.DeviceId))

	if query.ExtensionId != "" {
		ds = ds.Where(extensionsTable.Col("id").Eq(query.ExtensionId))
	}

	sql, _, _ := ds.ToSQL()

	result := h.Database.Raw(sql).Find(&presenter)

	if result.Error != nil {
		return nil, result.Error
	}

	return presenter, nil
}
