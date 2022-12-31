package adapters

import (
	"fmt"
	"github.com/2fas/api/internal/api/browser_extension/domain"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type TokenRequestCouldNotBeFound struct {
	RequestId string
}

func (e TokenRequestCouldNotBeFound) Error() string {
	return fmt.Sprintf("Token request could not be found: %s", e.RequestId)
}

type BrowserExtension2FaRequestsMysqlRepository struct {
	db *gorm.DB
}

func NewBrowserExtension2FaRequestsMysqlRepository(db *gorm.DB) *BrowserExtension2FaRequestsMysqlRepository {
	return &BrowserExtension2FaRequestsMysqlRepository{db: db}
}

func (r *BrowserExtension2FaRequestsMysqlRepository) Save(request *domain.BrowserExtension2FaRequest) error {
	if err := r.db.Create(request).Error; err != nil {
		return err
	}

	return nil
}

func (r *BrowserExtension2FaRequestsMysqlRepository) Update(request *domain.BrowserExtension2FaRequest) error {
	if err := r.db.Save(request).Error; err != nil {
		return err
	}

	return nil
}

func (r *BrowserExtension2FaRequestsMysqlRepository) Delete(request *domain.BrowserExtension2FaRequest) error {
	if err := r.db.Delete(request).Error; err != nil {
		return err
	}

	return nil
}

func (r *BrowserExtension2FaRequestsMysqlRepository) FindPendingByExtensionId(extensionId uuid.UUID) []*domain.BrowserExtension2FaRequest {
	var requests []*domain.BrowserExtension2FaRequest

	r.db.Find(&requests, "extension_id = ?", extensionId.String())

	return requests
}

func (r *BrowserExtension2FaRequestsMysqlRepository) FindById(tokenRequestId, extensionId uuid.UUID) (*domain.BrowserExtension2FaRequest, error) {
	var request *domain.BrowserExtension2FaRequest

	result := r.db.First(&request, "extension_id = ? AND id = ?", extensionId.String(), tokenRequestId.String())

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, TokenRequestCouldNotBeFound{RequestId: tokenRequestId.String()}
	}

	return request, nil
}
