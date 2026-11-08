package controllers

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

	"risk-service/models"
	"risk-service/pkg/masterclient"
	"risk-service/pkg/sqlitetest"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type stubDepartments struct {
	known map[uuid.UUID]string
	down  bool
	calls int
}

func (s *stubDepartments) LookupDepartment(_ context.Context, id uuid.UUID) (masterclient.Department, error) {
	s.calls++
	if s.down {
		return masterclient.Department{}, masterclient.ErrDepartmentsUnavailable
	}
	if name, ok := s.known[id]; ok {
		return masterclient.Department{ID: id, Name: name, IsActive: true}, nil
	}
	return masterclient.Department{}, masterclient.ErrDepartmentNotFound
}

type rcmEnv struct {
	db     *gorm.DB
	depts  *stubDepartments
	engine *gin.Engine
	finID  uuid.UUID
	itID   uuid.UUID
}

func newRCMEnv(t *testing.T) *rcmEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t,
		&models.RiskControlMatrix{}, &models.CosoComponent{}, &models.CosoPrinciple{}, &models.RcmCosoPrinciple{},
		&models.RiskProfile{}, &models.RiskRegister{}, &models.RiskAssessment{}, &models.RiskMitigation{},
	)
	if err := services.SeedCosoCatalogue(db); err != nil {
		t.Fatal(err)
	}
	env := &rcmEnv{db: db, finID: uuid.New(), itID: uuid.New()}
	env.depts = &stubDepartments{known: map[uuid.UUID]string{env.finID: "Finance & Accounting", env.itID: "Information Technology"}}
	ctrl := NewRCMController(db, env.depts, nil)
	ctrl.now = func() time.Time { return time.Date(2026, 8, 1, 3, 0, 0, 0, time.UTC) }
	r := gin.New()
	g := r.Group("/api/v1")
	g.GET("/rcm/coso-principles", ctrl.ListCosoPrinciples)
	g.GET("/rcm/coso-signals", ctrl.GetCosoSignals)
	g.GET("/rcm", ctrl.ListRCM)
	g.POST("/rcm", ctrl.CreateRCM)
	g.PUT("/rcm/:id", ctrl.UpdateRCM)
	g.DELETE("/rcm/:id", ctrl.DeleteRCM)
	g.GET("/rcm/summary", ctrl.GetRCMSummary)
	env.engine = r
	return env
}

type apiResp struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   struct {
		Code   string            `json:"code"`
		Fields map[string]string `json:"fields"`
	} `json:"error"`
}

func (e *rcmEnv) do(t *testing.T, method, path string, body any) (int, apiResp) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if s, ok := body.(string); ok {
			buf.WriteString(s)
		} else if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)
	var out apiResp
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

type rcmRow struct {
	ID                 string   `json:"id"`
	Department         string   `json:"department"`
	DepartmentID       *string  `json:"department_id"`
	CosoComponentCode  *string  `json:"coso_component_code"`
	CosoPrincipleCodes []string `json:"coso_principle_codes"`
	Notes              string   `json:"notes"`
}

func baseBody(extra map[string]any) map[string]any {
	b := map[string]any{
		"risk_code": "FIN-1", "risk_event": "e", "control_code": "CTL-1", "control_description": "d",
		"control_type": "Preventive", "control_owner": "o", "department": "Head Office", "year": 2026,
	}
	for k, v := range extra {
		b[k] = v
	}
	return b
}

func (e *rcmEnv) create(t *testing.T, extra map[string]any) rcmRow {
	t.Helper()
	code, resp := e.do(t, "POST", "/api/v1/rcm", baseBody(extra))
	if code != http.StatusCreated {
		t.Fatalf("create = %d %+v", code, resp.Error)
	}
	var row rcmRow
	if err := json.Unmarshal(resp.Data, &row); err != nil {
		t.Fatal(err)
	}
	return row
}

