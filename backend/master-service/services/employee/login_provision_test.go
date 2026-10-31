package employee

import (
	"context"
	"errors"
	"testing"

	"master-service/models"
	"master-service/pkg/authclient"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
)

// stubAccounts is a configurable auth-service double.
type stubAccounts struct {
	found     *authclient.User
	findErr   error
	createErr error
	created   []authclient.CreateUserInput
	linked    []string
}

func (s *stubAccounts) FindUser(context.Context, string, string, string) (*authclient.User, error) {
	return s.found, s.findErr
}
func (s *stubAccounts) CreateUser(_ context.Context, in authclient.CreateUserInput, _ string) (*authclient.CreatedUser, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	s.created = append(s.created, in)
	return &authclient.CreatedUser{Username: in.Username, TemporaryPassword: "Tmp-pass-1"}, nil
}
func (s *stubAccounts) LinkEmployee(_ context.Context, _, code, _ string) error {
	s.linked = append(s.linked, code)
	return nil
}

// A non-company email (the reported gmail case) must get an account too.
func TestCreate_AnyEmailGetsLinkedAccount(t *testing.T) {
	repo, acc := newFakeRepo(), &stubAccounts{}
	svc := NewEmployeeServiceWithAccess(repo, acc, nil)
	emp, err := svc.Create(base.NewBaseService(), &models.Employee{FullName: "Dimas", Email: "dimas@gmail.com"})
	if err != nil {
		t.Fatal(err)
	}
	if emp.Account == nil || emp.Account.Status != "created" || emp.Account.TemporaryPassword == "" {
		t.Fatalf("account = %+v", emp.Account)
	}
	if len(acc.created) != 1 || acc.created[0].EmployeeID != emp.EmployeeCode || acc.created[0].Email != "dimas@gmail.com" {
		t.Fatalf("auth-service got %+v, code %s", acc.created, emp.EmployeeCode)
	}
}

func TestCreate_AuthFailureRollsBackNotSilent(t *testing.T) {
	repo := newFakeRepo()
	svc := NewEmployeeServiceWithAccess(repo, &stubAccounts{createErr: errors.New("dial tcp")}, nil)
	_, err := svc.Create(base.NewBaseService(), &models.Employee{FullName: "D", Email: "d@gmail.com"})
	if ae, ok := apperrors.As(err); !ok || ae.StatusCode != 502 {
		t.Fatalf("want 502, got %v", err)
	}
	if len(repo.created) != 0 {
		t.Error("employee must be rolled back")
	}
}

func TestCreate_NoAuthClientRefused(t *testing.T) {
	repo := newFakeRepo()
	svc := NewEmployeeServiceWithAccess(repo, nil, nil)
	_, err := svc.Create(base.NewBaseService(), &models.Employee{FullName: "D", Email: "d@gmail.com"})
	if ae, ok := apperrors.As(err); !ok || ae.StatusCode != 503 {
		t.Fatalf("want 503, got %v", err)
	}
	if len(repo.created) != 0 {
		t.Error("nothing may be written")
	}
}

func seeded(repo *fakeRepo) *models.Employee {
	e := &models.Employee{ID: [16]byte{1}, EmployeeCode: "EMP-0008", Email: "dimas@gmail.com", FullName: "Dimas"}
	repo.employees = append(repo.employees, e)
	return e
}

func TestProvisionLoginAccount_CreatesThenIsIdempotent(t *testing.T) {
	repo, acc := newFakeRepo(), &stubAccounts{}
	e := seeded(repo)
	svc := NewEmployeeServiceWithAccess(repo, acc, nil)

	got, err := svc.ProvisionLoginAccount(base.NewBaseService(), e.ID.String(), "auditor")
	if err != nil || got.Status != "created" || got.TemporaryPassword == "" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if acc.created[0].Roles[0] != "AUDITOR" || acc.created[0].EmployeeID != "EMP-0008" {
		t.Fatalf("sent %+v", acc.created[0])
	}

	acc.found = &authclient.User{ID: "u1", EmployeeID: "EMP-0008", Username: "dimas@gmail.com"}
	got, err = svc.ProvisionLoginAccount(base.NewBaseService(), e.ID.String(), "")
	if err != nil || got.Status != "exists" || got.TemporaryPassword != "" || len(acc.created) != 1 {
		t.Fatalf("second call got %+v, %v, creates=%d", got, err, len(acc.created))
	}
}

func TestProvisionLoginAccount_LinksUnlinkedAndRejectsOthers(t *testing.T) {
	repo, acc := newFakeRepo(), &stubAccounts{found: &authclient.User{ID: "u1", Username: "x"}}
	e := seeded(repo)
	svc := NewEmployeeServiceWithAccess(repo, acc, nil)

	got, err := svc.ProvisionLoginAccount(base.NewBaseService(), e.ID.String(), "")
	if err != nil || got.Status != "linked" || len(acc.linked) != 1 || acc.linked[0] != "EMP-0008" {
		t.Fatalf("got %+v, %v, linked=%v", got, err, acc.linked)
	}

	acc.found = &authclient.User{ID: "u2", EmployeeID: "EMP-0001"}
	_, err = svc.ProvisionLoginAccount(base.NewBaseService(), e.ID.String(), "")
	if ae, ok := apperrors.As(err); !ok || ae.StatusCode != 409 {
		t.Fatalf("want 409, got %v", err)
	}
}

func TestProvisionLoginAccount_Errors(t *testing.T) {
	repo := newFakeRepo()
	e := seeded(repo)
	svc := NewEmployeeServiceWithAccess(repo, &stubAccounts{}, nil)
	if _, err := svc.ProvisionLoginAccount(base.NewBaseService(), "bad", ""); err == nil {
		t.Error("bad id must fail")
	}
	if _, err := svc.ProvisionLoginAccount(base.NewBaseService(), "00000000-0000-0000-0000-0000000000ff", ""); err == nil {
		t.Error("unknown employee must fail")
	}
	if _, err := svc.ProvisionLoginAccount(base.NewBaseService(), e.ID.String(), "ADMIN"); err == nil {
		t.Error("ADMIN override must be rejected")
	}
}
