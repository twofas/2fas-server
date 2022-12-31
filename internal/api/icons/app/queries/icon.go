package queries

import (
	"github.com/2fas/api/internal/api/icons/adapters"
	"github.com/doug-martin/goqu/v9"
	"gorm.io/gorm"
)

type IconPresenter struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	Type      string `json:"type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type IconQuery struct {
	Id string `uri:"icon_id" validate:"omitempty,uuid4"`
}

type IconQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *IconQueryHandler) FindOne(query *IconQuery) (*IconPresenter, error) {
	sql, _, _ := h.Qb.From("icons").Where(
		goqu.C("id").Eq(query.Id),
		goqu.C("deleted_at").IsNull(),
	).ToSQL()

	presenter := &IconPresenter{}

	result := h.Database.Raw(sql).First(&presenter)

	if result.Error != nil {
		return nil, adapters.IconCouldNotBeFound{IconId: query.Id}
	}

	return presenter, nil
}

func (h *IconQueryHandler) FindAll(query *IconQuery) ([]*IconPresenter, error) {
	var presenter []*IconPresenter

	ds := h.Qb.From("icons").Where(goqu.And(
		goqu.C("deleted_at").IsNull(),
	))

	sql, _, _ := ds.ToSQL()

	h.Database.Raw(sql).Find(&presenter)

	return presenter, nil
}
