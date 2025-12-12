package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	Light string = "light"
	Dark  string = "dark"
)

type Icon struct {
	gorm.Model

	Id     uuid.UUID `gorm:"primarykey"`
	Name   string
	Type   string
	Url    string
	Width  int
	Height int
}

func (Icon) TableName() string {
	return "icons"
}

type IconsRepository interface {
	Save(icon *Icon) error
	Update(icon *Icon) error
	Delete(icon *Icon) error
	FindById(id uuid.UUID) (*Icon, error)
	FindAll() []*Icon
}
