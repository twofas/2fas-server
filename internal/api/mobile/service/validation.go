package service

import (
	"github.com/2fas/api/internal/api/mobile/domain"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func DeviceIdExistsValidator(repository domain.MobileDeviceRepository) validator.Func {
	return func(fl validator.FieldLevel) bool {
		id, err := uuid.Parse(fl.Field().String())

		if err != nil {
			return false
		}

		_, err = repository.FindById(id)

		if err != nil {
			return false
		}

		return true
	}
}
