package systemlog

import (
	"net/http"

	"auth-service/pkg/base"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/response"
	"auth-service/pkg/validations"
	systemLogService "auth-service/services/systemlog"

	"github.com/gin-gonic/gin"
)

// ControllerInterface defines the system log controller.
type ControllerInterface interface {
	List(c *gin.Context)
}

// Controller handles GET /system-logs.
type Controller struct {
	*base.BaseController
	service systemLogService.ServiceInterface
}

// NewController creates the controller.
func NewController(validator *validations.Validator, service systemLogService.ServiceInterface) ControllerInterface {
	return &Controller{BaseController: base.NewBaseController(validator), service: service}
}

// listResponse is the 200 body.
type listResponse struct {
	Success      bool                       `json:"success"`
	Data         []systemLogService.LogItem `json:"data"`
	NextCursor   *string                    `json:"next_cursor"`
	LatestCursor *string                    `json:"latest_cursor"`
}

// List returns the admin user activity log, newest first.
// @Summary List system activity logs
// @Description Requires the view_system_logs permission. ip_address is masked (a.b.x.x) unless the caller also holds export_system_logs. Page older with cursor=next_cursor; poll for new rows with since_cursor=latest_cursor.
// @Tags system-logs
// @Produce json
// @Security Bearer
// @Param limit query int false "Page size (default 50, max 200)"
// @Param cursor query string false "next_cursor of the previous page"
// @Param since_cursor query string false "latest_cursor of the previous response; returns newer rows"
// @Param action query string false "Action, or comma-separated actions (e.g. user.login,user.login_failed)"
// @Param actor_user_id query string false "Actor user UUID"
// @Param result query string false "success or failure"
// @Param from query string false "RFC3339 lower bound on occurred_at"
// @Param to query string false "RFC3339 upper bound on occurred_at"
// @Param q query string false "Search actor_username and summary"
// @Success 200 {object} listResponse
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/system-logs [get]
func (ctrl *Controller) List(c *gin.Context) {
	// Activity data must never be stored by a browser or a gateway cache.
	c.Header("Cache-Control", "no-store")

	rolesValue, _ := c.Get("roles")
	roles, _ := rolesValue.([]string)
	access, err := ctrl.service.ResolveAccess(roles)
	if err != nil {
		writeError(c, err)
		return
	}
	if !access.CanView {
		response.Forbidden(c, "view_system_logs permission required")
		return
	}

	res, err := ctrl.service.List(c.Request.Context(), access, systemLogService.ListQuery{
		Limit:       c.Query("limit"),
		Cursor:      c.Query("cursor"),
		SinceCursor: c.Query("since_cursor"),
		Action:      c.Query("action"),
		ActorUserID: c.Query("actor_user_id"),
		Result:      c.Query("result"),
		From:        c.Query("from"),
		To:          c.Query("to"),
		Q:           c.Query("q"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, listResponse{
		Success:      true,
		Data:         res.Data,
		NextCursor:   res.NextCursor,
		LatestCursor: res.LatestCursor,
	})
}

func writeError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		msg := appErr.Message
		if appErr.StatusCode >= http.StatusInternalServerError {
			msg = "Failed to fetch system logs"
		}
		response.Error(c, appErr.StatusCode, appErr.Code, msg, "")
		return
	}
	response.InternalServerError(c, "Failed to fetch system logs")
}
