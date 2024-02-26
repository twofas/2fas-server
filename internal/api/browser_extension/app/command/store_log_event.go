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
		logging.FromContext(ctx).WithFields(context).Info(cmd.Message)
	case "warning":
		logging.FromContext(ctx).WithFields(context).Warning(cmd.Message)
	case "error":
		logging.FromContext(ctx).WithFields(context).Error(cmd.Message)
	case "debug":
		logging.FromContext(ctx).WithFields(context).Debug(cmd.Message)
	}
}
