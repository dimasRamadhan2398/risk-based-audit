package routes

import (
	"net/http"
	"strings"
	"testing"
)

type routeCase struct{ method, path string }

func isGuardStatus(c int) bool { return c == http.StatusForbidden || c == http.StatusUnauthorized }

// Master-data writes: ADMIN only. Listed by hand so a new route is added here
// on purpose; the sweep test below catches omissions.
var masterWriteRoutes = []routeCase{
	{"POST", "/api/v1/companies"}, {"PUT", "/api/v1/companies/x"}, {"DELETE", "/api/v1/companies/x"},
	{"POST", "/api/v1/business-units"}, {"PUT", "/api/v1/business-units/x"}, {"DELETE", "/api/v1/business-units/x"},
	{"POST", "/api/v1/departments"}, {"PUT", "/api/v1/departments/x"}, {"DELETE", "/api/v1/departments/x"},
	{"POST", "/api/v1/departments/events/resync"},
	{"POST", "/api/v1/employees"}, {"PUT", "/api/v1/employees/x"}, {"DELETE", "/api/v1/employees/x"},
	{"POST", "/api/v1/job-roles"}, {"PUT", "/api/v1/job-roles/x"}, {"DELETE", "/api/v1/job-roles/x"},
	{"POST", "/api/v1/locations"}, {"PUT", "/api/v1/locations/x"}, {"DELETE", "/api/v1/locations/x"},
	{"POST", "/api/v1/vision-mission-goals"}, {"PUT", "/api/v1/vision-mission-goals/x"}, {"DELETE", "/api/v1/vision-mission-goals/x"},
}

// QA reports and consulting services writes: ADMIN, EXECUTIVE,
// CHIEF_AUDIT_EXECUTIVE, AUDIT_MANAGER.
var reportWriteRoutes = []routeCase{
	{"POST", "/api/v1/quality-assurance"}, {"PUT", "/api/v1/quality-assurance/x"},
	{"DELETE", "/api/v1/quality-assurance/x"}, {"POST", "/api/v1/quality-assurance/import"},
	{"POST", "/api/v1/consulting-services"}, {"PUT", "/api/v1/consulting-services/x"},
	{"DELETE", "/api/v1/consulting-services/x"}, {"POST", "/api/v1/consulting-services/import"},
}

var readRoutes = []routeCase{
	{"GET", "/api/v1/companies"}, {"GET", "/api/v1/companies/x"},
	{"GET", "/api/v1/business-units"}, {"GET", "/api/v1/business-units/x"},
	{"GET", "/api/v1/departments"}, {"GET", "/api/v1/departments/x"},
	{"GET", "/api/v1/employees"}, {"GET", "/api/v1/employees/next-code"}, {"GET", "/api/v1/employees/by-code/E1"}, {"GET", "/api/v1/employees/x"},
	{"GET", "/api/v1/job-roles"}, {"GET", "/api/v1/job-roles/x"},
	{"GET", "/api/v1/locations"}, {"GET", "/api/v1/locations/x"},
	{"GET", "/api/v1/vision-mission-goals"}, {"GET", "/api/v1/vision-mission-goals/x"}, {"GET", "/api/v1/vision-mission-goals/company/x"},
	{"GET", "/api/v1/quality-assurance"}, {"GET", "/api/v1/consulting-services"}, {"GET", "/api/v1/data-sources"},
}

type roleSet map[string][]string // test name -> token roles

func (rs roleSet) tokens(t *testing.T) map[string]string {
	out := map[string]string{}
	for name, roles := range rs {
		out[name] = viewerTestToken(t, roles...)
	}
	return out
}

func checkWrites(t *testing.T, routes []routeCase, allowed, refused roleSet) {
	t.Helper()
	engine := viewerTestEngine(t)
	for _, tc := range routes {
		for name, tok := range allowed.tokens(t) {
			if c := doViewerReq(engine, tc.method, tc.path, tok); isGuardStatus(c) {
				t.Errorf("%s %s %s = %d, want it to pass the guards", name, tc.method, tc.path, c)
			}
		}
		for name, tok := range refused.tokens(t) {
			if c := doViewerReq(engine, tc.method, tc.path, tok); c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, tc.method, tc.path, c)
			}
		}
	}
}

// Roles that can never write under any rule here.
var nonWriters = roleSet{
	"AUDITOR":                 {"AUDITOR"},
	"auditor lower":           {"auditor"},
	"AUDITEE":                 {"AUDITEE"},
	"DEPARTMENT_HEAD":         {"DEPARTMENT_HEAD"},
	"AUDITOR+VIEWER":          {"AUDITOR", "VIEWER"},
	"VIEWER":                  {"VIEWER"},
	"AUDITEE+DEPARTMENT_HEAD": {"AUDITEE", "DEPARTMENT_HEAD"},
	"unknown role":            {"Administrator"},
	"no roles":                {},
}

