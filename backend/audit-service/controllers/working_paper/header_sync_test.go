package workingpaper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ctrlAssignmentLetter "audit-service/controllers/assignment_letter"
	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Routes as registered in routes/handler.go (auth omitted).
func router(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t, &models.AssignmentLetter{}, &models.WorkingPaperHeader{})
	newLetter := func() interface{} { return &models.AssignmentLetter{} }
	newLetters := func() interface{} { return &[]models.AssignmentLetter{} }
	newHeader := func() interface{} { return &models.WorkingPaperHeader{} }
	newHeaders := func() interface{} { return &[]models.WorkingPaperHeader{} }

	r := gin.New()
	al := r.Group("/assignment-letters")
	al.GET("", crud.List(db, "AssignmentLetter", newLetters))
	al.GET("/:id", crud.GetByID(db, "AssignmentLetter", newLetter))
	al.PUT("/:id", crud.UpdateWithHook(db, "AssignmentLetter", newLetter, ctrlAssignmentLetter.NormalizePurposeOnUpdate))
	wp := r.Group("/working-papers/headers")
	wp.GET("", crud.List(db, "WorkingPaperHeader", newHeaders))
	wp.GET("/:id", crud.GetByID(db, "WorkingPaperHeader", newHeader))
	wp.POST("", crud.Create(db, "WorkingPaperHeader", newHeader))
	wp.PUT("/:id", crud.UpdateWithHook(db, "WorkingPaperHeader", newHeader, SyncHeaderOnUpdate))
	return r, db
}

type item struct {
	ID                 string   `json:"id"`
	LetterNumber       string   `json:"letterNumber"`
	AssignmentLetterID string   `json:"assignmentLetterId"`
	AuditPurpose       *string  `json:"auditPurpose"`
	PurposeList        []string `json:"purposeList"`
}

func do(t *testing.T, r *gin.Engine, method, path, body string) (int, json.RawMessage) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code >= 300 {
		t.Fatalf("%s %s -> %d %s", method, path, w.Code, w.Body.String())
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("%s %s: bad body %s", method, path, w.Body.String())
	}
	return w.Code, env.Data
}

func one(t *testing.T, r *gin.Engine, method, path, body string) item {
	t.Helper()
	_, data := do(t, r, method, path, body)
	var it item
	if err := json.Unmarshal(data, &it); err != nil {
		t.Fatal(err)
	}
	return it
}

func list(t *testing.T, r *gin.Engine, path string) []item {
	t.Helper()
	_, data := do(t, r, http.MethodGet, path, "")
	var page struct {
		Items []item `json:"items"`
	}
	if err := json.Unmarshal(data, &page); err != nil {
		t.Fatal(err)
	}
	return page.Items
}

func purpose(t *testing.T, label string, it item, want string) {
	t.Helper()
	if it.AuditPurpose == nil {
		t.Fatalf("%s: auditPurpose missing from response", label)
	}
	if *it.AuditPurpose != want {
		t.Errorf("%s: auditPurpose = %q, want %q", label, *it.AuditPurpose, want)
	}
}

func storedPurpose(t *testing.T, db *gorm.DB, table, id string) string {
	t.Helper()
	var s string
	if err := db.Raw("SELECT COALESCE(audit_purpose, '') FROM "+table+" WHERE id = ?", id).Scan(&s).Error; err != nil {
		t.Fatal(err)
	}
	return s
}

const (
	uiLetter     = "ST-001/SKAI/2026" // created via the UI form: only purposeList
	legacyLetter = "ST-002/SKAI/2026" // auditPurpose column set (seeded letters)
	uiPurpose    = "Assess cash disbursement controls; Review payment approvals"
)

func seedLetters(t *testing.T, db *gorm.DB) (ui, legacy models.AssignmentLetter) {
	t.Helper()
	ui = models.AssignmentLetter{LetterNumber: uiLetter, AuditPurpose: "",
		PurposeList: []string{"Assess cash disbursement controls", " ", "Review payment approvals"}}
	legacy = models.AssignmentLetter{LetterNumber: legacyLetter, AuditPurpose: "IT Security Audit",
		PurposeList: []string{"Evaluate ERP access"}}
	for _, l := range []*models.AssignmentLetter{&ui, &legacy} {
		if err := db.Create(l).Error; err != nil {
			t.Fatal(err)
		}
	}
	return ui, legacy
}

func TestAssignmentLetterResponsesIncludeAuditPurpose(t *testing.T) {
	r, db := router(t)
	ui, legacy := seedLetters(t, db)

	byNumber := map[string]item{}
	for _, it := range list(t, r, "/assignment-letters") {
		byNumber[it.LetterNumber] = it
	}
	purpose(t, "list ui letter", byNumber[uiLetter], uiPurpose)
	purpose(t, "list legacy letter", byNumber[legacyLetter], "IT Security Audit")
	purpose(t, "detail ui letter", one(t, r, http.MethodGet, "/assignment-letters/"+ui.ID.String(), ""), uiPurpose)
	purpose(t, "detail legacy letter", one(t, r, http.MethodGet, "/assignment-letters/"+legacy.ID.String(), ""), "IT Security Audit")

	if got := storedPurpose(t, db, "assignment_letters", ui.ID.String()); got != "" {
		t.Errorf("read must not write audit_purpose, stored %q", got)
	}
}

