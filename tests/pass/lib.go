package pass

import (
	"os"

	"github.com/google/uuid"
)

func getDeviceID() string {
	deviceID := uuid.NewString()
	if deviceIDFromEnv := os.Getenv("TEST_DEVICE_ID"); deviceIDFromEnv != "" {
		deviceID = deviceIDFromEnv
	}
	return deviceID
}
