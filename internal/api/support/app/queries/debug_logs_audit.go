package queries

import (
	"github.com/2fas/api/internal/api/support/adapters"
	"github.com/doug-martin/goqu/v9"
	"gorm.io/gorm"
)

type DebugLogsAuditPresenter struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	Description string `json:"description"`
	File        string `json:"file"`
	ExpireAt    string `json:"expire_at"`
	CreatedAt   string `json:"created_at"`
}

type DebugLogsAuditQuery struct {
	Id string `uri:"audit_id"`
}

type DebugLogsAuditQueryHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *DebugLogsAuditQueryHandler) Find(query *DebugLogsAuditQuery) (*DebugLogsAuditPresenter, error) {
	sql, _, _ := h.Qb.From("mobile_debug_logs_audit").Where(
		goqu.C("id").Eq(query.Id),
		goqu.C("deleted_at").IsNull(),
	).ToSQL()

	presenter := &DebugLogsAuditPresenter{}

	result := h.Database.Raw(sql).First(&presenter)

	if result.Error != nil {
		return nil, adapters.DebugLogsAuditCouldNotBeFound{AuditId: query.Id}
	}

	return presenter, nil
}

func (h *DebugLogsAuditQueryHandler) FindAll(query *DebugLogsAuditQuery) ([]*DebugLogsAuditPresenter, error) {
	var presenter []*DebugLogsAuditPresenter

	sql, _, _ := h.Qb.From("mobile_debug_logs_audit").Where(
		goqu.C("deleted_at").IsNull(),
	).ToSQL()

	h.Database.Raw(sql).Find(&presenter)

	return presenter, nil
}
