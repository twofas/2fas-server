package pairing

import (
	"time"
)

type MobileDevice struct {
	DeviceID string
	FCMToken string
}

type PairingInfo struct {
	Device   MobileDevice
	PairedAt time.Time
}

func (pi *PairingInfo) IsPaired() bool {
	return !pi.PairedAt.IsZero()
}
