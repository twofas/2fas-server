package ports

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/twofas/2fas-server/internal/api/browser_extension/adapters"
	"github.com/twofas/2fas-server/internal/api/browser_extension/app"
	"github.com/twofas/2fas-server/internal/api/browser_extension/app/command"
	"github.com/twofas/2fas-server/internal/api/browser_extension/app/query"
	"github.com/twofas/2fas-server/internal/api/browser_extension/domain"
	"github.com/twofas/2fas-server/internal/common/api"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type RoutesHandler struct {
	cqrs      *app.Cqrs
	validator *validator.Validate
}

func NewRoutesHandler(cqrs *app.Cqrs, validate *validator.Validate) *RoutesHandler {
	return &RoutesHandler{
		cqrs:      cqrs,
		validator: validate,
	}
}

func (r *RoutesHandler) Log(c *gin.Context) {
	cmd := &command.StoreLogEvent{}

	c.ShouldBindUri(&cmd)
	c.ShouldBindJSON(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	r.cqrs.Commands.StoreLogEvent.Handle(c.Request.Context(), cmd)

	c.JSON(200, api.NewOk("Log has been stored"))
}

func (r *RoutesHandler) FindBrowserExtensionPairedMobileDevices(c *gin.Context) {
	cmd := &query.BrowserExtensionPairedDevicesQuery{}

	c.ShouldBindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	result := r.cqrs.Queries.BrowserExtensionPairedDevicesQuery.Handle(cmd)

	c.JSON(200, result)
}

func (r *RoutesHandler) GetBrowserExtensionPairedMobileDevice(c *gin.Context) {
	cmd := &query.BrowserExtensionPairedDeviceQuery{}

	c.ShouldBindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	result, err := r.cqrs.Queries.BrowserExtensionPairedDeviceQuery.Handle(cmd)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, api.NotFoundError(err))
			return
		}
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) RemovePairedDeviceFromExtension(c *gin.Context) {
	cmd := &command.RemoveExtensionPairedDevice{}

	c.BindUri(&cmd)

	logging.LogCommand(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	err = r.cqrs.Commands.RemoveExtensionPairedDevice.Handle(cmd)

	if err != nil {
		var extensionNotFoundErr adapters.BrowserExtensionsCouldNotBeFound
		var deviceNotFoundErr adapters.ExtensionDeviceCouldNotBeFound

		if errors.As(err, &deviceNotFoundErr) || errors.As(err, &extensionNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, api.NewOk("Extension device has been removed"))
}

func (r *RoutesHandler) RemoveAllExtensionPairedDevices(c *gin.Context) {
	cmd := &command.RemoveAllExtensionPairedDevices{}

	c.BindUri(&cmd)

	logging.LogCommand(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.RemoveAllExtensionPairedDevices.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.BrowserExtensionsCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, api.NewOk("Extension devices has been removed"))
}

func (r *RoutesHandler) RemoveAllBrowserExtensions(c *gin.Context) {
	cmd := &command.RemoveAllBrowserExtensions{}

	r.cqrs.Commands.RemoveAllBrowserExtensions.Handle(cmd)

	c.JSON(200, api.NewOk("Browser extensions have been removed."))
}

func (r *RoutesHandler) RemoveAllBrowserExtensionsDevices(c *gin.Context) {
	cmd := &command.RemoveAllBrowserExtensionsDevices{}

	r.cqrs.Commands.RemoveAllBrowserExtensionsDevices.Handle(cmd)

	c.JSON(200, api.NewOk("Browser extensions devices have been removed."))
}

func (r *RoutesHandler) FindBrowserExtension(c *gin.Context) {
	cmd := &query.BrowserExtensionQuery{}

	c.ShouldBindUri(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	browserExtension, err := r.cqrs.Queries.BrowserExtensionQuery.Handle(cmd)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, browserExtension)
}

func (r *RoutesHandler) RegisterBrowserExtension(c *gin.Context) {
	id := uuid.New()

	cmd := &command.RegisterBrowserExtension{
		BrowserExtensionId: id,
	}

	c.BindJSON(cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	err = r.cqrs.Commands.RegisterBrowserExtension.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	queryCmd := &query.BrowserExtensionQuery{Id: id.String()}

	presenter, err := r.cqrs.Queries.BrowserExtensionQuery.Handle(queryCmd)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) UpdateBrowserExtension(c *gin.Context) {
	cmd := &command.UpdateBrowserExtension{}

	c.BindJSON(&cmd)
	c.BindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	err = r.cqrs.Commands.UpdateBrowserExtension.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.BrowserExtensionsCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	queryCmd := &query.BrowserExtensionQuery{Id: c.Param("extension_id")}

	presenter, err := r.cqrs.Queries.BrowserExtensionQuery.Handle(queryCmd)

	if errors.Is(err, adapters.BrowserExtensionsCouldNotBeFound{}) {
		c.JSON(404, api.NotFoundError(err))

		return
	}

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) Request2FaToken(c *gin.Context) {
	cmd := command.New2FaTokenRequestFromGin(c)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		c.JSON(400, api.NewBadRequestError(validationErrors))

		return
	}

	pushResult, err := r.cqrs.Commands.Request2FaToken.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	q := &query.BrowserExtension2FaRequestQuery{
		ExtensionId:    c.Param("extension_id"),
		TokenRequestId: cmd.Id,
	}

	result, err := r.cqrs.Queries.BrowserExtension2FaRequestQuery.Handle(q)
	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	var jsonResult struct {
		query.BrowserExtension2FaRequestPresenter
		PushStatus map[string]command.PushNotificationStatus `json:"push_status"`
	}
	if len(result) >= 0 && result[0] != nil {
		jsonResult.BrowserExtension2FaRequestPresenter = *result[0]
	}
	jsonResult.PushStatus = pushResult

	c.JSON(200, jsonResult)
}

func (r *RoutesHandler) Close2FaRequest(c *gin.Context) {
	cmd := &command.Close2FaRequest{}

	c.BindJSON(&cmd)
	c.BindUri(&cmd)

	err := r.validator.Struct(cmd)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(400, api.NewBadRequestError(validationErrors))
		return
	}

	err = r.cqrs.Commands.Close2FaRequest.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.TokenRequestCouldNotBeFound

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	q := &query.BrowserExtension2FaRequestQuery{
		ExtensionId:    cmd.ExtensionId,
		TokenRequestId: cmd.TokenRequestId,
	}

	presenter, _ := r.cqrs.Queries.BrowserExtension2FaRequestQuery.Handle(q)

	c.JSON(200, presenter[0])
}

func (r *RoutesHandler) GetAllBrowserExtension2FaTokenRequests(c *gin.Context) {
	q := &query.BrowserExtension2FaRequestQuery{
		Status: domain.Pending,
	}
	c.BindUri(q)

	result, err := r.cqrs.Queries.BrowserExtension2FaRequestQuery.Handle(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) GetBrowserExtension2FaTokenRequest(c *gin.Context) {
	q := &query.BrowserExtension2FaRequestQuery{}
	c.BindUri(q)

	result, err := r.cqrs.Queries.BrowserExtension2FaRequestQuery.Handle(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	if len(result) == 0 {
		c.JSON(404, api.NotFoundError(errors.New("Token request could not be found")))
		return
	}

	c.JSON(200, result[0])
}
