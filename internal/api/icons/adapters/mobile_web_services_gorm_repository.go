package adapters

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/icons/domain"
	"github.com/twofas/2fas-server/internal/common/db"
)

type WebServiceCouldNotBeFound struct {
	WebServiceId string
}

func (e WebServiceCouldNotBeFound) Error() string {
	return fmt.Sprintf("Web service could not be found: %s", e.WebServiceId)
}

type WebServiceMysqlRepository struct {
	db *gorm.DB
}

func NewWebServiceMysqlRepository(db *gorm.DB) *WebServiceMysqlRepository {
	return &WebServiceMysqlRepository{db: db}
}

func (r *WebServiceMysqlRepository) Save(webService *domain.WebService) error {
	if err := r.db.Create(webService).Error; err != nil {
		return err
	}

	return nil
}

func (r *WebServiceMysqlRepository) Update(webService *domain.WebService) error {
	if err := r.db.Updates(webService).Error; err != nil {
		return err
	}

	return nil
}

func (r *WebServiceMysqlRepository) Delete(webService *domain.WebService) error {
	if err := r.db.Delete(webService).Error; err != nil {
		return err
	}

	return nil
}

func (r *WebServiceMysqlRepository) FindById(id uuid.UUID) (*domain.WebService, error) {
	webService := &domain.WebService{}

	result := r.db.First(&webService, "id = ?", id.String())

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, WebServiceCouldNotBeFound{WebServiceId: id.String()}
	} else if result.Error != nil {
		return nil, db.WrapError(result.Error)
	}

	return webService, nil
}

func (r *WebServiceMysqlRepository) FindByName(name string) (*domain.WebService, error) {
	webService := &domain.WebService{}

	result := r.db.First(&webService, "name = ?", name)

	if err := result.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("web service could not be found")
	} else if err != nil {
		return nil, db.WrapError(err)
	}

	return webService, nil
}

func (r *WebServiceMysqlRepository) FindAll() []*domain.WebService {
	var webServices []*domain.WebService

	r.db.Find(&webServices)

	return webServices
}
