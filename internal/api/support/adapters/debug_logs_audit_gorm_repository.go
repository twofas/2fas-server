package adapters

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/support/domain"
	"gorm.io/gorm"
)

type DebugLogsAuditCouldNotBeFound struct {
	AuditId string
}

func (e DebugLogsAuditCouldNotBeFound) Error() string {
	return fmt.Sprintf("Debug logs audit could not be found: %s", e.AuditId)
}

type DebugLogsAuditMysqlRepository struct {
	db *gorm.DB
}

func NewDebugLogsAuditMysqlRepository(db *gorm.DB) *DebugLogsAuditMysqlRepository {
	return &DebugLogsAuditMysqlRepository{db: db}
}

func (r *DebugLogsAuditMysqlRepository) Save(debugLogsAudit *domain.DebugLogsAudit) error {
	if err := r.db.Create(debugLogsAudit).Error; err != nil {
		return err
	}

	return nil
}

func (r *DebugLogsAuditMysqlRepository) Update(debugLogsAudit *domain.DebugLogsAudit) error {
	if err := r.db.Updates(debugLogsAudit).Error; err != nil {
		return err
	}

	return nil
}

func (r *DebugLogsAuditMysqlRepository) FindById(id uuid.UUID) (*domain.DebugLogsAudit, error) {
	audit := &domain.DebugLogsAudit{}

	result := r.db.First(&audit, "id = ?", id.String())

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, DebugLogsAuditCouldNotBeFound{AuditId: id.String()}
	}

	return audit, nil
}

func (r *DebugLogsAuditMysqlRepository) Delete(debugLogsAudit *domain.DebugLogsAudit) error {
	if err := r.db.Delete(debugLogsAudit).Error; err != nil {
		return err
	}

	return nil
}
