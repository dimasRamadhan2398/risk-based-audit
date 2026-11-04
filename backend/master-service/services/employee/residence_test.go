package employee

import (
	"errors"
	"testing"

	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
)

func residenceFields(t *testing.T, err error) map[string]string {
	t.Helper()
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeValidationFailed {
		t.Fatalf("err = %v, want VALIDATION_FAILED", err)
	}
	return appErr.Fields
}

func TestCreate_ResidenceCanonicalised(t *testing.T) {
	svc := NewEmployeeService(newFakeRepo())
	e := baseEmployee("EMP-0100")
	e.ResidenceProvince = "  jawa barat"
	e.ResidenceCity = "KOTA BANDUNG "

	got, err := svc.Create(base.NewBaseService(), e)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ResidenceProvince != "Jawa Barat" || got.ResidenceCity != "Kota Bandung" {
		t.Fatalf("stored %q / %q", got.ResidenceProvince, got.ResidenceCity)
	}
}

func TestCreate_ResidenceUnknownRejected(t *testing.T) {
	repo := newFakeRepo()
	svc := NewEmployeeService(repo)
	e := baseEmployee("EMP-0101")
	e.ResidenceProvince = "DKI Jakarta"
	e.ResidenceCity = "Jakarta"

	_, err := svc.Create(base.NewBaseService(), e)
	fields := residenceFields(t, err)
	if fields["residence_province"] != apperrors.FieldInvalidOption {
		t.Fatalf("fields = %v", fields)
	}
	if len(repo.created) != 0 {
		t.Fatal("rejected employee was stored")
	}

	e = baseEmployee("EMP-0102")
	e.ResidenceCity = "Bandung" // ambiguous short form: not a reference name
	fields = residenceFields(t, func() error { _, err := svc.Create(base.NewBaseService(), e); return err }())
	if fields["residence_city"] != apperrors.FieldInvalidOption {
		t.Fatalf("fields = %v", fields)
	}
}

func TestCreate_ResidenceCityNotInProvince(t *testing.T) {
	svc := NewEmployeeService(newFakeRepo())
	e := baseEmployee("EMP-0103")
	e.ResidenceProvince = "Jawa Timur"
	e.ResidenceCity = "Kota Bandung"

	_, err := svc.Create(base.NewBaseService(), e)
	if f := residenceFields(t, err); f["residence_city"] != apperrors.FieldMismatch {
		t.Fatalf("fields = %v", f)
	}
}

func TestCreate_ResidenceEmptyAllowed(t *testing.T) {
	svc := NewEmployeeService(newFakeRepo())
	if _, err := svc.Create(base.NewBaseService(), baseEmployee("EMP-0104")); err != nil {
		t.Fatalf("Create without residence: %v", err)
	}
}

// Records saved before validation keep their free text when edited without
// touching those fields; changing them to another free-text value is refused.
func TestUpdate_ResidenceLegacyValueAccepted(t *testing.T) {
	repo := newFakeRepo()
	stored := storedEmployee()
	stored.ResidenceProvince = "DKI Jakarta"
	stored.ResidenceCity = "Jakarta"
	repo.employees = []*models.Employee{stored}
	svc := NewEmployeeService(repo)

	req := func(province, city string) *models.Employee {
		return &models.Employee{
			EmployeeCode: "EMP-0001", Email: "budi@example.com", IsActive: true,
			ResidenceProvince: province, ResidenceCity: city,
		}
	}

	got, err := svc.Update(base.NewBaseService(), stored.ID.String(), req("DKI Jakarta", " Jakarta "))
	if err != nil {
		t.Fatalf("Update with unchanged legacy values: %v", err)
	}
	if got.ResidenceProvince != "DKI Jakarta" || got.ResidenceCity != "Jakarta" {
		t.Fatalf("legacy values changed: %q / %q", got.ResidenceProvince, got.ResidenceCity)
	}

	_, err = svc.Update(base.NewBaseService(), stored.ID.String(), req("DKI Jakarta", "Jakarta Pusat"))
	if f := residenceFields(t, err); f["residence_city"] != apperrors.FieldInvalidOption {
		t.Fatalf("fields = %v", f)
	}

	got, err = svc.Update(base.NewBaseService(), stored.ID.String(),
		req("daerah khusus ibukota jakarta", "kota administrasi jakarta pusat"))
	if err != nil {
		t.Fatalf("Update to reference names: %v", err)
	}
	if got.ResidenceProvince != "Daerah Khusus Ibukota Jakarta" || got.ResidenceCity != "Kota Administrasi Jakarta Pusat" {
		t.Fatalf("stored %q / %q", got.ResidenceProvince, got.ResidenceCity)
	}
}
