package command

import (
	"github.com/2fas/api/internal/api/browser_extension/domain"
	"github.com/google/uuid"
)

type RegisterBrowserExtension struct {
	BrowserExtensionId uuid.UUID
	Name               string `json:"name" validate:"required,lte=64"`
	BrowserName        string `json:"browser_name" validate:"required,lte=255"`
	BrowserVersion     string `json:"browser_version" validate:"required,lte=32"`
	PublicKey          string `json:"public_key" validate:"required,lte=768"`
}

type RegisterBrowserExtensionHandler struct {
	Repository domain.BrowserExtensionRepository
}

func (h *RegisterBrowserExtensionHandler) Handle(cmd *RegisterBrowserExtension) error {
	browserExtension := domain.NewBrowserExtension()

	browserExtension.Id = cmd.BrowserExtensionId
	browserExtension.Name = cmd.Name
	browserExtension.BrowserName = cmd.BrowserName
	browserExtension.BrowserVersion = cmd.BrowserVersion
	browserExtension.PublicKey = cmd.PublicKey

	return h.Repository.Save(browserExtension)
}
