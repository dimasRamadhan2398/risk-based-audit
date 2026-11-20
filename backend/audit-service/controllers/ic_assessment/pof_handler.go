package icassessment

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"audit-service/pkg/response"
	"audit-service/pkg/safefile"
	"audit-service/pkg/upload"
	svc "audit-service/services/ic_assessment"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// QUANTITATIVE_POF routes (docs/coso-pof-assessment-api.md). They live under the
// same /ic-assessments group, so the Kong prefix and the authenticated
// /api/v1 group are unchanged. Viewer rules come from the group middleware:
// writes are refused, and the evidence / download / export routes are refused
// by DenyViewerFiles (matched on the segment names used below).

func (ctl *Controller) registerPoF(g *gin.RouterGroup) {
	g.GET("/branch-methods", ctl.GetBranchMethods)
	g.PUT("/branch-methods", ctl.PutBranchMethods)
	g.GET("/pof-settings", ctl.GetPofSettings)
	g.PUT("/pof-settings", ctl.PutPofSettings)
	g.GET("/pof-export-template", ctl.ExportTemplate)
	g.GET("/pof-import-profiles", ctl.ListProfiles)
	g.POST("/pof-import-profiles", ctl.CreateProfile)
	g.PUT("/pof-import-profiles/:profile_id", ctl.UpdateProfile)
	g.DELETE("/pof-import-profiles/:profile_id", ctl.DeleteProfile)

	g.GET("/:id/pof", ctl.GetPoFTree)
	g.GET("/:id/pof/:code", ctl.GetPoFItem)
	g.GET("/:id/pof-summary", ctl.GetPoFSummary)
	g.PUT("/:id/pof-values", ctl.PutPoFValues)
	g.GET("/:id/pof-assignments", ctl.ListAssignments)
	g.PUT("/:id/pof-assignments", ctl.PutAssignment)
	g.DELETE("/:id/pof-assignments/:assignment_id", ctl.DeleteAssignment)
	g.GET("/:id/pof-evidence", ctl.ListPoFEvidence)
	g.POST("/:id/pof-evidence", ctl.AddPoFEvidence)
	g.GET("/:id/pof-evidence/:evidence_id/download", ctl.DownloadPoFEvidence)
	g.DELETE("/:id/pof-evidence/:evidence_id", ctl.DeletePoFEvidence)
	g.GET("/:id/pof-export", ctl.ExportAssessment)
	g.POST("/:id/pof-import/preview", ctl.PreviewImport)
	g.GET("/:id/pof-import/:import_id", ctl.GetImport)
	g.POST("/:id/pof-import/:import_id/remap", ctl.RemapImport)
	g.POST("/:id/pof-import/:import_id/apply", ctl.ApplyImport)
	g.GET("/:id/pof-import/:import_id/download", ctl.DownloadImport)
	g.DELETE("/:id/pof-import/:import_id", ctl.DiscardImport)
}

// writeFieldErrors answers 422 with one message per field
func writeFieldErrors(c *gin.Context, fe svc.FieldErrors) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"success": false,
		"error":   gin.H{"code": "VALIDATION_ERROR", "message": "Invalid request", "fields": fe},
	})
}

// failPoF is fail with the field-error and import-error cases
func failPoF(c *gin.Context, err error) {
	var fe svc.FieldErrors
	if errors.As(err, &fe) {
		writeFieldErrors(c, fe)
		return
	}
	var ie *svc.ImportError
	if errors.As(err, &ie) {
		response.Error(c, ie.Status, ie.Code, ie.Message, "")
		return
	}
	fail(c, err)
}

func parseUUIDParam(c *gin.Context, name, label string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		response.BadRequest(c, "Invalid "+label)
		return uuid.Nil, false
	}
	return id, true
}

// ---- branch methods and settings ----

