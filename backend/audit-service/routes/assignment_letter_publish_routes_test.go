package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/pkg/middleware"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PUT /api/v1/assignment-letters/:id with a new status (the frontend's publish
// button) is limited to Admin, Audit Manager and the CAE. The CAE is the seeded
// EXECUTIVE role, so it must pass; auditors may still edit without changing the
// status.
func TestAssignmentLetterPublishRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger.Log = zap.NewNop()
	const secret = "test-secret"
	db := testdb.New(t,
		&models.AuditExecution{},
		&models.AuditActivity{},
		&models.AssignmentLetter{},
		&models.FieldworkInterview{},
		&models.FieldworkObservation{},
		&models.FieldworkDocument{},
		&models.FieldworkSample{},
		&models.FieldworkTestControl{},
		&models.WorkingPaperHeader{},
		&models.WorkingPaperRisk{},
		&models.WorkingPaperSample{},
		&models.WorkingPaperCause{},
		&models.WorkingPaperPlan{},
		&models.AuditResultReport{},
	)
	engine := gin.New()
	h := NewRouteHandler(engine, middleware.NewAuthMiddleware(secret), db, nil)
	registry := NewRouteRegistry()
	registry.SetAuditCharterController(stubCharterCtrl{})
	registry.SetAuditMandateController(stubMandateCtrl{})
	registry.SetAuditAssignmentController(stubAssignmentCtrl{})
	registry.SetAuditActivityController(stubActivityCtrl{})
	h.SetRegistry(registry)
	h.RegisterRoutes()

	tokenFor := func(roles ...string) string {
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &middleware.Claims{
			UserID:   "u1",
			Username: "tester",
			Roles:    roles,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		})
		s, err := tok.SignedString([]byte(secret))
		if err != nil {
			t.Fatal(err)
		}
		return s
	}

	newDraft := func(t *testing.T) string {
		t.Helper()
		l := &models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-" + uuid.NewString()[:8], Status: "Draft", AuditTeam: "SKAI", AuditYear: "2026"}
		if err := db.Create(l).Error; err != nil {
			t.Fatal(err)
		}
		return l.ID.String()
	}
	statusOf := func(t *testing.T, id string) string {
		t.Helper()
		var l models.AssignmentLetter
		if err := db.First(&l, "id = ?", id).Error; err != nil {
			t.Fatal(err)
		}
		return l.Status
	}
	put := func(id, body string, roles ...string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/assignment-letters/"+id, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenFor(roles...))
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		return w
	}

	cases := []struct {
		name  string
		roles []string
		want  int
	}{
		{"admin", []string{"ADMIN"}, http.StatusOK},
		{"seeded CAE role EXECUTIVE", []string{"EXECUTIVE"}, http.StatusOK},
		{"CHIEF_AUDIT_EXECUTIVE", []string{"CHIEF_AUDIT_EXECUTIVE"}, http.StatusOK},
		{"CAE spelled with spaces", []string{"Chief Audit Executive"}, http.StatusOK},
		{"audit manager", []string{"AUDIT_MANAGER"}, http.StatusOK},
		{"audit manager (lowercase claim)", []string{"audit_manager"}, http.StatusOK},
		{"auditor", []string{"AUDITOR"}, http.StatusForbidden},
		{"lead auditor (team leader)", []string{"LEAD_AUDITOR"}, http.StatusForbidden},
		{"viewer", []string{"VIEWER"}, http.StatusForbidden},
		{"auditee", []string{"AUDITEE"}, http.StatusForbidden},
		{"department head", []string{"DEPARTMENT_HEAD"}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := newDraft(t)
			w := put(id, `{"status":"Published"}`, tc.roles...)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d; body = %s", w.Code, tc.want, w.Body.String())
			}
			wantStatus := "Draft"
			if tc.want == http.StatusOK {
				wantStatus = "Published"
			}
			if got := statusOf(t, id); got != wantStatus {
				t.Errorf("stored status = %q, want %q", got, wantStatus)
			}
		})
	}

	// An auditor may still edit a draft as long as the status stays the same
	t.Run("auditor edits draft keeping status", func(t *testing.T) {
		id := newDraft(t)
		if w := put(id, `{"status":"Draft","auditTitle":"Audit Kas"}`, "AUDITOR"); w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
		}
	})
}
