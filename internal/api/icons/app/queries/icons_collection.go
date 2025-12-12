package queries

import (
	"github.com/doug-martin/goqu/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/icons/adapters"
)

type IconsCollectionPresenter struct {
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Icons       datatypes.JSON `json:"icons"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

type IconsCollectionQuery struct {
	Id string `uri:"collection_id" validate:"omitempty,uuid4"`
}

type IconsCollectionQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *IconsCollectionQueryHandler) FindOne(query *IconsCollectionQuery) (*IconsCollectionPresenter, error) {
	sql, _, _ := h.Qb.From("icons_collections").Where(
		goqu.C("id").Eq(query.Id),
		goqu.C("deleted_at").IsNull(),
	).ToSQL()

	presenter := &IconsCollectionPresenter{}

	result := h.Database.Raw(sql).First(&presenter)

	if result.Error != nil {
		return nil, adapters.IconsCollectionCouldNotBeFoundError{IconsCollectionId: query.Id}
	}

	return presenter, nil
}

func (h *IconsCollectionQueryHandler) FindAll(query *IconsCollectionQuery) ([]*IconsCollectionPresenter, error) {
	var presenter []*IconsCollectionPresenter

	ds := h.Qb.From("icons_collections").Where(goqu.And(
		goqu.C("deleted_at").IsNull(),
	))

	sql, _, _ := ds.ToSQL()

	h.Database.Raw(sql).Find(&presenter)

	return presenter, nil
}
