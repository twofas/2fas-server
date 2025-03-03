package domain

import (
	"fmt"
	"os"

	"github.com/twofas/2fas-server/internal/common/storage"
)

type FcmPushConfig struct {
	FcmApiServiceAccountFile *os.File
}

func NewFcmPushConfig(fs storage.FileSystemStorage) (*FcmPushConfig, error) {
	file, err := fs.Get("/2fas-api/service_account_key.json")

	if err != nil {
		return nil, fmt.Errorf("failed to get FCM service account file: %w", err)
	}

	return &FcmPushConfig{
		FcmApiServiceAccountFile: file,
	}, nil
}