// GetBranchMethods handles GET /ic-assessments/branch-methods?location_id=
func (ctl *Controller) GetBranchMethods(c *gin.Context) {
	var key *string
	if _, present := c.GetQuery("location_id"); present {
		raw := strings.TrimSpace(c.Query("location_id"))
		if raw != "" {
			if _, err := uuid.Parse(raw); err != nil {
				response.BadRequest(c, "location_id must be a UUID")
				return
			}
		}
		key = &raw
	}
	res, err := ctl.Service.ListBranchMethods(actor(c), key)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Assessment methods fetched", res)
}

// PutBranchMethods handles PUT /ic-assessments/branch-methods
func (ctl *Controller) PutBranchMethods(c *gin.Context) {
	var req svc.BranchMethodRequest
	if !bind(c, &req) {
		return
	}
	res, err := ctl.Service.PutBranchMethod(c.Request.Context(), ctl.Sources(c), actor(c), req)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Assessment method saved", res)
}

// GetPofSettings handles GET /ic-assessments/pof-settings
func (ctl *Controller) GetPofSettings(c *gin.Context) {
	res, err := ctl.Service.GetPofSettings(actor(c))
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "PoF settings fetched", res)
}

// PutPofSettings handles PUT /ic-assessments/pof-settings (ADMIN)
func (ctl *Controller) PutPofSettings(c *gin.Context) {
	var req svc.PofSettingsUpdate
	if !bind(c, &req) {
		return
	}
	res, err := ctl.Service.UpdatePofSettings(actor(c), req)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "PoF settings saved", res)
}

// ---- reading ----

func boolQuery(c *gin.Context, key string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(c.Query(key))) {
	case "", "false", "0":
		return false, true
	case "true", "1":
		return true, true
	}
	response.BadRequest(c, key+" must be true or false")
	return false, false
}

// GetPoFTree handles GET /ic-assessments/:id/pof
func (ctl *Controller) GetPoFTree(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	q := svc.PoFQuery{Component: strings.ToUpper(strings.TrimSpace(c.Query("component"))),
		Principle: strings.ToUpper(strings.TrimSpace(c.Query("principle"))), Status: strings.ToUpper(strings.TrimSpace(c.Query("status")))}
	switch strings.TrimSpace(c.Query("expand")) {
	case "":
	case "elements":
		q.ExpandElements = true
	default:
		response.BadRequest(c, "expand must be elements")
		return
	}
	if q.Component != "" && !svc.IsComponentCode(q.Component) {
		response.BadRequest(c, "component must be CE, RA, CA, IC or MA")
		return
	}
	if q.Principle != "" {
		if _, ok := svc.PrincipleByCode(q.Principle); !ok {
			response.BadRequest(c, "principle must be P01..P17")
			return
		}
	}
	switch q.Status {
	case "", "NOT_ASSESSED", "IN_PROGRESS", "YES", "PARTIAL", "NO":
	default:
		response.BadRequest(c, "status must be NOT_ASSESSED, IN_PROGRESS, YES, PARTIAL or NO")
		return
	}
	switch a := strings.TrimSpace(c.Query("assignee")); a {
	case "":
	case "me":
		q.AssigneeMe = true
	default:
		response.BadRequest(c, "assignee must be me")
		return
	}
	switch n := strings.TrimSpace(c.Query("needs")); n {
	case "", "explanation", "evidence":
		q.Needs = n
	default:
		response.BadRequest(c, "needs must be explanation or evidence")
		return
	}
	res, err := ctl.Service.GetTree(c.Request.Context(), ctl.Sources(c), actor(c), id, q)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "PoF assessment fetched", res)
}

// GetPoFItem handles GET /ic-assessments/:id/pof/:code
func (ctl *Controller) GetPoFItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	res, err := ctl.Service.GetPoF(c.Request.Context(), ctl.Sources(c), actor(c), id, c.Param("code"))
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Point of focus fetched", res)
}

// GetPoFSummary handles GET /ic-assessments/:id/pof-summary
func (ctl *Controller) GetPoFSummary(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	res, err := ctl.Service.GetPoFSummary(c.Request.Context(), ctl.Sources(c), actor(c), id)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "PoF summary fetched", res)
}

