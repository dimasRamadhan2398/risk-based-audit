// Package provisioning proxies /api/v1/provisioning/* to the host provisioner
// (backend/provisioner-service). See docs/provisioning-job-contract.md.
//
// The provisioner is reached at PROVISIONER_URL with X-Provisioner-Token. The
// caller is identified to it by created_by (POST /jobs) and the
// X-Provisioner-User header (reads), both taken from the JWT, never from the
// request.
package provisioning

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"auth-service/pkg/config"
	"auth-service/pkg/logger"
	"auth-service/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	tokenHeader = "X-Provisioner-Token"
	userHeader  = "X-Provisioner-User"
	maxBody     = 64 * 1024
)

var jobIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ControllerInterface is the provisioning proxy.
type ControllerInterface interface {
	CreateJob(c *gin.Context)
	ListJobs(c *gin.Context)
	GetJob(c *gin.Context)
	StreamEvents(c *gin.Context)
	NextPorts(c *gin.Context)
}

// Controller forwards requests to the provisioner.
type Controller struct {
	baseURL string
	token   string
	client  *http.Client // JSON calls, bounded
	stream  *http.Client // SSE, no overall timeout
}

// NewController builds the proxy. An empty URL leaves provisioning disabled
// (every route answers 503), which is the normal state on tenant stacks.
func NewController(cfg config.ProvisionerConfig) *Controller {
	dial := &net.Dialer{Timeout: 5 * time.Second}
	return &Controller{
		baseURL: strings.TrimRight(strings.TrimSpace(cfg.URL), "/"),
		token:   cfg.Token,
		client: &http.Client{
			Timeout:   45 * time.Second, // GET /ports/next runs a script
			Transport: &http.Transport{DialContext: dial.DialContext},
		},
		stream: &http.Client{
			Transport: &http.Transport{
				DialContext:           dial.DialContext,
				ResponseHeaderTimeout: 15 * time.Second,
				DisableCompression:    true,
			},
		},
	}
}

func callerID(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		if s, _ := v.(string); s != "" {
			return s
		}
	}
	if v, ok := c.Get("username"); ok {
		if s, _ := v.(string); s != "" {
			return s
		}
	}
	return ""
}

func (ctl *Controller) unavailable(c *gin.Context, err error) {
	if err != nil {
		logger.Warn("provisioner unreachable", logger.LogField("error", err.Error()))
	}
	response.Error(c, http.StatusServiceUnavailable, "PROVISIONER_UNAVAILABLE",
		"The provisioning service is not reachable. Check that the provisioner is running on the host.", "")
}

// newRequest builds an upstream request. path must already be safe.
func (ctl *Controller) newRequest(c *gin.Context, method, path string, query url.Values, body []byte) (*http.Request, error) {
	u := ctl.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), method, u, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set(tokenHeader, ctl.token)
	req.Header.Set(userHeader, callerID(c))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// forward performs a JSON call and relays status and body.
