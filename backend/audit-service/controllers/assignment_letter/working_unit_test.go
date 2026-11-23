package assignmentletter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"audit-service/pkg/masterclient"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubDepts struct {
	known map[string]masterclient.DepartmentDetail
	err   error
}

func (s stubDepts) GetDepartment(_ context.Context, _, id string) (*masterclient.DepartmentDetail, error) {
	if s.err != nil {
		return nil, s.err
	}
	if d, ok := s.known[id]; ok {
		return &d, nil
	}
	return nil, masterclient.ErrNotFound
}

func runWorkingUnit(t *testing.T, depts DepartmentGetter, body string) (int, map[string]interface{}) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	var seen map[string]interface{}
	r := gin.New()
	r.POST("/letters", ResolveWorkingUnit(nil, depts), func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		_ = json.Unmarshal(b, &seen)
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/letters", strings.NewReader(body)))
	return w.Code, seen
}

func TestResolveWorkingUnit(t *testing.T) {
	fin, old := uuid.NewString(), uuid.NewString()
	depts := stubDepts{known: map[string]masterclient.DepartmentDetail{
		fin: {ID: fin, DepartmentName: "Finance", DepartmentCode: "FIN", IsActive: true},
		old: {ID: old, DepartmentName: "Old", IsActive: false},
	}}

	code, seen := runWorkingUnit(t, depts, `{"workingUnit":"typed","workingUnitDepartmentId":"`+fin+`"}`)
	if code != 200 || seen["workingUnit"] != "Finance (FIN)" || seen["workingUnitDepartmentId"] != fin {
		t.Errorf("linked = %d %v", code, seen)
	}
	code, seen = runWorkingUnit(t, depts, `{"workingUnit":"typed"}`)
	if code != 200 || seen["workingUnit"] != "typed" {
		t.Errorf("omitted = %d %v", code, seen)
	}
	code, seen = runWorkingUnit(t, depts, `{"workingUnitDepartmentId":""}`)
	if v, ok := seen["workingUnitDepartmentId"]; code != 200 || !ok || v != nil {
		t.Errorf("cleared = %d %v", code, seen)
	}
	for name, id := range map[string]string{"inactive": old, "unknown": uuid.NewString(), "not a uuid": "x"} {
		if code, _ := runWorkingUnit(t, depts, `{"workingUnitDepartmentId":"`+id+`"}`); code != http.StatusUnprocessableEntity {
			t.Errorf("%s = %d, want 422", name, code)
		}
	}
	if code, _ := runWorkingUnit(t, stubDepts{err: errors.New("down")}, `{"workingUnitDepartmentId":"`+fin+`"}`); code != http.StatusServiceUnavailable {
		t.Errorf("master down = %d, want 503", code)
	}
}
