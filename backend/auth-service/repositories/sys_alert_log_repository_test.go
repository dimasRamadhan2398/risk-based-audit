package repositories

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"auth-service/models"
	"auth-service/pkg/activitylog"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// sysLogDB is an in-memory sqlite DB with sys_alert_logs created by hand (the
// model's gen_random_uuid()/now() defaults are Postgres-only).
func sysLogDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Exec(`CREATE TABLE sys_alert_logs (
		id TEXT PRIMARY KEY, user_id TEXT, event_id TEXT, occurred_at DATETIME NOT NULL,
		actor_user_id TEXT, actor_username TEXT, action TEXT NOT NULL, entity_type TEXT, entity_id TEXT,
		result TEXT NOT NULL DEFAULT 'success', ip_address TEXT, user_agent TEXT, summary TEXT, details TEXT,
		created_at DATETIME, updated_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX idx_sys_alert_logs_event_id ON sys_alert_logs(event_id)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func pgDryRun(t *testing.T) *gorm.DB {
	t.Helper()
	pg, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=none"}),
		&gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	return pg
}

func entry(i int, at time.Time, action, actor string) *activitylog.Entry {
	e := &activitylog.Entry{
		EventID:       fmt.Sprintf("evt-%03d", i),
		OccurredAt:    at,
		ActorUserID:   actor,
		ActorUsername: fmt.Sprintf("user%d", i%3),
		Action:        action,
		Result:        activitylog.ResultSuccess,
		Summary:       fmt.Sprintf("summary %d", i),
	}
	if i%4 == 0 {
		e.Result = activitylog.ResultFailure
	}
	e.Normalize(at)
	return e
}

func TestInsertEntryIsIdempotentOnEventID(t *testing.T) {
	repo := NewSysAlertLogRepository(sysLogDB(t))
	e := entry(1, time.Now(), activitylog.ActionLogin, uuid.NewString())

	ins, err := repo.InsertEntry(context.Background(), e)
	if err != nil || !ins {
		t.Fatalf("first insert: inserted=%v err=%v", ins, err)
	}
	ins, err = repo.InsertEntry(context.Background(), e)
	if err != nil || ins {
		t.Fatalf("duplicate insert: inserted=%v err=%v, want false,nil", ins, err)
	}
	rows, _ := repo.List(context.Background(), SysAlertLogFilter{Limit: 10})
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
}

// Paging newest-first with a keyset cursor must visit every row exactly once,
// also when many rows share the same occurred_at.
func TestListCursorPagingHasNoDuplicatesOrGaps(t *testing.T) {
	repo := NewSysAlertLogRepository(sysLogDB(t))
	base := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	const n = 23
	for i := 0; i < n; i++ {
		// groups of three rows share one timestamp
		at := base.Add(time.Duration(i/3) * time.Second)
		if _, err := repo.InsertEntry(context.Background(), entry(i, at, activitylog.ActionLogin, "")); err != nil {
			t.Fatal(err)
		}
	}

	seen := map[uuid.UUID]bool{}
	var before *LogCursor
	var prev *models.SysAlertLog
	for pages := 0; pages < 20; pages++ {
		rows, err := repo.List(context.Background(), SysAlertLogFilter{Limit: 4, Before: before})
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) == 0 {
			break
		}
		for i := range rows {
			r := rows[i]
			if seen[r.ID] {
				t.Fatalf("row %s returned twice", r.ID)
			}
			seen[r.ID] = true
			if prev != nil && (r.OccurredAt.After(prev.OccurredAt) ||
				(r.OccurredAt.Equal(prev.OccurredAt) && r.ID.String() > prev.ID.String())) {
				t.Fatalf("not newest first: %v/%s after %v/%s", r.OccurredAt, r.ID, prev.OccurredAt, prev.ID)
			}
			prev = &rows[i]
		}
		last := rows[len(rows)-1]
		before = &LogCursor{Time: last.OccurredAt, ID: last.ID}
	}
	if len(seen) != n {
		t.Fatalf("visited %d rows, want %d", len(seen), n)
	}
}

func TestListFilters(t *testing.T) {
	repo := NewSysAlertLogRepository(sysLogDB(t))
	base := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	actor := uuid.New()
	for i := 0; i < 12; i++ {
		action := activitylog.ActionLogin
		if i%2 == 1 {
			action = activitylog.ActionLoginFailed
		}
		a := ""
		if i < 4 {
			a = actor.String()
		}
		if _, err := repo.InsertEntry(context.Background(), entry(i, base.Add(time.Duration(i)*time.Minute), action, a)); err != nil {
			t.Fatal(err)
		}
	}
	from, to := base.Add(2*time.Minute), base.Add(5*time.Minute)
	cases := []struct {
		name string
		f    SysAlertLogFilter
		want int
	}{
		{"action", SysAlertLogFilter{Actions: []string{activitylog.ActionLoginFailed}}, 6},
		{"actions", SysAlertLogFilter{Actions: []string{activitylog.ActionLoginFailed, activitylog.ActionLogin}}, 12},
		{"actor", SysAlertLogFilter{ActorUserID: &actor}, 4},
		{"result", SysAlertLogFilter{Result: activitylog.ResultFailure}, 3}, // i = 0,4,8
		{"from-to", SysAlertLogFilter{From: &from, To: &to}, 4},
		{"q username", SysAlertLogFilter{Q: "USER1"}, 4}, // i%3 == 1
		{"q summary", SysAlertLogFilter{Q: "summary 1"}, 3},   // 1, 10, 11
		{"q wildcard is literal", SysAlertLogFilter{Q: "%"}, 0},
		{"q underscore is literal", SysAlertLogFilter{Q: "_"}, 0},
	}
	for _, tc := range cases {
		rows, err := repo.List(context.Background(), tc.f)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if len(rows) != tc.want {
			t.Errorf("%s: %d rows, want %d", tc.name, len(rows), tc.want)
		}
	}
}

func TestListSincePollsInStoredOrder(t *testing.T) {
	db := sysLogDB(t)
	repo := NewSysAlertLogRepository(db)
	base := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		if _, err := repo.InsertEntry(context.Background(), entry(i, base.Add(time.Duration(i)*time.Minute), activitylog.ActionLogin, "")); err != nil {
			t.Fatal(err)
		}
	}
	all, _ := repo.List(context.Background(), SysAlertLogFilter{})
	newest := all[0]
	for _, r := range all {
		if r.CreatedAt.After(newest.CreatedAt) || (r.CreatedAt.Equal(newest.CreatedAt) && r.ID.String() > newest.ID.String()) {
			newest = r
		}
	}
	time.Sleep(2 * time.Millisecond)
	// A late delivery: occurred long ago, stored now. Polling must still see it.
	late := entry(99, base.Add(-time.Hour), activitylog.ActionLogout, "")
	if _, err := repo.InsertEntry(context.Background(), late); err != nil {
		t.Fatal(err)
	}
	rows, err := repo.List(context.Background(), SysAlertLogFilter{After: &LogCursor{Time: newest.CreatedAt, ID: newest.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].EventID != late.EventID {
		t.Fatalf("poll returned %d rows (%v), want only the late row", len(rows), rows)
	}
}

func TestDeleteOlderThan(t *testing.T) {
	repo := NewSysAlertLogRepository(sysLogDB(t))
	now := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	for i, age := range []int{1, 89, 91, 200} {
		if _, err := repo.InsertEntry(context.Background(), entry(i, now.AddDate(0, 0, -age), activitylog.ActionLogin, "")); err != nil {
			t.Fatal(err)
		}
	}
	n, err := repo.DeleteOlderThan(context.Background(), RetentionCutoff(now, 90))
	if err != nil || n != 2 {
		t.Fatalf("deleted %d err %v, want 2", n, err)
	}
	rows, _ := repo.List(context.Background(), SysAlertLogFilter{})
	if len(rows) != 2 {
		t.Fatalf("%d rows left, want 2", len(rows))
	}
}

// The retention delete binds the cutoff as a parameter. The old query put the
// placeholder inside INTERVAL '? days', a string literal, so it never bound.
func TestRetentionDeleteSQLBindsCutoff(t *testing.T) {
	pg := pgDryRun(t)
	cutoff := time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)
	q := deleteOlderThan(pg.Session(&gorm.Session{DryRun: true}), cutoff)
	sql := q.Statement.SQL.String()
	if sql != `DELETE FROM "sys_alert_logs" WHERE "occurred_at" < $1` {
		t.Fatalf("delete SQL = %s", sql)
	}
	if len(q.Statement.Vars) != 1 || !q.Statement.Vars[0].(time.Time).Equal(cutoff) {
		t.Fatalf("vars = %v", q.Statement.Vars)
	}
	if strings.Contains(sql, "INTERVAL") {
		t.Fatal("interval literal is back")
	}
}

func TestRetentionCutoff(t *testing.T) {
	now := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	if got := RetentionCutoff(now, 90); !got.Equal(time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("cutoff = %v", got)
	}
	repo := NewSysAlertLogRepository(sysLogDB(t))
	if err := repo.DeleteOldLogs(0); err == nil {
		t.Fatal("DeleteOldLogs(0) must be rejected")
	}
	if err := repo.DeleteOldLogs(90); err != nil {
		t.Fatalf("DeleteOldLogs(90): %v", err)
	}
}

// Every request-derived value is a bind parameter; column names are fixed.
func TestListSQLShapeBindsAllInput(t *testing.T) {
	pg := pgDryRun(t)
	r := &SysAlertLogRepository{BaseRepository: NewBaseRepository(pg)}
	actor := uuid.New()
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	evil := `x'; DROP TABLE users; --`
	f := SysAlertLogFilter{
		Limit:       51,
		Before:      &LogCursor{Time: from, ID: actor},
		Actions:     []string{"user.login", "user.logout"},
		ActorUserID: &actor,
		Result:      "failure",
		From:        &from,
		Q:           evil,
	}
	var rows []models.SysAlertLog
	q := r.listQuery(pg.Session(&gorm.Session{DryRun: true}), f).Find(&rows)
	sql := q.Statement.SQL.String()
	if strings.Contains(sql, "DROP TABLE") {
		t.Fatalf("search term interpolated: %s", sql)
	}
	for _, want := range []string{
		`"action" IN ($1,$2)`, `"actor_user_id" = $3`, `"result" = $4`, `"occurred_at" >= $5`,
		`LOWER("actor_username") LIKE $6 ESCAPE '\'`, `LOWER("summary") LIKE $7 ESCAPE '\'`,
		`("occurred_at" < $8 OR ("occurred_at" = $9 AND "id" < $10))`,
		`ORDER BY "occurred_at" DESC,"id" DESC LIMIT $11`,
	} {
		if !strings.Contains(sql, want) {
			t.Errorf("SQL missing %q:\n%s", want, sql)
		}
	}
}

