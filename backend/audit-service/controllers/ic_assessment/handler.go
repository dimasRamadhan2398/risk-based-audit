// Package icassessment serves the COSO 2013 internal control component
// assessment under /api/v1/ic-assessments (docs/coso-2013-component-
// assessment-plan.md section 8.2). The rules live in services/ic_assessment;
// this package only binds requests, builds the caller and writes responses.
package icassessment

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	ctrlExecutiveSummary "audit-service/controllers/executive_summary"
	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/logger"
	"audit-service/pkg/response"
	svc "audit-service/services/ic_assessment"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Controller serves the assessment routes
type Controller struct {
	Service *svc.Service
	// Sources builds the master/auth/risk readers for one request (the
	// caller's token is forwarded). Tests replace it.
	Sources func(c *gin.Context) svc.Sources
}

// NewController returns a controller on db
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		Service: svc.New(db),
		Sources: func(c *gin.Context) svc.Sources { return svc.NewHTTPSources(c.GetHeader("Authorization")) },
	}
}

// Register adds the routes to g (a group under the authenticated /api/v1).
// Static paths are registered before /:id.
func (ctl *Controller) Register(g *gin.RouterGroup) {
	g.GET("", ctl.List)
	g.POST("", ctl.Create)
	g.GET("/summary", ctl.Summary)
	g.GET("/settings", ctl.GetSettings)
	g.PUT("/settings", ctl.PutSettings)
	g.GET("/periods", ctl.ListPeriods)
	g.POST("/periods/lock", ctl.LockPeriod)
	g.POST("/periods/unlock", ctl.UnlockPeriod)
	g.GET("/:id", ctl.Get)
	g.PUT("/:id", ctl.Update)
	g.DELETE("/:id", ctl.Delete)
	g.POST("/:id/propose", ctl.Propose)
	g.POST("/:id/submit", ctl.Submit)
	g.POST("/:id/return", ctl.Return)
	g.POST("/:id/approve", ctl.Approve)
	g.POST("/:id/reopen", ctl.Reopen)
	g.GET("/:id/evidence", ctl.Evidence)
	g.GET("/:id/events", ctl.Events)
	g.GET("/:id/drift", ctl.Drift)
	// The signature image is a file: served only here, authenticated, never
	// as a public URL ("view" also keeps VIEWER out, see DenyViewerFiles)
	g.GET("/:id/signature/view", ctl.SignatureView)
}

// actor builds the caller from the JWT claims set by Authenticate
func actor(c *gin.Context) svc.Actor {
	str := func(k string) string {
		v, _ := c.Get(k)
		s, _ := v.(string)
		return strings.TrimSpace(s)
	}
	rv, _ := c.Get("roles")
	roles, _ := rv.([]string)
	name := str("username")
	if name == "" {
		name = str("user_id")
	}
	return svc.Actor{UserID: str("user_id"), Name: name, Roles: roles, EmployeeCode: str("employee_code"),
		IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
}

func fail(c *gin.Context, err error) {
	var ae *apperrors.AppError
	if errors.As(err, &ae) {
		response.Error(c, ae.StatusCode, ae.Code, ae.Message, ae.Details)
		return
	}
	if logger.Log != nil {
		logger.Error("internal control assessment request failed", zap.Error(err), zap.String("path", c.FullPath()))
	}
	response.InternalServerError(c, "Internal control assessment request failed")
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid assessment id")
		return uuid.Nil, false
	}
	return id, true
}

func bind(c *gin.Context, v interface{}) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return false
	}
	return true
}

// bindOptional binds a body that may be empty
func bindOptional(c *gin.Context, v interface{}) bool {
	if c.Request.ContentLength == 0 {
		return true
	}
	return bind(c, v)
}

func optInt(c *gin.Context, key string, min, max int) (*int, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < min || n > max {
		response.BadRequest(c, key+" must be a whole number between "+strconv.Itoa(min)+" and "+strconv.Itoa(max))
		return nil, false
	}
	return &n, true
}

func optUUID(c *gin.Context, key string) (*uuid.UUID, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, true
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		response.BadRequest(c, key+" must be a UUID")
		return nil, false
	}
	return &id, true
}

