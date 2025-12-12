package adapters

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/mobile/domain"
)

type MobileNotificationCouldNotBeFoundError struct {
	NotificationId string
}

func (e MobileNotificationCouldNotBeFoundError) Error() string {
	return fmt.Sprintf("Mobile notification could not be found: %s", e.NotificationId)
}

type MobileNotificationMysqlRepository struct {
	db *gorm.DB
}

func NewMobileNotificationMysqlRepository(db *gorm.DB) *MobileNotificationMysqlRepository {
	return &MobileNotificationMysqlRepository{db: db}
}

func (r *MobileNotificationMysqlRepository) Save(notification *domain.MobileNotification) error {
	if err := r.db.Create(notification).Error; err != nil {
		return err
	}

	return nil
}

func (r *MobileNotificationMysqlRepository) Update(notification *domain.MobileNotification) error {
	if err := r.db.Updates(notification).Error; err != nil {
		return err
	}

	return nil
}

func (r *MobileNotificationMysqlRepository) Delete(notification *domain.MobileNotification) error {
	if err := r.db.Delete(notification).Error; err != nil {
		return err
	}

	return nil
}

func (r *MobileNotificationMysqlRepository) FindById(id uuid.UUID) (*domain.MobileNotification, error) {
	notification := &domain.MobileNotification{}

	result := r.db.First(&notification, "id = ?", id.String())

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, MobileNotificationCouldNotBeFoundError{NotificationId: id.String()}
	}

	return notification, nil
}

func (r *MobileNotificationMysqlRepository) FindAll() []*domain.MobileNotification {
	var notifications []*domain.MobileNotification

	r.db.Find(&notifications)

	return notifications
}
