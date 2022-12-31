package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Status string

const (
	Pending   Status = "pending"
	Completed Status = "completed"
)

type BrowserExtension2FaRequest struct {
	gorm.Model

	Id          uuid.UUID `gorm:"primarykey"`
	ExtensionId string
	Domain      string
	Status      Status
}

func (BrowserExtension2FaRequest) TableName() string {
	return "browser_extensions_2fa_requests"
}

func (e *BrowserExtension2FaRequest) Close(status Status) {
	e.Status = status
}

func NewBrowserExtension2FaRequest(id, extensionId uuid.UUID, domain string) *BrowserExtension2FaRequest {
	return &BrowserExtension2FaRequest{
		Id:          id,
		ExtensionId: extensionId.String(),
		Domain:      domain,
		Status:      Pending,
	}
}

type BrowserExtension2FaRequestRepository interface {
	Save(request *BrowserExtension2FaRequest) error
	Update(request *BrowserExtension2FaRequest) error
	Delete(tokenRequest *BrowserExtension2FaRequest) error
	FindPendingByExtensionId(extensionId uuid.UUID) []*BrowserExtension2FaRequest
	FindById(tokenRequestId, extensionId uuid.UUID) (*BrowserExtension2FaRequest, error)
}
