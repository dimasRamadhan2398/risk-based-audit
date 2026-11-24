package provisioner

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// Header names of the internal API.
const (
	TokenHeader = "X-Provisioner-Token"
	// UserHeader carries the authenticated auth-service user on reads, so the
	// provisioner can tell whether the caller is the job's creator.
	UserHeader = "X-Provisioner-User"
)

// Server is the provisioner's internal HTTP API.
type Server struct {
	Manager   *Manager
	Ports     PortsFunc
	Token     string
	Logger    *slog.Logger
	Heartbeat time.Duration // SSE heartbeat interval (default 15s)
}

// Handler returns the routed handler.
func (s *Server) Handler() http.Handler {
	if s.Logger == nil {
		s.Logger = slog.Default()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.Handle("POST /jobs", s.auth(s.createJob))
	mux.Handle("GET /jobs", s.auth(s.listJobs))
	mux.Handle("GET /jobs/{id}", s.auth(s.getJob))
	mux.Handle("GET /jobs/{id}/events", s.auth(s.streamEvents))
	mux.Handle("GET /ports/next", s.auth(s.nextPorts))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no such endpoint", nil)
	})
	return mux
}

func (s *Server) auth(h http.HandlerFunc) http.Handler {
	want := sha256.Sum256([]byte(s.Token))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hash both sides so the comparison is constant-time regardless of length.
		got := sha256.Sum256([]byte(r.Header.Get(TokenHeader)))
		if s.Token == "" || subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid "+TokenHeader, nil)
			return
		}
		h(w, r)
	})
}

// Error bodies use auth-service's envelope so the proxy can pass them through.
type errorBody struct {
	Success bool `json:"success"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Data any `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string, data any) {
	var b errorBody
	b.Error.Code, b.Error.Message, b.Data = code, msg, data
	writeJSON(w, status, b)
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024))
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", nil)
		return
	}
	job, err := s.Manager.Create(req)
	var verr *ValidationError
	switch {
	case errors.As(err, &verr):
		status, code := http.StatusBadRequest, "VALIDATION_FAILED"
		if verr.Conflict {
			status, code = http.StatusConflict, "CONFLICT"
		}
		writeError(w, status, code, "the request is invalid", map[string]any{"fields": verr.Fields})
	case err != nil:
		s.Logger.Error("create job", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create the job", nil)
	default:
		w.Header().Set("Location", "/jobs/"+job.ID)
		writeJSON(w, http.StatusAccepted, job)
	}
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 500 {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "limit must be 1-500", nil)
			return
		}
		limit = n
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": s.Manager.List(limit)})
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	job, err := s.Manager.Get(r.PathValue("id"), r.Header.Get(UserHeader))
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "job not found", nil)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) nextPorts(w http.ResponseWriter, r *http.Request) {
	np, err := s.Ports(r.Context())
	if err != nil {
		s.Logger.Error("next ports", "error", err)
		writeError(w, http.StatusBadGateway, "NEXT_PORTS_FAILED", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, np)
}

// streamEvents replays the events after ?since= (or Last-Event-ID), then
// streams new ones, flushing each, until the final status event.
func (s *Server) streamEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var since int64
	sinceStr := r.URL.Query().Get("since")
	if sinceStr == "" {
		sinceStr = r.Header.Get("Last-Event-ID")
	}
	if sinceStr != "" {
		n, err := strconv.ParseInt(sinceStr, 10, 64)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "since must be a non-negative integer", nil)
			return
		}
		since = n
	}
	evs, done, changed, err := s.Manager.EventsAfter(id, since)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "job not found", nil)
		return
	}
	if err != nil {
		s.Logger.Error("read events", "job", id, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not read job events", nil)
		return
	}
	rc := http.NewResponseController(w)
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	_ = rc.Flush()

	hb := s.Heartbeat
	if hb <= 0 {
		hb = 15 * time.Second
	}
	ticker := time.NewTicker(hb)
	defer ticker.Stop()
	for {
		for _, ev := range evs {
			if _, err := fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", ev.Seq, ev.Type, ev.Data); err != nil {
				return
			}
			if err := rc.Flush(); err != nil {
				return
			}
			since = ev.Seq
		}
		if done {
			return
		}
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
			if err := rc.Flush(); err != nil {
				return
			}
			evs = nil
			continue
		case <-changed:
		}
		evs, done, changed, err = s.Manager.EventsAfter(id, since)
		if err != nil {
			return
		}
	}
}
