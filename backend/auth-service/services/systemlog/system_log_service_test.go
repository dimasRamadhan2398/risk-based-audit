package systemlog

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"auth-service/models"
	"auth-service/pkg/activitylog"
	apperrors "auth-service/pkg/errors"
	"auth-service/repositories"

	"github.com/google/uuid"
)

type fakeRepo struct {
	repositories.ISysAlertLogRepository
	rows  []models.SysAlertLog
	lastF repositories.SysAlertLogFilter
	calls int
}

func (f *fakeRepo) List(_ context.Context, flt repositories.SysAlertLogFilter) ([]models.SysAlertLog, error) {
	f.calls++
	f.lastF = flt
	out := f.rows
	if flt.Limit > 0 && len(out) > flt.Limit {
		out = out[:flt.Limit]
	}
	cp := make([]models.SysAlertLog, len(out))
	copy(cp, out)
	return cp, nil
}

type fakePerms map[string][]string

func (f fakePerms) NamesForRoles(roles []string) ([]string, error) {
	var out []string
	for _, r := range roles {
		out = append(out, f[r]...)
	}
	return out, nil
}

type captureRecorder struct {
	mu      sync.Mutex
	entries []activitylog.Entry
}

func (c *captureRecorder) Record(_ context.Context, e activitylog.Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, e)
}

var perms = fakePerms{
	"ADMIN":     {PermissionView, PermissionExport},
	"EXECUTIVE": {PermissionView},
	"VIEWER":    {"view_users"},
}

func row(i int, ip string) models.SysAlertLog {
	at := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Add(-time.Duration(i) * time.Minute)
	return models.SysAlertLog{
		ID: uuid.New(), OccurredAt: at, CreatedAt: at, Action: "user.login", Result: "success",
		IPAddress: ip, ActorUsername: "alice",
		Details: models.JSONMap{"password": "leak", "reason": "ok"},
	}
}

func TestResolveAccess(t *testing.T) {
	s := NewService(&fakeRepo{}, perms, nil)
	cases := []struct {
		roles        []string
		view, fullIP bool
	}{
		{[]string{"ADMIN"}, true, true},
		{[]string{"EXECUTIVE"}, true, false},
		{[]string{"VIEWER"}, false, false},
		{[]string{"VIEWER", "EXECUTIVE"}, true, false},
		{nil, false, false},
	}
	for _, tc := range cases {
		a, err := s.ResolveAccess(tc.roles)
		if err != nil {
			t.Fatal(err)
		}
		if a.CanView != tc.view || a.FullIP != tc.fullIP {
			t.Errorf("%v: %+v", tc.roles, a)
		}
	}
}

func TestListRequiresViewPermission(t *testing.T) {
	repo := &fakeRepo{}
	_, err := NewService(repo, perms, nil).List(context.Background(), Access{}, ListQuery{})
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.StatusCode != http.StatusForbidden || repo.calls != 0 {
		t.Fatalf("err = %v, repo calls %d", err, repo.calls)
	}
}

