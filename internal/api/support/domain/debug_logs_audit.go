package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type DebugLogsAudit struct {
	gorm.Model

	Id          uuid.UUID `gorm:"primarykey"`
	Username    string
	Description string
	File        string
	ExpireAt    time.Time
}

func (DebugLogsAudit) TableName() string {
	return "mobile_debug_logs_audit"
}

func NewDebugLogsAudit(id uuid.UUID, username, description string, expireAt time.Time) *DebugLogsAudit {
	return &DebugLogsAudit{
		Id:          id,
		Username:    username,
		Description: description,
		ExpireAt:    expireAt,
	}
}

type DebugLogAuditRepository interface {
	Save(audit *DebugLogsAudit) error
	Update(audit *DebugLogsAudit) error
	FindById(id uuid.UUID) (*DebugLogsAudit, error)
	Delete(audit *DebugLogsAudit) error
}
