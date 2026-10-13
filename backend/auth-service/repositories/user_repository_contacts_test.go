package repositories

import (
	"testing"

	"github.com/google/uuid"
)

func contactsDB(t *testing.T) (*UserRepository, func(id uuid.UUID, roleName string)) {
	t.Helper()
	db := assignableDB(t)
	if err := db.Exec(`CREATE TABLE roles (id TEXT PRIMARY KEY, name TEXT)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE user_roles (user_id TEXT, role_id TEXT)`).Error; err != nil {
		t.Fatal(err)
	}
	grant := func(id uuid.UUID, roleName string) {
		roleID := uuid.New().String()
		if err := db.Exec(`INSERT INTO roles (id, name) VALUES (?, ?)`, roleID, roleName).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, id.String(), roleID).Error; err != nil {
			t.Fatal(err)
		}
	}
	return &UserRepository{BaseRepository: NewBaseRepository(db)}, grant
}

func TestFindContactsByIDsAndDepartmentHead(t *testing.T) {
	repo, grant := contactsDB(t)
	db := repo.GetDB()
	insertUsers(t, db, []seedUser{
		{id: uid(1), name: "Budi", dept: "Finance", email: "budi@x.id", active: true},
		{id: uid(2), name: "Dewi", dept: " finance ", email: "dewi@x.id", active: true},
		{id: uid(3), name: "Eko", dept: "Finance", email: "eko@x.id", active: true},
		{id: uid(4), name: "Fitri", dept: "Finance", email: "fitri@x.id", active: false},
		{id: uid(5), name: "Gita", dept: "IT", email: "gita@x.id", active: true},
	})
	grant(uid(2), "DEPARTMENT_HEAD")
	grant(uid(3), "AUDITOR")
	grant(uid(4), "DEPARTMENT_HEAD")
	grant(uid(5), "DEPARTMENT_HEAD")

	// By id: exactly the requested active user, email included
	users, err := repo.FindContacts([]uuid.UUID{uid(1), uid(4)}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].Email != "budi@x.id" {
		t.Fatalf("by id = %+v", users)
	}

	// Department heads of Finance: trimmed, case-insensitive, active only
	heads, err := repo.FindContacts(nil, "  FINANCE ")
	if err != nil {
		t.Fatal(err)
	}
	if len(heads) != 1 || heads[0].Email != "dewi@x.id" {
		t.Fatalf("department heads = %+v", heads)
	}
}