// List handles GET /ic-assessments
func (ctl *Controller) List(c *gin.Context) {
	f := svc.ListFilter{}
	var ok bool
	if f.Year, ok = optInt(c, "year", 2000, 2100); !ok {
		return
	}
	if f.Quarter, ok = optInt(c, "quarter", 0, 4); !ok {
		return
	}
	if f.DepartmentID, ok = optUUID(c, "department_id"); !ok {
		return
	}
	if f.LocationID, ok = optUUID(c, "location_id"); !ok {
		return
	}
	if st := strings.ToUpper(strings.TrimSpace(c.Query("status"))); st != "" {
		valid := false
		for _, s := range svc.AssessmentStatuses {
			if s == st {
				valid = true
			}
		}
		if !valid {
			response.BadRequest(c, "status must be one of DRAFT, SUBMITTED, APPROVED, REOPENED")
			return
		}
		f.Status = st
	}
	if s := strings.TrimSpace(c.Query("sort")); s != "" {
		if !svc.SortColumns[s] {
			response.BadRequest(c, "Invalid sort field")
			return
		}
		f.Sort = s
	}
	switch strings.ToLower(strings.TrimSpace(c.DefaultQuery("order", "desc"))) {
	case "desc":
		f.Desc = true
	case "asc":
	default:
		response.BadRequest(c, "order must be asc or desc")
		return
	}
	page, ok := optInt(c, "page", 1, 1_000_000)
	if !ok {
		return
	}
	size, ok := optInt(c, "page_size", 1, 500)
	if !ok {
		return
	}
	if page != nil {
		f.Page = *page
	}
	if size != nil {
		f.PageSize = *size
	}
	res, err := ctl.Service.List(c.Request.Context(), ctl.Sources(c), actor(c), f)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Internal control assessments fetched", res)
}

// Summary handles GET /ic-assessments/summary?year=&quarter=&location_id=
func (ctl *Controller) Summary(c *gin.Context) {
	year, ok := optInt(c, "year", 2000, 2100)
	if !ok {
		return
	}
	if year == nil {
		response.BadRequest(c, "year is required")
		return
	}
	quarter, ok := optInt(c, "quarter", 0, 4)
	if !ok {
		return
	}
	q := 0
	if quarter != nil {
		q = *quarter
	}
	loc, ok := optUUID(c, "location_id")
	if !ok {
		return
	}
	res, err := ctl.Service.GetSummary(c.Request.Context(), ctl.Sources(c), actor(c), *year, q, loc)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Internal control assessment summary fetched", res)
}

// Create handles POST /ic-assessments
func (ctl *Controller) Create(c *gin.Context) {
	var req svc.CreateRequest
	if !bind(c, &req) {
		return
	}
	d, err := ctl.Service.Create(c.Request.Context(), ctl.Sources(c), actor(c), req)
	if err != nil {
		fail(c, err)
		return
	}
	response.Created(c, "Internal control assessment created", d)
}

// Get handles GET /ic-assessments/:id
func (ctl *Controller) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	d, err := ctl.Service.Get(c.Request.Context(), ctl.Sources(c), actor(c), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Internal control assessment fetched", d)
}

// Update handles PUT /ic-assessments/:id
func (ctl *Controller) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req svc.UpdateRequest
	if !bind(c, &req) {
		return
	}
	d, err := ctl.Service.Update(c.Request.Context(), actor(c), id, req)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Internal control assessment saved", d)
}

// Delete handles DELETE /ic-assessments/:id
func (ctl *Controller) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := ctl.Service.Delete(c.Request.Context(), actor(c), id); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Internal control assessment deleted", gin.H{"id": id})
}

// Propose handles POST /ic-assessments/:id/propose
func (ctl *Controller) Propose(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req svc.ProposeRequest
	if !bindOptional(c, &req) {
		return
	}
	d, err := ctl.Service.Propose(c.Request.Context(), ctl.Sources(c), actor(c), id, req)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Proposal computed", d)
}

func (ctl *Controller) transition(c *gin.Context, run func(id uuid.UUID, req svc.TransitionRequest) (*svc.Detail, error), msg string) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req svc.TransitionRequest
	if !bind(c, &req) {
		return
	}
	d, err := run(id, req)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, msg, d)
}

// Submit handles POST /ic-assessments/:id/submit
func (ctl *Controller) Submit(c *gin.Context) {
	ctl.transition(c, func(id uuid.UUID, req svc.TransitionRequest) (*svc.Detail, error) {
		return ctl.Service.Submit(c.Request.Context(), actor(c), id, req)
	}, "Internal control assessment submitted")
}

// Return handles POST /ic-assessments/:id/return
func (ctl *Controller) Return(c *gin.Context) {
	ctl.transition(c, func(id uuid.UUID, req svc.TransitionRequest) (*svc.Detail, error) {
		return ctl.Service.Return(c.Request.Context(), actor(c), id, req)
	}, "Internal control assessment returned")
}

// Reopen handles POST /ic-assessments/:id/reopen
func (ctl *Controller) Reopen(c *gin.Context) {
	ctl.transition(c, func(id uuid.UUID, req svc.TransitionRequest) (*svc.Detail, error) {
		return ctl.Service.Reopen(c.Request.Context(), actor(c), id, req)
	}, "Internal control assessment reopened")
}

