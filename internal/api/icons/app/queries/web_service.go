package queries

import (
	"errors"

	"github.com/doug-martin/goqu/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/icons/adapters"
	"github.com/twofas/2fas-server/internal/common/db"
)

type WebServicePresenter struct {
	Id               string         `json:"id"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Issuers          datatypes.JSON `json:"issuers"`
	Tags             datatypes.JSON `json:"tags"`
	IconsCollections datatypes.JSON `json:"icons_collections"`
	MatchRules       datatypes.JSON `json:"match_rules"`
	CreatedAt        string         `json:"created_at"`
	UpdatedAt        string         `json:"updated_at"`
}

type WebServiceQuery struct {
	Id     string `uri:"service_id" validate:"omitempty,uuid4"`
	Search string `uri:"search" validate:"omitempty,max=128"`
}

type WebServiceQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *WebServiceQueryHandler) FindOne(query *WebServiceQuery) (*WebServicePresenter, error) {
	ds := h.Qb.From("web_services")

	ds = ds.Where(
		goqu.C("id").Eq(query.Id),
		goqu.C("deleted_at").IsNull(),
	)

	sql, _, err := ds.ToSQL()
	if err != nil {
		return nil, db.QueryPrepError(err)
	}

	presenter := &WebServicePresenter{}

	result := h.Database.Raw(sql).First(&presenter)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, db.WrapError(err)
		}
		return nil, adapters.WebServiceCouldNotBeFound{Identifier: query.Id}
	}

	return presenter, nil
}

func (h *WebServiceQueryHandler) FindAll(query *WebServiceQuery) ([]*WebServicePresenter, error) {
	var presenter []*WebServicePresenter

	ds := h.Qb.From("web_services")

	ds = ds.Where(goqu.And(
		goqu.C("deleted_at").IsNull(),
	))

	if query.Search != "" {
		ds.Where(
			goqu.Or(
				goqu.C("name").Eq(query.Search),
				goqu.L(`JSON_CONTAINS('tags', '`+query.Search+`', '$')`),
				goqu.L(`JSON_CONTAINS('issuers', '`+query.Search+`', '$')`),
			),
		)
	}

	sql, _, err := ds.ToSQL()
	if err != nil {
		return nil, db.QueryPrepError(err)
	}

	result := h.Database.Raw(sql).Find(&presenter)
	if err := result.Error; err != nil {
		return nil, db.WrapError(err)
	}

	return presenter, nil
}
