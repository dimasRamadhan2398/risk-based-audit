package riskclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const token = "Bearer caller-token"

func newServer(t *testing.T, status int, bodies map[string]string, hits *int32, auth *string) *RiskClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(hits, 1)
		*auth = r.Header.Get("Authorization")
		if status != http.StatusOK {
			w.WriteHeader(status)
			_, _ = io.WriteString(w, `{"success":false,"error":"nope"}`)
			return
		}
		body, ok := bodies[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL+"/", nil)
}

func TestListCorporateEntitiesFlattensAndForwardsToken(t *testing.T) {
	var hits int32
	var auth string
	c := newServer(t, http.StatusOK, map[string]string{
		"/api/v1/audit-universe/corporate": `{"success":true,"data":[
			{"id":"r1","name":"Root","children":[
				{"id":"c1","name":"Child","parent_id":"r1","children":[{"id":"g1","name":"Grandchild","parent_id":"c1"}]}
			]},
			{"id":"r2","name":"Root 2"}
		]}`,
	}, &hits, &auth)

	got, err := c.ListCorporateEntities(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if auth != token {
		t.Fatalf("Authorization = %q", auth)
	}
	var ids []string
	for _, e := range got {
		if len(e.Children) != 0 {
			t.Fatalf("children not cleared on %s", e.ID)
		}
		ids = append(ids, e.ID)
	}
	if strings.Join(ids, ",") != "r1,c1,g1,r2" {
		t.Fatalf("ids = %v", ids)
	}
}

func TestListYearUniverse(t *testing.T) {
	var hits int32
	var auth string
	c := newServer(t, http.StatusOK, map[string]string{
		"/api/v1/audit-universe/year/2026": `{"success":true,"data":[
			{"corporate_audit_universe_id":"c1","year":2026,"risk_level":"High","scored_at":"2026-01-02T00:00:00Z",
			 "corporate_audit_universe":{"id":"c1","name":"Child"}},
			{"corporate_audit_universe_id":"c2","year":2026,"risk_level":"","scored_at":null}
		]}`,
	}, &hits, &auth)

	rows, err := c.ListYearUniverse(context.Background(), token, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].RiskLevel != "High" || rows[0].CorporateAuditUniverse == nil ||
		rows[0].CorporateAuditUniverse.Name != "Child" || rows[1].ScoredAt != nil {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestNoTokenOrNoURLMakesNoCall(t *testing.T) {
	var hits int32
	var auth string
	c := newServer(t, http.StatusOK, map[string]string{}, &hits, &auth)
	if _, err := c.ListCorporateEntities(context.Background(), "  "); !errors.Is(err, ErrNoAuthorization) {
		t.Fatalf("err = %v", err)
	}
	if hits != 0 {
		t.Fatalf("called risk-service without a token")
	}
	if _, err := New("", nil).ListYearUniverse(context.Background(), token, 2026); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err = %v", err)
	}
	if _, err := c.ListYearUniverse(context.Background(), token, 0); err == nil {
		t.Fatal("year 0 accepted")
	}
}

func TestNon200IsError(t *testing.T) {
	var hits int32
	var auth string
	c := newServer(t, http.StatusUnauthorized, nil, &hits, &auth)
	if _, err := c.ListCorporateEntities(context.Background(), token); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("err = %v, want 401", err)
	}
}

func TestSuccessFalseIsError(t *testing.T) {
	var hits int32
	var auth string
	c := newServer(t, http.StatusOK, map[string]string{
		"/api/v1/audit-universe/corporate": `{"success":false,"error":"db down","data":null}`,
	}, &hits, &auth)
	if _, err := c.ListCorporateEntities(context.Background(), token); err == nil {
		t.Fatal("success:false accepted")
	}
}
