package command

import (
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/api/support/domain"
	"github.com/twofas/2fas-server/internal/common/clock"
)

type CreateDebugLogsAuditClaim struct {
	Id          uuid.UUID
	Username    string `json:"username" validate:"required,max=128"`
	Description string `json:"description" validate:"required,max=2048"`
}

type CreateDebugLogsAuditClaimHandler struct {
	DebugLogsAuditRepository domain.DebugLogAuditRepository
	DebugLogsAuditConfig     domain.DebugLogsConfig
	Clock                    clock.Clock
}

func (h *CreateDebugLogsAuditClaimHandler) Handle(cmd *CreateDebugLogsAuditClaim) error {
	expirationTime := h.Clock.Now().Add(h.DebugLogsAuditConfig.ExpireAt)

	debugLogsAudit := domain.NewDebugLogsAudit(cmd.Id, cmd.Username, cmd.Description, expirationTime)

	err := h.DebugLogsAuditRepository.Save(debugLogsAudit)

	if err != nil {
		return err
	}

	return nil
}
