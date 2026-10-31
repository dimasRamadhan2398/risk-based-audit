package job_role

import (
	"testing"

	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"

	"github.com/google/uuid"
)

type fakeRepo struct{ rows []*models.JobRole }

func (f *fakeRepo) Create(r *models.JobRole) error {
	r.ID = uuid.New()
	f.rows = append(f.rows, r)
	return nil
}
func (f *fakeRepo) Update(r *models.JobRole) error { return nil }
func (f *fakeRepo) Delete(uuid.UUID) error         { return nil }
func (f *fakeRepo) FindByID(id uuid.UUID) (*models.JobRole, error) {
	for _, r := range f.rows {
		if r.ID == id {
			c := *r
			return &c, nil
		}
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeRepo) FindByCode(code string) (*models.JobRole, error) {
	for _, r := range f.rows {
		if r.JobRoleCode == code {
			return r, nil
		}
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeRepo) FindByName(name string) (*models.JobRole, error) {
	for _, r := range f.rows {
		if r.JobRoleName == name {
			return r, nil
		}
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeRepo) FindAll(activeOnly bool) ([]*models.JobRole, error) {
	var out []*models.JobRole
	for _, r := range f.rows {
		if !activeOnly || r.IsActive {
			out = append(out, r)
		}
	}
	return out, nil
}

func newRole(code, name, access string) *models.JobRole {
	return &models.JobRole{JobRoleCode: code, JobRoleName: name, JobPositionType: models.PositionTypeStaff, DefaultAccessRole: access, IsActive: true}
}

func TestCreateAccessRole(t *testing.T) {
	ctx := base.NewBaseService()
	svc := NewJobRoleService(&fakeRepo{})

	got, err := svc.Create(ctx, newRole("A", "Alpha", " executive "))
	if err != nil || got.DefaultAccessRole != "EXECUTIVE" {
		t.Fatalf("normalized create: %+v, %v", got, err)
	}
	if got, err := svc.Create(ctx, newRole("B", "Beta", "")); err != nil || got.DefaultAccessRole != "" {
		t.Fatalf("empty mapping must be allowed: %+v, %v", got, err)
	}
	for _, bad := range []string{"ADMIN", "SUPERUSER", "AUDIT_MANAGER"} {
		_, err := svc.Create(ctx, newRole("C-"+bad, "Gamma "+bad, bad))
		ae, ok := apperrors.As(err)
		if !ok || ae.StatusCode != 400 {
			t.Errorf("%s: want 400 validation error, got %v", bad, err)
		}
	}
}

func TestDuplicateCodeAndNameAre409(t *testing.T) {
	ctx := base.NewBaseService()
	svc := NewJobRoleService(&fakeRepo{})
	if _, err := svc.Create(ctx, newRole("A", "Alpha", "")); err != nil {
		t.Fatal(err)
	}
	for _, dup := range []*models.JobRole{newRole("A", "Other", ""), newRole("Z", " Alpha ", "")} {
		_, err := svc.Create(ctx, dup)
		if ae, ok := apperrors.As(err); !ok || ae.StatusCode != 409 {
			t.Errorf("%s/%s: want 409, got %v", dup.JobRoleCode, dup.JobRoleName, err)
		}
	}
}

func TestUpdateAccessRole(t *testing.T) {
	ctx := base.NewBaseService()
	svc := NewJobRoleService(&fakeRepo{})
	a, _ := svc.Create(ctx, newRole("A", "Alpha", "AUDITEE"))
	if _, err := svc.Create(ctx, newRole("B", "Beta", "")); err != nil {
		t.Fatal(err)
	}

	got, err := svc.Update(ctx, a.ID.String(), newRole("A", "Alpha", "auditor"))
	if err != nil || got.DefaultAccessRole != "AUDITOR" {
		t.Fatalf("update: %+v, %v", got, err)
	}
	if _, err := svc.Update(ctx, a.ID.String(), newRole("A", "Alpha", "ADMIN")); err == nil {
		t.Error("ADMIN must be rejected")
	}
	if _, err := svc.Update(ctx, a.ID.String(), newRole("A", "Beta", "")); err == nil {
		t.Error("renaming onto an existing name must conflict")
	}
}

func TestFindAllActiveFilter(t *testing.T) {
	ctx := base.NewBaseService()
	repo := &fakeRepo{}
	svc := NewJobRoleService(repo)
	svc.Create(ctx, newRole("A", "Alpha", ""))
	off := newRole("B", "Beta", "")
	off.IsActive = false
	svc.Create(ctx, off)

	if all, _ := svc.FindAll(ctx, false); len(*all) != 2 {
		t.Errorf("FindAll(false) = %d rows, want 2", len(*all))
	}
	if act, _ := svc.FindAll(ctx, true); len(*act) != 1 {
		t.Errorf("FindAll(true) = %d rows, want 1", len(*act))
	}
}
