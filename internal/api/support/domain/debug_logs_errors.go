package domain

import "github.com/google/uuid"

type DebugLogsAuditClaimIsHasBeenExpiredError struct {
	Id uuid.UUID
}

func (e DebugLogsAuditClaimIsHasBeenExpiredError) Error() string {
	return "Debug logs audit claim has been expired " + e.Id.String()
}

type DebugLogsAuditClaimIsAlreadyCompletedError struct {
	Id uuid.UUID
}

func (e DebugLogsAuditClaimIsAlreadyCompletedError) Error() string {
	return "Debug logs audit claim has been completed " + e.Id.String()
}
