package reference

import (
	"strings"

	apperrors "master-service/pkg/errors"
)

// Field checks for free-text name columns that should hold reference names
// (employees.residence_*, locations.province/city/country).
//
// Matching: a value matches when it equals a reference display name after
// Normalize (case-insensitive, whitespace trimmed and collapsed), or equals the
// code (province "32", regency "32.73", country "ID"/"IDN"). The canonical
// display name is stored, never the typed variant or the code.
//
// Legacy values: on update, a value equal (after trimming) to the value already
// stored on the record is accepted and kept unchanged, even if it is not in the
// list, so records saved before validation existed can still be edited.
// Legacy values are never used for the city-in-province check.
//
// Failures are added to fields: INVALID_OPTION when the value is not in the
// list, MISMATCH when the city exists but not in the chosen province.

// NameField is one column to check. Stored is nil on create.
type NameField struct {
	Key    string // JSON field name for the error map
	Value  *string
	Stored *string
}

// legacy reports whether the submitted value is the record's unchanged stored
// value; if so the stored value is written back untouched.
func (f NameField) legacy() bool {
	if f.Stored == nil {
		return false
	}
	if strings.TrimSpace(*f.Value) == strings.TrimSpace(*f.Stored) {
		*f.Value = *f.Stored
		return true
	}
	return false
}

// CheckRegion validates a province/kabupaten-kota pair in place.
func (d *Dataset) CheckRegion(province, city NameField, fields map[string]string) {
	provinceCode := ""
	if strings.TrimSpace(*province.Value) == "" {
		*province.Value = ""
	} else if province.legacy() {
		// Unchanged legacy text; still use it for the city check if it resolves.
		if p, ok := d.ResolveProvince(*province.Value); ok {
			provinceCode = p.Code
		}
	} else if p, ok := d.ResolveProvince(*province.Value); ok {
		*province.Value = p.Name
		provinceCode = p.Code
	} else {
		fields[province.Key] = apperrors.FieldInvalidOption
		return // city membership cannot be judged against an invalid province
	}

	if strings.TrimSpace(*city.Value) == "" {
		*city.Value = ""
		return
	}
	if city.legacy() {
		return
	}
	r, ok, inProvince := d.ResolveRegency(*city.Value, provinceCode)
	switch {
	case !ok:
		fields[city.Key] = apperrors.FieldInvalidOption
	case !inProvince:
		fields[city.Key] = apperrors.FieldMismatch
	default:
		*city.Value = r.Name
	}
}

// CheckCountry validates a country name in place. It reports whether the
// stored value denotes Indonesia, and whether it is a known country at all;
// an empty value is reported as Indonesia (the locations.country default).
func (d *Dataset) CheckCountry(country NameField, fields map[string]string) (isIndonesia, known bool) {
	if strings.TrimSpace(*country.Value) == "" {
		*country.Value = ""
		return true, true
	}
	if country.legacy() {
		c, _, ok := d.ResolveCountry(*country.Value)
		return ok && c.Alpha2 == "ID", ok
	}
	c, name, ok := d.ResolveCountry(*country.Value)
	if !ok {
		fields[country.Key] = apperrors.FieldInvalidOption
		return false, false
	}
	*country.Value = name
	return c.Alpha2 == "ID", true
}
