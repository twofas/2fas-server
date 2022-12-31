package app

import (
	"github.com/2fas/api/internal/api/browser_extension/app/command"
	"github.com/2fas/api/internal/api/browser_extension/app/query"
)

type Commands struct {
	RegisterBrowserExtension          command.RegisterBrowserExtensionHandler
	RemoveAllBrowserExtensions        command.RemoveAllBrowserExtensionsHandler
	RemoveAllBrowserExtensionsDevices command.RemoveAllBrowserExtensionsDevicesHandler
	UpdateBrowserExtension            command.UpdateBrowserExtensionHandler
	Request2FaToken                   command.Request2FaTokenHandler
	Close2FaRequest                   command.Close2FaRequestHandler
	RemoveExtensionPairedDevice       command.RemoveExtensionPairedDeviceHandler
	RemoveAllExtensionPairedDevices   command.RemoveALlExtensionPairedDevicesHandler
	StoreLogEvent                     command.StoreLogEventHandler
}

type Queries struct {
	BrowserExtensionQuery              query.BrowserExtensionQueryHandler
	BrowserExtensionPairedDevicesQuery query.BrowserExtensionPairedMobileDevicesQueryHandler
	BrowserExtensionPairedDeviceQuery  query.BrowserExtensionPairedMobileDeviceQueryHandler
	BrowserExtension2FaRequestQuery    query.BrowserExtension2FaRequestQueryHandler
}

type Cqrs struct {
	Commands Commands
	Queries  Queries
}
