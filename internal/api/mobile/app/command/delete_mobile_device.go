package command

import (
	"github.com/doug-martin/goqu/v9"
	"gorm.io/gorm"
)

type RemoveAllMobileDevices struct{}

type RemoveAllMobileDevicesHandler struct {
	Database *gorm.DB
	Qb       *goqu.Database
}

func (h *RemoveAllMobileDevicesHandler) Handle(cmd *RemoveAllMobileDevices) {
	sql, _, _ := h.Qb.Truncate("mobile_devices").ToSQL()

	h.Database.Exec(sql)
}
