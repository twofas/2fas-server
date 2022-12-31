package command

import (
	"github.com/doug-martin/goqu/v9"
	"gorm.io/gorm"
)

// RemoveAllBrowserExtensions command for tests
type RemoveAllBrowserExtensions struct{}

type RemoveAllBrowserExtensionsHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *RemoveAllBrowserExtensionsHandler) Handle(cmd *RemoveAllBrowserExtensions) {
	sql, _, _ := h.Qb.Truncate("browser_extensions").ToSQL()

	h.Database.Exec(sql)
}
