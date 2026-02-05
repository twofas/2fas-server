package adapters

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
)

type ExtensionDeviceCouldNotBeFoundError struct {
	DeviceId string
}

func (d ExtensionDeviceCouldNotBeFoundError) Error() string {
	return fmt.Sprintf("Extension device could not be found: %s", d.DeviceId)
}

type BrowserExtensionDevicesMysqlRepository struct {
	db *gorm.DB
	qb *goqu.Database
}

func NewBrowserExtensionDevicesMysqlRepository(db *gorm.DB, qb *goqu.Database) *BrowserExtensionDevicesMysqlRepository {
	return &BrowserExtensionDevicesMysqlRepository{
		db: db,
		qb: qb,
	}
}

func (r *BrowserExtensionDevicesMysqlRepository) FindAll(id uuid.UUID) []*domain.ExtensionDevice {
	var devices []*domain.ExtensionDevice

	mdbeT := goqu.T("mobile_device_browser_extension")
	mdT := goqu.T("mobile_devices")

	sql, _, _ := r.qb.From(mdT).
		Select(
			mdT.Col("id").As("id"),
			mdbeT.Col("extension_id").As("extension_id"),
			mdT.Col("name").As("name"),
			mdT.Col("platform").As("platform"),
			mdT.Col("fcm_token").As("fcm_token")).
		LeftJoin(mdbeT, goqu.On(mdbeT.Col("device_id").Eq(mdT.Col("id")))).
		Where(mdbeT.Col("extension_id").Eq(id.String())).
		ToSQL()

	r.db.Raw(sql).Find(&devices)

	return devices
}

func (r *BrowserExtensionDevicesMysqlRepository) Delete(pairedDevice *domain.ExtensionDevice) error {
	sql, _, _ := r.qb.From("mobile_device_browser_extension").
		Where(
			goqu.C("device_id").Eq(pairedDevice.Id.String()),
			goqu.C("extension_id").Eq(pairedDevice.ExtensionId.String()),
		).
		Delete().ToSQL()

	if result := r.db.Exec(sql); result.Error != nil {
		return errors.New("Could not delete device: " + result.Error.Error())
	}

	return nil
}

func (r *BrowserExtensionDevicesMysqlRepository) GetById(extensionId, deviceId uuid.UUID) (*domain.ExtensionDevice, error) {
	var device *domain.ExtensionDevice

	mdbeT := goqu.T("mobile_device_browser_extension")
	mdT := goqu.T("mobile_devices")

	sql, _, _ := r.qb.From(mdT).
		Select(
			mdT.Col("id").As("id"),
			mdbeT.Col("extension_id").As("extension_id"),
			mdT.Col("name").As("name"),
			mdT.Col("platform").As("platform"),
			mdT.Col("fcm_token").As("fcm_token")).
		LeftJoin(mdbeT, goqu.On(mdbeT.Col("device_id").Eq(mdT.Col("id")))).
		Where(
			mdbeT.Col("extension_id").Eq(extensionId.String()),
			mdbeT.Col("device_id").Eq(deviceId.String()),
		).
		ToSQL()

	result := r.db.Raw(sql).First(&device)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ExtensionDeviceCouldNotBeFoundError{DeviceId: deviceId.String()}
	}

	return device, nil
}
