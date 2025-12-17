package adapters

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/icons/domain"
)

type IconsCollectionsRelationsMysqlRepository struct {
	db *gorm.DB
}

func NewIconsCollectionsRelationsMysqlRepository(db *gorm.DB) *IconsCollectionsRelationsMysqlRepository {
	return &IconsCollectionsRelationsMysqlRepository{db: db}
}

func (r *IconsCollectionsRelationsMysqlRepository) DeleteAll(IconCollection *domain.IconsCollection) error {
	sql := fmt.Sprintf("UPDATE web_services SET icons_collections = %s WHERE \"%s\" MEMBER OF (icons_collections)",
		"JSON_REMOVE(`icons_collections`, JSON_UNQUOTE(JSON_SEARCH(`icons_collections`, 'one', '"+IconCollection.Id.String()+"')))",
		IconCollection.Id.String(),
	)

	res := r.db.Exec(sql)

	return res.Error
}