func (ctl *Controller) forward(c *gin.Context, method, path string, query url.Values, body []byte) {
	if ctl.baseURL == "" {
		ctl.unavailable(c, nil)
		return
	}
	req, err := ctl.newRequest(c, method, path, query, body)
	if err != nil {
		response.InternalServerError(c, "Could not build the provisioner request")
		return
	}
	resp, err := ctl.client.Do(req)
	if err != nil {
		ctl.unavailable(c, err)
		return
	}
	defer resp.Body.Close()
	if ctl.tokenRejected(c, resp) {
		return
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		ctl.unavailable(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Data(resp.StatusCode, "application/json", data)
}

// tokenRejected turns an upstream 401 into a 502: it means PROVISIONER_TOKEN
// differs between auth-service and the provisioner, and relaying a 401 would
// make the frontend sign the user out.
func (ctl *Controller) tokenRejected(c *gin.Context, resp *http.Response) bool {
	if resp.StatusCode != http.StatusUnauthorized {
		return false
	}
	logger.Error("provisioner rejected PROVISIONER_TOKEN; check that auth-service and the provisioner share the same token")
	response.Error(c, http.StatusBadGateway, "PROVISIONER_AUTH_FAILED",
		"auth-service is not authorised by the provisioning service (token mismatch).", "")
	return true
}

func (ctl *Controller) jobID(c *gin.Context) (string, bool) {
	id := c.Param("id")
	if !jobIDPattern.MatchString(id) {
		response.NotFound(c, "Job not found")
		return "", false
	}
	return id, true
}

// CreateJob godoc: POST /api/v1/provisioning/jobs. created_by is always set
// from the JWT; a value in the body is overwritten.
func (ctl *Controller) CreateJob(c *gin.Context) {
	if ctl.baseURL == "" {
		ctl.unavailable(c, nil)
		return
	}
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBody+1))
	if err != nil || len(raw) > maxBody {
		response.BadRequest(c, "Request body is missing or too large")
		return
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil || body == nil {
		response.BadRequest(c, "Request body must be a JSON object")
		return
	}
	creator := callerID(c)
	if creator == "" {
		response.Unauthorized(c, "Missing user identity")
		return
	}
	cb, _ := json.Marshal(creator)
	body["created_by"] = cb
	out, _ := json.Marshal(body)
	logger.Info("provisioning job requested", logger.LogField("created_by", creator))
	ctl.forward(c, http.MethodPost, "/jobs", nil, out)
}

// ListJobs godoc: GET /api/v1/provisioning/jobs?limit=
func (ctl *Controller) ListJobs(c *gin.Context) {
	q := url.Values{}
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 500 {
			response.BadRequest(c, "limit must be between 1 and 500")
			return
		}
		q.Set("limit", strconv.Itoa(n))
	}
	ctl.forward(c, http.MethodGet, "/jobs", q, nil)
}

// GetJob godoc: GET /api/v1/provisioning/jobs/:id. The initial admin
// password is included once, and only when the caller created the job.
func (ctl *Controller) GetJob(c *gin.Context) {
	id, ok := ctl.jobID(c)
	if !ok {
		return
	}
	ctl.forward(c, http.MethodGet, "/jobs/"+id, nil, nil)
}

// NextPorts godoc: GET /api/v1/provisioning/ports/next
func (ctl *Controller) NextPorts(c *gin.Context) {
	ctl.forward(c, http.MethodGet, "/ports/next", nil, nil)
}

// StreamEvents godoc: GET /api/v1/provisioning/jobs/:id/events?since=
// Relays the provisioner's SSE stream, flushing after every read so nothing
// is buffered in auth-service.
func (ctl *Controller) StreamEvents(c *gin.Context) {
	id, ok := ctl.jobID(c)
	if !ok {
		return
	}
	if ctl.baseURL == "" {
		ctl.unavailable(c, nil)
		return
	}
	q := url.Values{}
	since := c.Query("since")
	if since == "" {
		since = c.GetHeader("Last-Event-ID")
	}
	if since != "" {
		n, err := strconv.ParseInt(since, 10, 64)
		if err != nil || n < 0 {
			response.BadRequest(c, "since must be a non-negative integer")
			return
		}
		q.Set("since", strconv.FormatInt(n, 10))
	}
	req, err := ctl.newRequest(c, http.MethodGet, "/jobs/"+id+"/events", q, nil)
	if err != nil {
		response.InternalServerError(c, "Could not build the provisioner request")
		return
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := ctl.stream.Do(req)
	if err != nil {
		if c.Request.Context().Err() == nil {
			ctl.unavailable(c, err)
		}
		return
	}
	defer resp.Body.Close()
	if ctl.tokenRejected(c, resp) {
		return
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		c.Data(resp.StatusCode, "application/json", data)
		return
	}

	w := c.Writer
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	w.Flush()

	buf := make([]byte, 8*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return
			}
			w.Flush()
		}
		if rerr != nil {
			return
		}
	}
}
