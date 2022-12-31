package domain

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type IconRequest struct {
	gorm.Model

	Id           uuid.UUID `gorm:"primarykey"`
	CallerId     string
	ServiceName  string
	Issuers      datatypes.JSON
	Description  string
	LightIconUrl string
	DarkIconUrl  string
}

func (IconRequest) TableName() string {
	return "icons_requests"
}

type IconsRequestsRepository interface {
	Save(iconRequest *IconRequest) error
	Update(iconRequest *IconRequest) error
	Delete(iconRequest *IconRequest) error
	FindById(id uuid.UUID) (*IconRequest, error)
	FindAll() []*IconRequest
}
