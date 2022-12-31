package app

import (
	"github.com/2fas/api/internal/api/icons/app/command"
	"github.com/2fas/api/internal/api/icons/app/queries"
)

type Commands struct {
	CreateWebService     *command.CreateWebServiceHandler
	UpdateWebService     *command.UpdateWebServiceHandler
	RemoveWebService     *command.DeleteWebServiceHandler
	RemoveAllWebServices *command.DeleteAllWebServicesHandler

	CreateIcon     *command.CreateIconHandler
	UpdateIcon     *command.UpdateIconHandler
	RemoveIcon     *command.DeleteIconHandler
	RemoveAllIcons *command.DeleteAllIconsHandler

	CreateIconRequest                *command.CreateIconRequestHandler
	RemoveIconRequest                *command.DeleteIconRequestHandler
	RemoveAllIconsRequests           *command.DeleteAllIconsRequestsHandler
	UpdateWebServiceFromIconRequest  *command.UpdateWebServiceFromIconRequestHandler
	TransformIconRequestToWebService *command.TransformIconRequestToWebServiceHandler

	CreateIconsCollection     *command.CreateIconsCollectionHandler
	UpdateIconsCollection     *command.UpdateIconsCollectionHandler
	RemoveIconsCollection     *command.DeleteIconsCollectionHandler
	RemoveAllIconsCollections *command.DeleteAllIconsCollectionsHandler
}

type Queries struct {
	WebServiceQuery      *queries.WebServiceQueryHandler
	WebServicesDumpQuery *queries.WebServicesDumpQueryHandler
	IconQuery            *queries.IconQueryHandler
	IconRequestQuery     *queries.IconRequestQueryHandler
	IconsCollectionQuery *queries.IconsCollectionQueryHandler
}

type Cqrs struct {
	Commands Commands
	Queries  Queries
}
