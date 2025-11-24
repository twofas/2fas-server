package command

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
)

type RemoveAllExtensionPairedDevices struct {
	ExtensionId string `uri:"extension_id" validate:"required,uuid4"`
}

type RemoveALlExtensionPairedDevicesHandler struct {
	BrowserExtensionRepository              domain.BrowserExtensionRepository
	BrowserExtensionPairedDevicesRepository domain.BrowserExtensionDevicesRepository
}

func (h *RemoveALlExtensionPairedDevicesHandler) Handle(cmd *RemoveAllExtensionPairedDevices) error {
	extId, _ := uuid.Parse(cmd.ExtensionId)

	extension, err := h.BrowserExtensionRepository.FindById(extId)
	if err != nil {
		return err
	}

	pairedDevices := h.BrowserExtensionPairedDevicesRepository.FindAll(extension.Id)

	for _, device := range pairedDevices {
		err := h.BrowserExtensionPairedDevicesRepository.Delete(device)
		if err != nil {
			return fmt.Errorf("failed to remove paired device: %w", err)
		}
	}

	return nil
}

// RemoveAllBrowserExtensionsDevices command for tests
type RemoveAllBrowserExtensionsDevices struct{}

type RemoveAllBrowserExtensionsDevicesHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *RemoveAllBrowserExtensionsDevicesHandler) Handle(cmd *RemoveAllBrowserExtensionsDevices) {
	sql, _, _ := h.Qb.Truncate("mobile_device_browser_extension").ToSQL()

	h.Database.Exec(sql)
}