const (
	maxSignerNameLen = 200
	maxSignerRoleLen = 100
)

// Approve handles POST /ic-assessments/:id/approve (optional signature)
func (ctl *Controller) Approve(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, ctrlExecutiveSummary.MaxApproveBodyBytes)
	var req svc.ApproveRequest
	if !bind(c, &req) {
		return
	}
	if strings.TrimSpace(req.SignatureImage) != "" {
		img, appErr := ctrlExecutiveSummary.ValidateSignatureImage(req.SignatureImage)
		if appErr != nil {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
			return
		}
		req.SignatureImage = img
	}
	if utf8.RuneCountInString(strings.TrimSpace(req.SignerName)) > maxSignerNameLen {
		response.BadRequest(c, "signer_name is too long")
		return
	}
	if utf8.RuneCountInString(strings.TrimSpace(req.SignerRole)) > maxSignerRoleLen {
		response.BadRequest(c, "signer_role is too long")
		return
	}
	d, err := ctl.Service.Approve(c.Request.Context(), actor(c), id, req)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Internal control assessment approved", d)
}

// Evidence handles GET /ic-assessments/:id/evidence?run=
func (ctl *Controller) Evidence(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	run, ok := optInt(c, "run", 1, 1_000_000)
	if !ok {
		return
	}
	r := 0
	if run != nil {
		r = *run
	}
	ev, err := ctl.Service.Evidence(c.Request.Context(), ctl.Sources(c), actor(c), id, r)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Evidence fetched", ev)
}

// Events handles GET /ic-assessments/:id/events
func (ctl *Controller) Events(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	ev, err := ctl.Service.Events(c.Request.Context(), actor(c), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Events fetched", ev)
}

// Drift handles GET /ic-assessments/:id/drift
func (ctl *Controller) Drift(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	d, err := ctl.Service.CheckDrift(c.Request.Context(), ctl.Sources(c), actor(c), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Data change check done", d)
}

// SignatureView handles GET /ic-assessments/:id/signature/view: the
// approval signature image bytes, never cached
func (ctl *Controller) SignatureView(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	dataURL, err := ctl.Service.Signature(c.Request.Context(), ctl.Sources(c), actor(c), id)
	if err != nil {
		fail(c, err)
		return
	}
	comma := strings.IndexByte(dataURL, ',')
	mime := "image/png"
	if strings.HasPrefix(strings.ToLower(dataURL), "data:image/jpeg") {
		mime = "image/jpeg"
	}
	raw, derr := base64.StdEncoding.DecodeString(dataURL[comma+1:])
	if comma < 0 || derr != nil {
		response.InternalServerError(c, "Stored signature is not readable")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", "inline")
	c.Data(http.StatusOK, mime, raw)
}

// GetSettings handles GET /ic-assessments/settings (ADMIN)
func (ctl *Controller) GetSettings(c *gin.Context) {
	if !actor(c).IsAdmin() {
		response.Forbidden(c, "Only ADMIN can read the assessment settings")
		return
	}
	st, err := svc.LoadSettings(ctl.Service.DB)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Assessment settings fetched", st)
}

// PutSettings handles PUT /ic-assessments/settings (ADMIN)
func (ctl *Controller) PutSettings(c *gin.Context) {
	a := actor(c)
	if !a.IsAdmin() {
		response.Forbidden(c, "Only ADMIN can change the assessment settings")
		return
	}
	var req svc.SettingsUpdate
	if !bind(c, &req) {
		return
	}
	st, err := svc.UpdateSettings(ctl.Service.DB, a, req, ctl.Service.Now().UTC())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Assessment settings saved", st)
}

// ListPeriods handles GET /ic-assessments/periods?year=
func (ctl *Controller) ListPeriods(c *gin.Context) {
	year, ok := optInt(c, "year", 2000, 2100)
	if !ok {
		return
	}
	y := 0
	if year != nil {
		y = *year
	}
	rows, err := ctl.Service.ListPeriods(actor(c), y)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Periods fetched", rows)
}

// LockPeriod handles POST /ic-assessments/periods/lock
func (ctl *Controller) LockPeriod(c *gin.Context) {
	var req svc.PeriodRequest
	if !bind(c, &req) {
		return
	}
	row, err := ctl.Service.LockPeriod(actor(c), req)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Period locked", row)
}

// UnlockPeriod handles POST /ic-assessments/periods/unlock
func (ctl *Controller) UnlockPeriod(c *gin.Context) {
	var req svc.PeriodRequest
	if !bind(c, &req) {
		return
	}
	row, err := ctl.Service.UnlockPeriod(actor(c), req)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Period unlocked", row)
}
