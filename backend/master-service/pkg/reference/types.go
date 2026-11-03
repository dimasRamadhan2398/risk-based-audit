// Package reference serves read-only reference data (Indonesian provinces and
// kabupaten/kota, ISO 3166 countries) from JSON files embedded in the binary.
//
// The data is generated from pinned upstream sources by
// scripts/gen_reference_data and validated once at start-up; nothing is stored
// in the database. Codes are the stable keys: a code is never removed from a
// later dataset version, it is marked is_active=false with replaced_by set.
package reference

// Regency types (derived from the Kemendagri code and the official name).
const (
	TypeProvinsi  = "PROVINSI"
	TypeKabupaten = "KABUPATEN"
	TypeKota      = "KOTA"
)

// Expected row counts for the vendored dataset. The loader refuses data that
// does not match, so a bad regeneration cannot ship silently. Update these
// together with the data when the legal basis changes.
const (
	ExpectedProvinces = 38
	ExpectedKabupaten = 416
	ExpectedKota      = 98
	ExpectedCountries = 249
)

// Source records where a vendored file came from.
type Source struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Commit  string `json:"commit,omitempty"`
	Version string `json:"version,omitempty"`
	License string `json:"license"`
	SHA256  string `json:"sha256"`
}

// Header is the provenance block shared by both vendored files.
type Header struct {
	Dataset     string   `json:"dataset"`
	Version     string   `json:"version"`
	LegalBasis  string   `json:"legal_basis"`
	Sources     []Source `json:"sources"`
	GeneratedAt string   `json:"generated_at"`
	Generator   string   `json:"generator"`
}

type Province struct {
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	IsActive   bool     `json:"is_active"`
	ReplacedBy []string `json:"replaced_by"`
}

type Regency struct {
	Code         string   `json:"code"`
	ProvinceCode string   `json:"province_code"`
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	IsActive     bool     `json:"is_active"`
	ReplacedBy   []string `json:"replaced_by"`
}

type Country struct {
	Alpha2   string `json:"alpha2"`
	Alpha3   string `json:"alpha3"`
	Numeric  string `json:"numeric"`
	NameEN   string `json:"name_en"`
	NameID   string `json:"name_id"`
	IsActive bool   `json:"is_active"`
}

// RegionsFile is the on-disk shape of data/regions.json.
type RegionsFile struct {
	Header
	Counts    map[string]int `json:"counts"`
	Provinces []Province     `json:"provinces"`
	Regencies []Regency      `json:"regencies"`
}

// CountriesFile is the on-disk shape of data/countries.json.
type CountriesFile struct {
	Header
	Counts    map[string]int `json:"counts"`
	Countries []Country      `json:"countries"`
}
