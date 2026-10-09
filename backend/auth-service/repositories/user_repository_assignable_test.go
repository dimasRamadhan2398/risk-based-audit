package repositories

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func assignableDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	// users.id defaults to gen_random_uuid() in the model, which sqlite cannot
	// parse, so the table is created by hand with the columns the query uses
	if err := db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY, employee_id TEXT, username TEXT, email TEXT, password_hash TEXT,
		full_name TEXT, phone TEXT, department TEXT, position TEXT, is_active BOOLEAN,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

type seedUser struct {
	id                     uuid.UUID
	name, dept, pos, email string
	active, deleted        bool
}

func insertUsers(t *testing.T, db *gorm.DB, users []seedUser) {
	t.Helper()
	for i, u := range users {
		var deleted interface{}
		if u.deleted {
			deleted = "2026-01-01 00:00:00"
		}
		if err := db.Exec(`INSERT INTO users (id, employee_id, username, email, full_name, department, position, is_active, deleted_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			u.id.String(), "E"+string(rune('A'+i)), "user"+string(rune('a'+i)), u.email,
			u.name, u.dept, u.pos, u.active, deleted).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func uid(n byte) uuid.UUID {
	var u uuid.UUID
	u[15] = n
	u[6] = 0x40
	u[8] = 0x80
	return u
}

func TestFindAssignable(t *testing.T) {
	db := assignableDB(t)
	insertUsers(t, db, []seedUser{
		{id: uid(1), name: "Budi Santoso", dept: "Finance", pos: "Manager", email: "budi@x.id", active: true},
		{id: uid(2), name: "andi wijaya", dept: "IT", pos: "Engineer", email: "andi@x.id", active: true},
		{id: uid(3), name: "Citra", dept: "Internal Audit", pos: "Auditor", email: "citra@x.id", active: true},
		{id: uid(4), name: "Dewi Inactive", dept: "Finance", pos: "Staff", email: "dewi@x.id", active: false},
		{id: uid(5), name: "Eko Deleted", dept: "Finance", pos: "Staff", email: "eko@x.id", active: true, deleted: true},
		{id: uid(6), name: "Budi Santoso", dept: "Ops", pos: "50%_Lead", email: "budi2@x.id", active: true},
	})
	repo := NewUserRepository(db)

	names := func(t *testing.T, offset, limit int, search string, id *uuid.UUID) ([]string, int64) {
		t.Helper()
		users, total, err := repo.FindAssignable(offset, limit, search, id)
		if err != nil {
			t.Fatal(err)
		}
		out := []string{}
		for _, u := range users {
			out = append(out, u.FullName+"|"+u.ID.String()[34:])
			// Only the picker's columns are read
			if u.Email != "" || u.Username != "" || !u.CreatedAt.IsZero() {
				t.Fatalf("FindAssignable loaded more than id/full_name/department/position: %+v", u)
			}
		}
		return out, total
	}

	// Active, non-deleted only; ordered by full_name then id (same names
	// tie-break on id). sqlite compares bytes, so "andi" sorts after the
	// capitalised names; Postgres orders by the database collation.
	got, total := names(t, 0, 20, "", nil)
	want := []string{"Budi Santoso|01", "Budi Santoso|06", "Citra|03", "andi wijaya|02"}
	if total != 4 || !equal(got, want) {
		t.Fatalf("all: got %v total %d, want %v total 4", got, total, want)
	}

	// Case-insensitive search on full_name, department and position
	for search, wantTotal := range map[string]int64{
		"BUDI":      2, // full_name
		"finance":   1, // department (Dewi inactive, Eko deleted)
		"AUDITOR":   1, // position
		"internal":  1,
		"engineer":  1,
		"nobody":    0,
		"budi@x.id": 0, // email is not searched
	} {
		if _, n := names(t, 0, 20, search, nil); n != wantTotal {
			t.Errorf("search %q: total %d, want %d", search, n, wantTotal)
		}
	}

	// LIKE wildcards in the search are literal
	if got, n := names(t, 0, 20, "50%_", nil); n != 1 || len(got) != 1 {
		t.Errorf(`search "50%%_": got %v total %d, want only the "50%%_Lead" user`, got, n)
	}
	if _, n := names(t, 0, 20, "%", nil); n != 1 {
		t.Errorf(`search "%%": total %d, want 1`, n)
	}

	// Pagination: total is the full count, page holds limit rows
	page2, n := names(t, 2, 2, "", nil)
	if n != 4 || len(page2) != 2 {
		t.Fatalf("page 2: got %v total %d", page2, n)
	}
	page1, _ := names(t, 0, 2, "", nil)
	for _, a := range page1 {
		for _, b := range page2 {
			if a == b {
				t.Fatalf("row %s on two pages", a)
			}
		}
	}

	// id filter: active user found, inactive not
	id := uid(3)
	if got, n := names(t, 0, 20, "", &id); n != 1 || got[0] != "Citra|03" {
		t.Fatalf("id filter: got %v total %d", got, n)
	}
	inactive := uid(4)
	if _, n := names(t, 0, 20, "", &inactive); n != 0 {
		t.Fatalf("inactive user returned by id filter")
	}
}

func equal(a, b []string) bool {
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