func TestAssignmentLetterEditDoesNotFreezeDerivedPurpose(t *testing.T) {
	r, db := router(t)
	ui, legacy := seedLetters(t, db)

	// The edit form posts back the auditPurpose it read (derived) along with
	// the edited purposeList.
	got := one(t, r, http.MethodPut, "/assignment-letters/"+ui.ID.String(),
		`{"auditPurpose":"`+uiPurpose+`","purposeList":["Assess petty cash"]}`)
	purpose(t, "after purposeList edit", got, "Assess petty cash")
	if s := storedPurpose(t, db, "assignment_letters", ui.ID.String()); s != "" {
		t.Errorf("echoed derived purpose was stored: %q", s)
	}

	// A purpose that is not the joined list is stored as sent.
	got = one(t, r, http.MethodPut, "/assignment-letters/"+legacy.ID.String(), `{"auditPurpose":"Compliance Audit"}`)
	purpose(t, "explicit purpose", got, "Compliance Audit")
	if s := storedPurpose(t, db, "assignment_letters", legacy.ID.String()); s != "Compliance Audit" {
		t.Errorf("explicit purpose stored as %q", s)
	}
}

func TestHeaderCreateFallsBackToLetterPurpose(t *testing.T) {
	r, db := router(t)
	ui, _ := seedLetters(t, db)

	cases := []struct {
		name, body, want string
	}{
		{"ui letter by number, empty purpose", `{"assignmentLetterId":"` + uiLetter + `","auditPurpose":""}`, uiPurpose},
		{"ui letter by id, purpose omitted", `{"assignmentLetterId":"` + ui.ID.String() + `"}`, uiPurpose},
		{"legacy letter", `{"assignmentLetterId":"` + legacyLetter + `","auditPurpose":"  "}`, "IT Security Audit"},
		{"client purpose kept", `{"assignmentLetterId":"` + legacyLetter + `","auditPurpose":"Own purpose"}`, "Own purpose"},
		{"unknown letter", `{"assignmentLetterId":"ST-404/SKAI/2026","auditPurpose":""}`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			created := one(t, r, http.MethodPost, "/working-papers/headers", tc.body)
			purpose(t, "create response", created, tc.want)
			if s := storedPurpose(t, db, "working_paper_headers", created.ID); s != tc.want {
				t.Errorf("stored audit_purpose = %q, want %q", s, tc.want)
			}
		})
	}
}

func TestHeaderUpdateFillsAndRefreshesPurpose(t *testing.T) {
	r, db := router(t)
	seedLetters(t, db)

	h := one(t, r, http.MethodPost, "/working-papers/headers",
		`{"assignmentLetterId":"`+uiLetter+`","auditPurpose":"Own purpose","businessProcess":"Cash"}`)
	path := "/working-papers/headers/" + h.ID

	// Empty purpose for the same letter: filled from the letter and stored.
	got := one(t, r, http.MethodPut, path, `{"auditPurpose":"","businessProcess":"Cash 2"}`)
	purpose(t, "emptied", got, uiPurpose)
	if s := storedPurpose(t, db, "working_paper_headers", h.ID); s != uiPurpose {
		t.Errorf("emptied: stored %q", s)
	}

	// Non-empty purpose for the same letter is kept.
	got = one(t, r, http.MethodPut, path, `{"auditPurpose":"Edited"}`)
	purpose(t, "kept", got, "Edited")

	// Switching letters replaces the purpose, even when the client sends the
	// previous letter's value.
	got = one(t, r, http.MethodPut, path, `{"assignmentLetterId":"`+legacyLetter+`","auditPurpose":"Edited"}`)
	purpose(t, "letter changed", got, "IT Security Audit")
	if got.AssignmentLetterID != legacyLetter {
		t.Errorf("assignmentLetterId = %q", got.AssignmentLetterID)
	}
	if s := storedPurpose(t, db, "working_paper_headers", h.ID); s != "IT Security Audit" {
		t.Errorf("letter changed: stored %q", s)
	}

	// Switching letters without sending a purpose also refreshes it.
	got = one(t, r, http.MethodPut, path, `{"assignmentLetterId":"`+uiLetter+`"}`)
	purpose(t, "letter changed, purpose omitted", got, uiPurpose)

	// Updates that touch neither field leave the purpose alone.
	got = one(t, r, http.MethodPut, path, `{"location":"HQ"}`)
	purpose(t, "unrelated update", got, uiPurpose)
}

func TestHeaderReadShowsLetterPurposeForOldRecords(t *testing.T) {
	r, db := router(t)
	seedLetters(t, db)

	// A header stored before the fix, with no purpose.
	h := one(t, r, http.MethodPost, "/working-papers/headers", `{"assignmentLetterId":"`+uiLetter+`"}`)
	if err := db.Exec("UPDATE working_paper_headers SET audit_purpose = '' WHERE id = ?", h.ID).Error; err != nil {
		t.Fatal(err)
	}
	unlinked := one(t, r, http.MethodPost, "/working-papers/headers", `{"assignmentLetterId":"","auditPurpose":""}`)

	purpose(t, "detail", one(t, r, http.MethodGet, "/working-papers/headers/"+h.ID, ""), uiPurpose)
	byID := map[string]item{}
	for _, it := range list(t, r, "/working-papers/headers?assignmentLetterId="+uiLetter) {
		byID[it.ID] = it
	}
	purpose(t, "list filtered by letter", byID[h.ID], uiPurpose)
	purpose(t, "unlinked header", one(t, r, http.MethodGet, "/working-papers/headers/"+unlinked.ID, ""), "")

	if s := storedPurpose(t, db, "working_paper_headers", h.ID); s != "" {
		t.Errorf("read must not write audit_purpose, stored %q", s)
	}
}
