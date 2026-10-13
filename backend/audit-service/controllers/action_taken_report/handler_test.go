package actiontakenreport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/authclient"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var testNow = time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)

// fakeContacts answers the auth-service contact lookups from memory
type fakeContacts struct {
	byID  map[string]authclient.Contact
	heads map[string][]authclient.Contact // by department
	err   error
	calls int
}

func (f *fakeContacts) Contacts(_ context.Context, _ string, ids []string, dept string) ([]authclient.Contact, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if dept != "" {
		return f.heads[dept], nil
	}
	var out []authclient.Contact
	for _, id := range ids {
		if c, ok := f.byID[id]; ok {
			out = append(out, c)
		}
	}
	return out, nil
}

// fakeMailer records sends and fails the addresses in failFor
type fakeMailer struct {
	sent    []string
	subject []string
	text    []string
	failFor map[string]bool
}

func (m *fakeMailer) Send(_ context.Context, to []string, subject, text string) error {
	if len(to) != 1 {
		return errors.New("one recipient per email expected")
	}
	if m.failFor[to[0]] {
		return errors.New("provider down")
	}
	m.sent = append(m.sent, to[0])
	m.subject = append(m.subject, subject)
	m.text = append(m.text, text)
	return nil
}

type fixture struct {
	db       *gorm.DB
	ctl      *Controller
	contacts *fakeContacts
	mailer   *fakeMailer
	lha      *models.AuditResultReport
	atr      *models.ActionTakenReport
	pic      string
}

const (
	picID    = "11111111-1111-1111-1111-111111111111"
	headID   = "22222222-2222-2222-2222-222222222222"
	headID2  = "33333333-3333-3333-3333-333333333333"
	auditor  = "44444444-4444-4444-4444-444444444444"
	outsider = "55555555-5555-5555-5555-555555555555"
)

func newFixture(t *testing.T, status string, atrStatus string) *fixture {
	t.Helper()
	db := testdb.New(t, &models.AuditResultReport{}, &models.ActionTakenReport{}, &models.ActionTakenReportEvidence{})
	lha := &models.AuditResultReport{
		ID: uuid.New(), ReportNumber: "LHA-001/SKAI/2026", ReportTitle: "Audit Kas", Title: "Judul LHA",
		Department: "Finance", AuditObject: "Kas Kecil", AuditPeriod: "Q3 2026", Scope: "Kas dan bank",
		Conclusion: "Perlu perbaikan", PreparedBy: "Rina", Status: status,
		Findings: []models.AuditReportFinding{
			{ID: "f-1", Title: "Kas tidak sesuai", Category: "Major", Action: "Rekonsiliasi harian"},
			{ID: "f-2", Title: "Otorisasi lemah", Category: "Minor", Action: "Tetapkan approver"},
		},
		Signatures: []models.ReportSignature{{Name: "Dewi", Role: "CAE"}},
	}
	if err := db.Create(lha).Error; err != nil {
		t.Fatal(err)
	}
	due := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	atr := &models.ActionTakenReport{
		ID: uuid.New(), AuditResultReportID: &lha.ID, FindingID: lha.Findings[0].ID,
		ReportNumber: lha.ReportNumber, ReportTitle: lha.ReportTitle, FindingTitle: "Kas tidak sesuai",
		FindingCategory: "Major", Recommendation: "Rekonsiliasi harian", Status: atrStatus,
		PICUserID: picID, PICName: "Budi", DueDate: &due, CreatedAt: testNow, UpdatedAt: testNow,
	}
	if err := db.Create(atr).Error; err != nil {
		t.Fatal(err)
	}
	contacts := &fakeContacts{
		byID: map[string]authclient.Contact{
			picID: {ID: picID, FullName: "Budi", Email: "budi@example.id", Department: "Finance"},
		},
		heads: map[string][]authclient.Contact{
			"Finance": {{ID: headID, FullName: "Kepala Finance", Email: "head@example.id", Department: "Finance"}},
		},
	}
	mailer := &fakeMailer{failFor: map[string]bool{}}
	ctl := &Controller{DB: db, Now: func() time.Time { return testNow }, Contacts: contacts, Mailer: mailer}
	return &fixture{db: db, ctl: ctl, contacts: contacts, mailer: mailer, lha: lha, atr: atr, pic: picID}
}

