package employee

import (
	"context"
	"testing"

	"master-service/models"
	"master-service/pkg/authclient"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() {
	if logger.Log == nil {
		logger.Log = zap.NewNop()
	}
}

type fakeAccounts struct{ createdRoles [][]string }

func (f *fakeAccounts) FindUser(context.Context, string, string, string) (*authclient.User, error) {
	return nil, nil
}
func (f *fakeAccounts) CreateUser(_ context.Context, in authclient.CreateUserInput, _ string) (*authclient.CreatedUser, error) {
	f.createdRoles = append(f.createdRoles, in.Roles)
	return &authclient.CreatedUser{Username: in.Username}, nil
}
func (f *fakeAccounts) LinkEmployee(context.Context, string, string, string) error { return nil }

type fakeJobRoles map[uuid.UUID]*models.JobRole

func (f fakeJobRoles) FindByID(id uuid.UUID) (*models.JobRole, error) {
	if r, ok := f[id]; ok {
		return r, nil
	}
	return nil, apperrors.ErrNotFound
}

func TestLoginRoleFollowsJobRole(t *testing.T) {
	exec, unmapped := uuid.New(), uuid.New()
	jobRoles := fakeJobRoles{
		exec:     {DefaultAccessRole: "EXECUTIVE"},
		unmapped: {},
	}
	cases := []struct {
		name     string
		jobRole  uuid.UUID
		override string
		lookup   JobRoleLookup
		want     string
	}{
		{"mapped job role", exec, "", jobRoles, "EXECUTIVE"},
		{"override wins", exec, "auditor", jobRoles, "AUDITOR"},
		{"no mapping falls back to AUDITEE", unmapped, "", jobRoles, "AUDITEE"},
		{"unknown job role falls back", uuid.New(), "", jobRoles, "AUDITEE"},
		{"no lookup configured", exec, "", nil, "AUDITEE"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			accounts := &fakeAccounts{}
			svc := NewEmployeeServiceWithAccess(newFakeRepo(), accounts, c.lookup)
			_, err := svc.Create(base.NewBaseService(), &models.Employee{
				FullName: "A B", Email: "ab@aifl.co.id", JobRoleID: c.jobRole, AccessRole: c.override,
			})
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if len(accounts.createdRoles) != 1 || len(accounts.createdRoles[0]) != 1 || accounts.createdRoles[0][0] != c.want {
				t.Fatalf("roles sent to auth-service = %v, want [%s]", accounts.createdRoles, c.want)
			}
		})
	}
}

func TestInvalidAccessRoleOverrideRejectedBeforeWrite(t *testing.T) {
	repo := newFakeRepo()
	accounts := &fakeAccounts{}
	svc := NewEmployeeServiceWithAccess(repo, accounts, nil)
	_, err := svc.Create(base.NewBaseService(), &models.Employee{FullName: "A", Email: "a@aifl.co.id", AccessRole: "ADMIN"})
	if ae, ok := apperrors.As(err); !ok || ae.StatusCode != 400 {
		t.Fatalf("want 400, got %v", err)
	}
	if len(repo.created) != 0 || len(accounts.createdRoles) != 0 {
		t.Error("nothing may be written for a rejected override")
	}
}