func TestMasterDataWritesAdminOnly(t *testing.T) {
	checkWrites(t, masterWriteRoutes,
		roleSet{
			"ADMIN": {"ADMIN"}, "admin lower": {"admin"}, "padded": {" Admin "},
			"ADMIN+VIEWER": {"ADMIN", "VIEWER"}, "AUDITOR+ADMIN": {"AUDITOR", "ADMIN"},
		},
		merge(nonWriters, roleSet{
			"EXECUTIVE": {"EXECUTIVE"}, "CHIEF_AUDIT_EXECUTIVE": {"CHIEF_AUDIT_EXECUTIVE"}, "AUDIT_MANAGER": {"AUDIT_MANAGER"},
		}))
}

func TestReportWritesAllowList(t *testing.T) {
	checkWrites(t, reportWriteRoutes,
		roleSet{
			"ADMIN": {"ADMIN"}, "EXECUTIVE": {"EXECUTIVE"}, "executive lower": {"executive"},
			"CHIEF_AUDIT_EXECUTIVE": {"CHIEF_AUDIT_EXECUTIVE"}, "AUDIT_MANAGER": {"AUDIT_MANAGER"},
			"audit_manager lower": {"audit_manager"}, "AUDIT_MANAGER+VIEWER": {"AUDIT_MANAGER", "VIEWER"},
			"AUDITOR+EXECUTIVE": {"AUDITOR", "EXECUTIVE"},
		},
		nonWriters)
}

// Reads are unchanged: every role except VIEWER-only still reads.
func TestReadsUnchanged(t *testing.T) {
	engine := viewerTestEngine(t)
	readers := roleSet{
		"ADMIN": {"ADMIN"}, "AUDITOR": {"AUDITOR"}, "AUDITEE": {"AUDITEE"}, "DEPARTMENT_HEAD": {"DEPARTMENT_HEAD"},
		"EXECUTIVE": {"EXECUTIVE"}, "AUDITOR+VIEWER": {"AUDITOR", "VIEWER"},
	}
	for _, tc := range readRoutes {
		// The test DB has no master tables, so a handler may answer 404/500;
		// what matters is that the role guards did not stop the request.
		for name, tok := range readers.tokens(t) {
			if c := doViewerReq(engine, tc.method, tc.path, tok); isGuardStatus(c) {
				t.Errorf("%s %s %s = %d, want read access", name, tc.method, tc.path, c)
			}
		}
		if c := doViewerReq(engine, tc.method, tc.path, viewerTestToken(t, "VIEWER")); c != http.StatusForbidden {
			t.Errorf("VIEWER %s %s = %d, want 403", tc.method, tc.path, c)
		}
	}
}

// Data sources are undecided and unchanged: any non-VIEWER role may still write.
func TestDataSourceWritesUnchanged(t *testing.T) {
	engine := viewerTestEngine(t)
	for _, role := range []string{"ADMIN", "AUDITOR", "AUDITEE", "DEPARTMENT_HEAD", "EXECUTIVE"} {
		if c := doViewerReq(engine, "POST", "/api/v1/data-sources/sync-all", viewerTestToken(t, role)); isGuardStatus(c) {
			t.Errorf("%s POST data-sources/sync-all = %d, want unchanged (open)", role, c)
		}
	}
}

// Every registered non-GET route must be refused for AUDITOR and AUDITEE,
// except data-sources (undecided), so a route added later cannot escape.
func TestEveryWriteRouteRefusesNonWriters(t *testing.T) {
	engine := viewerTestEngine(t)
	tokens := map[string]string{
		"AUDITOR": viewerTestToken(t, "AUDITOR"), "AUDITEE": viewerTestToken(t, "AUDITEE"),
		"DEPARTMENT_HEAD": viewerTestToken(t, "DEPARTMENT_HEAD"),
	}
	seen := 0
	for _, r := range engine.Routes() {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			continue
		}
		if !strings.HasPrefix(r.Path, "/api/v1/") || strings.HasPrefix(r.Path, "/api/v1/data-sources") {
			continue
		}
		seen++
		path := r.Path
		for _, seg := range strings.Split(r.Path, "/") {
			if strings.HasPrefix(seg, ":") {
				path = strings.ReplaceAll(path, seg, "x")
			}
		}
		for name, tok := range tokens {
			if c := doViewerReq(engine, r.Method, path, tok); c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, r.Method, r.Path, c)
			}
		}
	}
	if seen < 25 {
		t.Fatalf("only %d write routes found; sweep is not covering the router", seen)
	}
}

func merge(a, b roleSet) roleSet {
	out := roleSet{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
