package command

import (
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/support/domain"
)

type UpdateDebugLogsAudit struct {
	Id          string `uri:"audit_id" validate:"required,uuid4"`
	Username    string `json:"username" validate:"required,max=128"`
	Description string `json:"description" validate:"required,max=2048"`
}

type UpdateDebugLogsAuditHandler struct {
	DebugLogsAuditRepository domain.DebugLogAuditRepository
}

func (h *UpdateDebugLogsAuditHandler) Handle(command *UpdateDebugLogsAudit) error {
	id, err := uuid.Parse(command.Id)

	if err != nil {
		return err
	}

	audit, err := h.DebugLogsAuditRepository.FindById(id)

	if err != nil {
		return err
	}

	if command.Username != "" {
		audit.Username = command.Username
	}

	if command.Description != "" {
		audit.Description = command.Description
	}

	return h.DebugLogsAuditRepository.Update(audit)
}
