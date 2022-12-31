package domain

import (
	"github.com/google/uuid"
)

type ExtensionHasAlreadyBeenPairedError struct {
	ExtensionId string `json:"extension_id"`
}

func (e ExtensionHasAlreadyBeenPairedError) Error() string {
	return "Extension has already been paired"
}

type MobileDeviceExtension struct {
	DeviceId    uuid.UUID `gorm:"primarykey"`
	ExtensionId uuid.UUID `gorm:"primarykey"`
	Name        string
	Platform    string
	FcmToken    string
}

func (MobileDeviceExtension) TableName() string {
	return "mobile_device_browser_extension"
}

type MobileDeviceExtensionsRepository interface {
	FindById(deviceId, extensionId uuid.UUID) (*MobileDeviceExtension, error)
	Delete(extension *MobileDeviceExtension) error
}
