package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"auth-service/models"
	"auth-service/pkg/logger"
	"auth-service/pkg/validations"
	"auth-service/repositories"
	userService "auth-service/services/user"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// seatEnv is a real UserRepository + UserService + UserController on an
// in-memory SQLite database. SQLite has no advisory locks, so WithSeatLock
// is a plain transaction there; with one open connection transactions run
// one after another.
type seatEnv struct {
	t     *testing.T
	db    *gorm.DB
	svc   userService.UserServiceInterface
	r     *gin.Engine
	roles map[string]string // name -> id
	n     int
}

func newSeatEnv(t *testing.T) *seatEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger.Log = zap.NewNop()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	// The models default ids to gen_random_uuid(), which SQLite cannot run, so
	// the tables are created by hand and new users get their id here.
	for _, ddl := range []string{
		`CREATE TABLE users (id TEXT PRIMARY KEY, employee_id TEXT UNIQUE, username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL, password_hash TEXT, full_name TEXT, phone TEXT, department TEXT,
			position TEXT, is_active BOOLEAN DEFAULT 1, avatar_url TEXT, must_change_password BOOLEAN DEFAULT 0,
			locked_until DATETIME, last_login_fingerprint TEXT, last_login_at DATETIME,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE roles (id TEXT PRIMARY KEY, name TEXT UNIQUE, description TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE user_roles (user_id TEXT, role_id TEXT, PRIMARY KEY (user_id, role_id))`,
		`CREATE TABLE mfa_setups (id TEXT PRIMARY KEY, user_id TEXT, mfa_type TEXT, secret_key TEXT, is_enabled BOOLEAN, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE trusted_devices (id TEXT PRIMARY KEY, user_id TEXT, device_fingerprint TEXT, created_at DATETIME)`,
		`CREATE TABLE refresh_tokens (id TEXT PRIMARY KEY, user_id TEXT, token TEXT, expires_at DATETIME, created_at DATETIME)`,
		`CREATE TABLE confidentiality_agreements (id TEXT PRIMARY KEY, user_id TEXT, title TEXT, deleted_at DATETIME)`,
		`CREATE TABLE sys_alert_logs (id TEXT PRIMARY KEY, user_id TEXT)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Callback().Create().Before("gorm:create").Register("test:user_id", func(tx *gorm.DB) {
		if u, ok := tx.Statement.Dest.(*models.User); ok && u.ID == uuid.Nil {
			u.ID = uuid.New()
		}
	}); err != nil {
		t.Fatal(err)
	}

	e := &seatEnv{t: t, db: db, roles: map[string]string{}}
	for _, name := range []string{"ADMIN", "AUDITOR", "EXECUTIVE", "AUDITEE", "VIEWER", "DEPARTMENT_HEAD"} {
		id := uuid.NewString()
		if err := db.Exec(`INSERT INTO roles (id, name) VALUES (?, ?)`, id, name).Error; err != nil {
			t.Fatal(err)
		}
		e.roles[name] = id
	}

	e.svc = userService.NewUserService(repositories.NewUserRepository(db), nil)
	ctrl := NewUserController(validations.New(), e.svc)
	r := gin.New()
	asAdmin := func(c *gin.Context) { c.Set("roles", []string{"ADMIN"}) }
	r.Use(asAdmin)
	r.POST("/users", ctrl.CreateUser)
	r.GET("/users/seats", ctrl.GetSeats)
	r.PUT("/users/:id", ctrl.UpdateUser)
	r.PATCH("/users/:id/status", ctrl.UpdateUserStatus)
	r.PUT("/users/:id/roles", ctrl.ReplaceUserRoles)
	r.DELETE("/users/:id", ctrl.DeleteUser)
	e.r = r
	return e
}

// user inserts a user holding roles.
func (e *seatEnv) user(active bool, roles ...string) uuid.UUID {
	e.t.Helper()
	e.n++
	id := uuid.New()
	if err := e.db.Exec(`INSERT INTO users (id, username, email, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id.String(), fmt.Sprintf("u%d", e.n), fmt.Sprintf("u%d@x.id", e.n), active, time.Now(), time.Now()).Error; err != nil {
		e.t.Fatal(err)
	}
	for _, r := range roles {
		if err := e.db.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, id.String(), e.roles[r]).Error; err != nil {
			e.t.Fatal(err)
		}
	}
	return id
}

