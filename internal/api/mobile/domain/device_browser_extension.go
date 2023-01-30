package domain

import (
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
)

type DeviceBrowserExtensionRepository interface {
	FindAllForDevice(deviceId uuid.UUID) ([]*domain.BrowserExtension, error)
}
