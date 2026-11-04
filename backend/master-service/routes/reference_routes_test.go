package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	referenceCtrl "master-service/controllers/reference"
	"master-service/pkg/reference"
)

type refEnvelope struct {
	Success bool              `json:"success"`
	Data    []json.RawMessage `json:"data"`
	Error   *struct {
		Code string `json:"code"`
	} `json:"error"`
}

func doRefReq(t *testing.T, path, token string, header map[string]string) (*httptest.ResponseRecorder, refEnvelope) {
	t.Helper()
	engine := viewerTestEngine(t)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	var env refEnvelope
	if w.Code != http.StatusNotModified && w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
			t.Fatalf("%s: decode %q: %v", path, w.Body.String(), err)
		}
	}
	return w, env
}

var refPaths = []string{
	"/api/v1/reference/provinces",
	"/api/v1/reference/provinces/32/regencies",
	"/api/v1/reference/regencies",
	"/api/v1/reference/countries",
}

func TestReferenceRequiresAuth(t *testing.T) {
	for _, p := range refPaths {
		if w, _ := doRefReq(t, p, "", nil); w.Code != http.StatusUnauthorized {
			t.Errorf("%s anonymous = %d, want 401", p, w.Code)
		}
	}
}

func TestReferenceViewerForbidden(t *testing.T) {
	tok := viewerTestToken(t, "viewer")
	for _, p := range refPaths {
		if w, _ := doRefReq(t, p, tok, nil); w.Code != http.StatusForbidden {
			t.Errorf("%s VIEWER = %d, want 403", p, w.Code)
		}
	}
}

func TestReferenceLists(t *testing.T) {
	tok := viewerTestToken(t, "AUDITEE")
	cases := []struct {
		path string
		want int
	}{
		{"/api/v1/reference/provinces", 38},
		{"/api/v1/reference/provinces?include_inactive=true", 38},
		{"/api/v1/reference/provinces/32/regencies", 27},
		{"/api/v1/reference/regencies", 514},
		{"/api/v1/reference/regencies?province=32&q=BANDUNG", 3},
		{"/api/v1/reference/regencies?q=kota%20administrasi", 5},
		{"/api/v1/reference/countries", 249},
		{"/api/v1/reference/countries?q=jepang", 1},
		{"/api/v1/reference/countries?q=zzzz", 0},
	}
	for _, tc := range cases {
		w, env := doRefReq(t, tc.path, tok, nil)
		if w.Code != http.StatusOK || !env.Success {
			t.Errorf("%s = %d %s", tc.path, w.Code, w.Body.String())
			continue
		}
		if len(env.Data) != tc.want {
			t.Errorf("%s returned %d items, want %d", tc.path, len(env.Data), tc.want)
		}
		if cc := w.Header().Get("Cache-Control"); cc != referenceCtrl.CacheControl {
			t.Errorf("%s Cache-Control = %q", tc.path, cc)
		}
	}

	_, env := doRefReq(t, "/api/v1/reference/provinces/31/regencies", tok, nil)
	var first reference.Regency
	if err := json.Unmarshal(env.Data[0], &first); err != nil || first.Code != "31.01" || first.Type != reference.TypeKabupaten || first.ProvinceCode != "31" {
		t.Errorf("first DKI regency = %+v (%v)", first, err)
	}
}

func TestReferenceNotFoundAndBadInput(t *testing.T) {
	tok := viewerTestToken(t, "ADMIN")
	for _, p := range []string{"/api/v1/reference/provinces/99/regencies", "/api/v1/reference/regencies?province=99"} {
		w, env := doRefReq(t, p, tok, nil)
		if w.Code != http.StatusNotFound || env.Error == nil || env.Error.Code != "NOT_FOUND" {
			t.Errorf("%s = %d %s, want 404 NOT_FOUND", p, w.Code, w.Body.String())
		}
	}
	if w, _ := doRefReq(t, "/api/v1/reference/provinces?include_inactive=maybe", tok, nil); w.Code != http.StatusBadRequest {
		t.Errorf("bad include_inactive = %d, want 400", w.Code)
	}
}

func TestReferenceETag(t *testing.T) {
	tok := viewerTestToken(t, "AUDITOR")
	d := reference.MustDefault()
	cases := map[string]string{
		"/api/v1/reference/provinces":              d.RegionsVersion(),
		"/api/v1/reference/provinces/32/regencies": d.RegionsVersion(),
		"/api/v1/reference/regencies?q=bandung":    d.RegionsVersion(),
		"/api/v1/reference/countries":              d.CountriesVersion(),
	}
	for path, version := range cases {
		w, _ := doRefReq(t, path, tok, nil)
		etag := w.Header().Get("ETag")
		if etag != `"`+version+`"` {
			t.Errorf("%s ETag = %q, want %q", path, etag, version)
		}
		w, _ = doRefReq(t, path, tok, map[string]string{"If-None-Match": `"old", ` + etag})
		if w.Code != http.StatusNotModified || w.Body.Len() != 0 {
			t.Errorf("%s revalidate = %d (body %d bytes), want 304 empty", path, w.Code, w.Body.Len())
		}
		if w, _ = doRefReq(t, path, tok, map[string]string{"If-None-Match": `"stale"`}); w.Code != http.StatusOK {
			t.Errorf("%s stale etag = %d, want 200", path, w.Code)
		}
	}
	// 304 must not bypass auth.
	if w, _ := doRefReq(t, "/api/v1/reference/provinces", "", map[string]string{"If-None-Match": `"` + d.RegionsVersion() + `"`}); w.Code != http.StatusUnauthorized {
		t.Errorf("anonymous revalidate = %d, want 401", w.Code)
	}
}
