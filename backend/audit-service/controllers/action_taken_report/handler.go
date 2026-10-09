// Package actiontakenreport serves /api/v1/action-taken-reports: the follow-up
// of LHA findings. ATRs are created by the server when an LHA is approved
// (services/action_taken_report); this package only lists them and moves them
// through their workflow:
//
//	PLANNED --action-plan--> IN_PROGRESS --submit--> PENDING_REVIEW
//	PENDING_REVIEW --review approve--> COMPLETED
//	PENDING_REVIEW --review reject---> IN_PROGRESS
//	PLANNED|IN_PROGRESS|PENDING_REVIEW --cancel--> CANCELLED
//
// Who may do what is decided by services/action_taken_report (RolePermissions
// and Actor). Every route must be mounted inside the authenticated /api/v1
// group: the handlers read user_id/username/roles set by Authenticate.
package actiontakenreport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"audit-service/models"
	"audit-service/pkg/response"
	atrsvc "audit-service/services/action_taken_report"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DefaultEvidenceDir is where evidence files are stored. It lies under
// ./uploads, which serve.go exposes as static files, so serve.go must keep
// excluding it (see routes.UploadsHandler); downloads go through
// DownloadEvidence only.
const DefaultEvidenceDir = "./uploads/action-taken-reports"

// DefaultMaxEvidenceBytes is the largest evidence file accepted (10 MB, as
// the other upload endpoints)
const DefaultMaxEvidenceBytes int64 = 10 << 20

const maxActionPlanLen = 20000
const maxNoteLen = 5000

// Controller handles the ATR endpoints
type Controller struct {
	DB               *gorm.DB
	EvidenceDir      string
	MaxEvidenceBytes int64
	Now              func() time.Time
}

// NewController creates the controller with the default evidence directory
func NewController(db *gorm.DB) *Controller {
	return &Controller{DB: db, EvidenceDir: DefaultEvidenceDir, MaxEvidenceBytes: DefaultMaxEvidenceBytes, Now: time.Now}
}

// Register mounts the routes on g, which must already require authentication
func (ctl *Controller) Register(g *gin.RouterGroup) {
	g.GET("", ctl.List)
	g.GET("/permissions", ctl.MyPermissions)
	g.GET("/:id", ctl.Get)
	g.PUT("/:id/assignment", ctl.Assign)
	g.PUT("/:id/action-plan", ctl.UpdateActionPlan)
	g.POST("/:id/evidence", ctl.UploadEvidence)
	g.GET("/:id/evidence/:evidenceId", ctl.DownloadEvidence)
	g.POST("/:id/submit", ctl.Submit)
	g.POST("/:id/review", ctl.Review)
	g.POST("/:id/cancel", ctl.Cancel)
	g.DELETE("/:id", ctl.Delete)
}

func (ctl *Controller) now() time.Time {
	if ctl.Now != nil {
		return ctl.Now()
	}
	return time.Now()
}

// actor builds the caller from the claims Authenticate stored in the context
func actor(c *gin.Context) atrsvc.Actor {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	roles, _ := c.Get("roles")
	uid, _ := userID.(string)
	uname, _ := username.(string)
	roleList, _ := roles.([]string)
	return atrsvc.NewActor(uid, uname, roleList)
}

// ---------------------------------------------------------------------------
// responses
// ---------------------------------------------------------------------------

