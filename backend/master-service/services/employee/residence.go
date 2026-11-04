package employee

import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/reference"
)

const msgResidenceInvalid = "Residence province and city must be chosen from the reference list (GET /api/v1/reference/provinces), and the city must be in the chosen province."

// checkResidence validates residence_province/residence_city against the
// embedded reference data and replaces them with the canonical names. existing
// is the stored record on update (nil on create): its unchanged values are
// accepted as they are. See pkg/reference/validate.go for the matching rules.
func checkResidence(employee, existing *models.Employee) error {
	province := reference.NameField{Key: "residence_province", Value: &employee.ResidenceProvince}
	city := reference.NameField{Key: "residence_city", Value: &employee.ResidenceCity}
	if existing != nil {
		province.Stored = &existing.ResidenceProvince
		city.Stored = &existing.ResidenceCity
	}
	fields := map[string]string{}
	reference.MustDefault().CheckRegion(province, city, fields)
	if len(fields) > 0 {
		return apperrors.ValidationFailed(msgResidenceInvalid, fields)
	}
	return nil
}