// fill adds n active AUDITOR users.
func (e *seatEnv) fill(n int) {
	for i := 0; i < n; i++ {
		e.user(true, "AUDITOR")
	}
}

func (e *seatEnv) do(method, path, body string) (int, map[string]any) {
	e.t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func (e *seatEnv) rolesOf(id uuid.UUID) []string {
	e.t.Helper()
	var names []string
	if err := e.db.Raw(`SELECT ro.name FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id WHERE ur.user_id = ? ORDER BY ro.name`, id.String()).
		Scan(&names).Error; err != nil {
		e.t.Fatal(err)
	}
	return names
}

func (e *seatEnv) isActive(id uuid.UUID) bool {
	e.t.Helper()
	var active bool
	if err := e.db.Raw(`SELECT is_active FROM users WHERE id = ?`, id.String()).Scan(&active).Error; err != nil {
		e.t.Fatal(err)
	}
	return active
}

func (e *seatEnv) seatsUsed() int64 {
	e.t.Helper()
	u, err := e.svc.GetSeatUsage(context.Background())
	if err != nil {
		e.t.Fatal(err)
	}
	return u.Used
}

func errCode(out map[string]any) string {
	if e, ok := out["error"].(map[string]any); ok {
		s, _ := e["code"].(string)
		return s
	}
	return ""
}

func data(out map[string]any) map[string]any {
	d, _ := out["data"].(map[string]any)
	return d
}

func strs(v any) []string {
	list, _ := v.([]any)
	out := make([]string, 0, len(list))
	for _, x := range list {
		s, _ := x.(string)
		out = append(out, s)
	}
	return out
}

func eqStrs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func createBody(username string, roles ...string) string {
	b, _ := json.Marshal(map[string]any{
		"username": username, "email": username + "@x.id", "password": "secret1", "roles": roles,
	})
	return string(b)
}

func TestSeatCreateAtLimitBecomesViewer(t *testing.T) {
	e := newSeatEnv(t)
	e.fill(49)

	// Seat 50 is free: full roles, no flag
	code, out := e.do(http.MethodPost, "/users", createBody("fiftieth", "AUDITOR"))
	if code != http.StatusCreated {
		t.Fatalf("create with room: %d %v", code, out)
	}
	d := data(out)
	if d["seat_downgraded"] != false || !eqStrs(strs(d["assigned_roles"]), []string{"AUDITOR"}) || !eqStrs(strs(d["requested_roles"]), []string{"AUDITOR"}) {
		t.Fatalf("create with room data = %v", d)
	}
	if got := e.seatsUsed(); got != 50 {
		t.Fatalf("used = %d, want 50", got)
	}

	// Full: still created, but VIEWER only, and the response says so
	code, out = e.do(http.MethodPost, "/users", createBody("fiftyfirst", "AUDITOR", "DEPARTMENT_HEAD"))
	if code != http.StatusCreated {
		t.Fatalf("create when full: %d %v", code, out)
	}
	d = data(out)
	if d["seat_downgraded"] != true {
		t.Fatalf("seat_downgraded = %v, want true", d["seat_downgraded"])
	}
	if !eqStrs(strs(d["requested_roles"]), []string{"AUDITOR", "DEPARTMENT_HEAD"}) {
		t.Fatalf("requested_roles = %v", d["requested_roles"])
	}
	if !eqStrs(strs(d["assigned_roles"]), []string{"VIEWER"}) {
		t.Fatalf("assigned_roles = %v", d["assigned_roles"])
	}
	id := uuid.MustParse(d["id"].(string))
	if got := e.rolesOf(id); !eqStrs(got, []string{"VIEWER"}) {
		t.Fatalf("stored roles = %v, want [VIEWER]", got)
	}
	if !e.isActive(id) {
		t.Fatal("downgraded user must be active")
	}
	if got := e.seatsUsed(); got != 50 {
		t.Fatalf("used after downgrade = %d, want 50", got)
	}
}

func TestSeatCreateViewerOnlyBypassesLimit(t *testing.T) {
	e := newSeatEnv(t)
	e.fill(53) // already over the limit

	code, out := e.do(http.MethodPost, "/users", createBody("watcher", "VIEWER"))
	if code != http.StatusCreated {
		t.Fatalf("viewer create: %d %v", code, out)
	}
	d := data(out)
	if d["seat_downgraded"] != false || !eqStrs(strs(d["assigned_roles"]), []string{"VIEWER"}) {
		t.Fatalf("viewer create data = %v", d)
	}
	if got := e.seatsUsed(); got != 53 {
		t.Fatalf("used = %d, want 53", got)
	}
}

func TestSeatCreateFailureLeavesNoAccount(t *testing.T) {
	e := newSeatEnv(t)
	code, out := e.do(http.MethodPost, "/users", createBody("ghost", "NO_SUCH_ROLE"))
	if code != http.StatusBadRequest || errCode(out) != "ROLE_NOT_FOUND" {
		t.Fatalf("unknown role: %d %v", code, out)
	}
	var n int64
	e.db.Raw(`SELECT COUNT(*) FROM users WHERE username = 'ghost'`).Scan(&n)
	if n != 0 {
		t.Fatal("the failed create must be rolled back")
	}
}

func TestSeatReactivationAtLimitIs409(t *testing.T) {
	e := newSeatEnv(t)
	e.fill(50)
	sleeper := e.user(false, "AUDITOR")
	viewer := e.user(false, "VIEWER")
	noRole := e.user(false)

	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPatch, "/users/" + sleeper.String() + "/status", `{"is_active":true}`},
		{http.MethodPut, "/users/" + sleeper.String(), `{"is_active":true}`},
	} {
		code, out := e.do(tc.method, tc.path, tc.body)
		if code != http.StatusConflict || errCode(out) != "SEAT_LIMIT_REACHED" {
			t.Fatalf("%s %s: %d %v", tc.method, tc.path, code, out)
		}
		d := data(out)
		if d["limit"] != float64(50) || d["used"] != float64(50) || d["remaining"] != float64(0) {
			t.Fatalf("details = %v", d)
		}
	}
	if e.isActive(sleeper) {
		t.Fatal("user must stay inactive")
	}
	if got := e.rolesOf(sleeper); !eqStrs(got, []string{"AUDITOR"}) {
		t.Fatalf("roles changed to %v", got)
	}

	// Users without a seat role reactivate freely
	for _, id := range []uuid.UUID{viewer, noRole} {
		if code, out := e.do(http.MethodPatch, "/users/"+id.String()+"/status", `{"is_active":true}`); code != http.StatusOK {
			t.Fatalf("reactivate seatless user: %d %v", code, out)
		}
	}

	// Deactivating a seat holder frees a seat; then the reactivation goes through
	holder := e.user(true, "AUDITOR") // 51st, as if the tenant was over the limit
	if code, _ := e.do(http.MethodPatch, "/users/"+holder.String()+"/status", `{"is_active":false}`); code != http.StatusOK {
		t.Fatalf("deactivate: %d", code)
	}
	if code, _ := e.do(http.MethodPatch, "/users/"+sleeper.String()+"/status", `{"is_active":true}`); code != http.StatusConflict {
		t.Fatalf("still 50 used: %d, want 409", code)
	}
	var someone string
	e.db.Raw(`SELECT ur.user_id FROM user_roles ur JOIN users u ON u.id = ur.user_id WHERE ur.role_id = ? AND u.is_active = 1 LIMIT 1`, e.roles["AUDITOR"]).Scan(&someone)
	if code, _ := e.do(http.MethodPatch, "/users/"+someone+"/status", `{"is_active":false}`); code != http.StatusOK {
		t.Fatalf("deactivate seat holder: %d", code)
	}
	if code, out := e.do(http.MethodPatch, "/users/"+sleeper.String()+"/status", `{"is_active":true}`); code != http.StatusOK {
		t.Fatalf("reactivate with room: %d %v", code, out)
	}
	if got := e.seatsUsed(); got != 50 {
		t.Fatalf("used = %d, want 50", got)
	}

	// Missing is_active is a 400
	if code, _ := e.do(http.MethodPatch, "/users/"+sleeper.String()+"/status", `{}`); code != http.StatusBadRequest {
		t.Fatalf("missing is_active: %d, want 400", code)
	}
}

