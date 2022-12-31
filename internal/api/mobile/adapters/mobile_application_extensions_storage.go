package adapters

import (
	"github.com/2fas/api/internal/common/clock"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceExtensionsService struct {
	db    *gorm.DB
	qb    *goqu.Database
	clock clock.Clock
}

func NewDeviceExtensionsService(db *gorm.DB, qb *goqu.Database, clock clock.Clock) *DeviceExtensionsService {
	return &DeviceExtensionsService{
		db:    db,
		qb:    qb,
		clock: clock,
	}
}

func (r *DeviceExtensionsService) PairDeviceWithBrowserExtension(deviceId string, extId uuid.UUID) error {
	ds := r.qb.Insert("mobile_device_browser_extension").OnConflict(goqu.DoNothing()).Rows(
		goqu.Record{
			"device_id":    deviceId,
			"extension_id": extId.String(),
			"created_at":   r.clock.Now(),
		},
	)

	sql, _, _ := ds.ToSQL()

	if err := r.db.Exec(sql).Error; err != nil {
		return err
	}

	return nil
}
