package activityplan

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/logger"
	"audit-service/pkg/response"
	form "audit-service/services/activity_plan_form"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FormController serves /activity-plans with the Formulir Rencana Kegiatan
// Audit contract (docs/audit-activity-plan-rework.md section 6). The rules
// live in services/activity_plan_form; this layer only parses and responds.
type FormController struct {
	DB *gorm.DB
	// Now is the clock (tests); nil means time.Now.
	Now func() time.Time
}

// NewFormController creates the activity plan controller.
func NewFormController(db *gorm.DB) *FormController {
	return &FormController{DB: db}
}

// Register mounts the routes on g (the /activity-plans group, which must sit
// under the authenticated /api/v1 group). annual-activity-options is
// registered before /:id.
func (ctl *FormController) Register(g *gin.RouterGroup) {
	g.GET("", ctl.List)
	g.GET("/annual-activity-options", ctl.Options)
	g.GET("/:id", ctl.Get)
	g.POST("", ctl.Create)
	g.PUT("/:id", ctl.Update)
	g.DELETE("/:id", ctl.Delete)
}

func (ctl *FormController) now() time.Time {
	if ctl.Now != nil {
		return ctl.Now()
	}
	return time.Now()
}

// List is GET /activity-plans.
func (ctl *FormController) List(c *gin.Context) {
	p := form.ListParams{IncludeLegacy: true, Search: c.Query("search"), ActivityType: c.Query("activity_type")}
	var err error
	if p.Page, err = intQuery(c, "page", 1); err != nil {
		writeError(c, err)
		return
	}
	if p.PageSize, err = intQuery(c, "page_size", form.DefaultPageSize); err != nil {
		writeError(c, err)
		return
	}
	if s := strings.TrimSpace(c.Query("year")); s != "" {
		y, err := strconv.Atoi(s)
		if err != nil {
			writeError(c, validation("year", "year must be a number"))
			return
		}
		p.Year = &y
	}
	if s := strings.TrimSpace(c.Query("annual_plan_id")); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			writeError(c, validation("annual_plan_id", "annual_plan_id must be a UUID"))
			return
		}
		p.AnnualPlanID = &id
	}
	if s := strings.TrimSpace(c.Query("include_legacy")); s != "" {
		b, err := strconv.ParseBool(s)
		if err != nil {
			writeError(c, validation("include_legacy", "include_legacy must be true or false"))
			return
		}
		p.IncludeLegacy = b
	}

	res, err := form.List(ctl.DB, p)
	if err != nil {
		writeError(c, err)
		return
	}
	totalPages := (res.Total + int64(res.PageSize) - 1) / int64(res.PageSize)
	response.OK(c, "Activity plans fetched successfully", gin.H{
		"items": res.Items,
		"pagination": gin.H{
			"page":        res.Page,
			"page_size":   res.PageSize,
			"total":       res.Total,
			"total_pages": totalPages,
		},
	})
}

// Get is GET /activity-plans/:id.
func (ctl *FormController) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	d, err := form.Get(ctl.DB, id)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, "Activity plan fetched successfully", d)
}

// Options is GET /activity-plans/annual-activity-options.
func (ctl *FormController) Options(c *gin.Context) {
	rawPlan := strings.TrimSpace(c.Query("annual_plan_id"))
	planID, err := uuid.Parse(rawPlan)
	if err != nil {
		writeError(c, validation("annual_plan_id", "annual_plan_id is required and must be a UUID"))
		return
	}
	var exclude *uuid.UUID
	if s := strings.TrimSpace(c.Query("exclude_plan_id")); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			writeError(c, validation("exclude_plan_id", "exclude_plan_id must be a UUID"))
			return
		}
		exclude = &id
	}
	items, err := form.Options(ctl.DB, planID, strings.TrimSpace(c.Query("activity_type")), exclude)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, "Annual activity options fetched successfully", gin.H{"items": items})
}

// Create is POST /activity-plans.
func (ctl *FormController) Create(c *gin.Context) {
	in, ok := bindSave(c)
	if !ok {
		return
	}
	id, err := form.Create(ctl.DB, in, actorName(c), ctl.now())
	if err != nil {
		writeError(c, err)
		return
	}
	d, err := form.Get(ctl.DB, id)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, "Activity plan created successfully", d)
}

// Update is PUT /activity-plans/:id.
func (ctl *FormController) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	in, ok := bindSave(c)
	if !ok {
		return
	}
	if err := form.Update(ctl.DB, id, in, actorName(c), ctl.now()); err != nil {
		writeError(c, err)
		return
	}
	d, err := form.Get(ctl.DB, id)
	if err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, "Activity plan updated successfully", d)
}

// Delete is DELETE /activity-plans/:id.
func (ctl *FormController) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := form.Delete(ctl.DB, id); err != nil {
		writeError(c, err)
		return
	}
	response.OK(c, "Activity plan deleted successfully", gin.H{"id": id.String()})
}

// bindSave decodes the POST/PUT body. Malformed JSON is 400 VALIDATION_ERROR.
func bindSave(c *gin.Context) (form.SaveInput, bool) {
	var in form.SaveInput
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writeError(c, validation("", "could not read the request body"))
		return in, false
	}
	if len(bytes.TrimSpace(body)) == 0 {
		writeError(c, validation("rows", "request body is required"))
		return in, false
	}
	if err := json.Unmarshal(body, &in); err != nil {
		writeError(c, validation("", "invalid JSON body: "+err.Error()))
		return in, false
	}
	return in, true
}

func pathID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, validation("id", "invalid activity plan id"))
		return uuid.Nil, false
	}
	return id, true
}

func intQuery(c *gin.Context, key string, def int) (int, error) {
	s := strings.TrimSpace(c.Query(key))
	if s == "" {
		return def, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, validation(key, key+" must be a number")
	}
	return n, nil
}

func validation(field, message string) error {
	return form.Fail(form.CodeValidation, http.StatusBadRequest, message, form.Details{Field: field})
}

// actorName is the signed-in user's name from the JWT claims.
func actorName(c *gin.Context) string {
	v, _ := c.Get("username")
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

// writeError sends an AppError with its status, code and details; anything
// else is logged and sent as 500.
func writeError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if stderrors.As(err, &appErr) {
		response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
		return
	}
	if logger.Log != nil {
		logger.Error("Activity plan request failed",
			logger.LogField("path", c.FullPath()), logger.LogField("error", err.Error()))
	}
	response.InternalServerError(c, "Failed to process the activity plan request")
}
