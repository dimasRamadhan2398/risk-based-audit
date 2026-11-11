// Package systemlog serves GET /api/v1/system-logs, the admin user activity log.
package systemlog

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"auth-service/models"
	"auth-service/pkg/activitylog"
	apperrors "auth-service/pkg/errors"
	"auth-service/repositories"

	"github.com/google/uuid"
)

// Paging limits.
const (
	DefaultLimit = 50
	MaxLimit     = 200
	maxQLength   = 100
	maxActions   = 20
)

// Permissions (seeded by pkg/database/seeders).
const (
	PermissionView   = "view_system_logs"
	PermissionExport = "export_system_logs"
)

// Cursor kinds. "o" pages by occurred_at (cursor / next_cursor), "c" polls by
// stored order (since_cursor / latest_cursor). They are not interchangeable.
const (
	cursorKindPage = "o"
	cursorKindPoll = "c"
)

var actionPattern = regexp.MustCompile(`^[a-z0-9_]+(\.[a-z0-9_]+)*$`)

// Access is what the caller may see.
type Access struct {
	CanView bool
	// FullIP: the caller holds export_system_logs; otherwise ip_address is masked.
	FullIP bool
}

// ListQuery is the raw query string of GET /system-logs.
type ListQuery struct {
	Limit       string
	Cursor      string
	SinceCursor string
	Action      string
	ActorUserID string
	Result      string
	From        string
	To          string
	Q           string
}

// LogItem is one row of the response.
type LogItem struct {
	ID            string                 `json:"id"`
	OccurredAt    time.Time              `json:"occurred_at"`
	ActorUserID   *string                `json:"actor_user_id"`
	ActorUsername string                 `json:"actor_username"`
	Action        string                 `json:"action"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	Result        string                 `json:"result"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	Summary       string                 `json:"summary"`
	Details       map[string]interface{} `json:"details"`
}

// ListResult is the body of a 200 response (besides "success").
type ListResult struct {
	Data         []LogItem `json:"data"`
	NextCursor   *string   `json:"next_cursor"`
	LatestCursor *string   `json:"latest_cursor"`
}

// PermissionLookup resolves role names to permission names.
type PermissionLookup interface {
	NamesForRoles(roleNames []string) ([]string, error)
}

// ServiceInterface is the system log service.
type ServiceInterface interface {
	ResolveAccess(roles []string) (Access, error)
	List(ctx context.Context, access Access, q ListQuery) (*ListResult, error)
}

// Service implements ServiceInterface.
type Service struct {
	repo     repositories.ISysAlertLogRepository
	perms    PermissionLookup
	recorder activitylog.Recorder
	now      func() time.Time
}

// NewService builds the service. recorder may be nil (no system_logs.viewed event).
func NewService(repo repositories.ISysAlertLogRepository, perms PermissionLookup, recorder activitylog.Recorder) *Service {
	if recorder == nil {
		recorder = activitylog.NopRecorder{}
	}
	return &Service{repo: repo, perms: perms, recorder: recorder, now: time.Now}
}

// ResolveAccess reads the caller's permissions from the roles in their token.
func (s *Service) ResolveAccess(roles []string) (Access, error) {
	names, err := s.perms.NamesForRoles(roles)
	if err != nil {
		return Access{}, apperrors.Wrap("PERMISSION_LOOKUP_FAILED", "Failed to resolve permissions", http.StatusInternalServerError, err)
	}
	var a Access
	for _, n := range names {
		switch strings.ToLower(n) {
		case PermissionView:
			a.CanView = true
		case PermissionExport:
			a.FullIP = true
		}
	}
	return a, nil
}

