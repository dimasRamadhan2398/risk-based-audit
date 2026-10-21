package executivesummary

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 1x1 PNG and a JPEG (SOI marker) as data URLs
const (
	testPNGDataURL  = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
	testJPEGDataURL = "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////wgALCAABAAEBAREA/8QAFBABAAAAAAAAAAAAAAAAAAAAAP/aAAgBAQABPxA="
)

// apprDB is an in-memory database with the executive summary table only
func apprDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.New(t, &models.ExecutiveSummary{})
}

// apprRouter mounts the executive summary routes as routes/handler.go does,
// for a caller with the given username and roles
func apprRouter(db *gorm.DB, username string, roles []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-"+username)
		c.Set("username", username)
		c.Set("roles", roles)
		c.Next()
	})
	g := r.Group("/api/v1/executive-summaries")
	g.GET("", crud.ListWithItems(db, "ExecutiveSummary", func() interface{} { return &[]models.ExecutiveSummary{} }, ListDecorator))
	g.GET("/:id", crud.GetByID(db, "ExecutiveSummary", func() interface{} { return &models.ExecutiveSummary{} }))
	g.POST("", crud.CreateWithHook(db, "ExecutiveSummary", func() interface{} { return &models.ExecutiveSummary{} }, GuardCreateHook))
	g.PUT("/:id", crud.UpdateWithHook(db, "ExecutiveSummary", func() interface{} { return &models.ExecutiveSummary{} }, UpdateHook))
	g.POST("/:id/approve", ApproveHandler(db))
	return r
}

func doAppr(t *testing.T, r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// apprErrorCode reads error.code from the response envelope
func apprErrorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var env struct {
		Error *struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode %q: %v", w.Body.String(), err)
	}
	if env.Error == nil {
		return ""
	}
	return env.Error.Code
}

// draftSummary creates a Draft summary with a little content
func draftSummary(t *testing.T, db *gorm.DB) *models.ExecutiveSummary {
	t.Helper()
	e := &models.ExecutiveSummary{ID: uuid.New(), Status: StatusDraft, Tahun: 2026, Narrative: "draft"}
	mustCreate(t, db, e)
	return e
}

