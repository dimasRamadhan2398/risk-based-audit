package routes

import (
	"strings"
	"testing"

	ctrlActivity "audit-service/controllers/audit_activity"
	ctrlAssignment "audit-service/controllers/audit_assignment"
	ctrlCharter "audit-service/controllers/audit_charter"
	ctrlMandate "audit-service/controllers/audit_mandate"
	"audit-service/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Stubs only satisfy the registry interfaces so RegisterRoutes can build the
// route table; their handlers are never called in these tests.
type stubCharterCtrl struct {
	ctrlCharter.AuditCharterControllerInterface
}
type stubMandateCtrl struct {
	ctrlMandate.AuditMandateControllerInterface
}
type stubAssignmentCtrl struct {
	ctrlAssignment.AuditAssignmentControllerInterface
}
type stubActivityCtrl struct {
	ctrlActivity.AuditActivityControllerInterface
}

func registeredRoutes(t *testing.T) gin.RoutesInfo {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	engine := gin.New()
	h := NewRouteHandler(engine, middleware.NewAuthMiddleware("test-secret"), db, nil)
	registry := NewRouteRegistry()
	registry.SetAuditCharterController(stubCharterCtrl{})
	registry.SetAuditMandateController(stubMandateCtrl{})
	registry.SetAuditAssignmentController(stubAssignmentCtrl{})
	registry.SetAuditActivityController(stubActivityCtrl{})
	h.SetRegistry(registry)
	h.RegisterRoutes()
	return engine.Routes()
}

// The frontend reaches audit-service through Kong, which only forwards the
// prefixes listed in kong.yml. /api/v1/executive-summary/* is not one of them
// (Kong answered 404 "no Route matched with those values", shown in the
// compilation form as "Failed to load source data"), so the statistics must be
// served under the routed /api/v1/executive-summaries prefix and must not be
// swallowed by GET /executive-summaries/:id.
func TestCompilationStatsIsServedUnderKongRoutedPrefix(t *testing.T) {
	var found bool
	for _, r := range registeredRoutes(t) {
		if r.Method == "GET" && r.Path == "/api/v1/executive-summaries/compilation-stats" {
			found = true
			if !strings.Contains(r.Handler, "CompilationStatsHandler") {
				t.Fatalf("GET /api/v1/executive-summaries/compilation-stats handled by %s, want CompilationStatsHandler", r.Handler)
			}
		}
	}
	if !found {
		t.Fatal("GET /api/v1/executive-summaries/compilation-stats is not registered")
	}
}
