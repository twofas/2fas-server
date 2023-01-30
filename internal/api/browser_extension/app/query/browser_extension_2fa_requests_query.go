package query

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
	"gorm.io/gorm"
)

type BrowserExtension2FaRequestPresenter struct {
	Id          string `json:"token_request_id"`
	ExtensionId string `json:"extension_id"`
	Domain      string `json:"domain"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type BrowserExtension2FaRequestQuery struct {
	ExtensionId    string `uri:"extension_id" validate:"required,uuid4"`
	TokenRequestId string `uri:"token_request_id" validate:"uuid4"`
	Status         domain.Status
}

type BrowserExtension2FaRequestQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *BrowserExtension2FaRequestQueryHandler) Handle(query *BrowserExtension2FaRequestQuery) ([]*BrowserExtension2FaRequestPresenter, error) {
	var presenter []*BrowserExtension2FaRequestPresenter

	ds := h.Qb.From("browser_extensions_2fa_requests")

	ds = ds.Where(goqu.C("extension_id").Eq(query.ExtensionId))

	if query.TokenRequestId != "" {
		ds = ds.Where(goqu.C("id").Eq(query.TokenRequestId))
	}

	if query.Status != "" {
		ds = ds.Where(goqu.C("status").Eq(query.Status))
	}

	ds = ds.Order(goqu.I("created_at").Desc())

	sql, _, _ := ds.ToSQL()

	result := h.Database.Raw(sql).Find(&presenter)

	if result.Error != nil {
		return nil, result.Error
	}

	return presenter, nil
}
