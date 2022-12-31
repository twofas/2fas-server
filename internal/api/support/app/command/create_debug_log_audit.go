package command

import (
	"bytes"
	"errors"
	"github.com/2fas/api/internal/api/support/adapters"
	"github.com/2fas/api/internal/api/support/domain"
	"github.com/2fas/api/internal/common/clock"
	"github.com/2fas/api/internal/common/storage"
	"github.com/google/uuid"
	"io/ioutil"
	"mime/multipart"
	"path/filepath"
)

type CreateDebugLogsAudit struct {
	Id   string                `uri:"audit_id" validate:"required"`
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type CreateDebugLogsAuditHandler struct {
	DebugLogsAuditRepository domain.DebugLogAuditRepository
	FileSystem               storage.FileSystemStorage
	Config                   domain.DebugLogsConfig
	Clock                    clock.Clock
}

func (h *CreateDebugLogsAuditHandler) Handle(command *CreateDebugLogsAudit) error {
	id, err := uuid.Parse(command.Id)

	if err != nil {
		return adapters.DebugLogsAuditCouldNotBeFound{AuditId: id.String()}
	}

	auditClaim, err := h.DebugLogsAuditRepository.FindById(id)

	if err != nil {
		return err
	}

	if auditClaim.File != "" {
		return domain.DebugLogsAuditClaimIsAlreadyCompletedError{Id: id}
	}

	if command.File == nil {
		return errors.New("logs file is required")
	}

	if h.Clock.Now().After(auditClaim.ExpireAt) {
		return domain.DebugLogsAuditClaimIsHasBeenExpiredError{Id: id}
	}

	logsFile, err := command.File.Open()

	if err != nil {
		return err
	}

	file, _ := ioutil.ReadAll(logsFile)

	logsFileReader := bytes.NewReader(file)

	logsPath := filepath.Join(h.Config.DebugLogsDirectory, auditClaim.Id.String())
	fileLocation, err := h.FileSystem.Save(logsPath, logsFileReader)

	if err != nil {
		return err
	}

	auditClaim.File = fileLocation

	err = h.DebugLogsAuditRepository.Update(auditClaim)

	if err != nil {
		return err
	}

	return nil
}