func TestSeatRoleUpgradeAtLimitIs409(t *testing.T) {
	e := newSeatEnv(t)
	e.fill(50)
	viewer := e.user(true, "VIEWER")
	noRole := e.user(true) // active account without any role, treated like a VIEWER

	for _, id := range []uuid.UUID{viewer, noRole} {
		before := e.rolesOf(id)
		code, out := e.do(http.MethodPut, "/users/"+id.String()+"/roles", `{"roles":["AUDITOR"]}`)
		if code != http.StatusConflict || errCode(out) != "SEAT_LIMIT_REACHED" {
			t.Fatalf("upgrade at limit: %d %v", code, out)
		}
		if d := data(out); d["used"] != float64(50) || d["remaining"] != float64(0) {
			t.Fatalf("details = %v", d)
		}
		if got := e.rolesOf(id); !eqStrs(got, before) {
			t.Fatalf("roles changed to %v", got)
		}
	}

	// An inactive user takes no seat, so its roles can change; reactivation checks later
	sleeper := e.user(false, "VIEWER")
	if code, out := e.do(http.MethodPut, "/users/"+sleeper.String()+"/roles", `{"roles":["AUDITOR"]}`); code != http.StatusOK {
		t.Fatalf("inactive upgrade: %d %v", code, out)
	}

	// Empty list is a 400
	if code, out := e.do(http.MethodPut, "/users/"+viewer.String()+"/roles", `{"roles":[]}`); code != http.StatusBadRequest || errCode(out) != "ROLE_REQUIRED" {
		t.Fatalf("empty roles: %d %v", code, out)
	}
}

