package adapters

import (
	"fmt"
	"github.com/2fas/api/internal/api/icons/domain"
	"gorm.io/gorm"
)

type IconsRelationsMysqlRepository struct {
	db *gorm.DB
}

func NewIconsRelationsMysqlRepository(db *gorm.DB) *IconsRelationsMysqlRepository {
	return &IconsRelationsMysqlRepository{db: db}
}

func (r *IconsRelationsMysqlRepository) DeleteAll(Icon *domain.Icon) error {
	sql := fmt.Sprintf("UPDATE icons_collections SET icons = %s WHERE \"%s\" MEMBER OF (icons)",
		"JSON_REMOVE(`icons`, JSON_UNQUOTE(JSON_SEARCH(`icons`, 'one', '"+Icon.Id.String()+"')))",
		Icon.Id.String(),
	)

	res := r.db.Exec(sql)

	return res.Error
}
