package routes

import (
	"net/http"
	"strings"
	"testing"
)

// Writes limited to ADMIN, EXECUTIVE, CHIEF_AUDIT_EXECUTIVE.
var universeWriteRoutes = [][2]string{
	{"POST", "/api/v1/audit-universe/corporate"},
	{"POST", "/api/v1/audit-universe/corporate/sync-master"},
	{"DELETE", "/api/v1/audit-universe/corporate/x"},
	{"POST", "/api/v1/audit-universe/year/2026"},
	{"POST", "/api/v1/audit-universe/year/2026/score"},
	{"POST", "/api/v1/audit-universe/events/resync"},
	{"POST", "/api/v1/risk-factors/standard"},
	{"PUT", "/api/v1/risk-factors/standard/x"},
	{"DELETE", "/api/v1/risk-factors/standard/x"},
}

var universeReadRoutes = [][2]string{
	{"GET", "/api/v1/audit-universe/standard"},
	{"GET", "/api/v1/audit-universe/corporate"},
	{"GET", "/api/v1/audit-universe/year/2026"},
	{"GET", "/api/v1/risk-factors/standard"},
}

func guarded(c int) bool { return c == http.StatusForbidden || c == http.StatusUnauthorized }

func TestUniverseWriteAllowList(t *testing.T) {
	engine := viewerTestEngine(t)
	allowed := map[string][]string{
		"ADMIN": {"ADMIN"}, "admin lower": {"admin"}, "EXECUTIVE": {"EXECUTIVE"},
		"CHIEF_AUDIT_EXECUTIVE": {"CHIEF_AUDIT_EXECUTIVE"}, "cae lower": {"chief_audit_executive"},
		"EXECUTIVE+VIEWER": {"EXECUTIVE", "VIEWER"}, "AUDITOR+ADMIN": {"AUDITOR", "ADMIN"},
	}
	refused := map[string][]string{
		"AUDITOR": {"AUDITOR"}, "AUDITEE": {"AUDITEE"}, "VIEWER": {"VIEWER"},
		"AUDITOR+VIEWER": {"AUDITOR", "VIEWER"},
		"AUDIT_MANAGER":          {"AUDIT_MANAGER"},
		"no roles":               {},
	}
	for _, tc := range universeWriteRoutes {
		for name, roles := range allowed {
			if c := doViewerReq(engine, tc[0], tc[1], viewerTestToken(t, roles...)); guarded(c) {
				t.Errorf("%s %s %s = %d, want it to pass the guards", name, tc[0], tc[1], c)
			}
		}
		for name, roles := range refused {
			if c := doViewerReq(engine, tc[0], tc[1], viewerTestToken(t, roles...)); c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, tc[0], tc[1], c)
			}
		}
		for name, roles := range map[string][]string{"DEPARTMENT_HEAD": {"DEPARTMENT_HEAD"}, "DEPARTMENT_HEAD+VIEWER": {"DEPARTMENT_HEAD", "VIEWER"}} {
			c := doViewerReq(engine, tc[0], tc[1], viewerTestToken(t, roles...))
			if deptHeadRoutes[tc[0]+" "+tc[1]] && guarded(c) {
				t.Errorf("%s %s %s = %d, want it to reach the controller", name, tc[0], tc[1], c)
			}
			if !deptHeadRoutes[tc[0]+" "+tc[1]] && c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, tc[0], tc[1], c)
			}
		}
	}
}

// The only write routes DEPARTMENT_HEAD reaches (department scoping happens in
// the controller). sync-master, year establish/score, events/resync and the
// standard risk factor library stay closed to it.
var deptHeadRoutes = map[string]bool{
	"POST /api/v1/audit-universe/corporate":     true,
	"DELETE /api/v1/audit-universe/corporate/x": true,
}

func TestUniverseReadsUnchanged(t *testing.T) {
	engine := viewerTestEngine(t)
	for _, tc := range universeReadRoutes {
		for _, roles := range [][]string{{"ADMIN"}, {"AUDITOR"}, {"AUDITEE"}, {"DEPARTMENT_HEAD"}, {"EXECUTIVE"}, {"VIEWER"}, {"AUDITOR", "VIEWER"}} {
			if c := doViewerReq(engine, tc[0], tc[1], viewerTestToken(t, roles...)); guarded(c) {
				t.Errorf("%v %s %s = %d, want read access", roles, tc[0], tc[1], c)
			}
		}
	}
}

// Every non-GET route under audit-universe and risk-factors/standard must
// refuse AUDITOR, AUDITEE and DEPARTMENT_HEAD (except deptHeadRoutes), including routes added later.
func TestEveryUniverseWriteRouteRefusesNonWriters(t *testing.T) {
	engine := viewerTestEngine(t)
	toks := map[string]string{
		"AUDITOR": viewerTestToken(t, "AUDITOR"), "AUDITEE": viewerTestToken(t, "AUDITEE"),
		"DEPARTMENT_HEAD": viewerTestToken(t, "DEPARTMENT_HEAD"),
	}
	deptHead := toks["DEPARTMENT_HEAD"]
	seen := 0
	for _, r := range engine.Routes() {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			continue
		}
		if !strings.HasPrefix(r.Path, "/api/v1/audit-universe") && !strings.HasPrefix(r.Path, "/api/v1/risk-factors/standard") {
			continue
		}
		seen++
		path := r.Path
		for _, seg := range strings.Split(r.Path, "/") {
			if strings.HasPrefix(seg, ":") {
				path = strings.ReplaceAll(path, seg, "x")
			}
		}
		for name, tok := range toks {
			if tok == deptHead && deptHeadRoutes[r.Method+" "+path] {
				continue
			}
			if c := doViewerReq(engine, r.Method, path, tok); c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, r.Method, r.Path, c)
			}
		}
	}
	if seen != len(universeWriteRoutes) {
		t.Errorf("found %d write routes, table lists %d; update universeWriteRoutes", seen, len(universeWriteRoutes))
	}
}