func TestSeatDowngradesAndLateralMovesAllowedAtLimit(t *testing.T) {
	e := newSeatEnv(t)
	e.fill(48)
	lateral := e.user(true, "AUDITOR")
	down := e.user(true, "AUDITOR", "DEPARTMENT_HEAD")

	code, out := e.do(http.MethodPut, "/users/"+lateral.String()+"/roles", `{"roles":["EXECUTIVE","AUDITEE"]}`)
	if code != http.StatusOK {
		t.Fatalf("lateral move: %d %v", code, out)
	}
	if got := strs(data(out)["roles"]); len(got) != 2 {
		t.Fatalf("response roles = %v", got)
	}
	if got := e.rolesOf(lateral); !eqStrs(got, []string{"AUDITEE", "EXECUTIVE"}) {
		t.Fatalf("lateral roles = %v", got)
	}

	if code, out := e.do(http.MethodPut, "/users/"+down.String()+"/roles", `{"roles":["VIEWER"]}`); code != http.StatusOK {
		t.Fatalf("downgrade: %d %v", code, out)
	}
	if got := e.rolesOf(down); !eqStrs(got, []string{"VIEWER"}) {
		t.Fatalf("downgraded roles = %v", got)
	}
	if got := e.seatsUsed(); got != 49 {
		t.Fatalf("used after downgrade = %d, want 49", got)
	}
}

