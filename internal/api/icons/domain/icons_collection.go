package domain

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type IconsCollection struct {
	gorm.Model

	Id          uuid.UUID `gorm:"primarykey"`
	Name        string
	Description string
	Icons       datatypes.JSON
}

func (IconsCollection) TableName() string {
	return "icons_collections"
}

type IconsCollectionRepository interface {
	Save(iconsCollection *IconsCollection) error
	Update(iconsCollection *IconsCollection) error
	Delete(iconsCollection *IconsCollection) error
	FindById(id uuid.UUID) (*IconsCollection, error)
	FindAll() []*IconsCollection
}