// EvidenceResponse is one evidence file of an ATR
type EvidenceResponse struct {
	ID         string    `json:"id"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	UploadedAt time.Time `json:"uploaded_at"`
	UploadedBy string    `json:"uploaded_by"`
}

// ATRResponse is the JSON shape of an ATR. audit_result_report_id,
// finding_id, pic_user_id, due_date and reviewed_at are null when unset.
type ATRResponse struct {
	ID                  string             `json:"id"`
	AuditResultReportID *string            `json:"audit_result_report_id"`
	ReportNumber        string             `json:"report_number"`
	ReportTitle         string             `json:"report_title"`
	FindingID           *string            `json:"finding_id"`
	FindingTitle        string             `json:"finding_title"`
	FindingCategory     string             `json:"finding_category"`
	Recommendation      string             `json:"recommendation"`
	ActionPlan          string             `json:"action_plan"`
	PICUserID           *string            `json:"pic_user_id"`
	PICName             string             `json:"pic_name"`
	DueDate             *string            `json:"due_date"`
	Progress            int                `json:"progress"`
	Status              string             `json:"status"`
	IsOverdue           bool               `json:"is_overdue"`
	OverdueDays         int                `json:"overdue_days"`
	Evidence            []EvidenceResponse `json:"evidence"`
	ReviewNote          string             `json:"review_note"`
	ReviewedBy          string             `json:"reviewed_by"`
	ReviewedAt          *time.Time         `json:"reviewed_at"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	// What the caller may do with this ATR now (see atrsvc.Actor.AllowedActions)
	AllowedActions []string `json:"allowed_actions"`
}

func strPtr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func toResponse(r *models.ActionTakenReport, a atrsvc.Actor, now time.Time) ATRResponse {
	out := ATRResponse{
		ID:              r.ID.String(),
		ReportNumber:    r.ReportNumber,
		ReportTitle:     r.ReportTitle,
		FindingID:       strPtr(r.FindingID),
		FindingTitle:    r.FindingTitle,
		FindingCategory: r.FindingCategory,
		Recommendation:  r.Recommendation,
		ActionPlan:      r.ActionPlan,
		PICUserID:       strPtr(r.PICUserID),
		PICName:         r.PICName,
		Progress:        r.Progress,
		Status:          r.Status,
		Evidence:        []EvidenceResponse{},
		ReviewNote:      r.ReviewNote,
		ReviewedBy:      r.ReviewedBy,
		ReviewedAt:      r.ReviewedAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		AllowedActions:  a.AllowedActions(r),
	}
	if r.AuditResultReportID != nil && *r.AuditResultReportID != uuid.Nil {
		s := r.AuditResultReportID.String()
		out.AuditResultReportID = &s
	}
	if r.DueDate != nil && !r.DueDate.IsZero() {
		s := models.ATRDueDate(r.DueDate.UTC()).Format(time.RFC3339)
		out.DueDate = &s
	}
	out.IsOverdue, out.OverdueDays = models.ComputeATROverdue(r.DueDate, r.Status, now)
	for _, e := range r.Evidence {
		out.Evidence = append(out.Evidence, EvidenceResponse{
			ID: e.ID.String(), FileName: e.FileName, FileSize: e.FileSize,
			UploadedAt: e.UploadedAt, UploadedBy: e.UploadedBy,
		})
	}
	return out
}

func forbidden(c *gin.Context, msg string) {
	response.Error(c, http.StatusForbidden, "FORBIDDEN", msg, "")
}

func conflict(c *gin.Context, msg string) {
	response.Error(c, http.StatusConflict, "INVALID_STATE", msg, "")
}

func invalid(c *gin.Context, msg string) {
	response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", msg, "")
}

// load fetches the ATR named by :id with its evidence; it writes the 400/404
// response and returns false when it cannot
func (ctl *Controller) load(c *gin.Context) (*models.ActionTakenReport, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid action taken report ID")
		return nil, false
	}
	var r models.ActionTakenReport
	err = ctl.DB.Preload("Evidence", func(db *gorm.DB) *gorm.DB {
		return db.Order("uploaded_at ASC").Order("id ASC")
	}).First(&r, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "Action taken report not found")
		return nil, false
	}
	if err != nil {
		response.InternalServerError(c, "Failed to fetch action taken report")
		return nil, false
	}
	return &r, true
}

// loadVisible is load plus the view check (403 when the caller may not see it)
func (ctl *Controller) loadVisible(c *gin.Context, a atrsvc.Actor) (*models.ActionTakenReport, bool) {
	r, ok := ctl.load(c)
	if !ok {
		return nil, false
	}
	if !a.CanView(r) {
		forbidden(c, "You are not allowed to view this action taken report")
		return nil, false
	}
	return r, true
}

