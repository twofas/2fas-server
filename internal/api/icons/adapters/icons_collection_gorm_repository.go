package adapters

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/icons/domain"
)

type IconsCollectionCouldNotBeFoundError struct {
	IconsCollectionId string
}

func (e IconsCollectionCouldNotBeFoundError) Error() string {
	return fmt.Sprintf("Icons collection could not be found: %s", e.IconsCollectionId)
}

type IconsCollectionMysqlRepository struct {
	db *gorm.DB
}

func NewIconsCollectionMysqlRepository(db *gorm.DB) *IconsCollectionMysqlRepository {
	return &IconsCollectionMysqlRepository{db: db}
}

func (r *IconsCollectionMysqlRepository) Save(IconsCollection *domain.IconsCollection) error {
	if err := r.db.Create(IconsCollection).Error; err != nil {
		return err
	}

	return nil
}

func (r *IconsCollectionMysqlRepository) Update(IconsCollection *domain.IconsCollection) error {
	if err := r.db.Updates(IconsCollection).Error; err != nil {
		return err
	}

	return nil
}

func (r *IconsCollectionMysqlRepository) Delete(IconsCollection *domain.IconsCollection) error {
	if err := r.db.Delete(IconsCollection).Error; err != nil {
		return err
	}

	return nil
}

func (r *IconsCollectionMysqlRepository) FindById(id uuid.UUID) (*domain.IconsCollection, error) {
	IconsCollection := &domain.IconsCollection{}

	result := r.db.First(&IconsCollection, "id = ?", id.String())

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, IconsCollectionCouldNotBeFoundError{IconsCollectionId: id.String()}
	}

	return IconsCollection, nil
}

func (r *IconsCollectionMysqlRepository) FindAll() []*domain.IconsCollection {
	var IconsCollections []*domain.IconsCollection

	r.db.Find(&IconsCollections)

	return IconsCollections
}
