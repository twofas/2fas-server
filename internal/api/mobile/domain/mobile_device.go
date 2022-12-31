package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	IOS     = "ios"
	Android = "android"
)

type MobileDevice struct {
	gorm.Model

	Id       uuid.UUID `gorm:"primarykey"`
	Name     string
	Platform string
	FcmToken string
}

func (MobileDevice) TableName() string {
	return "mobile_devices"
}

func NewMobileDevice(id uuid.UUID, name, platform, fcmToken string) *MobileDevice {
	return &MobileDevice{
		Id:       id,
		Name:     name,
		Platform: platform,
		FcmToken: fcmToken,
	}
}

type MobileDeviceRepository interface {
	Save(device *MobileDevice) error
	Update(device *MobileDevice) error
	FindById(device uuid.UUID) (*MobileDevice, error)
	FindAll() []*MobileDevice
}