// respondOne reloads the ATR and writes it
func (ctl *Controller) respondOne(c *gin.Context, a atrsvc.Actor, id uuid.UUID, msg string) {
	var r models.ActionTakenReport
	if err := ctl.DB.Preload("Evidence", func(db *gorm.DB) *gorm.DB {
		return db.Order("uploaded_at ASC").Order("id ASC")
	}).First(&r, "id = ?", id).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch action taken report")
		return
	}
	response.OK(c, msg, toResponse(&r, a, ctl.now()))
}

// transition applies updates only if the ATR is still in one of from, so two
// concurrent requests cannot both move it. Writes 409 and returns false when
// the status changed in between.
func (ctl *Controller) transition(c *gin.Context, id uuid.UUID, from []string, updates map[string]interface{}) bool {
	res := ctl.DB.Model(&models.ActionTakenReport{}).
		Where("id = ? AND status IN ?", id, from).
		Updates(updates)
	if res.Error != nil {
		response.InternalServerError(c, "Failed to update action taken report")
		return false
	}
	if res.RowsAffected == 0 {
		conflict(c, "The action taken report was changed by someone else; reload and try again")
		return false
	}
	return true
}

func bindJSON(c *gin.Context, v interface{}) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		invalid(c, "Invalid request body: "+err.Error())
		return false
	}
	return true
}