// The AutoMigrate indexes: cursor (occurred_at DESC, id DESC), actor
// (actor_user_id, occurred_at), poll (created_at, id), unique event_id.
func TestSysAlertLogIndexes(t *testing.T) {
	s, err := schema.Parse(&models.SysAlertLog{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	idx := map[string]*schema.Index{}
	for _, ix := range s.ParseIndexes() {
		idx[ix.Name] = ix
	}
	check := func(name, class string, cols ...string) {
		t.Helper()
		ix := idx[name]
		if ix == nil {
			t.Fatalf("index %s missing (have %v)", name, idx)
		}
		if ix.Class != class {
			t.Errorf("%s class %q, want %q", name, ix.Class, class)
		}
		var got []string
		for _, f := range ix.Fields {
			c := f.DBName
			if f.Sort != "" {
				c += " " + strings.ToLower(f.Sort)
			}
			got = append(got, c)
		}
		if strings.Join(got, ",") != strings.Join(cols, ",") {
			t.Errorf("%s columns %v, want %v", name, got, cols)
		}
	}
	check("idx_sys_alert_logs_occurred", "", "occurred_at desc", "id desc")
	check("idx_sys_alert_logs_actor_occurred", "", "actor_user_id", "occurred_at")
	check("idx_sys_alert_logs_created", "", "created_at", "id")
	check("idx_sys_alert_logs_event_id", "UNIQUE", "event_id")

	// user_id must be nullable now (failed logins have no user)
	if f := s.LookUpField("user_id"); f == nil || f.NotNull {
		t.Fatal("user_id must be nullable")
	}
}
