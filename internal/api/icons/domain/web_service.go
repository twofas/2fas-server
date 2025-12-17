package domain

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WebServiceAlreadyExistsError struct {
	Name string
}

func (e WebServiceAlreadyExistsError) Error() string {
	return fmt.Sprintf("Web service already exists: %s", e.Name)
}

type WebService struct {
	gorm.Model

	Id               uuid.UUID `gorm:"primarykey"`
	Name             string
	Description      string
	Issuers          datatypes.JSON
	Tags             datatypes.JSON
	IconsCollections datatypes.JSON
	MatchRules       datatypes.JSON
}

func (WebService) TableName() string {
	return "web_services"
}

type WebServicesRepository interface {
	Save(webService *WebService) error
	Update(webService *WebService) error
	Delete(webService *WebService) error
	FindById(webService uuid.UUID) (*WebService, error)
	FindByName(name string) (*WebService, error)
	FindAll() []*WebService
}