func statusIn(status string, allowed ...string) bool {
	for _, s := range allowed {
		if status == s {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// GET /action-taken-reports
// ---------------------------------------------------------------------------

var orderColumns = map[string]bool{
	"created_at": true, "updated_at": true, "due_date": true, "status": true,
	"progress": true, "report_number": true, "finding_title": true,
}

// parseOrder reads "column [asc|desc]" against orderColumns
func parseOrder(raw string) (clause.OrderByColumn, bool) {
	fields := strings.Fields(raw)
	if len(fields) == 0 || len(fields) > 2 {
		return clause.OrderByColumn{}, false
	}
	col := strings.ToLower(fields[0])
	if !orderColumns[col] {
		return clause.OrderByColumn{}, false
	}
	desc := false
	if len(fields) == 2 {
		switch strings.ToLower(fields[1]) {
		case "asc":
		case "desc":
			desc = true
		default:
			return clause.OrderByColumn{}, false
		}
	}
	return clause.OrderByColumn{Column: clause.Column{Name: col}, Desc: desc}, true
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func parseBool(raw string) (value bool, set bool, ok bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return false, false, true
	case "true", "1", "yes":
		return true, true, true
	case "false", "0", "no":
		return false, true, true
	}
	return false, false, false
}

// List returns a page of ATRs. Callers without view-all only get the ATRs
// they are the PIC of. Filters: search, status (comma separated),
// audit_result_report_id, pic_user_id, overdue (true/false); order is
// "<column> [asc|desc]" over a fixed set of columns.
func (ctl *Controller) List(c *gin.Context) {
	a := actor(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	q := ctl.DB.Model(&models.ActionTakenReport{})
	if !a.Perms[atrsvc.PermViewAll] {
		if a.UserID == "" {
			q = q.Where("1 = 0")
		} else {
			q = q.Where("pic_user_id = ?", a.UserID)
		}
	}

	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		var statuses []string
		for _, part := range strings.Split(raw, ",") {
			if strings.TrimSpace(part) == "" {
				continue
			}
			s, ok := models.NormalizeATRStatus(part)
			if !ok {
				invalid(c, "Invalid status \""+strings.TrimSpace(part)+"\": must be one of "+strings.Join(models.ATRStatuses, ", "))
				return
			}
			statuses = append(statuses, s)
		}
		if len(statuses) > 0 {
			q = q.Where("status IN ?", statuses)
		}
	}

	if raw := strings.TrimSpace(c.Query("audit_result_report_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			invalid(c, "Invalid audit_result_report_id")
			return
		}
		q = q.Where("audit_result_report_id = ?", id)
	}

	if raw := strings.TrimSpace(c.Query("pic_user_id")); raw != "" {
		q = q.Where("pic_user_id = ?", raw)
	}

	overdue, overdueSet, ok := parseBool(c.Query("overdue"))
	if !ok {
		invalid(c, "Invalid overdue: must be true or false")
		return
	}
	if overdueSet {
		today := models.ATRToday(ctl.now())
		closed := []string{models.ATRStatusCompleted, models.ATRStatusCancelled}
		if overdue {
			q = q.Where("due_date IS NOT NULL AND due_date < ? AND status NOT IN ?", today, closed)
		} else {
			q = q.Where("(due_date IS NULL OR due_date >= ? OR status IN ?)", today, closed)
		}
	}

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		pattern := "%" + escapeLike(strings.ToLower(search)) + "%"
		q = q.Where(`(LOWER(finding_title) LIKE ? ESCAPE '\' OR LOWER(report_number) LIKE ? ESCAPE '\' `+
			`OR LOWER(report_title) LIKE ? ESCAPE '\' OR LOWER(pic_name) LIKE ? ESCAPE '\' `+
			`OR LOWER(recommendation) LIKE ? ESCAPE '\' OR LOWER(action_plan) LIKE ? ESCAPE '\')`,
			pattern, pattern, pattern, pattern, pattern, pattern)
	}

	order := clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}
	if raw := strings.TrimSpace(c.Query("order")); raw != "" {
		o, ok := parseOrder(raw)
		if !ok {
			invalid(c, "Invalid order parameter")
			return
		}
		order = o
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch action taken reports")
		return
	}

	var rows []models.ActionTakenReport
	if err := q.Preload("Evidence", func(db *gorm.DB) *gorm.DB {
		return db.Order("uploaded_at ASC").Order("id ASC")
	}).
		Order(clause.OrderBy{Columns: []clause.OrderByColumn{order, {Column: clause.Column{Name: "id"}}}}).
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&rows).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch action taken reports")
		return
	}

	now := ctl.now()
	items := make([]ATRResponse, 0, len(rows))
	for i := range rows {
		items = append(items, toResponse(&rows[i], a, now))
	}
	response.OK(c, "Action taken reports fetched successfully", gin.H{
		"items": items,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// MyPermissions returns the caller's ATR permission keys
// (GET /action-taken-reports/permissions)
func (ctl *Controller) MyPermissions(c *gin.Context) {
	a := actor(c)
	response.OK(c, "Action taken report permissions", gin.H{
		"user_id":     a.UserID,
		"permissions": a.Perms.List(),
	})
}

// Get returns one ATR
func (ctl *Controller) Get(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.loadVisible(c, a)
	if !ok {
		return
	}
	response.OK(c, "Action taken report fetched successfully", toResponse(r, a, ctl.now()))
}

// ---------------------------------------------------------------------------
// PUT /action-taken-reports/:id/assignment
// ---------------------------------------------------------------------------

type assignRequest struct {
	PICUserID string          `json:"pic_user_id"`
	PICName   *string         `json:"pic_name"`
	DueDate   json.RawMessage `json:"due_date"`
}

// Assign sets the PIC and due date (assign permission; open ATRs only).
// pic_user_id is the auth-service user id; pic_name is stored as sent (the
// UI fills it from GET /users/assignable). due_date: "YYYY-MM-DD" or RFC3339;
// null or "" clears it; omitted leaves it unchanged.
func (ctl *Controller) Assign(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.load(c)
	if !ok {
		return
	}
	if !a.Perms[atrsvc.PermAssign] {
		forbidden(c, "You are not allowed to assign action taken reports")
		return
	}
	if models.IsATRClosed(r.Status) {
		conflict(c, "A "+r.Status+" action taken report cannot be reassigned")
		return
	}
	var req assignRequest
	if !bindJSON(c, &req) {
		return
	}
	picID, err := uuid.Parse(strings.TrimSpace(req.PICUserID))
	if err != nil || picID == uuid.Nil {
		invalid(c, "pic_user_id must be a user id")
		return
	}
	updates := map[string]interface{}{"pic_user_id": picID.String()}
	if req.PICName != nil {
		name := strings.TrimSpace(*req.PICName)
		if len([]rune(name)) > 200 {
			invalid(c, "pic_name is too long (max 200 characters)")
			return
		}
		updates["pic_name"] = name
	} else if !strings.EqualFold(picID.String(), r.PICUserID) {
		// A new PIC without a name: do not keep the previous PIC's name
		updates["pic_name"] = ""
	}
	if len(req.DueDate) > 0 {
		raw := strings.TrimSpace(string(req.DueDate))
		if raw == "null" || raw == `""` {
			updates["due_date"] = nil
		} else {
			var s string
			if err := json.Unmarshal(req.DueDate, &s); err != nil {
				invalid(c, "due_date must be a date string (YYYY-MM-DD or RFC3339) or null")
				return
			}
			d, ok := models.ParseATRDueDate(s)
			if !ok {
				invalid(c, "due_date must be a date (YYYY-MM-DD or RFC3339)")
				return
			}
			updates["due_date"] = d
		}
	}
	if !ctl.transition(c, r.ID, models.ATROpenStatuses, updates) {
		return
	}
	ctl.respondOne(c, a, r.ID, "Action taken report assigned")
}

// ---------------------------------------------------------------------------
// PUT /action-taken-reports/:id/action-plan
// ---------------------------------------------------------------------------

type actionPlanRequest struct {
	ActionPlan *string `json:"action_plan"`
	Progress   *int    `json:"progress"`
}

// UpdateActionPlan saves the PIC's action plan and progress (the PIC or an
// admin; PLANNED/IN_PROGRESS only). The first save moves PLANNED to
// IN_PROGRESS.
func (ctl *Controller) UpdateActionPlan(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.load(c)
	if !ok {
		return
	}
	if !a.CanActAsPIC(r) {
		forbidden(c, "Only the assigned PIC can update the action plan")
		return
	}
	if !statusIn(r.Status, models.ATRStatusPlanned, models.ATRStatusInProgress) {
		conflict(c, "The action plan cannot be changed while the action taken report is "+r.Status)
		return
	}
	var req actionPlanRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.ActionPlan == nil || strings.TrimSpace(*req.ActionPlan) == "" {
		invalid(c, "action_plan is required")
		return
	}
	plan := strings.TrimSpace(*req.ActionPlan)
	if len([]rune(plan)) > maxActionPlanLen {
		invalid(c, "action_plan is too long")
		return
	}
	updates := map[string]interface{}{"action_plan": plan, "status": models.ATRStatusInProgress}
	if req.Progress != nil {
		if *req.Progress < 0 || *req.Progress > 100 {
			invalid(c, "progress must be between 0 and 100")
			return
		}
		updates["progress"] = *req.Progress
	}
	if !ctl.transition(c, r.ID, []string{models.ATRStatusPlanned, models.ATRStatusInProgress}, updates) {
		return
	}
	ctl.respondOne(c, a, r.ID, "Action plan saved")
}

// ---------------------------------------------------------------------------
// POST /action-taken-reports/:id/submit
// ---------------------------------------------------------------------------

// Submit hands the ATR to the auditors for review (the PIC or an admin;
// IN_PROGRESS with an action plan only)
func (ctl *Controller) Submit(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.load(c)
	if !ok {
		return
	}
	if !a.CanActAsPIC(r) {
		forbidden(c, "Only the assigned PIC can submit the action taken report")
		return
	}
	if r.Status != models.ATRStatusInProgress {
		conflict(c, "Only an IN_PROGRESS action taken report can be submitted (current status "+r.Status+")")
		return
	}
	if strings.TrimSpace(r.ActionPlan) == "" {
		invalid(c, "Fill in the action plan before submitting")
		return
	}
	if !ctl.transition(c, r.ID, []string{models.ATRStatusInProgress}, map[string]interface{}{
		"status": models.ATRStatusPendingReview,
	}) {
		return
	}
	ctl.respondOne(c, a, r.ID, "Action taken report submitted for review")
}

// ---------------------------------------------------------------------------
// POST /action-taken-reports/:id/review
// ---------------------------------------------------------------------------

type reviewRequest struct {
	Decision string `json:"decision"`
	Note     string `json:"note"`
}

// Review approves (-> COMPLETED, progress 100) or rejects (-> IN_PROGRESS,
// note required) a submitted ATR. Review permission; a reviewer cannot review
// an ATR they are the PIC of (admins excepted).
func (ctl *Controller) Review(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.load(c)
	if !ok {
		return
	}
	if !a.Perms[atrsvc.PermReview] {
		forbidden(c, "You are not allowed to review action taken reports")
		return
	}
	if !a.CanReview(r) {
		forbidden(c, "You cannot review an action taken report you are the PIC of")
		return
	}
	if r.Status != models.ATRStatusPendingReview {
		conflict(c, "Only a PENDING_REVIEW action taken report can be reviewed (current status "+r.Status+")")
		return
	}
	var req reviewRequest
	if !bindJSON(c, &req) {
		return
	}
	note := strings.TrimSpace(req.Note)
	if len([]rune(note)) > maxNoteLen {
		invalid(c, "note is too long")
		return
	}
	now := ctl.now()
	updates := map[string]interface{}{
		"review_note": note,
		"reviewed_by": a.DisplayName(),
		"reviewed_at": now,
	}
	msg := ""
	switch strings.ToLower(strings.TrimSpace(req.Decision)) {
	case "approve":
		updates["status"] = models.ATRStatusCompleted
		updates["progress"] = 100
		msg = "Action taken report approved"
	case "reject":
		if note == "" {
			invalid(c, "A note is required when rejecting")
			return
		}
		updates["status"] = models.ATRStatusInProgress
		msg = "Action taken report returned to the PIC"
	default:
		invalid(c, `decision must be "approve" or "reject"`)
		return
	}
	if !ctl.transition(c, r.ID, []string{models.ATRStatusPendingReview}, updates) {
		return
	}
	ctl.respondOne(c, a, r.ID, msg)
}

// ---------------------------------------------------------------------------
// POST /action-taken-reports/:id/cancel
// ---------------------------------------------------------------------------

type cancelRequest struct {
	Note string `json:"note"`
}

// Cancel cancels an open ATR (cancel permission; a reason is required and
// stored as review_note)
func (ctl *Controller) Cancel(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.load(c)
	if !ok {
		return
	}
	if !a.Perms[atrsvc.PermCancel] {
		forbidden(c, "You are not allowed to cancel action taken reports")
		return
	}
	if models.IsATRClosed(r.Status) {
		conflict(c, "The action taken report is already "+r.Status)
		return
	}
	var req cancelRequest
	if !bindJSON(c, &req) {
		return
	}
	note := strings.TrimSpace(req.Note)
	if note == "" {
		invalid(c, "A cancellation note is required")
		return
	}
	if len([]rune(note)) > maxNoteLen {
		invalid(c, "note is too long")
		return
	}
	if !ctl.transition(c, r.ID, models.ATROpenStatuses, map[string]interface{}{
		"status":      models.ATRStatusCancelled,
		"review_note": note,
		"reviewed_by": a.DisplayName(),
		"reviewed_at": ctl.now(),
	}) {
		return
	}
	ctl.respondOne(c, a, r.ID, "Action taken report cancelled")
}

// ---------------------------------------------------------------------------
// DELETE /action-taken-reports/:id
// ---------------------------------------------------------------------------

// Delete soft-deletes an ATR (admin only). Its (LHA, finding) slot stays
// taken, so re-approving the LHA does not recreate it.
func (ctl *Controller) Delete(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.load(c)
	if !ok {
		return
	}
	if !a.Perms[atrsvc.PermDelete] {
		forbidden(c, "Only an admin can delete action taken reports")
		return
	}
	if err := ctl.DB.Delete(&models.ActionTakenReport{}, "id = ?", r.ID).Error; err != nil {
		response.InternalServerError(c, "Failed to delete action taken report")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Action taken report deleted successfully"})
}