// List returns one page (or one poll) of the activity log. The first page load
// (no cursor, no since_cursor) records a system_logs.viewed event.
func (s *Service) List(ctx context.Context, access Access, q ListQuery) (*ListResult, error) {
	if !access.CanView {
		return nil, apperrors.Wrap("FORBIDDEN", "view_system_logs permission required", http.StatusForbidden, nil)
	}
	f, err := ParseFilter(q)
	if err != nil {
		return nil, err
	}

	var res *ListResult
	if f.After != nil {
		res, err = s.poll(ctx, f, q.SinceCursor)
	} else {
		res, err = s.page(ctx, f)
	}
	if err != nil {
		return nil, err
	}

	for i := range res.Data {
		if !access.FullIP {
			res.Data[i].IPAddress = activitylog.MaskIP(res.Data[i].IPAddress)
		}
	}

	if q.Cursor == "" && q.SinceCursor == "" {
		s.recorder.Record(ctx, activitylog.Entry{
			Action:     activitylog.ActionSystemLogsViewed,
			EntityType: "system_logs",
			Result:     activitylog.ResultSuccess,
			Summary:    viewedSummary(ctx),
			Details:    appliedFilters(q),
		})
	}
	return res, nil
}

func (s *Service) page(ctx context.Context, f repositories.SysAlertLogFilter) (*ListResult, error) {
	limit := f.Limit
	f.Limit = limit + 1
	rows, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, apperrors.Wrap("FETCH_SYSTEM_LOGS_FAILED", "Failed to fetch system logs", http.StatusInternalServerError, err)
	}
	res := &ListResult{Data: make([]LogItem, 0, limit)}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	for i := range rows {
		res.Data = append(res.Data, toItem(&rows[i]))
	}
	if hasMore {
		last := rows[len(rows)-1]
		c := EncodeCursor(cursorKindPage, last.OccurredAt, last.ID)
		res.NextCursor = &c
	}
	res.LatestCursor = s.latestOf(rows)
	return res, nil
}

func (s *Service) poll(ctx context.Context, f repositories.SysAlertLogFilter, since string) (*ListResult, error) {
	rows, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, apperrors.Wrap("FETCH_SYSTEM_LOGS_FAILED", "Failed to fetch system logs", http.StatusInternalServerError, err)
	}
	res := &ListResult{Data: make([]LogItem, 0, len(rows))}
	if len(rows) == 0 {
		res.LatestCursor = &since
		return res, nil
	}
	// Rows come in stored order; the newest stored one is the next poll point.
	last := rows[len(rows)-1]
	c := EncodeCursor(cursorKindPoll, last.CreatedAt, last.ID)
	res.LatestCursor = &c
	sort.SliceStable(rows, func(i, j int) bool { return newerFirst(&rows[i], &rows[j]) })
	for i := range rows {
		res.Data = append(res.Data, toItem(&rows[i]))
	}
	return res, nil
}

// latestOf is the poll cursor after the newest stored row of a page. For an
// empty page it is "just before now", so polling still works.
func (s *Service) latestOf(rows []models.SysAlertLog) *string {
	var best *models.SysAlertLog
	for i := range rows {
		r := &rows[i]
		if best == nil || r.CreatedAt.After(best.CreatedAt) ||
			(r.CreatedAt.Equal(best.CreatedAt) && r.ID.String() > best.ID.String()) {
			best = r
		}
	}
	var c string
	if best == nil {
		c = EncodeCursor(cursorKindPoll, s.now().Add(-time.Second), uuid.Nil)
	} else {
		c = EncodeCursor(cursorKindPoll, best.CreatedAt, best.ID)
	}
	return &c
}

func newerFirst(a, b *models.SysAlertLog) bool {
	if !a.OccurredAt.Equal(b.OccurredAt) {
		return a.OccurredAt.After(b.OccurredAt)
	}
	return a.ID.String() > b.ID.String()
}

func toItem(r *models.SysAlertLog) LogItem {
	it := LogItem{
		ID:            r.ID.String(),
		OccurredAt:    r.OccurredAt.UTC(),
		ActorUsername: r.ActorUsername,
		Action:        string(r.Action),
		EntityType:    r.EntityType,
		EntityID:      r.EntityID,
		Result:        r.Result,
		IPAddress:     r.IPAddress,
		UserAgent:     r.UserAgent,
		Summary:       r.Summary,
		// Redacted again on read: rows written before redaction, or by hand,
		// must not leak through the API either.
		Details: activitylog.Redact(map[string]interface{}(r.Details)),
	}
	if it.Details == nil {
		it.Details = map[string]interface{}{}
	}
	if r.ActorUserID != nil {
		id := r.ActorUserID.String()
		it.ActorUserID = &id
	}
	return it
}

