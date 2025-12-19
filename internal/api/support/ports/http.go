package ports

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	adapters3 "github.com/twofas/2fas-server/internal/api/support/adapters"
	"github.com/twofas/2fas-server/internal/api/support/app"
	"github.com/twofas/2fas-server/internal/api/support/app/command"
	"github.com/twofas/2fas-server/internal/api/support/app/queries"
	"github.com/twofas/2fas-server/internal/api/support/domain"
	"github.com/twofas/2fas-server/internal/common/api"
	"github.com/twofas/2fas-server/internal/common/http"
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

func (r *RoutesHandler) CreateDebugLogsAuditClaim(c *gin.Context) {
	id := uuid.New()

	cmd := &command.CreateDebugLogsAuditClaim{}
	cmd.Id = id

	logging.LogCommand(cmd)

	if err := c.BindJSON(cmd); err != nil {
		// c.BindJSON already returned 400 and error.
		return
	}

	err := r.cqrs.Commands.CreateDebugLogsAuditClain.Handle(cmd)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))

		return
	}

	q := &queries.DebugLogsAuditQuery{
		Id: id.String(),
	}

	presenter, err := r.cqrs.Queries.DebugLogsAuditQuery.Find(q)

	if err != nil {
		c.JSON(500, api.NewInternalServerError(err))

		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) CreateDebugLogsAudit(c *gin.Context) {
	cmd := &command.CreateDebugLogsAudit{}

	if err := c.BindUri(cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}
	if err := c.Bind(cmd); err != nil {
		// c.Bind already returned 400 and error.
		return
	}
	if cmd.File == nil {
		c.JSON(400, api.NewBadRequestError(errors.New("logs file is required")))
		return
	}
	if cmd.Id == "" {
		c.JSON(400, api.NewBadRequestError(errors.New("audit id is required")))
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	logging.LogCommand(cmd)

	err := r.cqrs.Commands.CreateDebugLogsAudit.Handle(cmd)
	if err != nil {
		r.handleError(c, err)
		return
	}

	q := &queries.DebugLogsAuditQuery{
		Id: cmd.Id,
	}

	presenter, err := r.cqrs.Queries.DebugLogsAuditQuery.Find(q)
	if err != nil {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) handleError(c *gin.Context, err error) {
	var notFoundErr adapters3.DebugLogsAuditCouldNotBeFoundError

	if errors.As(err, &notFoundErr) {
		c.JSON(404, api.NotFoundError(err))
		return
	}

	var expiredErr domain.DebugLogsAuditClaimIsHasBeenExpiredError

	if errors.As(err, &expiredErr) {
		c.JSON(410, api.GoneError(err))
		return
	}

	var completedErr domain.DebugLogsAuditClaimIsAlreadyCompletedError

	if errors.As(err, &completedErr) {
		c.JSON(410, api.GoneError(err))
		return
	}

	c.JSON(400, api.NewBadRequestError(err))
	return
}

func (r *RoutesHandler) UpdateDebugLogsAuditClaim(c *gin.Context) {
	cmd := &command.UpdateDebugLogsAudit{}

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

	logging.LogCommand(cmd)

	err := r.cqrs.Commands.UpdateDebugLogsAudit.Handle(cmd)

	if err != nil {
		var notFoundErr adapters3.DebugLogsAuditCouldNotBeFoundError

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	q := &queries.DebugLogsAuditQuery{
		Id: cmd.Id,
	}
	presenter, err := r.cqrs.Queries.DebugLogsAuditQuery.Find(q)
	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) DeleteDebugLogsAudit(c *gin.Context) {
	cmd := &command.DeleteDebugLogsAudit{}

	if err := c.BindUri(cmd); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	if ok := http.Validate(c, r.validator, cmd); !ok {
		// http.Validate already returned 400 and error.
		return
	}

	logging.LogCommand(cmd)

	err := r.cqrs.Commands.DeleteDebugLogsAudit.Handle(cmd)

	if err != nil {
		var notFoundErr adapters3.DebugLogsAuditCouldNotBeFoundError

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, api.NewOk("Debug logs audit has been removed."))
}

func (r *RoutesHandler) DeleteAllDebugLogsAudit(c *gin.Context) {
	cmd := &command.DeleteAllDebugLogsAudit{}

	r.cqrs.Commands.DeleteAllDebugLogsAudit.Handle(cmd)

	c.JSON(200, api.NewOk("Debug logs audit has been removed."))
}

func (r *RoutesHandler) GetDebugLogsAudit(c *gin.Context) {
	q := &queries.DebugLogsAuditQuery{}

	if err := c.BindUri(q); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	presenter, err := r.cqrs.Queries.DebugLogsAuditQuery.Find(q)

	if err != nil {
		var notFoundErr adapters3.DebugLogsAuditCouldNotBeFoundError

		if errors.As(err, &notFoundErr) {
			c.JSON(404, api.NotFoundError(err))
			return
		}

		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, presenter)
}

func (r *RoutesHandler) GetDebugAllLogsAudit(c *gin.Context) {
	q := &queries.DebugLogsAuditQuery{}
	if err := c.BindUri(q); err != nil {
		// c.BindUri already returned 400 and error.
		return
	}

	presenter, err := r.cqrs.Queries.DebugLogsAuditQuery.FindAll(q)

	if err != nil {
		c.JSON(400, api.NewBadRequestError(err))
		return
	}

	c.JSON(200, presenter)
}
