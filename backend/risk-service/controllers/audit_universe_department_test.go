package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"risk-service/models"
	"risk-service/pkg/masterclient"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fakeDepartments struct {
	byCode map[string]uuid.UUID
	names  map[uuid.UUID]string
}

func (f fakeDepartments) CallerDepartment(_ context.Context, code, _ string) (uuid.UUID, error) {
	if d, ok := f.byCode[code]; ok {
		return d, nil
	}
	return uuid.Nil, errors.Join(masterclient.ErrDepartmentUnresolved, errors.New("unknown"))
}

func (f fakeDepartments) DepartmentName(_ context.Context, id uuid.UUID) (string, bool) {
	n, ok := f.names[id]
	return n, ok
}

type deptFixture struct {
	db           *gorm.DB
	router       *gin.Engine
	finance, hr  uuid.UUID
	finNode      models.CorporateAuditUniverse // finance dept
	hrNode       models.CorporateAuditUniverse // hr dept
	nullNode     models.CorporateAuditUniverse // no department
	finChild     models.CorporateAuditUniverse // finance parent, hr child (mixed subtree)
	finParentMix models.CorporateAuditUniverse
}

func newDeptFixture(t *testing.T) *deptFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := openTestDB(t)
	now := time.Now()
	f := &deptFixture{db: db, finance: uuid.New(), hr: uuid.New()}
	mk := func(name string, dept *uuid.UUID, parent *uuid.UUID) models.CorporateAuditUniverse {
		n := models.CorporateAuditUniverse{ID: uuid.New(), Name: name, DepartmentID: dept, ParentID: parent, CreatedAt: now, UpdatedAt: now}
		mustCreate(t, db, &n)
		return n
	}
	fin, hr := f.finance, f.hr
	f.finNode = mk("Finance Ops", &fin, nil)
	f.hrNode = mk("HR Ops", &hr, nil)
	f.nullNode = mk("Legacy", nil, nil)
	f.finParentMix = mk("Finance Mixed", &fin, nil)
	f.finChild = mk("HR under Finance", &hr, &f.finParentMix.ID)

	deps := fakeDepartments{
		byCode: map[string]uuid.UUID{"EMP-FIN": f.finance, "EMP-HR": f.hr},
		names:  map[uuid.UUID]string{f.finance: "Finance", f.hr: "Human Resources"},
	}
	ctrl := NewAuditUniverseController(db, nil, deps)
	r := gin.New()
	r.Use(gin.Recovery())
	// Stand in for Authenticate: roles and employee link come from headers.
	r.Use(func(c *gin.Context) {
		if v := c.GetHeader("X-Roles"); v != "" {
			c.Set("roles", splitComma(v))
		}
		c.Set("employee_code", c.GetHeader("X-Emp"))
	})
	r.GET("/corporate", ctrl.ListCorporateAuditUniverse)
	r.POST("/corporate", ctrl.CreateOrUpdateCorporateNode)
	r.DELETE("/corporate/:id", ctrl.DeleteCorporateNode)
	f.router = r
	return f
}

func splitComma(s string) []string {
	var out []string
	cur := ""
	for _, ch := range s {
		if ch == ',' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(ch)
	}
	return append(out, cur)
}

