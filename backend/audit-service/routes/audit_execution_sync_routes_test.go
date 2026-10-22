package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/pkg/middleware"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// POST /api/v1/audit-executions/sync creates executions from assignment letters,
// so only the roles that manage letters may call it. The frontend shows the sync
// to ADMIN and EXECUTIVE (canManageAssignmentLetter); those must pass, roles
// without letter access must get 403.
func TestAuditExecutionSyncRequiresLetterManagerRole(t *testing.T) {
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

	cases := []struct {
		name  string
		roles []string
		want  int
	}{
		{"admin", []string{"ADMIN"}, http.StatusOK},
		{"executive (lowercase claim)", []string{"executive"}, http.StatusOK},
		{"auditor", []string{"AUDITOR"}, http.StatusForbidden},
		{"viewer", []string{"VIEWER"}, http.StatusForbidden},
		{"auditee", []string{"AUDITEE"}, http.StatusForbidden},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/audit-executions/sync", nil)
		req.Header.Set("Authorization", "Bearer "+tokenFor(tc.roles...))
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s: status = %d, want %d; body = %s", tc.name, w.Code, tc.want, w.Body.String())
		}
	}

	// No token at all: Authenticate rejects it before the role check.
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/audit-executions/sync", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("no token: status = %d, want 401", w.Code)
	}
}
