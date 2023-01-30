package command

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/support/domain"
	"gorm.io/gorm"
)

type DeleteDebugLogsAudit struct {
	Id string `uri:"audit_id" validate:"required,uuid4"`
}

type DeleteDebugLogsAuditHandler struct {
	DebugLogsAuditRepository domain.DebugLogAuditRepository
}

func (h *DeleteDebugLogsAuditHandler) Handle(cmd *DeleteDebugLogsAudit) error {
	auditId, _ := uuid.Parse(cmd.Id)

	audit, err := h.DebugLogsAuditRepository.FindById(auditId)

	if err != nil {
		return err
	}

	return h.DebugLogsAuditRepository.Delete(audit)
}

type DeleteAllDebugLogsAudit struct{}

type DeleteAllDebugLogsAuditHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *DeleteAllDebugLogsAuditHandler) Handle(cmd *DeleteAllDebugLogsAudit) {
	sql, _, _ := h.Qb.Truncate("mobile_debug_logs_audit").ToSQL()

	h.Database.Exec(sql)
}