// ParseFilter validates the query string and maps it onto the whitelisted
// repository filter. Every error is a 400.
func ParseFilter(q ListQuery) (repositories.SysAlertLogFilter, error) {
	f := repositories.SysAlertLogFilter{Limit: DefaultLimit}

	if v := strings.TrimSpace(q.Limit); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return f, badRequest("limit must be an integer")
		}
		switch {
		case n < 1:
			n = DefaultLimit
		case n > MaxLimit:
			n = MaxLimit
		}
		f.Limit = n
	}

	if q.Cursor != "" && q.SinceCursor != "" {
		return f, badRequest("cursor and since_cursor cannot be combined")
	}
	if q.Cursor != "" {
		c, err := DecodeCursor(q.Cursor, cursorKindPage)
		if err != nil {
			return f, badRequest("invalid cursor")
		}
		f.Before = c
	}
	if q.SinceCursor != "" {
		c, err := DecodeCursor(q.SinceCursor, cursorKindPoll)
		if err != nil {
			return f, badRequest("invalid since_cursor")
		}
		f.After = c
	}

	if v := strings.TrimSpace(q.Action); v != "" {
		for _, a := range strings.Split(v, ",") {
			a = strings.ToLower(strings.TrimSpace(a))
			if a == "" {
				continue
			}
			if len(a) > 100 || !actionPattern.MatchString(a) {
				return f, badRequest("invalid action")
			}
			f.Actions = append(f.Actions, a)
		}
		if len(f.Actions) > maxActions {
			return f, badRequest("too many actions")
		}
	}

	if v := strings.TrimSpace(q.ActorUserID); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return f, badRequest("actor_user_id must be a UUID")
		}
		f.ActorUserID = &id
	}

	switch v := strings.ToLower(strings.TrimSpace(q.Result)); v {
	case "":
	case activitylog.ResultSuccess, activitylog.ResultFailure:
		f.Result = v
	default:
		return f, badRequest("result must be success or failure")
	}

	if v := strings.TrimSpace(q.From); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, badRequest("from must be RFC3339")
		}
		f.From = &t
	}
	if v := strings.TrimSpace(q.To); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, badRequest("to must be RFC3339")
		}
		f.To = &t
	}
	if f.From != nil && f.To != nil && f.From.After(*f.To) {
		return f, badRequest("from must not be after to")
	}

	if v := strings.TrimSpace(q.Q); v != "" {
		if len([]rune(v)) > maxQLength {
			return f, badRequest("q is too long")
		}
		f.Q = v
	}
	return f, nil
}

type cursorPayload struct {
	K  string `json:"k"`
	T  string `json:"t"`
	ID string `json:"id"`
}

// EncodeCursor returns an opaque cursor.
func EncodeCursor(kind string, t time.Time, id uuid.UUID) string {
	b, _ := json.Marshal(cursorPayload{K: kind, T: t.UTC().Format(time.RFC3339Nano), ID: id.String()})
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor parses a cursor of the given kind.
func DecodeCursor(s, kind string) (*repositories.LogCursor, error) {
	if len(s) > 256 {
		return nil, apperrors.ErrBadRequest
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var p cursorPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	if p.K != kind {
		return nil, apperrors.ErrBadRequest
	}
	t, err := time.Parse(time.RFC3339Nano, p.T)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, err
	}
	return &repositories.LogCursor{Time: t, ID: id}, nil
}

func badRequest(msg string) error {
	return apperrors.Wrap("VALIDATION_ERROR", msg, http.StatusBadRequest, nil)
}

func viewedSummary(ctx context.Context) string {
	if a, ok := activitylog.ActorFrom(ctx); ok && a.Username != "" {
		return a.Username + " viewed the activity log"
	}
	return "Activity log viewed"
}

func appliedFilters(q ListQuery) map[string]interface{} {
	out := map[string]interface{}{}
	add := func(k, v string) {
		if v = strings.TrimSpace(v); v != "" {
			out[k] = v
		}
	}
	add("action", q.Action)
	add("actor_user_id", q.ActorUserID)
	add("result", q.Result)
	add("from", q.From)
	add("to", q.To)
	add("q", q.Q)
	if len(out) == 0 {
		return nil
	}
	return map[string]interface{}{"filters": out}
}