func TestLastAdminGuard(t *testing.T) {
	e := newSeatEnv(t)
	admin := e.user(true, "ADMIN", "AUDITOR")
	e.user(false, "ADMIN") // an inactive admin does not count
	e.fill(3)

	cases := []struct{ name, method, path, body string }{
		{"deactivate", http.MethodPatch, "/users/" + admin.String() + "/status", `{"is_active":false}`},
		{"deactivate via PUT", http.MethodPut, "/users/" + admin.String(), `{"is_active":false}`},
		{"soft delete", http.MethodDelete, "/users/" + admin.String(), ``},
		{"hard delete", http.MethodDelete, "/users/" + admin.String() + "?hard=true", ``},
		{"remove ADMIN", http.MethodPut, "/users/" + admin.String() + "/roles", `{"roles":["AUDITOR"]}`},
	}
	for _, tc := range cases {
		code, out := e.do(tc.method, tc.path, tc.body)
		if code != http.StatusConflict || errCode(out) != "LAST_ADMIN" {
			t.Fatalf("%s: %d %v, want 409 LAST_ADMIN", tc.name, code, out)
		}
	}
	if !e.isActive(admin) || !eqStrs(e.rolesOf(admin), []string{"ADMIN", "AUDITOR"}) {
		t.Fatal("the last admin must be unchanged")
	}
	var n int64
	e.db.Raw(`SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL`, admin.String()).Scan(&n)
	if n != 1 {
		t.Fatal("the last admin must not be deleted")
	}

	// Keeping ADMIN while changing other roles is fine
	if code, out := e.do(http.MethodPut, "/users/"+admin.String()+"/roles", `{"roles":["ADMIN"]}`); code != http.StatusOK {
		t.Fatalf("keep ADMIN: %d %v", code, out)
	}

	// With a second active admin each of these is allowed
	second := e.user(true, "ADMIN")
	if code, out := e.do(http.MethodPut, "/users/"+admin.String()+"/roles", `{"roles":["AUDITOR"]}`); code != http.StatusOK {
		t.Fatalf("remove ADMIN with two admins: %d %v", code, out)
	}
	// second is now the last one
	if code, out := e.do(http.MethodPatch, "/users/"+second.String()+"/status", `{"is_active":false}`); code != http.StatusConflict || errCode(out) != "LAST_ADMIN" {
		t.Fatalf("new last admin: %d %v", code, out)
	}
	third := e.user(true, "ADMIN")
	if code, out := e.do(http.MethodDelete, "/users/"+second.String(), ``); code != http.StatusOK {
		t.Fatalf("soft delete with two admins: %d %v", code, out)
	}
	fourth := e.user(true, "ADMIN")
	if code, out := e.do(http.MethodDelete, "/users/"+third.String()+"?hard=true", ``); code != http.StatusOK {
		t.Fatalf("hard delete with two admins: %d %v", code, out)
	}
	if code, out := e.do(http.MethodPatch, "/users/"+fourth.String()+"/status", `{"is_active":false}`); code != http.StatusConflict {
		t.Fatalf("fourth is the last admin: %d %v", code, out)
	}
}

func TestSeatsEndpointNumbers(t *testing.T) {
	e := newSeatEnv(t)
	e.user(true, "ADMIN")                      // counts (ADMIN is not exempt)
	e.user(true, "AUDITOR", "DEPARTMENT_HEAD") // counts once
	e.user(true, "VIEWER", "AUDITEE")          // counts: AUDITEE is not VIEWER
	e.user(true, "VIEWER")                     // no seat
	e.user(true)                               // no roles: no seat
	e.user(false, "AUDITOR")                   // inactive: no seat
	deleted := e.user(true, "AUDITOR")         // soft-deleted: no seat
	e.db.Exec(`UPDATE users SET deleted_at = ? WHERE id = ?`, time.Now(), deleted.String())

	code, out := e.do(http.MethodGet, "/users/seats", "")
	if code != http.StatusOK {
		t.Fatalf("seats: %d %v", code, out)
	}
	d := data(out)
	if d["limit"] != float64(50) || d["used"] != float64(3) || d["remaining"] != float64(47) || d["over_limit"] != false {
		t.Fatalf("seats = %v", d)
	}

	e.fill(47) // exactly at the limit
	_, out = e.do(http.MethodGet, "/users/seats", "")
	if d := data(out); d["used"] != float64(50) || d["remaining"] != float64(0) || d["over_limit"] != false {
		t.Fatalf("at limit = %v", d)
	}

	e.fill(2) // over the limit (tenant that had more users already)
	_, out = e.do(http.MethodGet, "/users/seats", "")
	if d := data(out); d["used"] != float64(52) || d["remaining"] != float64(0) || d["over_limit"] != true {
		t.Fatalf("over limit = %v", d)
	}
}