// PutPoFValues handles PUT /ic-assessments/:id/pof-values
func (ctl *Controller) PutPoFValues(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4<<20)
	var req svc.ValuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if upload.IsTooLarge(err) {
			response.Error(c, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "The request body is too large", "")
			return
		}
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}
	res, err := ctl.Service.SaveValues(c.Request.Context(), ctl.Sources(c), actor(c), id, req)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Values saved", res)
}

// ---- assignments ----

// ListAssignments handles GET /ic-assessments/:id/pof-assignments
func (ctl *Controller) ListAssignments(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	mine := strings.TrimSpace(c.Query("assignee")) == "me"
	res, err := ctl.Service.ListAssignments(c.Request.Context(), ctl.Sources(c), actor(c), id, mine)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Assignments fetched", res)
}

// PutAssignment handles PUT /ic-assessments/:id/pof-assignments
func (ctl *Controller) PutAssignment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req svc.AssignmentRequest
	if !bind(c, &req) {
		return
	}
	res, err := ctl.Service.PutAssignment(c.Request.Context(), actor(c), id, req)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Assignment saved", res)
}

// DeleteAssignment handles DELETE /ic-assessments/:id/pof-assignments/:assignment_id
func (ctl *Controller) DeleteAssignment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	aid, ok := parseUUIDParam(c, "assignment_id", "assignment id")
	if !ok {
		return
	}
	if err := ctl.Service.DeleteAssignment(c.Request.Context(), actor(c), id, aid); err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Assignment removed", gin.H{"id": aid})
}

// ---- evidence ----

// ListPoFEvidence handles GET /ic-assessments/:id/pof-evidence
func (ctl *Controller) ListPoFEvidence(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	rows, err := ctl.Service.ListPoFEvidence(c.Request.Context(), ctl.Sources(c), actor(c), id, c.Query("scope_code"))
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Evidence fetched", rows)
}