func (f *deptFixture) do(method, path, roles, emp string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Roles", roles)
	req.Header.Set("X-Emp", emp)
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func (f *deptFixture) deptOf(t *testing.T, id uuid.UUID) *uuid.UUID {
	t.Helper()
	var n models.CorporateAuditUniverse
	if err := f.db.First(&n, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return n.DepartmentID
}

func (f *deptFixture) exists(id uuid.UUID) bool {
	var c int64
	f.db.Model(&models.CorporateAuditUniverse{}).Where("id = ?", id).Count(&c)
	return c == 1
}

func TestDeptHeadUpdate(t *testing.T) {
	f := newDeptFixture(t)
	rename := func(n models.CorporateAuditUniverse, name string) map[string]any {
		return map[string]any{"id": n.ID.String(), "name": name}
	}
	cases := []struct {
		name       string
		roles, emp string
		body       map[string]any
		want       int
	}{
		{"own department", "DEPARTMENT_HEAD", "EMP-FIN", rename(f.finNode, "Renamed"), 200},
		{"other department", "DEPARTMENT_HEAD", "EMP-HR", rename(f.finNode, "Hijack"), 403},
		{"NULL department entry", "DEPARTMENT_HEAD", "EMP-FIN", rename(f.nullNode, "Hijack"), 403},
		{"unresolved department", "DEPARTMENT_HEAD", "EMP-UNKNOWN", rename(f.finNode, "x"), 403},
		{"no employee link", "DEPARTMENT_HEAD", "", rename(f.finNode, "x"), 403},
		{"change department away", "DEPARTMENT_HEAD", "EMP-FIN", map[string]any{"id": f.finNode.ID.String(), "name": "n", "department_id": f.hr.String()}, 403},
		{"clear department", "DEPARTMENT_HEAD", "EMP-FIN", map[string]any{"id": f.finNode.ID.String(), "name": "n", "department_id": ""}, 403},
		{"same department explicitly", "DEPARTMENT_HEAD", "EMP-FIN", map[string]any{"id": f.finNode.ID.String(), "name": "n2", "department_id": f.finance.String()}, 200},
		{"ADMIN other dept", "ADMIN", "", rename(f.hrNode, "Admin edit"), 200},
		{"ADMIN NULL entry", "ADMIN", "", rename(f.nullNode, "Admin edit"), 200},
		{"EXECUTIVE NULL entry", "EXECUTIVE", "", rename(f.nullNode, "Exec edit"), 200},
		{"CAE NULL entry", "CHIEF_AUDIT_EXECUTIVE", "", rename(f.nullNode, "Cae edit"), 200},
		{"ADMIN+DEPARTMENT_HEAD is unscoped", "ADMIN,DEPARTMENT_HEAD", "EMP-FIN", rename(f.hrNode, "union"), 200},
		{"AUDITOR refused", "AUDITOR", "EMP-FIN", rename(f.finNode, "x"), 403},
		{"DEPARTMENT_HEAD+VIEWER own dept", "DEPARTMENT_HEAD,VIEWER", "EMP-FIN", rename(f.finNode, "dh+v"), 200},
	}
	for _, tc := range cases {
		if w := f.do("POST", "/corporate", tc.roles, tc.emp, tc.body); w.Code != tc.want {
			t.Errorf("%s: status %d, want %d (%s)", tc.name, w.Code, tc.want, w.Body.String())
		}
	}
	if d := f.deptOf(t, f.finNode.ID); d == nil || *d != f.finance {
		t.Errorf("finance node department changed: %v", d)
	}
}

func TestDeptHeadCreateForcedToOwnDepartment(t *testing.T) {
	f := newDeptFixture(t)

	w := f.do("POST", "/corporate", "DEPARTMENT_HEAD", "EMP-FIN", map[string]any{"name": "New Finance Entity"})
	if w.Code != 200 {
		t.Fatalf("create = %d (%s)", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			ID             string  `json:"id"`
			DepartmentID   *string `json:"department_id"`
			DepartmentName *string `json:"department_name"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.DepartmentID == nil || *resp.Data.DepartmentID != f.finance.String() {
		t.Fatalf("department_id = %v, want own department", resp.Data.DepartmentID)
	}
	if resp.Data.DepartmentName == nil || *resp.Data.DepartmentName != "Finance" {
		t.Fatalf("department_name = %v, want Finance", resp.Data.DepartmentName)
	}

	// Explicitly asking for another department is refused, not rewritten.
	if w := f.do("POST", "/corporate", "DEPARTMENT_HEAD", "EMP-FIN", map[string]any{"name": "x", "department_id": f.hr.String()}); w.Code != 403 {
		t.Errorf("create in other department = %d, want 403", w.Code)
	}
	// Asking for the own department is fine.
	if w := f.do("POST", "/corporate", "DEPARTMENT_HEAD", "EMP-FIN", map[string]any{"name": "y", "department_id": f.finance.String()}); w.Code != 200 {
		t.Errorf("create in own department = %d, want 200", w.Code)
	}
	// Under a parent of another department / NULL department: refused.
	for name, parent := range map[string]uuid.UUID{"other dept parent": f.hrNode.ID, "NULL dept parent": f.nullNode.ID} {
		if w := f.do("POST", "/corporate", "DEPARTMENT_HEAD", "EMP-FIN", map[string]any{"name": "z", "parent_id": parent.String()}); w.Code != 403 {
			t.Errorf("create under %s = %d, want 403", name, w.Code)
		}
	}
	// Under an own-department parent: fine.
	if w := f.do("POST", "/corporate", "DEPARTMENT_HEAD", "EMP-FIN", map[string]any{"name": "child", "parent_id": f.finNode.ID.String()}); w.Code != 200 {
		t.Errorf("create under own parent = %d, want 200", w.Code)
	}
	// Unresolved department: refused.
	if w := f.do("POST", "/corporate", "DEPARTMENT_HEAD", "EMP-UNKNOWN", map[string]any{"name": "q"}); w.Code != 403 {
		t.Errorf("create with unresolved department = %d, want 403", w.Code)
	}
	// A full writer may pick any department, or none.
	w = f.do("POST", "/corporate", "ADMIN", "", map[string]any{"name": "admin-made", "department_id": f.hr.String()})
	if w.Code != 200 {
		t.Fatalf("admin create = %d", w.Code)
	}
	w = f.do("POST", "/corporate", "ADMIN", "", map[string]any{"name": "admin-null"})
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if w.Code != 200 || resp.Data.DepartmentID != nil {
		t.Errorf("admin create without department: %d, department_id=%v", w.Code, resp.Data.DepartmentID)
	}
}

func TestDeptHeadDelete(t *testing.T) {
	f := newDeptFixture(t)
	del := func(id uuid.UUID, roles, emp string) int {
		return f.do("DELETE", "/corporate/"+id.String(), roles, emp, nil).Code
	}
	if c := del(f.hrNode.ID, "DEPARTMENT_HEAD", "EMP-FIN"); c != 403 || !f.exists(f.hrNode.ID) {
		t.Errorf("delete other department = %d (exists=%v)", c, f.exists(f.hrNode.ID))
	}
	if c := del(f.nullNode.ID, "DEPARTMENT_HEAD", "EMP-FIN"); c != 403 || !f.exists(f.nullNode.ID) {
		t.Errorf("delete NULL department = %d", c)
	}
	if c := del(f.finNode.ID, "DEPARTMENT_HEAD", "EMP-UNKNOWN"); c != 403 || !f.exists(f.finNode.ID) {
		t.Errorf("delete with unresolved department = %d", c)
	}
	// Own entry whose subtree contains another department's child: refused and
	// rolled back, nothing is deleted.
	if c := del(f.finParentMix.ID, "DEPARTMENT_HEAD", "EMP-FIN"); c != 403 || !f.exists(f.finParentMix.ID) || !f.exists(f.finChild.ID) {
		t.Errorf("delete mixed subtree = %d, parent exists=%v child exists=%v", c, f.exists(f.finParentMix.ID), f.exists(f.finChild.ID))
	}
	if c := del(f.finNode.ID, "DEPARTMENT_HEAD", "EMP-FIN"); c != 200 || f.exists(f.finNode.ID) {
		t.Errorf("delete own = %d (still exists=%v)", c, f.exists(f.finNode.ID))
	}
	// Full writers are unscoped, including NULL-department and mixed subtrees.
	if c := del(f.nullNode.ID, "ADMIN", ""); c != 200 {
		t.Errorf("admin delete NULL = %d", c)
	}
	if c := del(f.finParentMix.ID, "EXECUTIVE", ""); c != 200 || f.exists(f.finChild.ID) {
		t.Errorf("executive delete mixed subtree = %d", c)
	}
}

func TestCorporateListCarriesDepartmentFields(t *testing.T) {
	f := newDeptFixture(t)
	w := f.do("GET", "/corporate", "AUDITOR", "", nil)
	if w.Code != 200 {
		t.Fatalf("list = %d", w.Code)
	}
	var resp struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	seen := map[string]map[string]any{}
	for _, n := range resp.Data {
		seen[n["name"].(string)] = n
	}
	for _, n := range resp.Data {
		if _, ok := n["department_id"]; !ok {
			t.Errorf("%v: department_id key missing (must be present, null when unset)", n["name"])
		}
		if _, ok := n["department_name"]; !ok {
			t.Errorf("%v: department_name key missing", n["name"])
		}
	}
	if seen["Legacy"]["department_id"] != nil || seen["Legacy"]["department_name"] != nil {
		t.Errorf("NULL-department entry must have null department fields: %v", seen["Legacy"])
	}
	if seen["Finance Ops"]["department_id"] != f.finance.String() || seen["Finance Ops"]["department_name"] != "Finance" {
		t.Errorf("Finance Ops department fields wrong: %v", seen["Finance Ops"])
	}
}

// A nil department source (misconfiguration) must refuse scoped callers.
func TestDeptHeadRefusedWithoutDepartmentSource(t *testing.T) {
	f := newDeptFixture(t)
	ctrl := NewAuditUniverseController(f.db, nil, nil)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("roles", []string{"DEPARTMENT_HEAD"}); c.Set("employee_code", "EMP-FIN") })
	r.POST("/corporate", ctrl.CreateOrUpdateCorporateNode)
	req := httptest.NewRequest("POST", "/corporate", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("status %d, want 403", w.Code)
	}
}
