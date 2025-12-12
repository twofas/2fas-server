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
	"github.com/twofas/2fas-server/internal/common/http"
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

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}
	if err := c.BindJSON(&cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	r.cqrs.Commands.StoreLogEvent.Handle(c.Request.Context(), cmd)

	c.JSON(200, api.NewOk("Log has been stored"))
}

func (r *RoutesHandler) FindBrowserExtensionPairedMobileDevices(c *gin.Context) {
	cmd := &query.BrowserExtensionPairedDevicesQuery{}

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	result := r.cqrs.Queries.BrowserExtensionPairedDevicesQuery.Handle(cmd)

	c.JSON(200, result)
}

func (r *RoutesHandler) GetBrowserExtensionPairedMobileDevice(c *gin.Context) {
	cmd := &query.BrowserExtensionPairedDeviceQuery{}

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
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

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	logging.LogCommand(cmd)

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.RemoveExtensionPairedDevice.Handle(cmd)

	if err != nil {
		var extensionNotFoundErr adapters.BrowserExtensionsCouldNotBeFoundError
		var deviceNotFoundErr adapters.ExtensionDeviceCouldNotBeFoundError

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

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	logging.LogCommand(cmd)

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.RemoveAllExtensionPairedDevices.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.BrowserExtensionsCouldNotBeFoundError

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

	if err := c.BindUri(cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
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

	if err := c.BindJSON(cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.RegisterBrowserExtension.Handle(cmd)

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

	if err := c.BindJSON(&cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return
	}
	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.UpdateBrowserExtension.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.BrowserExtensionsCouldNotBeFoundError

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	queryCmd := &query.BrowserExtensionQuery{Id: c.Param("extension_id")}

	presenter, err := r.cqrs.Queries.BrowserExtensionQuery.Handle(queryCmd)

	if errors.Is(err, adapters.BrowserExtensionsCouldNotBeFoundError{}) {
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
	cmd, ok := command.New2FaTokenRequestFromGin(c)
	if !ok {
		// New2FaTokenRequestFromGin already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
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

	if err := c.BindJSON(&cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return
	}
	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.Close2FaRequest.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.TokenRequestCouldNotBeFoundError

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
	if err := c.BindUri(q); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	result, err := r.cqrs.Queries.BrowserExtension2FaRequestQuery.Handle(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) GetBrowserExtension2FaTokenRequest(c *gin.Context) {
	q := &query.BrowserExtension2FaRequestQuery{}
	if err := c.BindUri(q); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

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
