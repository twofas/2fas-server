package queries

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/twofas/2fas-server/internal/api/icons/adapters"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type IconRequestPresenter struct {
	Id           string         `json:"id"`
	CallerId     string         `json:"caller_id"`
	ServiceName  string         `json:"name"`
	Issuers      datatypes.JSON `json:"issuers"`
	Description  string         `json:"description"`
	LightIconUrl string         `json:"light_icon_url"`
	DarkIconUrl  string         `json:"dark_icon_url"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
	DeletedAt    string         `json:"deleted_at"`
}

type IconRequestQuery struct {
	Id string `uri:"icon_request_id" validate:"omitempty,uuid4"`
}

type IconRequestQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *IconRequestQueryHandler) FindOne(query *IconRequestQuery) (*IconRequestPresenter, error) {
	irT := goqu.T("icons_requests")

	sql, _, _ := h.Qb.From(irT).
		Select(
			irT.Col("id").As("id"),
			irT.Col("caller_id").As("caller_id"),
			irT.Col("service_name").As("service_name"),
			irT.Col("description").As("description"),
			irT.Col("issuers").As("issuers"),
			irT.Col("created_at").As("created_at"),
			irT.Col("updated_at").As("updated_at"),
			irT.Col("deleted_at").As("deleted_at"),
			irT.Col("light_icon_url").As("light_icon_url"),
			irT.Col("dark_icon_url").As("dark_icon_url"),
		).
		Where(
			irT.Col("id").Eq(query.Id),
			irT.Col("deleted_at").IsNull(),
		).
		ToSQL()

	presenter := &IconRequestPresenter{}

	result := h.Database.Raw(sql).First(presenter)

	if result.Error != nil {
		return nil, adapters.IconRequestCouldNotBeFound{IconRequestId: query.Id}
	}

	return presenter, nil
}

func (h *IconRequestQueryHandler) FindAll(query *IconRequestQuery) ([]*IconRequestPresenter, error) {
	var presenter []*IconRequestPresenter

	irT := goqu.T("icons_requests")

	sql, _, _ := h.Qb.From(irT).
		Select(
			irT.Col("id").As("id"),
			irT.Col("caller_id").As("caller_id"),
			irT.Col("service_name").As("service_name"),
			irT.Col("description").As("description"),
			irT.Col("issuers").As("issuers"),
			irT.Col("created_at").As("created_at"),
			irT.Col("updated_at").As("updated_at"),
			irT.Col("light_icon_url").As("light_icon_url"),
			irT.Col("dark_icon_url").As("dark_icon_url"),
		).
		Where(irT.Col("deleted_at").IsNull()).
		ToSQL()

	h.Database.Raw(sql).Find(&presenter)

	return presenter, nil
}