// AddPoFEvidence handles POST /ic-assessments/:id/pof-evidence (multipart file
// or a JSON reference)
func (ctl *Controller) AddPoFEvidence(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	in := svc.EvidenceInput{}
	if e := upload.ParseForm(c, upload.DefaultMaxBytes); e != nil {
		upload.Respond(c, e)
		return
	}
	if c.ContentType() == "multipart/form-data" {
		in.ScopeCode, in.Title, in.DocNumber, in.Note = c.PostForm("scope_code"), c.PostForm("title"), c.PostForm("doc_number"), c.PostForm("note")
		fh, e := upload.FormFile(c, "file", upload.DefaultMaxBytes)
		if e != nil {
			upload.Respond(c, e)
			return
		}
		data, e := upload.ReadFormFile(fh, upload.DefaultMaxBytes)
		if e != nil {
			upload.Respond(c, e)
			return
		}
		name, e := upload.Validate(c.Request.Context(), fh.Filename, data, upload.DefaultMaxBytes, upload.DocumentExtensions)
		if e != nil {
			upload.Respond(c, e)
			return
		}
		in.FileName, in.Content, in.ContentType = name, data, safefile.ContentTypeFor(name)
	} else {
		var body struct {
			ScopeCode string `json:"scope_code"`
			Title     string `json:"title"`
			DocNumber string `json:"doc_number"`
			Note      string `json:"note"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			if upload.IsTooLarge(err) {
				upload.Respond(c, upload.ErrTooLarge(upload.DefaultMaxBytes))
				return
			}
			response.BadRequest(c, "Invalid request body: "+err.Error())
			return
		}
		in.ScopeCode, in.Title, in.DocNumber, in.Note = body.ScopeCode, body.Title, body.DocNumber, body.Note
	}
	row, err := ctl.Service.AddPoFEvidence(c.Request.Context(), actor(c), id, in)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.Created(c, "Evidence added", row)
}

// DownloadPoFEvidence handles GET /ic-assessments/:id/pof-evidence/:evidence_id/download
func (ctl *Controller) DownloadPoFEvidence(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	eid, ok := parseUUIDParam(c, "evidence_id", "evidence id")
	if !ok {
		return
	}
	row, err := ctl.Service.GetPoFEvidenceFile(c.Request.Context(), ctl.Sources(c), actor(c), id, eid)
	if err != nil {
		failPoF(c, err)
		return
	}
	upload.Serve(c, row.FileName, row.Content, "", false)
}

// DeletePoFEvidence handles DELETE /ic-assessments/:id/pof-evidence/:evidence_id
func (ctl *Controller) DeletePoFEvidence(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	eid, ok := parseUUIDParam(c, "evidence_id", "evidence id")
	if !ok {
		return
	}
	if err := ctl.Service.DeletePoFEvidence(c.Request.Context(), actor(c), id, eid); err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Evidence removed", gin.H{"id": eid})
}

// ---- Excel ----

const xlsxMime = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

func sendWorkbook(c *gin.Context, b []byte, name string) {
	c.Header("Content-Disposition", safefile.DispositionHeader("attachment", name))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "private, no-store")
	c.Data(http.StatusOK, xlsxMime, b)
}

func langQuery(c *gin.Context) (string, bool) {
	switch l := strings.ToLower(strings.TrimSpace(c.DefaultQuery("lang", "id"))); l {
	case "id", "en":
		return l, true
	}
	response.BadRequest(c, "lang must be id or en")
	return "", false
}

// ExportAssessment handles GET /ic-assessments/:id/pof-export
func (ctl *Controller) ExportAssessment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	lang, ok := langQuery(c)
	if !ok {
		return
	}
	blank, ok := boolQuery(c, "blank")
	if !ok {
		return
	}
	b, name, err := ctl.Service.ExportAssessment(c.Request.Context(), ctl.Sources(c), actor(c), id, lang, blank)
	if err != nil {
		failPoF(c, err)
		return
	}
	sendWorkbook(c, b, name)
}

// ExportTemplate handles GET /ic-assessments/pof-export-template
func (ctl *Controller) ExportTemplate(c *gin.Context) {
	lang, ok := langQuery(c)
	if !ok {
		return
	}
	b, name, err := ctl.Service.ExportTemplate(c.Request.Context(), ctl.Sources(c), actor(c), lang)
	if err != nil {
		failPoF(c, err)
		return
	}
	sendWorkbook(c, b, name)
}

var xlsxOnly = map[string]bool{".xlsx": true}

// PreviewImport handles POST /ic-assessments/:id/pof-import/preview
func (ctl *Controller) PreviewImport(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if e := upload.ParseForm(c, svc.MaxImportBytes); e != nil {
		upload.Respond(c, e)
		return
	}
	fh, e := upload.FormFile(c, "file", svc.MaxImportBytes)
	if e != nil {
		upload.Respond(c, e)
		return
	}
	data, e := upload.ReadFormFile(fh, svc.MaxImportBytes)
	if e != nil {
		upload.Respond(c, e)
		return
	}
	name, e := upload.Validate(c.Request.Context(), fh.Filename, data, svc.MaxImportBytes, xlsxOnly)
	if e != nil {
		upload.Respond(c, e)
		return
	}
	in := svc.PreviewInput{FileName: name, Content: data, Format: c.PostForm("format"), Sheet: strings.TrimSpace(c.PostForm("sheet"))}
	if raw := strings.TrimSpace(c.PostForm("profile_id")); raw != "" {
		pid, err := uuid.Parse(raw)
		if err != nil {
			writeFieldErrors(c, svc.FieldErrors{"profile_id": "must be a UUID"})
			return
		}
		in.ProfileID = &pid
	}
	if raw := strings.TrimSpace(c.PostForm("mapping")); raw != "" {
		var m svc.ImportMapping
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			writeFieldErrors(c, svc.FieldErrors{"mapping": "must be a JSON object"})
			return
		}
		in.Mapping = &m
	}
	res, err := ctl.Service.PreviewImport(c.Request.Context(), actor(c), id, in)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.Created(c, "Import preview ready", res)
}

// GetImport handles GET /ic-assessments/:id/pof-import/:import_id
func (ctl *Controller) GetImport(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	iid, ok := parseUUIDParam(c, "import_id", "import id")
	if !ok {
		return
	}
	res, err := ctl.Service.GetImportPreview(c.Request.Context(), actor(c), id, iid)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Import preview fetched", res)
}

// RemapImport handles POST /ic-assessments/:id/pof-import/:import_id/remap
func (ctl *Controller) RemapImport(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	iid, ok := parseUUIDParam(c, "import_id", "import id")
	if !ok {
		return
	}
	var req svc.RemapRequest
	if !bind(c, &req) {
		return
	}
	res, err := ctl.Service.RemapImport(c.Request.Context(), actor(c), id, iid, req)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Import preview updated", res)
}

// ApplyImport handles POST /ic-assessments/:id/pof-import/:import_id/apply
func (ctl *Controller) ApplyImport(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	iid, ok := parseUUIDParam(c, "import_id", "import id")
	if !ok {
		return
	}
	var req svc.ApplyRequest
	if !bindOptional(c, &req) {
		return
	}
	res, err := ctl.Service.ApplyImport(c.Request.Context(), ctl.Sources(c), actor(c), id, iid, req)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Import applied", res)
}

// DownloadImport handles GET /ic-assessments/:id/pof-import/:import_id/download
func (ctl *Controller) DownloadImport(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	iid, ok := parseUUIDParam(c, "import_id", "import id")
	if !ok {
		return
	}
	imp, err := ctl.Service.DownloadImport(c.Request.Context(), actor(c), id, iid)
	if err != nil {
		failPoF(c, err)
		return
	}
	upload.Serve(c, imp.FileName, imp.Content, "", false)
}

// DiscardImport handles DELETE /ic-assessments/:id/pof-import/:import_id
func (ctl *Controller) DiscardImport(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	iid, ok := parseUUIDParam(c, "import_id", "import id")
	if !ok {
		return
	}
	if err := ctl.Service.DiscardImport(c.Request.Context(), actor(c), id, iid); err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Import discarded", gin.H{"id": iid})
}

// ---- import profiles ----

// ListProfiles handles GET /ic-assessments/pof-import-profiles
func (ctl *Controller) ListProfiles(c *gin.Context) {
	rows, err := ctl.Service.ListProfiles(actor(c))
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Import profiles fetched", rows)
}

// CreateProfile handles POST /ic-assessments/pof-import-profiles
func (ctl *Controller) CreateProfile(c *gin.Context) {
	var req svc.ProfileRequest
	if !bind(c, &req) {
		return
	}
	p, err := ctl.Service.CreateProfile(actor(c), req)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.Created(c, "Import profile saved", p)
}

// UpdateProfile handles PUT /ic-assessments/pof-import-profiles/:profile_id
func (ctl *Controller) UpdateProfile(c *gin.Context) {
	pid, ok := parseUUIDParam(c, "profile_id", "profile id")
	if !ok {
		return
	}
	var req svc.ProfileRequest
	if !bind(c, &req) {
		return
	}
	p, err := ctl.Service.UpdateProfile(actor(c), pid, req)
	if err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Import profile saved", p)
}

// DeleteProfile handles DELETE /ic-assessments/pof-import-profiles/:profile_id
func (ctl *Controller) DeleteProfile(c *gin.Context) {
	pid, ok := parseUUIDParam(c, "profile_id", "profile id")
	if !ok {
		return
	}
	if err := ctl.Service.DeleteProfile(actor(c), pid); err != nil {
		failPoF(c, err)
		return
	}
	response.OK(c, "Import profile deleted", gin.H{"id": pid})
}
