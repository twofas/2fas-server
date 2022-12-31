package command

import (
	"encoding/json"
	"github.com/2fas/api/internal/api/browser_extension/domain"
	"github.com/2fas/api/internal/common/logging"
	"github.com/google/uuid"
)

type StoreLogEvent struct {
	ExtensionId string `uri:"extension_id" validate:"required,uuid4"`
	Level       string `json:"level" validate:"required,oneof=info warning error"`
	Message     string `json:"message" validate:"required"`
	Context     string `json:"context" validate:"omitempty,json"`
}

type StoreLogEventHandler struct {
	BrowserExtensionsRepository domain.BrowserExtensionRepository
}

func (h *StoreLogEventHandler) Handle(cmd *StoreLogEvent) {
	extId, _ := uuid.Parse(cmd.ExtensionId)

	_, err := h.BrowserExtensionsRepository.FindById(extId)

	if err != nil {
		return
	}

	context := logging.Fields{}

	json.Unmarshal([]byte(cmd.Context), &context)

	context["source"] = "browser_extension"

	switch cmd.Level {
	case "info":
		logging.WithFields(context).Info(cmd.Message)
	case "warning":
		logging.WithFields(context).Warning(cmd.Message)
	case "error":
		logging.WithFields(context).Error(cmd.Message)
	case "debug":
		logging.WithFields(context).Debug(cmd.Message)
	}
}
