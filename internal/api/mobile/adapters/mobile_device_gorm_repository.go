package adapters

import (
	"errors"
	"fmt"
	"github.com/2fas/api/internal/api/mobile/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MobileDeviceCouldNotBeFound struct {
	DeviceId string
}

func (e MobileDeviceCouldNotBeFound) Error() string {
	return fmt.Sprintf("Mobile device could not be found: %s", e.DeviceId)
}

type MobileDeviceMysqlRepository struct {
	db *gorm.DB
}

func NewMobileDeviceMysqlRepository(db *gorm.DB) *MobileDeviceMysqlRepository {
	return &MobileDeviceMysqlRepository{db: db}
}

func (r *MobileDeviceMysqlRepository) Save(device *domain.MobileDevice) error {
	if err := r.db.Create(device).Error; err != nil {
		return err
	}

	return nil
}

func (r *MobileDeviceMysqlRepository) Update(device *domain.MobileDevice) error {
	if err := r.db.Updates(device).Error; err != nil {
		return err
	}

	return nil
}

func (r *MobileDeviceMysqlRepository) FindById(id uuid.UUID) (*domain.MobileDevice, error) {
	device := &domain.MobileDevice{}

	result := r.db.First(&device, "id = ?", id.String())

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, MobileDeviceCouldNotBeFound{DeviceId: id.String()}
	}

	return device, nil
}

func (r *MobileDeviceMysqlRepository) FindAll() []*domain.MobileDevice {
	var devices []*domain.MobileDevice

	r.db.Find(&devices)

	return devices
}