// router mounts the controller behind a stand-in for Authenticate that reads
// the caller from the X-Test-User and X-Test-Roles headers
func (f *fixture) router() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v1/action-taken-reports")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", c.GetHeader("X-Test-User"))
		c.Set("username", "tester")
		var roles []string
		if raw := c.GetHeader("X-Test-Roles"); raw != "" {
			roles = strings.Split(raw, ",")
		}
		c.Set("roles", roles)
		c.Next()
	})
	f.ctl.Register(g)
	return r
}

func do(t *testing.T, r http.Handler, method, path, user, roles string, body interface{}) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", user)
	req.Header.Set("X-Test-Roles", roles)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func errCode(out map[string]any) string {
	e, _ := out["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

func dataOf(out map[string]any) map[string]any {
	d, _ := out["data"].(map[string]any)
	return d
}

func (f *fixture) atrPath(suffix string) string {
	return "/api/v1/action-taken-reports/" + f.atr.ID.String() + suffix
}

// ---------------------------------------------------------------------------
// LHA read-merge
// ---------------------------------------------------------------------------

func TestGetMergesLHAFields(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusPlanned)
	code, out := do(t, f.router(), http.MethodGet, f.atrPath(""), picID, "AUDITEE", nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, out)
	}
	d := dataOf(out)
	lha, ok := d["lha"].(map[string]any)
	if !ok {
		t.Fatalf("lha missing: %v", d)
	}
	if lha["department"] != "Finance" || lha["scope"] != "Kas dan bank" || lha["conclusion"] != "Perlu perbaikan" {
		t.Fatalf("LHA fields not merged: %v", lha)
	}
	if lha["findings_count"] != float64(2) || len(lha["findings"].([]any)) != 2 {
		t.Fatalf("findings not merged: %v", lha)
	}
	finding := lha["finding"].(map[string]any)
	if finding["id"] != f.atr.FindingID || finding["action"] != "Rekonsiliasi harian" {
		t.Fatalf("finding of this ATR wrong: %v", finding)
	}
	if sigs := lha["signatures"].([]any); len(sigs) != 1 {
		t.Fatalf("signatures = %v", sigs)
	}
}

func TestGetLHANullWhenReportDeleted(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusPlanned)
	if err := f.db.Delete(&models.AuditResultReport{}, "id = ?", f.lha.ID).Error; err != nil {
		t.Fatal(err)
	}
	code, out := do(t, f.router(), http.MethodGet, f.atrPath(""), auditor, "AUDITOR", nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	if v, ok := dataOf(out)["lha"]; !ok || v != nil {
		t.Fatalf("lha = %v, want null", v)
	}
}

// ---------------------------------------------------------------------------
// PUT /:id
// ---------------------------------------------------------------------------

func TestUpdateActionPlanByPIC(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusPlanned)
	code, out := do(t, f.router(), http.MethodPut, f.atrPath(""), picID, "AUDITEE",
		map[string]any{"action_plan": "  Rekonsiliasi kas setiap hari  ", "progress": 30})
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, out)
	}
	d := dataOf(out)
	if d["status"] != models.ATRStatusInProgress || d["action_plan"] != "Rekonsiliasi kas setiap hari" || d["progress"] != float64(30) {
		t.Fatalf("unexpected ATR after update: %v", d)
	}
	if d["lha"] == nil {
		t.Fatal("update response should carry the LHA block")
	}
}

func TestUpdateDueDateNeedsAssignPermission(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	r := f.router()
	body := map[string]any{"due_date": "2026-12-31"}
	if code, out := do(t, r, http.MethodPut, f.atrPath(""), picID, "AUDITEE", body); code != http.StatusForbidden {
		t.Fatalf("PIC changing due date: status = %d, body = %v", code, out)
	}
	code, out := do(t, r, http.MethodPut, f.atrPath(""), auditor, "AUDITOR", body)
	if code != http.StatusOK {
		t.Fatalf("auditor changing due date: status = %d, body = %v", code, out)
	}
	if got := dataOf(out)["due_date"]; got != "2026-12-31T00:00:00Z" {
		t.Fatalf("due_date = %v", got)
	}
}

