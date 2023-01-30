package domain

import (
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/storage"
	"os"
)

type FcmPushConfig struct {
	FcmApiServiceAccountFile *os.File
}

func NewFcmPushConfig(fs storage.FileSystemStorage) *FcmPushConfig {
	file, err := fs.Get("/2fas-api/service_account_key.json")

	if err != nil {
		logging.Fatal(err)
	}

	return &FcmPushConfig{
		FcmApiServiceAccountFile: file,
	}
}