func reloadSummary(t *testing.T, db *gorm.DB, id uuid.UUID) models.ExecutiveSummary {
	t.Helper()
	var e models.ExecutiveSummary
	if err := db.First(&e, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return e
}

func approvePayload(image string) map[string]interface{} {
	return map[string]interface{}{
		"signaturePlace": "Jakarta",
		"signatureDate":  "2026-10-20",
		"signatureImage": image,
		"signerName":     "Dewi Lestari",
		"signerRole":     "Kepala Audit Internal",
	}
}

func TestApproveSuccessStoresSignatureAndApprover(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	r := apprRouter(db, "admin.user", []string{"ADMIN"})

	w := doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries/"+e.ID.String()+"/approve", approvePayload(testPNGDataURL))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}

	got := reloadSummary(t, db, e.ID)
	if got.Status != StatusApproved {
		t.Errorf("status = %q, want Approved", got.Status)
	}
	if got.SignatureImage != testPNGDataURL {
		t.Errorf("signature image not stored")
	}
	if got.SignatureName != "Dewi Lestari" || got.SignatureRole != "Kepala Audit Internal" {
		t.Errorf("signer = %q / %q", got.SignatureName, got.SignatureRole)
	}
	if got.SignatureTempat != "Jakarta" || got.SignatureTanggal != "2026-10-20" {
		t.Errorf("place/date = %q / %q", got.SignatureTempat, got.SignatureTanggal)
	}
	if got.ApprovedBy != "admin.user" || got.ApprovedByRole != "ADMIN" {
		t.Errorf("approved by = %q (%q)", got.ApprovedBy, got.ApprovedByRole)
	}
	if got.ApprovedAt == nil || time.Since(*got.ApprovedAt) > time.Minute {
		t.Errorf("approved_at = %v", got.ApprovedAt)
	}
	if got.Narrative != "draft" {
		t.Errorf("narrative changed: %q", got.Narrative)
	}
}

func TestApproveAllowedRoles(t *testing.T) {
	for _, role := range []string{"ADMIN", "EXECUTIVE", "CHIEF_AUDIT_EXECUTIVE", "chief audit executive"} {
		t.Run(role, func(t *testing.T) {
			db := apprDB(t)
			e := draftSummary(t, db)
			r := apprRouter(db, "approver", []string{role})
			w := doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries/"+e.ID.String()+"/approve", approvePayload(testPNGDataURL))
			if w.Code != http.StatusOK {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if got := reloadSummary(t, db, e.ID); got.ApprovedByRole != role {
				t.Errorf("approved_by_role = %q, want %q", got.ApprovedByRole, role)
			}
		})
	}
}

func TestApproveForbiddenForOtherRoles(t *testing.T) {
	for _, roles := range [][]string{{"AUDITOR"}, {"AUDIT_MANAGER"}, {"DEPARTMENT_HEAD"}, {"AUDITEE"}, {"VIEWER"}, {}} {
		db := apprDB(t)
		e := draftSummary(t, db)
		r := apprRouter(db, "someone", roles)
		w := doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries/"+e.ID.String()+"/approve", approvePayload(testPNGDataURL))
		if w.Code != http.StatusForbidden {
			t.Errorf("roles %v: status %d, want 403: %s", roles, w.Code, w.Body.String())
		}
		if got := reloadSummary(t, db, e.ID); got.Status != StatusDraft || got.SignatureImage != "" || got.ApprovedAt != nil {
			t.Errorf("roles %v: summary changed: status=%q", roles, got.Status)
		}
	}
}

func TestApproveRequiresSignatureImage(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	r := apprRouter(db, "admin", []string{"ADMIN"})
	url := "/api/v1/executive-summaries/" + e.ID.String() + "/approve"

	// Missing, blank, and a payload without the field at all
	for name, body := range map[string]map[string]interface{}{
		"absent": {"signerName": "Dewi"},
		"blank":  approvePayload("   "),
	} {
		w := doAppr(t, r, http.MethodPost, url, body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400: %s", name, w.Code, w.Body.String())
		}
		if code := apprErrorCode(t, w); code != "SIGNATURE_REQUIRED" {
			t.Errorf("%s: code %q", name, code)
		}
	}
	if got := reloadSummary(t, db, e.ID); got.Status != StatusDraft {
		t.Errorf("status = %q after rejected approvals", got.Status)
	}
}

func TestApproveRejectsInvalidSignatureImage(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	r := apprRouter(db, "admin", []string{"ADMIN"})
	url := "/api/v1/executive-summaries/" + e.ID.String() + "/approve"

	// PNG bytes declared as JPEG
	jpegDeclaredPNG := "data:image/jpeg;base64," + strings.TrimPrefix(testPNGDataURL, "data:image/png;base64,")
	cases := map[string]string{
		"not an image":     "data:text/plain;base64,aGVsbG8=",
		"plain base64":     "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
		"bad base64":       "data:image/png;base64,%%%not-base64%%%",
		"mismatched bytes": jpegDeclaredPNG,
	}
	for name, img := range cases {
		w := doAppr(t, r, http.MethodPost, url, approvePayload(img))
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400: %s", name, w.Code, w.Body.String())
		}
	}

	// A real JPEG is accepted
	w := doAppr(t, r, http.MethodPost, url, approvePayload(testJPEGDataURL))
	if w.Code != http.StatusOK {
		t.Errorf("jpeg: status %d: %s", w.Code, w.Body.String())
	}
}

func TestApproveRejectsOversizedSignatureImage(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	r := apprRouter(db, "admin", []string{"ADMIN"})

	// PNG signature followed by padding, one byte over 2 MB
	raw := append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, make([]byte, MaxSignatureImageBytes-7)...)
	raw = append(raw, 0)
	big := "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)

	w := doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries/"+e.ID.String()+"/approve", approvePayload(big))
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status %d, want 413: %s", w.Code, w.Body.String())
	}
	if code := apprErrorCode(t, w); code != "SIGNATURE_TOO_LARGE" {
		t.Errorf("code %q", code)
	}
	if got := reloadSummary(t, db, e.ID); got.Status != StatusDraft {
		t.Errorf("status = %q", got.Status)
	}
}

