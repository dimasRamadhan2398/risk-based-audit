package employee

import (
	"testing"

	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
)

func TestFindByCode_FoundAndMissing(t *testing.T) {
	repo := &fakeRepo{byCode: map[string]bool{"EMP001": true}}
	svc := &EmployeeService{employeeRepo: repo}
	ctx := base.NewBaseService()

	got, err := svc.FindByCode(ctx, "  EMP001 ")
	if err != nil {
		t.Fatalf("FindByCode(EMP001) returned error: %v", err)
	}
	if got == (*models.Employee)(nil) || got.EmployeeCode != "EMP001" {
		t.Fatalf("FindByCode(EMP001) returned %+v", got)
	}

	if _, err := svc.FindByCode(ctx, "EMP999"); err != apperrors.ErrEmployeeNotFound {
		t.Fatalf("FindByCode(EMP999) error = %v, want ErrEmployeeNotFound", err)
	}
}

func TestFindByCode_BlankCodeIsNotFoundWithoutQuery(t *testing.T) {
	svc := &EmployeeService{employeeRepo: &fakeRepo{}}
	if _, err := svc.FindByCode(base.NewBaseService(), "   "); err != apperrors.ErrEmployeeNotFound {
		t.Fatalf("blank code error = %v, want ErrEmployeeNotFound", err)
	}
}
