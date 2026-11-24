package provisioning

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"auth-service/pkg/config"
	"auth-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const jobID = "0b6a3c1e-1d2f-4a5b-8c7d-9e0f1a2b3c4d"

// engine mounts the controller with a fake authenticated caller.
func engine(ctl *Controller) *gin.Engine {
	logger.Log = zap.NewNop()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("user_id", "admin-uuid"); c.Set("username", "admin"); c.Next() })
	g := r.Group("/api/v1/provisioning")
	g.POST("/jobs", ctl.CreateJob)
	g.GET("/jobs", ctl.ListJobs)
	g.GET("/jobs/:id", ctl.GetJob)
	g.GET("/jobs/:id/events", ctl.StreamEvents)
	g.GET("/ports/next", ctl.NextPorts)
	return r
}

func serve(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestNotConfiguredIs503(t *testing.T) {
	r := engine(NewController(config.ProvisionerConfig{}))
	for _, p := range []string{"/api/v1/provisioning/jobs", "/api/v1/provisioning/ports/next", "/api/v1/provisioning/jobs/" + jobID + "/events"} {
		if w := serve(r, "GET", p, ""); w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "PROVISIONER_UNAVAILABLE") {
			t.Errorf("%s: %d %s", p, w.Code, w.Body.String())
		}
	}
}

func TestUnreachableIs503(t *testing.T) {
	up := httptest.NewServer(http.NotFoundHandler())
	url := up.URL
	up.Close() // nothing listens there now
	r := engine(NewController(config.ProvisionerConfig{URL: url, Token: "t"}))
	if w := serve(r, "GET", "/api/v1/provisioning/jobs", ""); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d %s", w.Code, w.Body.String())
	}
	if w := serve(r, "POST", "/api/v1/provisioning/jobs", `{"slug":"bai"}`); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("create status = %d", w.Code)
	}
}

func TestForwardsWithTokenAndCreator(t *testing.T) {
	var gotBody map[string]any
	var gotToken, gotUser, gotPath string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken, gotUser, gotPath = r.Header.Get(tokenHeader), r.Header.Get(userHeader), r.URL.RequestURI()
		if r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(w, `{"id":"`+jobID+`","status":"queued"}`)
			return
		}
		_, _ = io.WriteString(w, `{"jobs":[]}`)
	}))
	defer up.Close()
	r := engine(NewController(config.ProvisionerConfig{URL: up.URL + "/", Token: "secret-token"}))

	w := serve(r, "POST", "/api/v1/provisioning/jobs", `{"slug":"bai","client_name":"PT \"B\" $(id)","created_by":"forged"}`)
	if w.Code != http.StatusAccepted || !strings.Contains(w.Body.String(), jobID) {
		t.Fatalf("create = %d %s", w.Code, w.Body.String())
	}
	if gotToken != "secret-token" || gotUser != "admin-uuid" || gotPath != "/jobs" {
		t.Errorf("token %q user %q path %q", gotToken, gotUser, gotPath)
	}
	if gotBody["created_by"] != "admin-uuid" || gotBody["client_name"] != `PT "B" $(id)` {
		t.Errorf("forwarded body = %v", gotBody)
	}

	if w := serve(r, "GET", "/api/v1/provisioning/jobs?limit=5&evil=1", ""); w.Code != http.StatusOK || gotPath != "/jobs?limit=5" {
		t.Errorf("list = %d, upstream path %q", w.Code, gotPath)
	}
	if w := serve(r, "GET", "/api/v1/provisioning/jobs/..%2f..%2fetc", ""); w.Code != http.StatusNotFound {
		t.Errorf("bad id = %d", w.Code)
	}
	if w := serve(r, "POST", "/api/v1/provisioning/jobs", `[1,2]`); w.Code != http.StatusBadRequest {
		t.Errorf("non-object body = %d", w.Code)
	}
}

func TestUpstream401BecomesBadGateway(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer up.Close()
	r := engine(NewController(config.ProvisionerConfig{URL: up.URL, Token: "wrong"}))
	if w := serve(r, "GET", "/api/v1/provisioning/jobs", ""); w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestSSEIsStreamedUnbuffered(t *testing.T) {
	release := make(chan struct{})
	var gotSince string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSince = r.URL.Query().Get("since")
		w.Header().Set("Content-Type", "text/event-stream")
		fl := w.(http.Flusher)
		fmt.Fprint(w, "id: 5\nevent: progress\ndata: {\"key\":\"build\"}\n\n")
		fl.Flush()
		<-release // the second event only comes after the client saw the first
		fmt.Fprint(w, "id: 6\nevent: status\ndata: {\"status\":\"succeeded\"}\n\n")
		fl.Flush()
	}))
	defer up.Close()
	defer close(release)

	front := httptest.NewServer(engine(NewController(config.ProvisionerConfig{URL: up.URL, Token: "t"})))
	defer front.Close()
	resp, err := http.Get(front.URL + "/api/v1/provisioning/jobs/" + jobID + "/events?since=4")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.Header.Get("Content-Type") != "text/event-stream" || resp.Header.Get("X-Accel-Buffering") != "no" {
		t.Errorf("headers = %v", resp.Header)
	}
	lines := make(chan string, 16)
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	select {
	case l := <-lines:
		if l != "id: 5" {
			t.Fatalf("first line = %q", l)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("first event was buffered: not received before the upstream finished")
	}
	if gotSince != "4" {
		t.Errorf("since forwarded as %q", gotSince)
	}
	release <- struct{}{}
	var rest []string
	for l := range lines {
		rest = append(rest, l)
	}
	if !strings.Contains(strings.Join(rest, "\n"), "succeeded") {
		t.Errorf("rest = %q", rest)
	}

	if w := serve(engine(NewController(config.ProvisionerConfig{URL: up.URL, Token: "t"})), "GET", "/api/v1/provisioning/jobs/"+jobID+"/events?since=abc", ""); w.Code != http.StatusBadRequest {
		t.Errorf("since=abc = %d", w.Code)
	}
}