func TestUpdateValidation(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	r := f.router()
	type caseT struct {
		body  map[string]any
		user  string
		roles string
	}
	// Plan and progress go to the PIC; a due date needs the assign permission
	cases := map[string]caseT{
		"empty body":       {body: map[string]any{}, user: picID, roles: "AUDITEE"},
		"progress over":    {body: map[string]any{"progress": 101}, user: picID, roles: "AUDITEE"},
		"progress under":   {body: map[string]any{"progress": -1}, user: picID, roles: "AUDITEE"},
		"blank plan":       {body: map[string]any{"action_plan": "   "}, user: picID, roles: "AUDITEE"},
		"bad due date":     {body: map[string]any{"due_date": "31/12/2026"}, user: auditor, roles: "AUDITOR"},
		"due date not str": {body: map[string]any{"due_date": 20261231}, user: auditor, roles: "AUDITOR"},
	}
	for name, c := range cases {
		if code, out := do(t, r, http.MethodPut, f.atrPath(""), c.user, c.roles, c.body); code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, body = %v", name, code, out)
		}
	}
	if code, _ := do(t, r, http.MethodPut, f.atrPath(""), picID, "AUDITEE", "not json"); code != http.StatusBadRequest {
		t.Errorf("malformed JSON: status = %d", code)
	}
}

func TestUpdateRejectsOutsiderAndClosedATR(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	if code, _ := do(t, f.router(), http.MethodPut, f.atrPath(""), outsider, "AUDITEE", map[string]any{"progress": 5}); code != http.StatusForbidden {
		t.Fatalf("non-PIC status = %d, want 403", code)
	}

	closed := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusCancelled)
	code, out := do(t, closed.router(), http.MethodPut, closed.atrPath(""), picID, "AUDITEE", map[string]any{"progress": 5})
	if code != http.StatusConflict {
		t.Fatalf("cancelled ATR update: status = %d, body = %v", code, out)
	}
}

// ---------------------------------------------------------------------------
// POST /:id/reminder
// ---------------------------------------------------------------------------

func TestReminderSendsToAuditeeAndDeptHead(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	f.contacts.heads["Finance"] = append(f.contacts.heads["Finance"],
		authclient.Contact{ID: headID2, FullName: "Wakil", Email: "wakil@example.id", Department: "Finance"})

	code, out := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, out)
	}
	want := []string{"budi@example.id", "head@example.id", "wakil@example.id"}
	if strings.Join(f.mailer.sent, ",") != strings.Join(want, ",") {
		t.Fatalf("sent to %v, want %v", f.mailer.sent, want)
	}
	if !strings.Contains(f.mailer.subject[0], "Kas tidak sesuai") {
		t.Fatalf("subject = %q", f.mailer.subject[0])
	}
	if !strings.Contains(f.mailer.text[0], "Yth. Budi") || !strings.Contains(f.mailer.text[1], "Yth. Kepala Finance") {
		t.Fatalf("greeting missing: %q / %q", f.mailer.text[0], f.mailer.text[1])
	}
	d := dataOf(out)
	if sent, _ := d["sent_to"].([]any); len(sent) != 3 {
		t.Fatalf("sent_to = %v", d["sent_to"])
	}
}

func TestReminderDedupesSameAddress(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusPlanned)
	// The PIC is also the department head: one e-mail only
	f.contacts.heads["Finance"] = []authclient.Contact{{ID: picID, FullName: "Budi", Email: "BUDI@example.id", Department: "Finance"}}
	code, out := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, out)
	}
	if len(f.mailer.sent) != 1 {
		t.Fatalf("sent %d emails, want 1", len(f.mailer.sent))
	}
}

func TestReminderCancelledIsRejectedWithoutSending(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusCancelled)
	code, out := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusConflict {
		t.Fatalf("status = %d, body = %v", code, out)
	}
	if errCode(out) != ErrCodeATRCancelled {
		t.Fatalf("error code = %q", errCode(out))
	}
	if len(f.mailer.sent) != 0 || f.contacts.calls != 0 {
		t.Fatalf("cancelled ATR must not look up or send: sent=%v lookups=%d", f.mailer.sent, f.contacts.calls)
	}
}

