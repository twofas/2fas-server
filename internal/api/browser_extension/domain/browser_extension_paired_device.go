package domain

import (
	"github.com/google/uuid"
)

const (
	IOS     = "ios"
	Android = "android"
)

type ExtensionDevice struct {
	Id          uuid.UUID
	ExtensionId uuid.UUID
	Name        string
	Platform    string
	FcmToken    string
}

func (e *ExtensionDevice) IsAndroid() bool {
	return e.Platform == Android
}

func (e *ExtensionDevice) IsiOS() bool {
	return e.Platform == IOS
}

type BrowserExtensionDevicesRepository interface {
	FindAll(extensionId uuid.UUID) []*ExtensionDevice
	GetById(extensionId, deviceId uuid.UUID) (*ExtensionDevice, error)
	Delete(device *ExtensionDevice) error
}
