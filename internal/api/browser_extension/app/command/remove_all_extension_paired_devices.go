package command

import (
	"github.com/2fas/api/internal/api/browser_extension/domain"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
		h.BrowserExtensionPairedDevicesRepository.Delete(device)
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
