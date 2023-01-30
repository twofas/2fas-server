package app

import (
	"github.com/twofas/2fas-server/internal/api/support/app/command"
	"github.com/twofas/2fas-server/internal/api/support/app/queries"
)

type Commands struct {
	CreateDebugLogsAudit      *command.CreateDebugLogsAuditHandler
	CreateDebugLogsAuditClain *command.CreateDebugLogsAuditClaimHandler
	UpdateDebugLogsAudit      *command.UpdateDebugLogsAuditHandler
	DeleteDebugLogsAudit      *command.DeleteDebugLogsAuditHandler
	DeleteAllDebugLogsAudit   *command.DeleteAllDebugLogsAuditHandler
}

type Queries struct {
	DebugLogsAuditQuery *queries.DebugLogsAuditQueryHandler
}

type Cqrs struct {
	Commands Commands
	Queries  Queries
}