func (e *rcmEnv) list(t *testing.T, query string) []rcmRow {
	t.Helper()
	code, resp := e.do(t, "GET", "/api/v1/rcm"+query, nil)
	if code != http.StatusOK {
		t.Fatalf("list%s = %d %+v", query, code, resp.Error)
	}
	var rows []rcmRow
	if err := json.Unmarshal(resp.Data, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

func eqCodes(a []string, b ...string) bool { return strings.Join(a, ",") == strings.Join(b, ",") }

func TestRCMCreateWithDepartmentAndTags(t *testing.T) {
	e := newRCMEnv(t)
	row := e.create(t, map[string]any{
		"department_id": e.finID.String(), "coso_component_code": "ca",
		"coso_principle_codes": []string{"p12", "P10", "P10", " P08 "},
	})
	if row.DepartmentID == nil || *row.DepartmentID != e.finID.String() || row.Department != "Finance & Accounting" {
		t.Errorf("department = %v %q", row.DepartmentID, row.Department)
	}
	if row.CosoComponentCode == nil || *row.CosoComponentCode != "CA" {
		t.Errorf("component = %v", row.CosoComponentCode)
	}
	if !eqCodes(row.CosoPrincipleCodes, "P08", "P10", "P12") {
		t.Errorf("codes = %v", row.CosoPrincipleCodes)
	}
	var n int64
	e.db.Model(&models.RcmCosoPrinciple{}).Count(&n)
	if n != 3 {
		t.Errorf("tag rows = %d, want 3", n)
	}
	got := e.list(t, "")
	if len(got) != 1 || !eqCodes(got[0].CosoPrincipleCodes, "P08", "P10", "P12") {
		t.Errorf("list = %+v", got)
	}
}

func TestRCMValidation(t *testing.T) {
	e := newRCMEnv(t)
	cases := []struct {
		name  string
		extra map[string]any
		field string
	}{
		{"unknown principle", map[string]any{"coso_principle_codes": []string{"P10", "P99"}}, "coso_principle_codes"},
		{"unknown component", map[string]any{"coso_component_code": "XX", "coso_principle_codes": []string{"P10"}}, "coso_component_code"},
		{"component without matching tag", map[string]any{"coso_component_code": "CE", "coso_principle_codes": []string{"P10"}}, "coso_component_code"},
		{"component without tags", map[string]any{"coso_component_code": "CA"}, "coso_component_code"},
		{"bad department uuid", map[string]any{"department_id": "Head Office"}, "department_id"},
		{"unknown department", map[string]any{"department_id": uuid.NewString()}, "department_id"},
	}
	for _, c := range cases {
		code, resp := e.do(t, "POST", "/api/v1/rcm", baseBody(c.extra))
		if code != http.StatusUnprocessableEntity || resp.Error.Code != "VALIDATION_ERROR" || resp.Error.Fields[c.field] == "" {
			t.Errorf("%s: %d %+v", c.name, code, resp.Error)
		}
	}
	var n int64
	e.db.Model(&models.RiskControlMatrix{}).Count(&n)
	if n != 0 {
		t.Errorf("%d rows written by refused requests", n)
	}

	e.depts.down = true
	code, resp := e.do(t, "POST", "/api/v1/rcm", baseBody(map[string]any{"department_id": e.finID.String()}))
	if code != http.StatusServiceUnavailable || resp.Error.Code != "DEPARTMENT_MASTER_UNAVAILABLE" {
		t.Errorf("master down: %d %+v", code, resp.Error)
	}
}

func TestRCMUpdateTagSemantics(t *testing.T) {
	e := newRCMEnv(t)
	row := e.create(t, map[string]any{
		"department_id": e.finID.String(), "coso_component_code": "CA", "coso_principle_codes": []string{"P10", "P11"},
	})
	put := func(extra map[string]any) (int, rcmRow, apiResp) {
		code, resp := e.do(t, "PUT", "/api/v1/rcm/"+row.ID, baseBody(extra))
		var out rcmRow
		_ = json.Unmarshal(resp.Data, &out)
		return code, out, resp
	}

	// Old client: no COSO/department fields -> everything kept, master name kept.
	code, got, _ := put(map[string]any{"notes": "edited", "department": "Head Office"})
	if code != 200 || !eqCodes(got.CosoPrincipleCodes, "P10", "P11") || got.CosoComponentCode == nil ||
		got.DepartmentID == nil || got.Department != "Finance & Accounting" || got.Notes != "edited" {
		t.Fatalf("omitted fields: %d %+v", code, got)
	}

	// Master down while the department is unchanged: accepted, name kept.
	e.depts.down = true
	code, got, _ = put(map[string]any{"department_id": e.finID.String()})
	if code != 200 || got.Department != "Finance & Accounting" {
		t.Errorf("unchanged department with master down: %d %+v", code, got)
	}
	e.depts.down = false

	// Removing the tag the component relies on -> 422, tags untouched.
	code, _, resp := put(map[string]any{"coso_principle_codes": []string{"P13"}})
	if code != 422 || resp.Error.Fields["coso_component_code"] == "" {
		t.Errorf("orphan component: %d %+v", code, resp.Error)
	}
	if rows := e.list(t, ""); !eqCodes(rows[0].CosoPrincipleCodes, "P10", "P11") {
		t.Errorf("refused update changed tags: %v", rows[0].CosoPrincipleCodes)
	}

	// Replace set and move department.
	code, got, _ = put(map[string]any{"coso_principle_codes": []string{"P12"}, "department_id": e.itID.String()})
	if code != 200 || !eqCodes(got.CosoPrincipleCodes, "P12") || got.Department != "Information Technology" {
		t.Errorf("replace: %d %+v", code, got)
	}

	// [] clears tags (component must be cleared with it); null clears department.
	code, got, _ = put(map[string]any{"coso_principle_codes": []string{}, "coso_component_code": nil, "department_id": nil, "department": "Head Office"})
	if code != 200 || len(got.CosoPrincipleCodes) != 0 || got.CosoPrincipleCodes == nil || got.CosoComponentCode != nil ||
		got.DepartmentID != nil || got.Department != "Head Office" {
		t.Errorf("clear: %d %+v", code, got)
	}
	var n int64
	e.db.Model(&models.RcmCosoPrinciple{}).Count(&n)
	if n != 0 {
		t.Errorf("tag rows = %d after clear", n)
	}

	// null tags = unchanged.
	put(map[string]any{"coso_principle_codes": []string{"P01"}})
	code, got, _ = put(map[string]any{"coso_principle_codes": nil})
	if code != 200 || !eqCodes(got.CosoPrincipleCodes, "P01") {
		t.Errorf("null tags: %d %+v", code, got)
	}
}

func TestRCMListLegacyRowsAndFilters(t *testing.T) {
	e := newRCMEnv(t)
	// Legacy row written as before the change: no department_id, no tags.
	legacy := models.RiskControlMatrix{ID: uuid.New(), RiskCode: "L", RiskEvent: "e", ControlCode: "L-1", ControlDescription: "d",
		ControlType: "Preventive", ControlOwner: "o", Department: "Bali Branch", Year: 2026}
	if err := e.db.Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	fin := e.create(t, map[string]any{"department_id": e.finID.String(), "coso_component_code": "CA", "coso_principle_codes": []string{"P10"}})
	it := e.create(t, map[string]any{"department_id": e.itID.String(), "coso_component_code": "MA", "coso_principle_codes": []string{"P16", "P11"}})

	all := e.list(t, "")
	if len(all) != 3 {
		t.Fatalf("list = %d rows", len(all))
	}
	for _, r := range all {
		if r.ID == legacy.ID.String() && (r.DepartmentID != nil || r.CosoComponentCode != nil || r.CosoPrincipleCodes == nil || len(r.CosoPrincipleCodes) != 0) {
			t.Errorf("legacy row = %+v", r)
		}
	}
	ids := func(rows []rcmRow) string {
		var s []string
		for _, r := range rows {
			s = append(s, r.ID)
		}
		return strings.Join(s, ",")
	}
	if got := e.list(t, "?department_id="+e.finID.String()); ids(got) != fin.ID {
		t.Errorf("department_id filter = %s", ids(got))
	}
	if got := e.list(t, "?department_id=unassigned"); ids(got) != legacy.ID.String() {
		t.Errorf("unassigned filter = %s", ids(got))
	}
	if got := e.list(t, "?component=ma"); ids(got) != it.ID {
		t.Errorf("component filter = %s", ids(got))
	}
	if got := e.list(t, "?principle=P11"); ids(got) != it.ID {
		t.Errorf("principle filter = %s", ids(got))
	}
	if got := e.list(t, "?principle=P10&department_id="+e.itID.String()); len(got) != 0 {
		t.Errorf("combined filter = %s", ids(got))
	}
	if got := e.list(t, "?department=Bali%20Branch"); ids(got) != legacy.ID.String() {
		t.Errorf("legacy department text filter = %s", ids(got))
	}
	for _, q := range []string{"?department_id=x", "?component=ZZ", "?principle=P18", "?principle=1%3BDROP"} {
		if code, resp := e.do(t, "GET", "/api/v1/rcm"+q, nil); code != 422 || resp.Error.Code != "VALIDATION_ERROR" {
			t.Errorf("GET %s = %d %+v", q, code, resp.Error)
		}
	}

	// Soft-deleted rows disappear from filters; their tags are never read.
	if code, _ := e.do(t, "DELETE", "/api/v1/rcm/"+it.ID, nil); code != 200 {
		t.Fatal("delete failed")
	}
	if got := e.list(t, "?principle=P11"); len(got) != 0 {
		t.Errorf("deleted row still listed: %s", ids(got))
	}
}

func TestRCMSummaryEmptyIsNull(t *testing.T) {
	e := newRCMEnv(t)
	code, resp := e.do(t, "GET", "/api/v1/rcm/summary?year=2030", nil)
	if code != 200 {
		t.Fatal(code)
	}
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"internal_control_effectiveness", "rating", "interpretation", "avg_total_weighted_score",
		"avg_design_effectiveness", "avg_operating_effectiveness", "avg_coverage_completeness", "avg_timeliness", "avg_automation_monitoring"} {
		v, ok := data[k]
		if !ok || v != nil {
			t.Errorf("%s = %v (present %v), want null", k, v, ok)
		}
	}
	if data["total_controls"].(float64) != 0 || data["total_inherent_risk"].(float64) != 0 {
		t.Errorf("counts = %v", data)
	}

	e.create(t, map[string]any{"department_id": e.finID.String(), "design_effectiveness_rating": 5})
	_, resp = e.do(t, "GET", "/api/v1/rcm/summary?year=2026&department_id="+e.finID.String(), nil)
	_ = json.Unmarshal(resp.Data, &data)
	if data["total_controls"].(float64) != 1 || data["internal_control_effectiveness"].(float64) != 60 || data["rating"] != "Weak" {
		t.Errorf("non-empty summary = %v", data)
	}
	_, resp = e.do(t, "GET", "/api/v1/rcm/summary?department_id=unassigned", nil)
	_ = json.Unmarshal(resp.Data, &data)
	if data["total_controls"].(float64) != 0 || data["rating"] != nil {
		t.Errorf("unassigned summary = %v", data)
	}
}

func TestListCosoPrinciples(t *testing.T) {
	e := newRCMEnv(t)
	code, resp := e.do(t, "GET", "/api/v1/rcm/coso-principles", nil)
	if code != 200 {
		t.Fatal(code)
	}
	var data struct {
		Framework  string `json:"framework"`
		Components []struct {
			Code       string `json:"code"`
			SortOrder  int    `json:"sort_order"`
			NameEn     string `json:"name_en"`
			Principles []struct {
				Code          string `json:"code"`
				Number        int    `json:"number"`
				ComponentCode string `json:"component_code"`
				IsActive      bool   `json:"is_active"`
			} `json:"principles"`
		} `json:"components"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Framework != "COSO_2013" || len(data.Components) != 5 || data.Components[0].Code != "CE" || data.Components[4].Code != "MA" {
		t.Fatalf("catalogue = %+v", data)
	}
	total := 0
	for _, c := range data.Components {
		for _, p := range c.Principles {
			if p.ComponentCode != c.Code || !p.IsActive {
				t.Errorf("principle %+v under %s", p, c.Code)
			}
			total++
		}
	}
	if total != 17 || data.Components[0].Principles[0].Code != "P01" {
		t.Errorf("principles = %d", total)
	}
	code, resp = e.do(t, "GET", "/api/v1/rcm/coso-principles?framework=COSO_ERM_2017", nil)
	if code != 200 || !strings.Contains(string(resp.Data), `"components":[]`) {
		t.Errorf("other framework = %d %s", code, resp.Data)
	}
	if code, _ := e.do(t, "GET", "/api/v1/rcm/coso-principles?framework=x%27%3B--", nil); code != 422 {
		t.Errorf("bad framework = %d", code)
	}
}

func TestGetCosoSignalsEndpoint(t *testing.T) {
	e := newRCMEnv(t)
	for _, q := range []string{"", "?year=2026", "?department_id=" + e.finID.String(),
		"?department_id=" + e.finID.String() + "&year=2026&quarter=5",
		"?department_id=" + e.finID.String() + "&year=2026&as_of=2026/01/01",
		"?department_id=" + e.finID.String() + "&year=2026&location_id=jkt"} {
		if code, resp := e.do(t, "GET", "/api/v1/rcm/coso-signals"+q, nil); code != 422 || resp.Error.Code != "VALIDATION_ERROR" {
			t.Errorf("GET coso-signals%s = %d", q, code)
		}
	}

	e.create(t, map[string]any{"department_id": e.finID.String(), "coso_component_code": "CA", "coso_principle_codes": []string{"P10"}})
	code, resp := e.do(t, "GET", "/api/v1/rcm/coso-signals?department_id="+e.finID.String()+"&year=2026&quarter=2", nil)
	if code != 200 {
		t.Fatalf("signals = %d %+v", code, resp.Error)
	}
	var data services.CosoSignalsResult
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatal(err)
	}
	// ctrl.now is 2026-08-01, past Q2's end: as_of is clamped to 2026-06-30.
	if data.Scope.AsOf != "2026-06-30" || data.Scope.PeriodStart != "2026-04-01" || data.RCMCount != 1 || len(data.Signals) != 3 {
		t.Errorf("scope = %+v count=%d signals=%d", data.Scope, data.RCMCount, len(data.Signals))
	}
	for _, s := range data.Signals {
		switch s.SignalKey {
		case services.SignalCARCMEffectiveness:
			if s.State != "OK" || s.ValuePct == nil || *s.ValuePct != 60 {
				t.Errorf("ca = %+v", s)
			}
		default:
			if s.State != "NO_DATA" { // the row has no linked risk
				t.Errorf("%s = %s", s.SignalKey, s.State)
			}
		}
	}
	code, resp = e.do(t, "GET", "/api/v1/rcm/coso-signals?department_id="+e.finID.String()+"&year=2026&as_of=2026-03-31", nil)
	_ = json.Unmarshal(resp.Data, &data)
	if code != 200 || data.Scope.AsOf != "2026-03-31" || strings.Trim(strings.Join(strings.Fields(strings.Trim(jsonOf(data.Scope.DueQuarters), "[]")), ""), ",") != "1" {
		t.Errorf("explicit as_of: %d %+v", code, data.Scope)
	}
}

func jsonOf(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// The optionalString decoder tells omitted, null and value apart.
func TestOptionalStringDecoding(t *testing.T) {
	var req RCMRequest
	if err := json.Unmarshal([]byte(`{"department_id":null,"coso_component_code":"CA"}`), &req); err != nil {
		t.Fatal(err)
	}
	if !req.DepartmentID.Set || req.DepartmentID.Value != nil || !req.CosoComponentCode.Set || *req.CosoComponentCode.Value != "CA" || req.CosoPrincipleCodes != nil {
		t.Errorf("decoded %+v", req)
	}
	var omitted RCMRequest
	_ = json.Unmarshal([]byte(`{}`), &omitted)
	if omitted.DepartmentID.Set || omitted.CosoComponentCode.Set {
		t.Error("omitted fields reported as set")
	}
	if err := json.Unmarshal([]byte(`{"department_id":5}`), &omitted); err == nil {
		t.Error("a number department_id must not decode")
	}
	_ = errors.New
}
