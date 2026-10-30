package department

import (
	"encoding/json"
	"testing"

	"master-service/internal/orgtest"
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/kafka"
	repo "master-service/repositories/department"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fixture struct {
	db      *gorm.DB
	svc     OrgDepartmentServiceInterface
	company uuid.UUID
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db := orgtest.DB(t)
	orgtest.BankLevels(t, db)
	pub := kafka.NewDomainPublisher(db, "master-service-test")
	svc := NewOrgDepartmentService(repo.NewDepartmentRepository(db), repo.NewOrgRepository(db), pub)
	return &fixture{db: db, svc: svc, company: orgtest.Company(t, db)}
}

func (f *fixture) create(t *testing.T, code, name string, level int, parent *uuid.UUID) (*models.Department, error) {
	t.Helper()
	return f.svc.Create(base.NewBaseService(), &models.Department{
		DepartmentCode: code, DepartmentName: name, Level: level,
		CompanyID: f.company, IsActive: true, ParentDepartmentID: parent,
	})
}

func (f *fixture) mustCreate(t *testing.T, code, name string, level int, parent *uuid.UUID) *models.Department {
	t.Helper()
	d, err := f.create(t, code, name, level, parent)
	if err != nil {
		t.Fatalf("create %s: %v", code, err)
	}
	return d
}

func wantCode(t *testing.T, err error, code string) {
	t.Helper()
	appErr, ok := apperrors.As(err)
	if !ok || appErr.Code != code {
		t.Fatalf("want %s, got %v", code, err)
	}
}

func TestChildLevelMustBeBelowParent_SkipAllowed(t *testing.T) {
	f := newFixture(t)
	dir := f.mustCreate(t, "DIR", "Direktorat Utama", 10, nil)
	// Level skip: Departemen (30) directly under a Direktorat (10).
	audit := f.mustCreate(t, "DEPT-IA", "Internal Audit", 30, &dir.ID)
	if audit.ParentDepartmentID == nil || *audit.ParentDepartmentID != dir.ID {
		t.Fatalf("parent not stored: %+v", audit.ParentDepartmentID)
	}
	if len(audit.Path) != 2 || audit.Path[0].Name != "Direktorat Utama" || audit.Path[1].LevelName != "Departemen" {
		t.Fatalf("bad path: %+v", audit.Path)
	}
	if audit.LevelName != "Departemen" || audit.NameKey == nil || *audit.NameKey != "department" {
		t.Fatalf("level name/key not set: %q %v", audit.LevelName, audit.NameKey)
	}
	// Same level as the parent is rejected.
	_, err := f.create(t, "DIR2", "Direktorat Lain", 10, &dir.ID)
	wantCode(t, err, apperrors.CodeDepartmentLevelNotBelowParent)
	// Higher (shallower) level than the parent is rejected.
	_, err = f.create(t, "X", "Divisi X", 20, &audit.ID)
	wantCode(t, err, apperrors.CodeDepartmentLevelNotBelowParent)
}

func TestLevelMustBeAnActiveOrgLevel(t *testing.T) {
	f := newFixture(t)
	_, err := f.create(t, "L1", "Legacy level", 1, nil)
	wantCode(t, err, apperrors.CodeOrgLevelInvalid)
}

func TestCrossCompanyParentRejected(t *testing.T) {
	f := newFixture(t)
	dir := f.mustCreate(t, "DIR", "Direktorat", 10, nil)
	other := orgtest.Company(t, f.db)
	_, err := f.svc.Create(base.NewBaseService(), &models.Department{
		DepartmentCode: "OTHER", DepartmentName: "Divisi", Level: 20, CompanyID: other, ParentDepartmentID: &dir.ID,
	})
	wantCode(t, err, apperrors.CodeParentDepartmentCompanyMismatch)
}

func TestUnknownParentRejected(t *testing.T) {
	f := newFixture(t)
	missing := uuid.New()
	_, err := f.create(t, "D", "Divisi", 20, &missing)
	wantCode(t, err, apperrors.CodeParentDepartmentNotFound)
}

func TestDepthSixAllowedSevenRejected(t *testing.T) {
	db := orgtest.DB(t)
	for i := 1; i <= 6; i++ {
		orgtest.Level(t, db, i*10, "L"+string(rune('0'+i)), "")
	}
	// A 7th level number is needed to even attempt depth 7.
	if err := db.Exec(`INSERT INTO org_levels (id, level, name, is_active) VALUES (?, 70, 'L7', 1)`, uuid.New()).Error; err != nil {
		t.Fatal(err)
	}
	f := &fixture{db: db, svc: NewOrgDepartmentService(repo.NewDepartmentRepository(db), repo.NewOrgRepository(db), nil), company: orgtest.Company(t, db)}
	var parent *uuid.UUID
	for i := 1; i <= 6; i++ {
		d := f.mustCreate(t, "C"+string(rune('0'+i)), "Unit "+string(rune('0'+i)), i*10, parent)
		parent = &d.ID
	}
	_, err := f.create(t, "C7", "Unit 7", 70, parent)
	wantCode(t, err, apperrors.CodeOrgDepthExceeded)
}

func TestDuplicateLeafNameDifferentParentsAllowedSameParentRejected(t *testing.T) {
	f := newFixture(t)
	akt := f.mustCreate(t, "DIV-AKT", "Divisi Akuntansi", 20, nil)
	trs := f.mustCreate(t, "DIV-TRS", "Divisi Treasury", 20, nil)
	f.mustCreate(t, "PJK-AKT", "Departemen Pajak", 30, &akt.ID)
	f.mustCreate(t, "PJK-TRS", "Departemen Pajak", 30, &trs.ID)
	// Same parent, same name (case-insensitive): rejected.
	_, err := f.create(t, "PJK-AKT2", "departemen pajak", 30, &akt.ID)
	wantCode(t, err, apperrors.CodeDepartmentNameAlreadyExists)
	// Code stays unique (D10 default: global among live rows).
	_, err = f.create(t, "PJK-AKT", "Departemen Lain", 30, &akt.ID)
	wantCode(t, err, apperrors.CodeDepartmentCodeAlreadyExists)
}

func TestReparentUnderDescendantRejected(t *testing.T) {
	f := newFixture(t)
	dir := f.mustCreate(t, "DIR", "Direktorat", 10, nil)
	div := f.mustCreate(t, "DIV", "Divisi", 20, &dir.ID)
	upd := *dir
	upd.ParentDepartmentID = &div.ID
	_, err := f.svc.Update(base.NewBaseService(), dir.ID.String(), &upd)
	wantCode(t, err, apperrors.CodeDepartmentParentCycle)
	// Changing the level so the child is no longer below.
	upd = *dir
	upd.Level = 30
	_, err = f.svc.Update(base.NewBaseService(), dir.ID.String(), &upd)
	wantCode(t, err, apperrors.CodeDepartmentLevelChildConflict)
}

func TestDeleteWithChildrenRejected(t *testing.T) {
	f := newFixture(t)
	dir := f.mustCreate(t, "DIR", "Direktorat", 10, nil)
	div := f.mustCreate(t, "DIV", "Divisi", 20, &dir.ID)
	err := f.svc.Delete(base.NewBaseService(), dir.ID.String())
	wantCode(t, err, apperrors.CodeDepartmentHasChildren)
	if err := f.svc.Delete(base.NewBaseService(), div.ID.String()); err != nil {
		t.Fatalf("delete leaf: %v", err)
	}
	if err := f.svc.Delete(base.NewBaseService(), dir.ID.String()); err != nil {
		t.Fatalf("delete now-empty parent: %v", err)
	}
}

func TestRenameRepublishesWholeSubtree(t *testing.T) {
	f := newFixture(t)
	dir := f.mustCreate(t, "DIR-KEU", "Direktorat Keuangan", 10, nil)
	div := f.mustCreate(t, "DIV-AKT", "Divisi Akuntansi", 20, &dir.ID)
	dept := f.mustCreate(t, "DEPT-PJK", "Departemen Pajak", 30, &div.ID)
	sec := f.mustCreate(t, "SKS", "Seksi PPh", 40, &dept.ID)
	unrelated := f.mustCreate(t, "DIR-OPS", "Direktorat Operasional", 10, nil)
	before := len(orgtest.Outbox(t, f.db))

	upd := *div
	upd.DepartmentName = "Divisi Akuntansi & Pajak"
	if _, err := f.svc.Update(base.NewBaseService(), div.ID.String(), &upd); err != nil {
		t.Fatalf("rename: %v", err)
	}
	rows := orgtest.Outbox(t, f.db)[before:]
	got := map[string][]string{}
	for _, r := range rows {
		if r.EventType != kafka.DomainEventDepartmentUpdated {
			t.Fatalf("unexpected event %s", r.EventType)
		}
		var env struct {
			Data struct {
				ID                 string  `json:"id"`
				ParentDepartmentID *string `json:"parent_department_id"`
				Level              int     `json:"level"`
				LevelName          string  `json:"level_name"`
				Path               []struct {
					Name string `json:"name"`
				} `json:"path"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(r.Payload), &env); err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, p := range env.Data.Path {
			names = append(names, p.Name)
		}
		got[env.Data.ID] = names
	}
	if len(got) != 3 {
		t.Fatalf("want events for the unit and its 2 descendants, got %d: %v", len(got), got)
	}
	if _, ok := got[unrelated.ID.String()]; ok {
		t.Fatal("an unrelated unit was republished")
	}
	if p := got[sec.ID.String()]; len(p) != 4 || p[1] != "Divisi Akuntansi & Pajak" {
		t.Fatalf("descendant path not refreshed: %v", p)
	}
	if p := got[dept.ID.String()]; len(p) != 3 || p[1] != "Divisi Akuntansi & Pajak" {
		t.Fatalf("descendant path not refreshed: %v", p)
	}
}

func TestMoveToRootAndTree(t *testing.T) {
	f := newFixture(t)
	dir := f.mustCreate(t, "DIR", "Direktorat", 10, nil)
	div := f.mustCreate(t, "DIV", "Divisi", 20, &dir.ID)
	f.mustCreate(t, "DEPT", "Departemen", 30, &div.ID)

	tree, err := f.svc.Tree(base.NewBaseService(), &f.company)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 1 || len(tree[0].Children) != 1 || len(tree[0].Children[0].Children) != 1 {
		t.Fatalf("bad tree shape")
	}
	upd := *div
	upd.ParentDepartmentID = nil
	moved, err := f.svc.Update(base.NewBaseService(), div.ID.String(), &upd)
	if err != nil {
		t.Fatalf("move to root: %v", err)
	}
	if moved.ParentDepartmentID != nil || len(moved.Path) != 1 {
		t.Fatalf("not moved to root: %+v", moved.Path)
	}
	tree, _ = f.svc.Tree(base.NewBaseService(), &f.company)
	if len(tree) != 2 {
		t.Fatalf("want 2 roots after the move, got %d", len(tree))
	}
}

func TestListFiltersAndOrderWhitelist(t *testing.T) {
	f := newFixture(t)
	dir := f.mustCreate(t, "DIR", "Direktorat", 10, nil)
	f.mustCreate(t, "B", "B Divisi", 20, &dir.ID)
	f.mustCreate(t, "A", "A Divisi", 20, &dir.ID)
	order, ok := repo.ParseOrder("department_code DESC")
	if !ok {
		t.Fatal("whitelisted order rejected")
	}
	depts, total, err := f.svc.List(base.NewBaseService(), repo.ListFilter{Limit: 10, ParentID: &dir.ID, Order: order})
	if err != nil || total != 2 || depts[0].DepartmentCode != "B" {
		t.Fatalf("filtered list: total=%d err=%v %+v", total, err, depts)
	}
	for _, bad := range []string{"id; DROP TABLE departments", "pic_id", "department_name SIDEWAYS", "deleted_at"} {
		if _, ok := repo.ParseOrder(bad); ok {
			t.Fatalf("order %q accepted", bad)
		}
	}
}