func TestListMasksIPWithoutExport(t *testing.T) {
	repo := &fakeRepo{rows: []models.SysAlertLog{row(0, "192.168.10.25")}}
	s := NewService(repo, perms, nil)

	res, err := s.List(context.Background(), Access{CanView: true}, ListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Data[0].IPAddress != "192.168.x.x" {
		t.Fatalf("masked ip = %q", res.Data[0].IPAddress)
	}
	res, _ = s.List(context.Background(), Access{CanView: true, FullIP: true}, ListQuery{})
	if res.Data[0].IPAddress != "192.168.10.25" {
		t.Fatalf("full ip = %q", res.Data[0].IPAddress)
	}
	if res.Data[0].Details["password"] != activitylog.Redacted || res.Data[0].Details["reason"] != "ok" {
		t.Fatalf("details not redacted on read: %v", res.Data[0].Details)
	}
}

func TestListPagingCursors(t *testing.T) {
	var rows []models.SysAlertLog
	for i := 0; i < 5; i++ {
		rows = append(rows, row(i, ""))
	}
	repo := &fakeRepo{rows: rows}
	s := NewService(repo, perms, nil)

	res, err := s.List(context.Background(), Access{CanView: true}, ListQuery{Limit: "2"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.lastF.Limit != 3 {
		t.Fatalf("repo asked for %d rows, want limit+1", repo.lastF.Limit)
	}
	if len(res.Data) != 2 || res.NextCursor == nil || res.LatestCursor == nil {
		t.Fatalf("page = %d rows next=%v latest=%v", len(res.Data), res.NextCursor, res.LatestCursor)
	}
	c, err := DecodeCursor(*res.NextCursor, cursorKindPage)
	if err != nil || c.ID != rows[1].ID || !c.Time.Equal(rows[1].OccurredAt) {
		t.Fatalf("next cursor = %+v %v", c, err)
	}
	lc, err := DecodeCursor(*res.LatestCursor, cursorKindPoll)
	if err != nil || lc.ID != rows[0].ID {
		t.Fatalf("latest cursor = %+v %v", lc, err)
	}

	// Following next_cursor passes Before to the repository.
	if _, err := s.List(context.Background(), Access{CanView: true}, ListQuery{Cursor: *res.NextCursor}); err != nil {
		t.Fatal(err)
	}
	if repo.lastF.Before == nil || repo.lastF.Before.ID != rows[1].ID || repo.lastF.After != nil {
		t.Fatalf("filter = %+v", repo.lastF)
	}

	// Last page: no next_cursor.
	repo.rows = rows[:1]
	res, _ = s.List(context.Background(), Access{CanView: true}, ListQuery{Limit: "2"})
	if res.NextCursor != nil {
		t.Fatal("next_cursor on the last page")
	}

	// Empty first page still gives a poll cursor.
	repo.rows = nil
	res, _ = s.List(context.Background(), Access{CanView: true}, ListQuery{})
	if res.LatestCursor == nil || res.NextCursor != nil || res.Data == nil {
		t.Fatalf("empty page = %+v", res)
	}
}

func TestListPoll(t *testing.T) {
	older, newer := row(2, ""), row(1, "")
	repo := &fakeRepo{rows: []models.SysAlertLog{older, newer}} // stored order (asc)
	s := NewService(repo, perms, nil)
	since := EncodeCursor(cursorKindPoll, time.Now().Add(-time.Hour), uuid.Nil)

	res, err := s.List(context.Background(), Access{CanView: true}, ListQuery{SinceCursor: since})
	if err != nil {
		t.Fatal(err)
	}
	if repo.lastF.After == nil || repo.lastF.Before != nil {
		t.Fatalf("filter = %+v", repo.lastF)
	}
	if len(res.Data) != 2 || res.Data[0].ID != newer.ID.String() {
		t.Fatalf("poll not newest first: %+v", res.Data)
	}
	lc, _ := DecodeCursor(*res.LatestCursor, cursorKindPoll)
	if lc.ID != newer.ID || res.NextCursor != nil {
		t.Fatalf("latest = %+v next = %v", lc, res.NextCursor)
	}

	repo.rows = nil
	res, _ = s.List(context.Background(), Access{CanView: true}, ListQuery{SinceCursor: since})
	if res.LatestCursor == nil || *res.LatestCursor != since {
		t.Fatal("empty poll must echo since_cursor")
	}
}

func TestViewedEventOnlyOnFirstPageLoad(t *testing.T) {
	rec := &captureRecorder{}
	repo := &fakeRepo{rows: []models.SysAlertLog{row(0, ""), row(1, ""), row(2, "")}}
	s := NewService(repo, perms, rec)
	ctx := activitylog.WithActor(context.Background(), activitylog.Actor{UserID: uuid.NewString(), Username: "admin"})

	res, _ := s.List(ctx, Access{CanView: true}, ListQuery{Limit: "1", Action: "user.login"})
	_, _ = s.List(ctx, Access{CanView: true}, ListQuery{SinceCursor: *res.LatestCursor})
	_, _ = s.List(ctx, Access{CanView: true}, ListQuery{Cursor: *res.NextCursor})

	if len(rec.entries) != 1 {
		t.Fatalf("%d viewed events, want 1", len(rec.entries))
	}
	e := rec.entries[0]
	if e.Action != activitylog.ActionSystemLogsViewed || e.Summary != "admin viewed the activity log" {
		t.Fatalf("viewed = %+v", e)
	}
	if f := e.Details["filters"].(map[string]interface{}); f["action"] != "user.login" {
		t.Fatalf("filters = %v", e.Details)
	}
}

func TestParseFilterWhitelist(t *testing.T) {
	good := EncodeCursor(cursorKindPage, time.Now(), uuid.New())
	poll := EncodeCursor(cursorKindPoll, time.Now(), uuid.New())
	bad := []ListQuery{
		{Limit: "abc"},
		{Cursor: "!!!"},
		{Cursor: poll},      // poll cursor used for paging
		{SinceCursor: good}, // page cursor used for polling
		{Cursor: good, SinceCursor: poll},
		{Action: "user.login; DROP TABLE users"},
		{Action: "USER LOGIN"},
		{ActorUserID: "1 OR 1=1"},
		{Result: "maybe"},
		{From: "yesterday"},
		{From: "2026-11-02T00:00:00Z", To: "2026-11-01T00:00:00Z"},
		{Q: string(make([]byte, 101))},
	}
	for _, q := range bad {
		_, err := ParseFilter(q)
		appErr, ok := err.(*apperrors.AppError)
		if !ok || appErr.StatusCode != http.StatusBadRequest {
			t.Errorf("%+v: err = %v, want 400", q, err)
		}
	}

	f, err := ParseFilter(ListQuery{})
	if err != nil || f.Limit != DefaultLimit {
		t.Fatalf("default limit = %d %v", f.Limit, err)
	}
	f, _ = ParseFilter(ListQuery{Limit: "1000"})
	if f.Limit != MaxLimit {
		t.Fatalf("limit not capped: %d", f.Limit)
	}
	f, _ = ParseFilter(ListQuery{Limit: "0"})
	if f.Limit != DefaultLimit {
		t.Fatalf("limit 0 = %d", f.Limit)
	}
	actor := uuid.New()
	f, err = ParseFilter(ListQuery{
		Action: "User.Login, user.login_failed", ActorUserID: actor.String(), Result: "FAILURE",
		From: "2026-11-01T00:00:00+07:00", To: "2026-11-02T00:00:00Z", Q: " bob ", Cursor: good,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Actions) != 2 || f.Actions[0] != "user.login" || *f.ActorUserID != actor || f.Result != "failure" ||
		f.From == nil || f.To == nil || f.Q != "bob" || f.Before == nil {
		t.Fatalf("parsed = %+v", f)
	}
}
