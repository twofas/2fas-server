package domain

import (
	"github.com/2fas/api/internal/api/browser_extension/domain"
	"github.com/google/uuid"
)

type DeviceBrowserExtensionRepository interface {
	FindAllForDevice(deviceId uuid.UUID) ([]*domain.BrowserExtension, error)
}
