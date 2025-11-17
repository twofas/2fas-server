package command

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
	"github.com/twofas/2fas-server/internal/common/logging"
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

func (h *StoreLogEventHandler) Handle(ctx context.Context, cmd *StoreLogEvent) {
	log := logging.FromContext(ctx)

	extId, err := uuid.Parse(cmd.ExtensionId)
	if err != nil {
		log.Error("Failed to parse extension id: %v", err)
		return
	}
	_, err = h.BrowserExtensionsRepository.FindById(extId)
	if err != nil {
		return
	}

	context := logging.Fields{}
	if err := json.Unmarshal([]byte(cmd.Context), &context); err != nil {
		log.Error("Failed to unmarshal log context: %v", err)
		return
	}

	context["source"] = "browser_extension"

	switch cmd.Level {
	case "info":
		log.WithFields(context).Info(cmd.Message)
	case "warning":
		log.WithFields(context).Warning(cmd.Message)
	case "error":
		log.WithFields(context).Error(cmd.Message)
	case "debug":
		log.WithFields(context).Debug(cmd.Message)
	}
}