func TestApproveRequiresDraftStatus(t *testing.T) {
	db := apprDB(t)
	r := apprRouter(db, "admin", []string{"ADMIN"})

	for _, status := range []string{StatusApproved, StatusRejected} {
		e := &models.ExecutiveSummary{ID: uuid.New(), Status: status}
		mustCreate(t, db, e)
		w := doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries/"+e.ID.String()+"/approve", approvePayload(testPNGDataURL))
		if w.Code != http.StatusConflict {
			t.Errorf("%s: status %d, want 409: %s", status, w.Code, w.Body.String())
		}
		if got := reloadSummary(t, db, e.ID); got.Status != status || got.SignatureImage != "" {
			t.Errorf("%s: row changed: status=%q image=%d bytes", status, got.Status, len(got.SignatureImage))
		}
	}
}

func TestApproveTwiceIsConflict(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	r := apprRouter(db, "admin", []string{"ADMIN"})
	url := "/api/v1/executive-summaries/" + e.ID.String() + "/approve"

	if w := doAppr(t, r, http.MethodPost, url, approvePayload(testPNGDataURL)); w.Code != http.StatusOK {
		t.Fatalf("first approve: %d: %s", w.Code, w.Body.String())
	}
	first := reloadSummary(t, db, e.ID)

	w := doAppr(t, r, http.MethodPost, url, approvePayload(testJPEGDataURL))
	if w.Code != http.StatusConflict {
		t.Fatalf("second approve: status %d, want 409", w.Code)
	}
	if got := reloadSummary(t, db, e.ID); got.SignatureImage != first.SignatureImage || got.ApprovedBy != first.ApprovedBy {
		t.Errorf("second approval overwrote the signature")
	}
}

func TestApproveNotFoundAndInvalidID(t *testing.T) {
	db := apprDB(t)
	r := apprRouter(db, "admin", []string{"ADMIN"})

	w := doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries/"+uuid.New().String()+"/approve", approvePayload(testPNGDataURL))
	if w.Code != http.StatusNotFound {
		t.Errorf("missing id: status %d, want 404", w.Code)
	}
	w = doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries/not-a-uuid/approve", approvePayload(testPNGDataURL))
	if w.Code != http.StatusBadRequest {
		t.Errorf("invalid id: status %d, want 400", w.Code)
	}
}

func TestApproveRequiresSignerName(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	r := apprRouter(db, "admin", []string{"ADMIN"})

	body := approvePayload(testPNGDataURL)
	body["signerName"] = "  "
	w := doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries/"+e.ID.String()+"/approve", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400: %s", w.Code, w.Body.String())
	}
	if got := reloadSummary(t, db, e.ID); got.Status != StatusDraft {
		t.Errorf("status = %q", got.Status)
	}
}

func TestPutCannotSetApproved(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	r := apprRouter(db, "admin", []string{"ADMIN"})
	url := "/api/v1/executive-summaries/" + e.ID.String()

	for _, status := range []string{"Approved", "approved", " APPROVED "} {
		w := doAppr(t, r, http.MethodPut, url, map[string]interface{}{"status": status})
		if w.Code != http.StatusBadRequest {
			t.Errorf("PUT status %q: %d, want 400: %s", status, w.Code, w.Body.String())
		}
	}
	if got := reloadSummary(t, db, e.ID); got.Status != StatusDraft {
		t.Errorf("status = %q after rejected PUT", got.Status)
	}
}

func TestPutCanStillSetDraftOrRejected(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	r := apprRouter(db, "auditor", []string{"AUDITOR"})
	url := "/api/v1/executive-summaries/" + e.ID.String()

	for _, status := range []string{StatusRejected, StatusDraft} {
		w := doAppr(t, r, http.MethodPut, url, map[string]interface{}{"status": status})
		if w.Code != http.StatusOK {
			t.Fatalf("PUT status %q: %d: %s", status, w.Code, w.Body.String())
		}
		if got := reloadSummary(t, db, e.ID); got.Status != status {
			t.Errorf("status = %q, want %q", got.Status, status)
		}
	}
}