func TestReminderPermissionAndConfig(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	if code, _ := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), picID, "AUDITEE", nil); code != http.StatusForbidden {
		t.Fatalf("PIC sending reminder: status = %d, want 403", code)
	}
	if len(f.mailer.sent) != 0 {
		t.Fatal("forbidden caller must not send")
	}

	f.ctl.Mailer = nil
	code, out := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusServiceUnavailable || errCode(out) != ErrCodeEmailNotConfigured {
		t.Fatalf("no mailer: status = %d, code = %q", code, errCode(out))
	}
}

func TestReminderMissingEmailsAndRecipients(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	f.contacts.byID[picID] = authclient.Contact{ID: picID, FullName: "Budi", Email: "  ", Department: "Finance"}
	f.contacts.heads["Finance"] = []authclient.Contact{{ID: headID, FullName: "Kepala Finance", Department: "Finance"}}
	code, out := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusUnprocessableEntity || errCode(out) != ErrCodeNoReminderRecipients {
		t.Fatalf("no emails: status = %d, code = %q, body = %v", code, errCode(out), out)
	}
	if len(f.mailer.sent) != 0 {
		t.Fatal("nothing must be sent without a recipient")
	}
}

func TestReminderPartialMissingEmailIsSkipped(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	f.contacts.heads["Finance"] = []authclient.Contact{{ID: headID, FullName: "Kepala Finance", Department: "Finance"}}
	code, out := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, out)
	}
	d := dataOf(out)
	skipped, _ := d["skipped"].([]any)
	if len(skipped) != 1 || skipped[0].(map[string]any)["role"] != reminderRoleDeptHead {
		t.Fatalf("skipped = %v", d["skipped"])
	}
	if strings.Join(f.mailer.sent, ",") != "budi@example.id" {
		t.Fatalf("sent = %v", f.mailer.sent)
	}
}

func TestReminderNoDeptHeadIsReportedAsSkipped(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	f.contacts.heads = map[string][]authclient.Contact{}
	code, out := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, out)
	}
	skipped, _ := dataOf(out)["skipped"].([]any)
	if len(skipped) != 1 {
		t.Fatalf("skipped = %v", skipped)
	}
}

func TestReminderLookupFailureIs503(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	f.contacts.err = errors.New("auth-service unreachable")
	code, out := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusServiceUnavailable || errCode(out) != ErrCodeRecipientLookup {
		t.Fatalf("status = %d, code = %q", code, errCode(out))
	}
	if len(f.mailer.sent) != 0 {
		t.Fatal("nothing must be sent when recipients are unknown")
	}
}

func TestReminderSendFailures(t *testing.T) {
	f := newFixture(t, models.LHAApprovedStatuses[0], models.ATRStatusInProgress)
	f.mailer.failFor["budi@example.id"] = true
	f.contacts.heads["Finance"] = []authclient.Contact{{ID: headID, FullName: "Kepala", Email: "head@example.id", Department: "Finance"}}

	// One failure: partial success, the failed address is listed
	code, out := do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusOK {
		t.Fatalf("partial failure: status = %d, body = %v", code, out)
	}
	failed, _ := dataOf(out)["failed"].([]any)
	if len(failed) != 1 || failed[0].(map[string]any)["email"] != "budi@example.id" {
		t.Fatalf("failed = %v", failed)
	}

	// Everything failed: 502
	f.mailer.failFor["head@example.id"] = true
	code, out = do(t, f.router(), http.MethodPost, f.atrPath("/reminder"), auditor, "AUDITOR", nil)
	if code != http.StatusBadGateway || errCode(out) != ErrCodeEmailSendFailed {
		t.Fatalf("all failed: status = %d, code = %q", code, errCode(out))
	}
}

func TestReminderMessageHasNoRawFindingBreaksInSubject(t *testing.T) {
	due := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	r := &models.ActionTakenReport{FindingTitle: "Judul\r\nBcc: x@y", Status: models.ATRStatusPlanned, DueDate: &due, ReportNumber: "LHA-1"}
	subject, text := reminderMessage("Budi", r, nil, testNow)
	if strings.ContainsAny(subject, "\r\n") {
		t.Fatalf("subject has line breaks: %q", subject)
	}
	if !strings.Contains(text, "terlewat 8 hari") {
		t.Fatalf("overdue line missing: %q", text)
	}
}