func TestHardDeleteRemovesRelatedRows(t *testing.T) {
	e := newSeatEnv(t)
	e.user(true, "ADMIN")
	victim := e.user(true, "AUDITOR", "VIEWER")
	other := e.user(true, "AUDITOR")
	vid, oid := victim.String(), other.String()
	for _, uid := range []string{vid, oid} {
		e.db.Exec(`INSERT INTO mfa_setups (id, user_id, mfa_type, secret_key) VALUES (?, ?, 'TOTP', 's')`, uuid.NewString(), uid)
		e.db.Exec(`INSERT INTO trusted_devices (id, user_id, device_fingerprint) VALUES (?, ?, 'fp')`, uuid.NewString(), uid)
		e.db.Exec(`INSERT INTO refresh_tokens (id, user_id, token) VALUES (?, ?, ?)`, uuid.NewString(), uid, uuid.NewString())
		e.db.Exec(`INSERT INTO confidentiality_agreements (id, user_id, title) VALUES (?, ?, 't')`, uuid.NewString(), uid)
	}
	logID := uuid.NewString()
	e.db.Exec(`INSERT INTO sys_alert_logs (id, user_id) VALUES (?, ?)`, logID, vid)

	// Soft delete first: the row stays, hidden by deleted_at
	if code, out := e.do(http.MethodDelete, "/users/"+vid, ""); code != http.StatusOK {
		t.Fatalf("soft delete: %d %v", code, out)
	}
	count := func(q string, args ...any) int64 {
		var n int64
		if err := e.db.Raw(q, args...).Scan(&n).Error; err != nil {
			t.Fatal(err)
		}
		return n
	}
	if count(`SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NOT NULL`, vid) != 1 {
		t.Fatal("soft delete must keep the row with deleted_at set")
	}
	// A soft-deleted user is gone for a second soft delete
	if code, _ := e.do(http.MethodDelete, "/users/"+vid, ""); code != http.StatusNotFound {
		t.Fatalf("second soft delete: %d, want 404", code)
	}

	// Hard delete also works on a soft-deleted user
	if code, out := e.do(http.MethodDelete, "/users/"+vid+"?hard=true", ""); code != http.StatusOK {
		t.Fatalf("hard delete: %d %v", code, out)
	}
	for _, table := range []string{"user_roles", "mfa_setups", "trusted_devices", "refresh_tokens", "confidentiality_agreements"} {
		if n := count(`SELECT COUNT(*) FROM `+table+` WHERE user_id = ?`, vid); n != 0 {
			t.Fatalf("%s still has %d rows of the deleted user", table, n)
		}
		if n := count(`SELECT COUNT(*) FROM `+table+` WHERE user_id = ?`, oid); n == 0 {
			t.Fatalf("%s lost the rows of another user", table)
		}
	}
	if count(`SELECT COUNT(*) FROM users WHERE id = ?`, vid) != 0 {
		t.Fatal("user row must be gone")
	}
	if count(`SELECT COUNT(*) FROM sys_alert_logs WHERE id = ? AND user_id IS NULL`, logID) != 1 {
		t.Fatal("activity log row must be kept with user_id cleared")
	}

	if code, _ := e.do(http.MethodDelete, "/users/"+vid+"?hard=true", ""); code != http.StatusNotFound {
		t.Fatalf("hard delete of a missing user: %d, want 404", code)
	}
	if code, _ := e.do(http.MethodDelete, "/users/"+oid+"?hard=maybe", ""); code != http.StatusBadRequest {
		t.Fatalf("bad hard value: %d, want 400", code)
	}
}

// Concurrent creates through the service: only one of them may take the last
// seat. On SQLite the transactions are serialised by the single connection;
// in production the pg_advisory_xact_lock in WithSeatLock does it (the lock
// is a no-op on non-Postgres dialects, so this test does not exercise it).
func TestSeatConcurrentCreatesTakeLastSeatOnce(t *testing.T) {
	e := newSeatEnv(t)
	e.fill(49)

	const workers = 12
	var wg sync.WaitGroup
	results := make([]*models.CreateUserResponse, workers)
	errs := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = e.svc.CreateUser(context.Background(), &models.CreateUserRequest{
				Username: fmt.Sprintf("racer%d", i), Email: fmt.Sprintf("racer%d@x.id", i),
				Password: "secret1", Roles: []string{"AUDITOR"},
			})
		}(i)
	}
	wg.Wait()

	full, downgraded := 0, 0
	for i := range results {
		if errs[i] != nil {
			t.Fatalf("create %d: %v", i, errs[i])
		}
		if results[i].SeatDowngraded {
			downgraded++
		} else {
			full++
		}
	}
	if full != 1 || downgraded != workers-1 {
		t.Fatalf("full = %d, downgraded = %d; want 1 and %d", full, downgraded, workers-1)
	}
	if got := e.seatsUsed(); got != 50 {
		t.Fatalf("used = %d, want 50", got)
	}
}
