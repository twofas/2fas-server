package adapters

import (
	"errors"
	"fmt"
	"github.com/2fas/api/internal/api/icons/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IconRequestCouldNotBeFound struct {
	IconRequestId string
}

func (e IconRequestCouldNotBeFound) Error() string {
	return fmt.Sprintf("Icon request could not be found: %s", e.IconRequestId)
}

type IconRequestMysqlRepository struct {
	db *gorm.DB
}

func NewIconRequestMysqlRepository(db *gorm.DB) *IconRequestMysqlRepository {
	return &IconRequestMysqlRepository{db: db}
}

func (r *IconRequestMysqlRepository) Save(iconRequest *domain.IconRequest) error {
	if err := r.db.Create(iconRequest).Error; err != nil {
		return err
	}

	return nil
}

func (r *IconRequestMysqlRepository) Update(iconRequest *domain.IconRequest) error {
	if err := r.db.Updates(iconRequest).Error; err != nil {
		return err
	}

	return nil
}

func (r *IconRequestMysqlRepository) Delete(iconRequest *domain.IconRequest) error {
	if err := r.db.Delete(iconRequest).Error; err != nil {
		return err
	}

	return nil
}

func (r *IconRequestMysqlRepository) FindById(id uuid.UUID) (*domain.IconRequest, error) {
	iconRequest := &domain.IconRequest{}

	result := r.db.First(&iconRequest, "id = ?", id.String())

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, IconRequestCouldNotBeFound{IconRequestId: id.String()}
	}

	return iconRequest, nil
}

func (r *IconRequestMysqlRepository) FindAll() []*domain.IconRequest {
	var iconsRequests []*domain.IconRequest

	r.db.Find(&iconsRequests)

	return iconsRequests
}
