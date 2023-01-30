package adapters

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
	"gorm.io/gorm"
)

type BrowserExtensionsCouldNotBeFound struct {
	ExtensionId string
}

func (e BrowserExtensionsCouldNotBeFound) Error() string {
	return fmt.Sprintf("Extension could not be found: %s", e.ExtensionId)
}

type BrowserExtensionsMysqlRepository struct {
	db *gorm.DB
}

func NewBrowserExtensionsMysqlRepository(db *gorm.DB) *BrowserExtensionsMysqlRepository {
	return &BrowserExtensionsMysqlRepository{db: db}
}

func (r *BrowserExtensionsMysqlRepository) Save(browserExtension *domain.BrowserExtension) error {
	if err := r.db.Create(browserExtension).Error; err != nil {
		return err
	}

	return nil
}

func (r *BrowserExtensionsMysqlRepository) Update(browserExtension *domain.BrowserExtension) error {
	if err := r.db.Updates(browserExtension).Error; err != nil {
		return err
	}

	return nil
}

func (r *BrowserExtensionsMysqlRepository) FindById(id uuid.UUID) (*domain.BrowserExtension, error) {
	var extension *domain.BrowserExtension

	result := r.db.First(&extension, "id = ?", id.String())

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, BrowserExtensionsCouldNotBeFound{ExtensionId: id.String()}
	}

	return extension, nil
}
