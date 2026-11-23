package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"risk-service/pkg/masterclient"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubLocations struct {
	locs []masterclient.Location
	err  error
}

func (s stubLocations) ListLocations(context.Context) ([]masterclient.Location, error) {
	return s.locs, s.err
}

func TestRCMLocationIDValidation(t *testing.T) {
	e := newRCMEnv(t)
	jkt, closed := uuid.New(), uuid.New()
	ctrl := NewRCMController(e.db, e.depts, stubLocations{locs: []masterclient.Location{
		{ID: jkt, Name: "Jakarta", IsActive: true}, {ID: closed, Name: "Closed", IsActive: false},
	}})
	r := gin.New()
	r.POST("/api/v1/rcm", ctrl.CreateRCM)
	r.PUT("/api/v1/rcm/:id", ctrl.UpdateRCM)
	e.engine = r

	code, resp := e.do(t, "POST", "/api/v1/rcm", baseBody(map[string]any{"department_id": e.finID.String(), "location_id": jkt.String()}))
	if code != http.StatusCreated {
		t.Fatalf("create with location = %d %+v", code, resp.Error)
	}
	var row struct {
		ID         string  `json:"id"`
		LocationID *string `json:"location_id"`
	}
	if err := json.Unmarshal(resp.Data, &row); err != nil {
		t.Fatal(err)
	}
	if row.LocationID == nil || *row.LocationID != jkt.String() {
		t.Fatalf("location_id = %v", row.LocationID)
	}

	for name, loc := range map[string]string{"inactive": closed.String(), "unknown": uuid.NewString(), "not a uuid": "x"} {
		code, resp := e.do(t, "POST", "/api/v1/rcm", baseBody(map[string]any{"department_id": e.finID.String(), "location_id": loc}))
		if code != http.StatusUnprocessableEntity || resp.Error.Fields["location_id"] == "" {
			t.Errorf("%s: %d %+v", name, code, resp.Error)
		}
	}

	// Omitted on PUT keeps it; null clears it.
	code, resp = e.do(t, "PUT", "/api/v1/rcm/"+row.ID, baseBody(map[string]any{"notes": "n"}))
	if code != http.StatusOK {
		t.Fatalf("put = %d", code)
	}
	_ = json.Unmarshal(resp.Data, &row)
	if row.LocationID == nil {
		t.Errorf("omitted location_id cleared the branch")
	}
	code, resp = e.do(t, "PUT", "/api/v1/rcm/"+row.ID, baseBody(map[string]any{"location_id": nil}))
	if code != http.StatusOK {
		t.Fatalf("put null = %d", code)
	}
	_ = json.Unmarshal(resp.Data, &row)
	if row.LocationID != nil {
		t.Errorf("null location_id kept %v", *row.LocationID)
	}

	// Location master down: a new link is refused with 503.
	down := NewRCMController(e.db, e.depts, stubLocations{err: errors.New("down")})
	r2 := gin.New()
	r2.POST("/api/v1/rcm", down.CreateRCM)
	e.engine = r2
	if code, resp := e.do(t, "POST", "/api/v1/rcm", baseBody(map[string]any{"location_id": jkt.String()})); code != http.StatusServiceUnavailable || resp.Error.Code != "LOCATION_MASTER_UNAVAILABLE" {
		t.Errorf("down = %d %s", code, resp.Error.Code)
	}
}
