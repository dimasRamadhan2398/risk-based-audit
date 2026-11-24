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

// Approving an LHA (status "Final") is limited to Admin, Audit Manager and the
// CAE (the seeded EXECUTIVE role). Auditors can still create and edit drafts.
func TestAuditResultReportApprovalRoles(t *testing.T) {
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
		&models.ActionTakenReport{},
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

	newReport := func(t *testing.T, status string) string {
		t.Helper()
		r := &models.AuditResultReport{ID: uuid.New(), ReportNumber: "LHA-" + uuid.NewString()[:8], Status: status}
		if err := db.Create(r).Error; err != nil {
			t.Fatal(err)
		}
		return r.ID.String()
	}
	statusOf := func(t *testing.T, id string) string {
		t.Helper()
		var r models.AuditResultReport
		if err := db.First(&r, "id = ?", id).Error; err != nil {
			t.Fatal(err)
		}
		return r.Status
	}
	send := func(method, path, body string, roles ...string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenFor(roles...))
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		return w
	}
	put := func(id, body string, roles ...string) *httptest.ResponseRecorder {
		return send(http.MethodPut, "/api/v1/audit-result-reports/"+id, body, roles...)
	}

	approvers := []struct {
		name  string
		roles []string
	}{
		{"admin", []string{"ADMIN"}},
		{"seeded CAE role EXECUTIVE", []string{"EXECUTIVE"}},
		{"CHIEF_AUDIT_EXECUTIVE", []string{"CHIEF_AUDIT_EXECUTIVE"}},
		{"audit manager", []string{"AUDIT_MANAGER"}},
		{"audit manager (lowercase claim)", []string{"audit_manager"}},
	}
	others := []struct {
		name  string
		roles []string
	}{
		{"auditor", []string{"AUDITOR"}},
		{"lead auditor (team leader)", []string{"LEAD_AUDITOR"}},
		{"auditee", []string{"AUDITEE"}},
		{"department head", []string{"DEPARTMENT_HEAD"}},
		{"viewer", []string{"VIEWER"}},
	}

	for _, tc := range approvers {
		t.Run("approver can finalise: "+tc.name, func(t *testing.T) {
			id := newReport(t, "Draft")
			if w := put(id, `{"status":"Final"}`, tc.roles...); w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
			}
			if got := statusOf(t, id); got != "Final" {
				t.Errorf("stored status = %q, want Final", got)
			}
		})
	}

	for _, tc := range others {
		t.Run("cannot finalise: "+tc.name, func(t *testing.T) {
			id := newReport(t, "Draft")
			if w := put(id, `{"status":"Final"}`, tc.roles...); w.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body = %s", w.Code, w.Body.String())
			}
			if got := statusOf(t, id); got != "Draft" {
				t.Errorf("stored status = %q, want Draft", got)
			}
		})
	}

	// Other spellings of an approved status are guarded too
	for _, status := range []string{"final", "FINAL", "Approved", "Published"} {
		t.Run("auditor cannot set "+status, func(t *testing.T) {
			id := newReport(t, "Draft")
			if w := put(id, `{"status":"`+status+`"}`, "AUDITOR"); w.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body = %s", w.Code, w.Body.String())
			}
		})
	}

	t.Run("auditor cannot create an already final report", func(t *testing.T) {
		w := send(http.MethodPost, "/api/v1/audit-result-reports", `{"reportNumber":"LHA-NEW-1","status":"Final"}`, "AUDITOR")
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body = %s", w.Code, w.Body.String())
		}
		var n int64
		db.Model(&models.AuditResultReport{}).Where("report_number = ?", "LHA-NEW-1").Count(&n)
		if n != 0 {
			t.Errorf("report was created despite the 403")
		}
	})

	t.Run("auditor can create a draft", func(t *testing.T) {
		w := send(http.MethodPost, "/api/v1/audit-result-reports", `{"reportNumber":"LHA-NEW-2","status":"Draft"}`, "AUDITOR")
		if w.Code != http.StatusCreated && w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 201/200; body = %s", w.Code, w.Body.String())
		}
	})

	// The status column defaults to "DRAFT" while the form sends "Draft"
	t.Run("auditor edits a DRAFT report sending Draft", func(t *testing.T) {
		id := newReport(t, "DRAFT")
		if w := put(id, `{"status":"Draft","reportTitle":"LHA Kas"}`, "AUDITOR"); w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
		}
	})

	t.Run("auditor edits without sending a status", func(t *testing.T) {
		id := newReport(t, "Draft")
		if w := put(id, `{"reportTitle":"LHA Kas"}`, "AUDITOR"); w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
		}
	})

	t.Run("auditor cannot re-open a final report", func(t *testing.T) {
		id := newReport(t, "Final")
		if w := put(id, `{"status":"Draft"}`, "AUDITOR"); w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body = %s", w.Code, w.Body.String())
		}
		if got := statusOf(t, id); got != "Final" {
			t.Errorf("stored status = %q, want Final", got)
		}
	})

	t.Run("auditor resending Final on a final report is not a change", func(t *testing.T) {
		id := newReport(t, "Final")
		if w := put(id, `{"status":"final"}`, "AUDITOR"); w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
		}
	})

	t.Run("admin can re-open a final report", func(t *testing.T) {
		id := newReport(t, "Final")
		if w := put(id, `{"status":"Draft"}`, "ADMIN"); w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
		}
	})
}
