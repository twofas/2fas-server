package query

import (
	"github.com/doug-martin/goqu/v9"
	"gorm.io/gorm"
)

type PairedBrowserExtensionPresenter struct {
	Id        string `json:"extension_id"`
	Name      string `json:"extension_name"`
	PublicKey string `json:"extension_public_key"`
}

type PairedBrowserExtensionQuery struct {
	ExtensionId string `uri:"id"`
}

type PairedBrowserExtensionQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *PairedBrowserExtensionQueryHandler) Handle(q *PairedBrowserExtensionQuery) (*PairedBrowserExtensionPresenter, error) {
	sql, _, _ := h.Qb.From("browser_extensions").Where(goqu.Ex{
		"id": q.ExtensionId,
	}).ToSQL()

	presenter := &PairedBrowserExtensionPresenter{}

	result := h.Database.Raw(sql).First(&presenter)

	if result.Error != nil {
		return nil, result.Error
	}

	return presenter, nil
}
