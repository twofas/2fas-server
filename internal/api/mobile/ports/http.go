package ports

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	browser_adapters "github.com/twofas/2fas-server/internal/api/browser_extension/adapters"
	"github.com/twofas/2fas-server/internal/api/mobile/adapters"
	"github.com/twofas/2fas-server/internal/api/mobile/app"
	"github.com/twofas/2fas-server/internal/api/mobile/app/command"
	query "github.com/twofas/2fas-server/internal/api/mobile/app/queries"
	"github.com/twofas/2fas-server/internal/api/mobile/domain"
	"github.com/twofas/2fas-server/internal/common/api"
	"github.com/twofas/2fas-server/internal/common/http"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type RoutesHandler struct {
	cqrs                   *app.Cqrs
	validator              *validator.Validate
	mobileDeviceRepository domain.MobileDeviceRepository
}

func NewRoutesHandler(
	cqrs *app.Cqrs,
	validate *validator.Validate,
	repository domain.MobileDeviceRepository,
) *RoutesHandler {
	return &RoutesHandler{
		cqrs:                   cqrs,
		validator:              validate,
		mobileDeviceRepository: repository,
	}
}

func (r *RoutesHandler) UpdateMobileDevice(c *gin.Context) {
	cmd := &command.UpdateMobileDevice{}

	if err := c.BindUri(cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}
	if err := c.BindJSON(cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	logging.Info("Start command", cmd)

	err := r.cqrs.Commands.UpdateMobileDevice.Handle(cmd)

	if err != nil {
		var deviceNotFoundErr adapters.MobileDeviceCouldNotBeFoundError

		if errors.As(err, &deviceNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileDeviceQuery{
		Id: cmd.Id,
	}

	presenter, err := r.cqrs.Queries.MobileDeviceQuery.Handle(q)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RegisterMobileDevice(c *gin.Context) {
	id := uuid.New()

	cmd := &command.RegisterMobileDevice{
		Id: id,
	}

	if err := c.BindJSON(cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	logging.Info("Start command", cmd)

	err := r.cqrs.Commands.RegisterMobileDevice.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileDeviceQuery{
		Id: id.String(),
	}

	presenter, err := r.cqrs.Queries.MobileDeviceQuery.Handle(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RemoveAllMobileDevices(c *gin.Context) {
	cmd := &command.RemoveAllMobileDevices{}

	r.cqrs.Commands.RemoveAllMobileDevices.Handle(cmd)

	c.JSON(200, api.NewOk("Mobile devices have been removed."))
}

func (r *RoutesHandler) PairMobileWithExtension(c *gin.Context) {
	cmd := &command.PairMobileWithBrowserExtension{}

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

	err := r.cqrs.Commands.PairMobileWithExtension.Handle(c.Request.Context(), cmd)

	if err != nil {
		var conflictError domain.ExtensionHasAlreadyBeenPairedError

		if errors.As(err, &conflictError) {
			c.JSON(409, api.ConflictError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.PairedBrowserExtensionQuery{ExtensionId: cmd.ExtensionId}

	presenter, err := r.cqrs.Queries.PairedBrowserExtensionQuery.Handle(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) RemovePairingWithExtension(c *gin.Context) {
	cmd := &command.RemoveDevicePairedExtension{}

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.RemovePairingWithExtension.Handle(cmd)

	if err != nil {
		var deviceNotFoundErr *adapters.MobileDeviceCouldNotBeFoundError
		var extensionsNotFoundErr *browser_adapters.BrowserExtensionsCouldNotBeFoundError

		if errors.As(err, &deviceNotFoundErr) || errors.As(err, &extensionsNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, api.NewOk("Extension has been disconnected from device."))
}

func (r *RoutesHandler) FindAllMobileAppExtensions(c *gin.Context) {
	cmd := &query.DeviceBrowserExtensionsQuery{}

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	deviceId, _ := uuid.Parse(cmd.DeviceId)
	_, err := r.mobileDeviceRepository.FindById(deviceId)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	result, err := r.cqrs.Queries.DeviceBrowserExtensionsQuery.Handle(cmd)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindMobileAppExtensionById(c *gin.Context) {
	cmd := &query.DeviceBrowserExtensionsQuery{}

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	deviceId, _ := uuid.Parse(cmd.DeviceId)
	_, err := r.mobileDeviceRepository.FindById(deviceId)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	result, err := r.cqrs.Queries.DeviceBrowserExtensionsQuery.Handle(cmd)

	if len(result) == 0 {
		c.JSON(404, api.NotFoundError(browser_adapters.BrowserExtensionsCouldNotBeFoundError{ExtensionId: cmd.ExtensionId}))
		return
	}

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, result[0])
}

func (r *RoutesHandler) Send2FaToken(c *gin.Context) {
	cmd := &command.Send2FaToken{}

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

	deviceId, _ := uuid.Parse(cmd.DeviceId)
	_, err := r.mobileDeviceRepository.FindById(deviceId)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	err = r.cqrs.Commands.Send2FaToken.Handle(c.Request.Context(), cmd)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))
		return
	}

	c.JSON(200, api.NewOk("Token has been sent to browser extension"))
}

func (r *RoutesHandler) GetAll2FaTokenRequests(c *gin.Context) {
	q := &query.DeviceBrowserExtension2FaRequestQuery{}
	if err := c.BindUri(&q); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, q); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	deviceId, _ := uuid.Parse(q.DeviceId)
	_, err := r.mobileDeviceRepository.FindById(deviceId)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	result, err := r.cqrs.Queries.DeviceBrowserExtension2FaRequestQuery.Handle(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) CreateMobileNotification(c *gin.Context) {
	id := uuid.New()

	cmd := &command.CreateNotification{Id: id}

	if err := c.BindJSON(&cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.CreateNotification.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileNotificationsQuery{Id: id.String()}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindOne(q)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) UpdateMobileNotification(c *gin.Context) {
	cmd := &command.UpdateNotification{}

	if err := c.BindUri(cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}
	if err := c.BindJSON(cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.UpdateNotification.Handle(cmd)

	if err != nil {
		var notificationNotFoundErr *adapters.MobileNotificationCouldNotBeFoundError

		if errors.As(err, &notificationNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileNotificationsQuery{Id: cmd.Id}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindOne(q)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindAllMobileNotifications(c *gin.Context) {
	q := &query.MobileNotificationsQuery{}

	if err := c.BindUri(&q); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}
	if err := c.BindQuery(&q); err != nil {
		// c.BindQuery already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, q); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindAll(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) FindMobileNotification(c *gin.Context) {
	q := &query.MobileNotificationsQuery{}

	if err := c.BindUri(&q); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, q); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindOne(q)

	if err != nil {
		var notificationNotFoundErr adapters.MobileNotificationCouldNotBeFoundError

		if errors.As(err, &notificationNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, result)
}

func (r *RoutesHandler) RemoveMobileNotification(c *gin.Context) {
	cmd := &command.DeleteNotification{}

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.DeleteNotification.Handle(cmd)

	if err != nil {
		var notFoundErr adapters.MobileNotificationCouldNotBeFoundError

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, api.NewOk("Notification has been removed."))
}

func (r *RoutesHandler) RemoveAllMobileNotifications(c *gin.Context) {
	cmd := &command.DeleteAllNotifications{}

	r.cqrs.Commands.RemoveAllMobileNotifications.Handle(cmd)

	c.JSON(200, api.NewOk("Mobile notifications has been removed."))
}

func (r *RoutesHandler) PublishMobileNotification(c *gin.Context) {
	cmd := &command.PublishNotification{}

	if err := c.BindUri(&cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.PublishNotification.Handle(cmd)

	if err != nil {
		var notificationNotFoundErr adapters.MobileNotificationCouldNotBeFoundError

		if errors.As(err, &notificationNotFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &query.MobileNotificationsQuery{Id: cmd.Id}

	result, err := r.cqrs.Queries.MobileNotificationsQuery.FindOne(q)

	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	c.JSON(200, result)
}
