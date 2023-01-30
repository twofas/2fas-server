package app

import (
	"github.com/twofas/2fas-server/internal/api/mobile/app/command"
	"github.com/twofas/2fas-server/internal/api/mobile/app/queries"
)

type Commands struct {
	RegisterMobileDevice   *command.RegisterMobileDeviceHandler
	RemoveAllMobileDevices *command.RemoveAllMobileDevicesHandler
	UpdateMobileDevice     *command.UpdateMobileDeviceHandler

	CreateNotification           *command.CreateNotificationHandler
	UpdateNotification           *command.UpdateNotificationHandler
	DeleteNotification           *command.DeleteNotificationHandler
	RemoveAllMobileNotifications *command.DeleteAllNotificationsHandler
	PublishNotification          *command.PublishNotificationHandler

	PairMobileWithExtension    *command.PairMobileWithExtensionHandler
	RemovePairingWithExtension *command.RemoveDeviceExtensionHandler
	Send2FaToken               *command.Send2FaTokenHandler
}

type Queries struct {
	MobileDeviceQuery                     *query.MobileDeviceQueryHandler
	DeviceBrowserExtensionsQuery          *query.DeviceBrowserExtensionsQueryHandler
	DeviceBrowserExtension2FaRequestQuery *query.DeviceBrowserExtension2FaRequestQueryHandler
	PairedBrowserExtensionQuery           *query.PairedBrowserExtensionQueryHandler
	MobileNotificationsQuery              *query.MobileNotificationsQueryHandler
}

type Cqrs struct {
	Commands Commands
	Queries  Queries
}
