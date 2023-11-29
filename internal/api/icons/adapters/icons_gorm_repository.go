package adapters

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/icons/domain"
	"github.com/twofas/2fas-server/internal/common/db"
)

type IconCouldNotBeFound struct {
	IconId string
}

func (e IconCouldNotBeFound) Error() string {
	return fmt.Sprintf("Icon could not be found: %s", e.IconId)
}

type IconMysqlRepository struct {
	db *gorm.DB
}

func NewIconMysqlRepository(db *gorm.DB) *IconMysqlRepository {
	return &IconMysqlRepository{db: db}
}

func (r *IconMysqlRepository) Save(Icon *domain.Icon) error {
	if err := r.db.Create(Icon).Error; err != nil {
		return err
	}

	return nil
}

func (r *IconMysqlRepository) Update(Icon *domain.Icon) error {
	if err := r.db.Updates(Icon).Error; err != nil {
		return err
	}

	return nil
}

func (r *IconMysqlRepository) Delete(Icon *domain.Icon) error {
	if err := r.db.Delete(Icon).Error; err != nil {
		return err
	}

	return nil
}

func (r *IconMysqlRepository) FindById(id uuid.UUID) (*domain.Icon, error) {
	Icon := &domain.Icon{}

	result := r.db.First(&Icon, "id = ?", id.String())

	if err := result.Error; err != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, IconCouldNotBeFound{IconId: id.String()}
		} else {
			return nil, db.WrapError(err)
		}
	}

	return Icon, nil
}

func (r *IconMysqlRepository) FindAll() []*domain.Icon {
	var Icons []*domain.Icon

	r.db.Find(&Icons)

	return Icons
}