func TestPutOnApprovedIsLocked(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	admin := apprRouter(db, "admin", []string{"ADMIN"})
	url := "/api/v1/executive-summaries/" + e.ID.String()
	if w := doAppr(t, admin, http.MethodPost, url+"/approve", approvePayload(testPNGDataURL)); w.Code != http.StatusOK {
		t.Fatalf("approve: %d: %s", w.Code, w.Body.String())
	}
	signed := reloadSummary(t, db, e.ID)

	// Any update is refused, including one that tries to replace the signature
	// or to move the summary back to Draft
	for _, body := range []map[string]interface{}{
		{"narrative": "changed"},
		{"status": StatusDraft},
		{"status": StatusRejected},
		{"signatureImage": testJPEGDataURL},
	} {
		w := doAppr(t, admin, http.MethodPut, url, body)
		if w.Code != http.StatusConflict {
			t.Errorf("PUT %v: status %d, want 409: %s", body, w.Code, w.Body.String())
		}
	}
	got := reloadSummary(t, db, e.ID)
	if got.Status != StatusApproved || got.SignatureImage != signed.SignatureImage || got.ApprovedBy != signed.ApprovedBy || got.Narrative != "draft" {
		t.Errorf("approved summary changed: status=%q narrative=%q", got.Status, got.Narrative)
	}
}

func TestPutIgnoresServerManagedApprovalFields(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	r := apprRouter(db, "auditor", []string{"AUDITOR"})

	w := doAppr(t, r, http.MethodPut, "/api/v1/executive-summaries/"+e.ID.String(), map[string]interface{}{
		"narrative":      "edited",
		"signatureImage": testPNGDataURL,
		"signerName":     "Forged",
		"approvedBy":     "forged",
		"approvedByRole": "ADMIN",
		"approvedAt":     "2026-01-01T00:00:00Z",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	got := reloadSummary(t, db, e.ID)
	if got.Narrative != "edited" {
		t.Errorf("narrative = %q", got.Narrative)
	}
	if got.SignatureImage != "" || got.SignatureName != "" || got.ApprovedBy != "" || got.ApprovedByRole != "" || got.ApprovedAt != nil || got.Status != StatusDraft {
		t.Errorf("server-managed fields written by PUT: %+v", got)
	}
}

func TestCreateCannotStartApprovedOrForgeSignature(t *testing.T) {
	db := apprDB(t)
	r := apprRouter(db, "admin", []string{"ADMIN"})

	w := doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries", map[string]interface{}{"status": StatusApproved, "tahun": 2026})
	if w.Code != http.StatusBadRequest {
		t.Errorf("POST approved: status %d, want 400: %s", w.Code, w.Body.String())
	}

	w = doAppr(t, r, http.MethodPost, "/api/v1/executive-summaries", map[string]interface{}{
		"status":         StatusDraft,
		"tahun":          2026,
		"signatureImage": testPNGDataURL,
		"approvedBy":     "forged",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("POST draft: status %d: %s", w.Code, w.Body.String())
	}
	var created models.ExecutiveSummary
	if err := db.Where("tahun = ?", 2026).First(&created).Error; err != nil {
		t.Fatal(err)
	}
	if created.Status != StatusDraft || created.SignatureImage != "" || created.ApprovedBy != "" {
		t.Errorf("created row carries approval data: %+v", created)
	}
}

func TestListOmitsSignatureImageDetailKeepsIt(t *testing.T) {
	db := apprDB(t)
	e := draftSummary(t, db)
	admin := apprRouter(db, "admin", []string{"ADMIN"})
	if w := doAppr(t, admin, http.MethodPost, "/api/v1/executive-summaries/"+e.ID.String()+"/approve", approvePayload(testPNGDataURL)); w.Code != http.StatusOK {
		t.Fatalf("approve: %d: %s", w.Code, w.Body.String())
	}

	list := doAppr(t, admin, http.MethodGet, "/api/v1/executive-summaries", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list: %d: %s", list.Code, list.Body.String())
	}
	if strings.Contains(list.Body.String(), "iVBORw0KGgo") {
		t.Errorf("list response contains the signature image")
	}
	if !strings.Contains(list.Body.String(), `"approvedBy":"admin"`) {
		t.Errorf("list response lacks approval record: %s", list.Body.String())
	}

	detail := doAppr(t, admin, http.MethodGet, "/api/v1/executive-summaries/"+e.ID.String(), nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail: %d: %s", detail.Code, detail.Body.String())
	}
	if !strings.Contains(detail.Body.String(), testPNGDataURL) {
		t.Errorf("detail response lacks the signature image")
	}
}
