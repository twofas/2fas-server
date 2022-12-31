package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BrowserExtension struct {
	gorm.Model

	Id             uuid.UUID `gorm:"primarykey"`
	Name           string
	BrowserName    string
	BrowserVersion string
	PublicKey      string
}

func NewBrowserExtension() *BrowserExtension {
	return &BrowserExtension{}
}

type BrowserExtensionRepository interface {
	Save(extension *BrowserExtension) error
	Update(extension *BrowserExtension) error
	FindById(id uuid.UUID) (*BrowserExtension, error)
}
