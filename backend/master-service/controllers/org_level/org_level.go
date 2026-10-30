package org_level

import (
	"errors"
	"net/http"
	"strconv"

	"master-service/pkg/base"
	"master-service/pkg/orgmode"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	orgLevelSvc "master-service/services/org_level"

	"github.com/gin-gonic/gin"
)

// OrgLevelControllerInterface serves /org-levels (flexible tenants only; the
// routes are guarded by middleware.RequireFlexible).
type OrgLevelControllerInterface interface {
	List(ctx *gin.Context)
	Preview(ctx *gin.Context)
	Apply(ctx *gin.Context)
	ChangeLog(ctx *gin.Context)
}

type OrgLevelController struct {
	*base.BaseController
	svc orgLevelSvc.ServiceInterface
}

func NewOrgLevelController(svc orgLevelSvc.ServiceInterface, validator *validations.Validator) OrgLevelControllerInterface {
	return &OrgLevelController{BaseController: base.NewBaseController(validator), svc: svc}
}

// List: GET /org-levels -> {levels:[{id,level,name,name_key,is_active,...}], max_levels}
func (c *OrgLevelController) List(ctx *gin.Context) {
	levels, err := c.svc.List(ctx.Request.Context())
	if err != nil {
		c.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Organisation levels fetched successfully", gin.H{"levels": levels, "max_levels": orgmode.MaxLevels})
}

// Preview: POST /org-levels/changes/preview (read-only dry run).
func (c *OrgLevelController) Preview(ctx *gin.Context) {
	var req orgLevelSvc.ChangeRequest
	if !c.ValidateRequest(ctx, &req) {
		return
	}
	pv, err := c.svc.Preview(ctx.Request.Context(), req)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Organisation level change previewed", pv)
}

// Apply: POST /org-levels/changes/apply (ADMIN/EXECUTIVE + X-Step-Up-Token).
func (c *OrgLevelController) Apply(ctx *gin.Context) {
	var req orgLevelSvc.ApplyRequest
	if !c.ValidateRequest(ctx, &req) {
		return
	}
	actor := orgLevelSvc.Actor{
		Username:  c.GetUsername(ctx),
		Roles:     c.GetRoles(ctx),
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	}
	actor.UserID, _ = c.GetUserID(ctx)
	res, err := c.svc.Apply(ctx.Request.Context(), req, actor)
	if err != nil {
		var verr *orgLevelSvc.ViolationsError
		if errors.As(err, &verr) {
			// 422 with the violations list, so the UI can show why.
			ctx.JSON(http.StatusUnprocessableEntity, response.Response{
				Success: false,
				Code:    verr.Code,
				Data:    gin.H{"violations": verr.Violations},
				Error:   &response.ErrorDetail{Code: verr.Code, Message: verr.Message},
			})
			return
		}
		c.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Organisation levels updated", res)
}

// ChangeLog: GET /org-levels/changes?page=&page_size= (newest first).
func (c *OrgLevelController) ChangeLog(ctx *gin.Context) {
	page, pageSize := 1, 20
	if p, err := strconv.Atoi(ctx.Query("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(ctx.Query("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}
	rows, total, err := c.svc.ChangeLog(ctx.Request.Context(), (page-1)*pageSize, pageSize)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Organisation change history fetched successfully", gin.H{
		"changes": rows,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}
