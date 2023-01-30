package adapters

import (
	"errors"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/mobile/domain"
	"gorm.io/gorm"
)

type MobileDeviceExtensionCouldNotBeFound struct {
	DeviceId    string
	ExtensionId string
}

func (e MobileDeviceExtensionCouldNotBeFound) Error() string {
	return fmt.Sprintf("Mobile device could not be found: %s", e.DeviceId)
}

type MobileDeviceExtensionsGormRepository struct {
	db *gorm.DB
	qb *goqu.Database
}

func NewMobileDeviceExtensionsGormRepository(db *gorm.DB, qb *goqu.Database) *MobileDeviceExtensionsGormRepository {
	return &MobileDeviceExtensionsGormRepository{
		db: db,
		qb: qb,
	}
}

func (r *MobileDeviceExtensionsGormRepository) FindById(deviceId, extensionId uuid.UUID) (*domain.MobileDeviceExtension, error) {
	var pairedExtension *domain.MobileDeviceExtension

	result := r.db.First(&pairedExtension, "device_id = ? and extension_id = ?", deviceId.String(), extensionId.String())

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, MobileDeviceExtensionCouldNotBeFound{DeviceId: deviceId.String(), ExtensionId: extensionId.String()}
	}

	return pairedExtension, nil
}

func (r *MobileDeviceExtensionsGormRepository) Delete(extension *domain.MobileDeviceExtension) error {
	sql, _, _ := r.qb.From("mobile_device_browser_extension").
		Where(
			goqu.C("device_id").Eq(extension.DeviceId.String()),
			goqu.C("extension_id").Eq(extension.ExtensionId.String()),
		).
		Delete().ToSQL()

	if result := r.db.Exec(sql); result.Error != nil {
		return errors.New("Could not delete extension: " + result.Error.Error())
	}

	return nil
}
