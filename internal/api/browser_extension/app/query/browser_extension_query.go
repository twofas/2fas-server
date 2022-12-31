package query

import (
	"gorm.io/gorm"
)

type BrowserExtensionPresenter struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	BrowserName    string `json:"browser_name"`
	BrowserVersion string `json:"browser_version"`
}

type BrowserExtensionQuery struct {
	Id string `uri:"extension_id" validate:"required,uuid4"`
}

type BrowserExtensionQueryHandler struct {
	Database *gorm.DB
}

func (h *BrowserExtensionQueryHandler) Handle(query *BrowserExtensionQuery) (*BrowserExtensionPresenter, error) {
	var presenter *BrowserExtensionPresenter

	result := h.Database.Raw("SELECT * FROM browser_extensions WHERE id = ?", query.Id).First(&presenter)

	if result.Error != nil {
		return nil, result.Error
	}

	return presenter, nil
}
