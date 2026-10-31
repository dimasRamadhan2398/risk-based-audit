package location

import (
	"errors"
	"testing"

	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"

	"github.com/google/uuid"
)

type fakeLocationRepo struct {
	rows map[uuid.UUID]models.Location
}

func newFakeLocationRepo(seed ...models.Location) *fakeLocationRepo {
	f := &fakeLocationRepo{rows: map[uuid.UUID]models.Location{}}
	for _, l := range seed {
		f.rows[l.ID] = l
	}
	return f
}

func (f *fakeLocationRepo) Create(l *models.Location) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	f.rows[l.ID] = *l
	return nil
}
func (f *fakeLocationRepo) Update(l *models.Location) error { f.rows[l.ID] = *l; return nil }
func (f *fakeLocationRepo) Delete(id uuid.UUID) error       { delete(f.rows, id); return nil }
func (f *fakeLocationRepo) FindByID(id uuid.UUID) (*models.Location, error) {
	l, ok := f.rows[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return &l, nil
}
func (f *fakeLocationRepo) FindByName(string) (*models.Location, error) {
	return nil, apperrors.ErrNotFound
}
func (f *fakeLocationRepo) FindByResourceAndAction(string, string) (*models.Location, error) {
	return nil, apperrors.ErrNotFound
}
func (f *fakeLocationRepo) FindAll() ([]*models.Location, error)                  { return nil, nil }
func (f *fakeLocationRepo) FindMany(int, int, string) ([]*models.Location, error) { return nil, nil }
func (f *fakeLocationRepo) Count(string) (int64, error)                           { return 0, nil }

func validationFields(t *testing.T, err error) map[string]string {
	t.Helper()
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeValidationFailed {
		t.Fatalf("err = %v, want VALIDATION_FAILED", err)
	}
	return appErr.Fields
}

func TestCreate_ValidNamesCanonicalised(t *testing.T) {
	svc := NewLocationService(newFakeLocationRepo())
	got, err := svc.Create(base.NewBaseService(), &models.Location{
		Name: "HQ", Address: "Jl. Asia Afrika", City: "kota bandung", Province: "JAWA BARAT", Country: "indonesia",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.City != "Kota Bandung" || got.Province != "Jawa Barat" || got.Country != "Indonesia" {
		t.Fatalf("stored %q / %q / %q", got.City, got.Province, got.Country)
	}
}

func TestCreate_EmptyCountryIsIndonesia(t *testing.T) {
	svc := NewLocationService(newFakeLocationRepo())
	_, err := svc.Create(base.NewBaseService(), &models.Location{Name: "HQ", City: "Surabaya", Province: "Jawa Timur"})
	if f := validationFields(t, err); f["city"] != apperrors.FieldInvalidOption {
		t.Fatalf("fields = %v", f)
	}
}

func TestCreate_InvalidNamesRejected(t *testing.T) {
	svc := NewLocationService(newFakeLocationRepo())
	_, err := svc.Create(base.NewBaseService(), &models.Location{Name: "HQ", City: "Kota Bandung", Province: "Jawa Timur", Country: "Indonesia"})
	if f := validationFields(t, err); f["city"] != apperrors.FieldMismatch {
		t.Fatalf("city not in province: fields = %v", f)
	}
	_, err = svc.Create(base.NewBaseService(), &models.Location{Name: "HQ", Country: "Atlantis"})
	if f := validationFields(t, err); f["country"] != apperrors.FieldInvalidOption {
		t.Fatalf("unknown country: fields = %v", f)
	}
}

func TestCreate_NonIndonesiaSkipsRegionCheck(t *testing.T) {
	svc := NewLocationService(newFakeLocationRepo())
	got, err := svc.Create(base.NewBaseService(), &models.Location{
		Name: "SG Office", City: "Singapore", Province: "Central Region", Country: "sg",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.Country != "Singapore" || got.City != "Singapore" || got.Province != "Central Region" {
		t.Fatalf("stored %q / %q / %q", got.City, got.Province, got.Country)
	}
}

func TestUpdate_LegacyValuesAccepted(t *testing.T) {
	stored := models.Location{ID: uuid.New(), Name: "HQ", City: "Jakarta Selatan", Province: "DKI Jakarta", Country: "Indonesia"}
	svc := NewLocationService(newFakeLocationRepo(stored))

	got, err := svc.Update(base.NewBaseService(), stored.ID.String(), &models.Location{
		Name: "HQ Renamed", City: "Jakarta Selatan", Province: "DKI Jakarta", Country: "Indonesia", IsActive: true,
	})
	if err != nil {
		t.Fatalf("Update with unchanged legacy values: %v", err)
	}
	if got.Name != "HQ Renamed" || got.City != "Jakarta Selatan" || got.Province != "DKI Jakarta" {
		t.Fatalf("got %+v", got)
	}

	_, err = svc.Update(base.NewBaseService(), stored.ID.String(), &models.Location{
		Name: "HQ", City: "Jakarta Timur", Province: "DKI Jakarta", Country: "Indonesia",
	})
	if f := validationFields(t, err); f["city"] != apperrors.FieldInvalidOption {
		t.Fatalf("changed free-text city: fields = %v", f)
	}
}
