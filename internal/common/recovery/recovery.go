package recovery

import (
	"github.com/twofas/2fas-server/internal/common/logging"
)

func DoNotPanic(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			stack := stack(3)

			logging.WithFields(logging.Fields{
				"stack": string(stack),
				"error": err,
			}).Error("Panic")
		}
	}()

	fn()
}
