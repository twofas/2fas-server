package domain

import (
	"database/sql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MobileNotification struct {
	gorm.Model

	Id          uuid.UUID `gorm:"primarykey"`
	Icon        string
	Link        string
	Message     string
	Push        bool
	Platform    string
	Version     string
	PublishedAt sql.NullTime
}

func (MobileNotification) TableName() string {
	return "mobile_notifications"
}

type MobileNotificationsRepository interface {
	Save(notification *MobileNotification) error
	Update(notification *MobileNotification) error
	Delete(notification *MobileNotification) error
	FindById(id uuid.UUID) (*MobileNotification, error)
	FindAll() []*MobileNotification
}
