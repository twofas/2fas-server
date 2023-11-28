package ports

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/api/icons/adapters"
	"github.com/twofas/2fas-server/internal/api/icons/app"
	"github.com/twofas/2fas-server/internal/api/icons/app/command"
	"github.com/twofas/2fas-server/internal/api/icons/app/queries"
	"github.com/twofas/2fas-server/internal/api/icons/domain"
	"github.com/twofas/2fas-server/internal/common/api"
	"github.com/twofas/2fas-server/internal/common/db"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type RoutesHandler struct {
	cqrs      *app.Cqrs
	validator *validator.Validate
}

func NewRoutesHandler(
	cqrs *app.Cqrs,
	validate *validator.Validate,
) *RoutesHandler {
	return &RoutesHandler{
		cqrs:      cqrs,
		validator: validate,
	}
}

func (r *RoutesHandler) CreateWebService(c *gin.Context) {
	id := uuid.New()

	cmd := &command.CreateWebService{
		Id: id,
	}

	c.BindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.CreateWebService.Handle(cmd)
	if err != nil {
		if db.IsDBError(err) {
			logging.LogCommandFailed(cmd, err)
			c.JSON(500, api.NewInternalServerError(err))
			return
		}

		var conflictErr domain.WebServiceAlreadyExistsError

		if errors.As(err, &conflictErr) {
			c.JSON(409, api.ConflictError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &queries.WebServiceQuery{
		Id: id.String(),
	}

	presenter, err := r.cqrs.Queries.WebServiceQuery.FindOne(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) UpdateWebService(c *gin.Context) {
	cmd := &command.UpdateWebService{}

	c.ShouldBindUri(cmd)
	c.ShouldBindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		var notFoundErr adapters.WebServiceCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.UpdateWebService.Handle(cmd)
	if err != nil {
		if db.IsDBError(err) {
			logging.LogCommandFailed(cmd, err)
			c.JSON(500, api.NewInternalServerError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &queries.WebServiceQuery{
		Id: cmd.Id,
	}

	presenter, err := r.cqrs.Queries.WebServiceQuery.FindOne(q)
	if err != nil {
		if db.IsDBError(err) {
			logging.LogCommandFailed(cmd, err)
			c.JSON(500, api.NewInternalServerError(err))
			return
		}

		c.JSON(404, api.NotFoundError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RemoveWebService(c *gin.Context) {
	cmd := &command.DeleteWebService{}

	c.BindUri(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.RemoveWebService.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.WebServiceCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, api.NewOk("Web service has been removed."))
}

func (r *RoutesHandler) RemoveAllWebServices(c *gin.Context) {
	cmd := &command.DeleteAllWebServices{}

	r.cqrs.Commands.RemoveAllWebServices.Handle(cmd)

	c.JSON(200, api.NewOk("Web services have been removed."))
}

func (r *RoutesHandler) FindWebService(c *gin.Context) {
	q := &queries.WebServiceQuery{}

	c.BindUri(q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.WebServiceQuery.FindOne(q)

	if err != nil {
		var notFoundErr adapters.WebServiceCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindAllWebServices(c *gin.Context) {
	q := &queries.WebServiceQuery{}

	c.BindQuery(q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.WebServiceQuery.FindAll(q)

	if err != nil {
		if db.IsDBError(err) {
			c.JSON(500, api.NewInternalServerError(err))
			return
		}
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) DumpWebServices(c *gin.Context) {
	q := &queries.WebServicesDumpQuery{}

	c.BindQuery(q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.WebServicesDumpQuery.Dump(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) CreateIcon(c *gin.Context) {
	id := uuid.New()

	cmd := &command.CreateIcon{
		Id: id,
	}

	c.BindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.CreateIcon.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	q := &queries.IconQuery{
		Id: id.String(),
	}

	presenter, err := r.cqrs.Queries.IconQuery.FindOne(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) UpdateIcon(c *gin.Context) {
	cmd := &command.UpdateIcon{}

	c.BindUri(cmd)
	c.BindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	logging.LogCommand(cmd)

	if err != nil {
		var iconNotFound *adapters.IconCouldNotBeFound

		if errors.As(err, &iconNotFound) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.UpdateIcon.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	q := &queries.IconQuery{
		Id: cmd.Id,
	}

	presenter, err := r.cqrs.Queries.IconQuery.FindOne(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RemoveIcon(c *gin.Context) {
	cmd := &command.DeleteIcon{}

	c.BindUri(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.RemoveIcon.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.IconCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, api.NewOk("Icon has been removed."))
}

func (r *RoutesHandler) RemoveAllIcons(c *gin.Context) {
	cmd := &command.DeleteAllIcons{}

	r.cqrs.Commands.RemoveAllIcons.Handle(cmd)

	c.JSON(200, api.NewOk("Icons have been removed."))
}

func (r *RoutesHandler) FindIcon(c *gin.Context) {
	q := &queries.IconQuery{}

	c.BindUri(q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.IconQuery.FindOne(q)

	if err != nil {
		var notFoundErr adapters.IconCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindAllIcons(c *gin.Context) {
	q := &queries.IconQuery{}

	c.BindQuery(q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.IconQuery.FindAll(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) CreateIconRequest(c *gin.Context) {
	id := uuid.New()

	cmd := &command.CreateIconRequest{
		Id: id,
	}

	c.BindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.CreateIconRequest.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	q := &queries.IconRequestQuery{
		Id: id.String(),
	}

	presenter, err := r.cqrs.Queries.IconRequestQuery.FindOne(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RemoveIconRequest(c *gin.Context) {
	cmd := &command.DeleteIconRequest{}

	c.BindUri(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.RemoveIconRequest.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.IconRequestCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, api.NewOk("Icon has been removed."))
}

func (r *RoutesHandler) RemoveAllIconsRequests(c *gin.Context) {
	r.cqrs.Commands.RemoveAllIconsRequests.Handle()

	c.JSON(200, api.NewOk("Icons requests have been removed."))
}

func (r *RoutesHandler) UpdateWebServiceFromIconRequest(c *gin.Context) {
	cmd := &command.UpdateWebServiceFromIconRequest{}

	c.BindUri(cmd)
	c.BindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		logging.LogCommandFailed(cmd, err)
		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.UpdateWebServiceFromIconRequest.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		logging.LogCommandFailed(cmd, err)
		return
	}

	q := &queries.WebServiceQuery{
		Id: cmd.WebServiceId,
	}

	presenter, err := r.cqrs.Queries.WebServiceQuery.FindOne(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) TransformToWebService(c *gin.Context) {
	webServiceId := uuid.New()

	cmd := &command.TransformIconRequestToWebService{
		WebServiceId: webServiceId,
	}

	c.BindUri(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.TransformIconRequestToWebService.Handle(cmd)

	if err != nil {
		var conflictErr domain.WebServiceAlreadyExistsError

		if errors.As(err, &conflictErr) {
			c.JSON(409, api.ConflictError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &queries.WebServiceQuery{
		Id: webServiceId.String(),
	}

	presenter, err := r.cqrs.Queries.WebServiceQuery.FindOne(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) FindIconRequest(c *gin.Context) {
	q := &queries.IconRequestQuery{}

	c.BindUri(q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.IconRequestQuery.FindOne(q)

	if err != nil {
		var notFoundErr adapters.IconRequestCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindAllIconsRequests(c *gin.Context) {
	q := &queries.IconRequestQuery{}

	c.BindQuery(q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.IconRequestQuery.FindAll(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) CreateIconsCollection(c *gin.Context) {
	id := uuid.New()

	cmd := &command.CreateIconsCollection{
		Id: id,
	}

	c.BindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	logging.LogCommand(cmd)

	err = r.cqrs.Commands.CreateIconsCollection.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &queries.IconsCollectionQuery{
		Id: id.String(),
	}

	presenter, err := r.cqrs.Queries.IconsCollectionQuery.FindOne(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) UpdateIconsCollection(c *gin.Context) {
	cmd := &command.UpdateIconsCollection{}

	c.BindUri(&cmd)
	c.BindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	logging.LogCommand(cmd)

	if err != nil {
		var notFound *adapters.IconsCollectionCouldNotBeFound

		if errors.As(err, &notFound) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.UpdateIconsCollection.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &queries.IconsCollectionQuery{
		Id: cmd.Id,
	}

	presenter, err := r.cqrs.Queries.IconsCollectionQuery.FindOne(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RemoveIconsCollection(c *gin.Context) {
	cmd := &command.DeleteIconsCollection{}

	c.BindUri(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.RemoveIconsCollection.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.IconsCollectionCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, api.NewOk("Icons collection has been removed."))
}

func (r *RoutesHandler) FindIconsCollection(c *gin.Context) {
	q := &queries.IconsCollectionQuery{}

	c.BindUri(q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.IconsCollectionQuery.FindOne(q)

	if err != nil {
		var notFoundErr adapters.IconsCollectionCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindAllIconsCollection(c *gin.Context) {
	q := &queries.IconsCollectionQuery{}

	c.BindQuery(q)

	err := r.validator.Struct(q)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	result, err := r.cqrs.Queries.IconsCollectionQuery.FindAll(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) RemoveAllIconsCollections(c *gin.Context) {
	cmd := &command.DeleteAllIconsCollections{}

	r.cqrs.Commands.RemoveAllIconsCollections.Handle(cmd)

	c.JSON(200, api.NewOk("Icons collections has been removed."))
}
