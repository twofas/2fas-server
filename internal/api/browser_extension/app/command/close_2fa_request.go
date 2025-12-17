package command

import (
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
)

type Close2FaRequest struct {
	ExtensionId    string `uri:"extension_id" validate:"required,uuid4"`
	TokenRequestId string `uri:"token_request_id" validate:"required,uuid4"`
	Status         string `json:"status" validate:"required,oneof=completed terminated"`
}

type Close2FaRequestHandler struct {
	BrowserExtensionsRepository          domain.BrowserExtensionRepository
	BrowserExtension2FaRequestRepository domain.BrowserExtension2FaRequestRepository
}

func (h *Close2FaRequestHandler) Handle(cmd *Close2FaRequest) error {
	extId, _ := uuid.Parse(cmd.ExtensionId)

	browserExtension, err := h.BrowserExtensionsRepository.FindById(extId)

	if err != nil {
		return err
	}

	tokenRequestId, _ := uuid.Parse(cmd.TokenRequestId)

	tokenRequest, err := h.BrowserExtension2FaRequestRepository.FindById(tokenRequestId, browserExtension.Id)

	if err != nil {
		return err
	}

	tokenRequest.Close(domain.Status(cmd.Status))

	err = h.BrowserExtension2FaRequestRepository.Update(tokenRequest)

	if err != nil {
		return err
	}

	err = h.BrowserExtension2FaRequestRepository.Delete(tokenRequest)

	if err != nil {
		return err
	}

	return nil
}
