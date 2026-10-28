package routes

import (
	"net/http"
	"strings"
	"testing"
)

type auditorCase struct{ method, path string }

// Master-data routes, enumerated by hand from routes.go so that a new route
// has to be added here on purpose (the sweep test below catches omissions).
var masterWriteRoutes = []auditorCase{
	{"POST", "/api/v1/companies"}, {"PUT", "/api/v1/companies/x"}, {"DELETE", "/api/v1/companies/x"},
	{"POST", "/api/v1/business-units"}, {"PUT", "/api/v1/business-units/x"}, {"DELETE", "/api/v1/business-units/x"},
	{"POST", "/api/v1/departments"}, {"PUT", "/api/v1/departments/x"}, {"DELETE", "/api/v1/departments/x"},
	{"POST", "/api/v1/departments/events/resync"},
	{"POST", "/api/v1/employees"}, {"PUT", "/api/v1/employees/x"}, {"DELETE", "/api/v1/employees/x"},
	{"POST", "/api/v1/job-roles"}, {"PUT", "/api/v1/job-roles/x"}, {"DELETE", "/api/v1/job-roles/x"},
	{"POST", "/api/v1/locations"}, {"PUT", "/api/v1/locations/x"}, {"DELETE", "/api/v1/locations/x"},
	{"POST", "/api/v1/vision-mission-goals"}, {"PUT", "/api/v1/vision-mission-goals/x"}, {"DELETE", "/api/v1/vision-mission-goals/x"},
}

var masterReadRoutes = []auditorCase{
	{"GET", "/api/v1/companies"}, {"GET", "/api/v1/companies/x"},
	{"GET", "/api/v1/business-units"}, {"GET", "/api/v1/business-units/x"},
	{"GET", "/api/v1/departments"}, {"GET", "/api/v1/departments/x"},
	{"GET", "/api/v1/employees"}, {"GET", "/api/v1/employees/next-code"}, {"GET", "/api/v1/employees/by-code/E1"}, {"GET", "/api/v1/employees/x"},
	{"GET", "/api/v1/job-roles"}, {"GET", "/api/v1/job-roles/x"},
	{"GET", "/api/v1/locations"}, {"GET", "/api/v1/locations/x"},
	{"GET", "/api/v1/vision-mission-goals"}, {"GET", "/api/v1/vision-mission-goals/x"}, {"GET", "/api/v1/vision-mission-goals/company/x"},
}

func isGuardStatus(c int) bool { return c == http.StatusForbidden || c == http.StatusUnauthorized }

func TestAuditorCannotWriteMasterData(t *testing.T) {
	engine := viewerTestEngine(t)
	tokens := map[string]string{
		"AUDITOR":         viewerTestToken(t, "AUDITOR"),
		"auditor":         viewerTestToken(t, "auditor"),
		"padded":          viewerTestToken(t, " Auditor "),
		"AUDITOR+auditor": viewerTestToken(t, "AUDITOR", "auditor"),
	}
	for _, tc := range masterWriteRoutes {
		for name, tok := range tokens {
			if c := doViewerReq(engine, tc.method, tc.path, tok); c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, tc.method, tc.path, c)
			}
		}
	}
}

func TestAuditorKeepsMasterReads(t *testing.T) {
	engine := viewerTestEngine(t)
	auditor := viewerTestToken(t, "AUDITOR")
	for _, tc := range masterReadRoutes {
		// The test DB has no master tables, so a handler may answer 404/500;
		// what matters is that the role guards did not stop the request.
		if c := doViewerReq(engine, tc.method, tc.path, auditor); isGuardStatus(c) {
			t.Errorf("AUDITOR %s %s = %d, want read access", tc.method, tc.path, c)
		}
	}
}

func TestOtherRolesStillPassWriteGuard(t *testing.T) {
	engine := viewerTestEngine(t)
	passing := map[string]string{
		"ADMIN":           viewerTestToken(t, "ADMIN"),
		"admin":           viewerTestToken(t, "admin"),
		"AUDITOR+ADMIN":   viewerTestToken(t, "AUDITOR", "ADMIN"),
		"AUDITEE":         viewerTestToken(t, "AUDITEE"),
		"DEPARTMENT_HEAD": viewerTestToken(t, "DEPARTMENT_HEAD"),
		"EXECUTIVE":       viewerTestToken(t, "EXECUTIVE"),
		"AUDITOR+AUDITEE": viewerTestToken(t, "AUDITOR", "AUDITEE"),
	}
	for _, tc := range masterWriteRoutes {
		for name, tok := range passing {
			if c := doViewerReq(engine, tc.method, tc.path, tok); isGuardStatus(c) {
				t.Errorf("%s %s %s = %d, want it to pass the guards", name, tc.method, tc.path, c)
			}
		}
	}
	for _, tc := range masterReadRoutes {
		if c := doViewerReq(engine, tc.method, tc.path, passing["ADMIN"]); isGuardStatus(c) {
			t.Errorf("ADMIN %s %s = %d", tc.method, tc.path, c)
		}
	}
}

// Every registered write route under a master-data prefix must be refused for
// AUDITOR, so a route added later cannot silently escape the guard.
func TestAuditorRefusedOnEveryMasterWriteRoute(t *testing.T) {
	engine := viewerTestEngine(t)
	auditor := viewerTestToken(t, "AUDITOR")
	prefixes := []string{"/api/v1/companies", "/api/v1/business-units", "/api/v1/departments",
		"/api/v1/employees", "/api/v1/job-roles", "/api/v1/locations", "/api/v1/vision-mission-goals"}
	seen := 0
	for _, r := range engine.Routes() {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			continue
		}
		master := false
		for _, p := range prefixes {
			if r.Path == p || strings.HasPrefix(r.Path, p+"/") {
				master = true
			}
		}
		if !master {
			continue
		}
		seen++
		path := r.Path
		for _, seg := range strings.Split(r.Path, "/") {
			if strings.HasPrefix(seg, ":") {
				path = strings.ReplaceAll(path, seg, "x")
			}
		}
		if c := doViewerReq(engine, r.Method, path, auditor); c != http.StatusForbidden {
			t.Errorf("AUDITOR %s %s = %d, want 403", r.Method, r.Path, c)
		}
	}
	if seen < 20 {
		t.Fatalf("only %d master write routes found; sweep is not covering the router", seen)
	}
}
