package reference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	provinceCodeRe = regexp.MustCompile(`^[0-9]{2}$`)
	regencyCodeRe  = regexp.MustCompile(`^([0-9]{2})\.([0-9]{2})$`)
	alpha2Re       = regexp.MustCompile(`^[A-Z]{2}$`)
	alpha3Re       = regexp.MustCompile(`^[A-Z]{3}$`)
	numericRe      = regexp.MustCompile(`^[0-9]{3}$`)
)

// Dataset is the validated, indexed reference data. It is immutable after
// Load; every accessor returns copies or read-only slices that callers must
// not modify.
type Dataset struct {
	regionsHeader   Header
	countriesHeader Header

	provinces []Province
	regencies []Regency
	countries []Country

	provinceByCode      map[string]int
	regenciesByProvince map[string][]int
	countryByAlpha2     map[string]int

	// Lookup keys are Normalize()d. Only active entries are indexed, so a
	// retired name is never accepted as a new value.
	provinceByName map[string]int
	regencyByName  map[string][]int
	countryByKey   map[string]countryMatch
}

type countryMatch struct {
	index int
	// name is the display name the key came from (English or Indonesian), so a
	// value typed in one language is stored in that language.
	name string
}

// Normalize is the comparison key for names: case-insensitive, surrounding
// whitespace removed and inner runs of whitespace collapsed to one space.
func Normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// Load parses and validates both vendored files. Any inconsistency is an error:
// the service must not start with partial or corrupt reference data.
func Load(regionsJSON, countriesJSON []byte) (*Dataset, error) {
	var rf RegionsFile
	if err := decodeStrict(regionsJSON, &rf); err != nil {
		return nil, fmt.Errorf("reference: decode regions: %w", err)
	}
	var cf CountriesFile
	if err := decodeStrict(countriesJSON, &cf); err != nil {
		return nil, fmt.Errorf("reference: decode countries: %w", err)
	}

	d := &Dataset{
		regionsHeader:       rf.Header,
		countriesHeader:     cf.Header,
		provinces:           rf.Provinces,
		regencies:           rf.Regencies,
		countries:           cf.Countries,
		provinceByCode:      map[string]int{},
		regenciesByProvince: map[string][]int{},
		countryByAlpha2:     map[string]int{},
		provinceByName:      map[string]int{},
		regencyByName:       map[string][]int{},
		countryByKey:        map[string]countryMatch{},
	}
	if err := d.indexRegions(rf.Counts); err != nil {
		return nil, fmt.Errorf("reference: regions: %w", err)
	}
	if err := d.indexCountries(cf.Counts); err != nil {
		return nil, fmt.Errorf("reference: countries: %w", err)
	}
	return d, nil
}

func decodeStrict(b []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func checkHeader(h Header) error {
	if strings.TrimSpace(h.Version) == "" {
		return fmt.Errorf("missing version")
	}
	if len(h.Sources) == 0 {
		return fmt.Errorf("missing sources")
	}
	return nil
}

func (d *Dataset) indexRegions(counts map[string]int) error {
	if err := checkHeader(d.regionsHeader); err != nil {
		return err
	}
	for i := range d.provinces {
		p := &d.provinces[i]
		if !provinceCodeRe.MatchString(p.Code) {
			return fmt.Errorf("province code %q is not two digits", p.Code)
		}
		if _, dup := d.provinceByCode[p.Code]; dup {
			return fmt.Errorf("duplicate province code %q", p.Code)
		}
		if p.Name == "" || p.Name != strings.TrimSpace(p.Name) {
			return fmt.Errorf("province %s has an empty or untrimmed name", p.Code)
		}
		d.provinceByCode[p.Code] = i
		if p.IsActive {
			key := Normalize(p.Name)
			if _, dup := d.provinceByName[key]; dup {
				return fmt.Errorf("duplicate active province name %q", p.Name)
			}
			d.provinceByName[key] = i
		}
	}

	seen := map[string]bool{}
	activeKab, activeKota := 0, 0
	for i := range d.regencies {
		r := &d.regencies[i]
		m := regencyCodeRe.FindStringSubmatch(r.Code)
		if m == nil {
			return fmt.Errorf("regency code %q is not NN.NN", r.Code)
		}
		if seen[r.Code] {
			return fmt.Errorf("duplicate regency code %q", r.Code)
		}
		seen[r.Code] = true
		if m[1] != r.ProvinceCode {
			return fmt.Errorf("regency %s has province_code %q, want %q", r.Code, r.ProvinceCode, m[1])
		}
		if _, ok := d.provinceByCode[r.ProvinceCode]; !ok {
			return fmt.Errorf("regency %s references unknown province %q", r.Code, r.ProvinceCode)
		}
		if r.Name == "" || r.Name != strings.TrimSpace(r.Name) {
			return fmt.Errorf("regency %s has an empty or untrimmed name", r.Code)
		}
		switch r.Type {
		case TypeKabupaten:
			if !strings.HasPrefix(r.Name, "Kabupaten ") {
				return fmt.Errorf("regency %s is KABUPATEN but named %q", r.Code, r.Name)
			}
		case TypeKota:
			if !strings.HasPrefix(r.Name, "Kota ") {
				return fmt.Errorf("regency %s is KOTA but named %q", r.Code, r.Name)
			}
		default:
			return fmt.Errorf("regency %s has unknown type %q", r.Code, r.Type)
		}
		d.regenciesByProvince[r.ProvinceCode] = append(d.regenciesByProvince[r.ProvinceCode], i)
		if r.IsActive {
			if r.Type == TypeKabupaten {
				activeKab++
			} else {
				activeKota++
			}
			key := Normalize(r.Name)
			d.regencyByName[key] = append(d.regencyByName[key], i)
		}
	}
	for _, r := range d.regencies {
		for _, to := range r.ReplacedBy {
			if !seen[to] {
				return fmt.Errorf("regency %s replaced_by unknown code %q", r.Code, to)
			}
		}
	}
	for _, p := range d.provinces {
		for _, to := range p.ReplacedBy {
			if _, ok := d.provinceByCode[to]; !ok {
				return fmt.Errorf("province %s replaced_by unknown code %q", p.Code, to)
			}
		}
	}

	activeProv := 0
	for _, p := range d.provinces {
		if p.IsActive {
			activeProv++
		}
	}
	if activeProv != ExpectedProvinces || activeKab != ExpectedKabupaten || activeKota != ExpectedKota {
		return fmt.Errorf("active counts provinces=%d kabupaten=%d kota=%d, want %d/%d/%d",
			activeProv, activeKab, activeKota, ExpectedProvinces, ExpectedKabupaten, ExpectedKota)
	}
	if counts["provinces"] != len(d.provinces) || counts["regencies"] != len(d.regencies) {
		return fmt.Errorf("header counts %v do not match %d provinces / %d regencies", counts, len(d.provinces), len(d.regencies))
	}
	return nil
}

func (d *Dataset) indexCountries(counts map[string]int) error {
	if err := checkHeader(d.countriesHeader); err != nil {
		return err
	}
	alpha3 := map[string]bool{}
	numeric := map[string]bool{}
	active := 0
	for i := range d.countries {
		c := &d.countries[i]
		if !alpha2Re.MatchString(c.Alpha2) || !alpha3Re.MatchString(c.Alpha3) || !numericRe.MatchString(c.Numeric) {
			return fmt.Errorf("country %q/%q/%q has a malformed code", c.Alpha2, c.Alpha3, c.Numeric)
		}
		if _, dup := d.countryByAlpha2[c.Alpha2]; dup {
			return fmt.Errorf("duplicate alpha2 %q", c.Alpha2)
		}
		if alpha3[c.Alpha3] || numeric[c.Numeric] {
			return fmt.Errorf("duplicate alpha3/numeric for %s", c.Alpha2)
		}
		alpha3[c.Alpha3], numeric[c.Numeric] = true, true
		if c.NameEN == "" || c.NameID == "" || c.NameEN != strings.TrimSpace(c.NameEN) || c.NameID != strings.TrimSpace(c.NameID) {
			return fmt.Errorf("country %s has an empty or untrimmed name", c.Alpha2)
		}
		d.countryByAlpha2[c.Alpha2] = i
		if !c.IsActive {
			continue
		}
		active++
		// Codes resolve to the English name.
		keys := []struct{ key, name string }{
			{c.Alpha2, c.NameEN},
			{c.Alpha3, c.NameEN},
			{c.NameEN, c.NameEN},
			{c.NameID, c.NameID},
		}
		for _, k := range keys {
			nk := Normalize(k.key)
			if prev, dup := d.countryByKey[nk]; dup && prev.index != i {
				return fmt.Errorf("country key %q is ambiguous (%s and %s)", k.key, d.countries[prev.index].Alpha2, c.Alpha2)
			}
			if _, dup := d.countryByKey[nk]; !dup {
				d.countryByKey[nk] = countryMatch{index: i, name: k.name}
			}
		}
	}
	if active != ExpectedCountries {
		return fmt.Errorf("active countries=%d, want %d", active, ExpectedCountries)
	}
	if counts["countries"] != len(d.countries) {
		return fmt.Errorf("header counts %v do not match %d countries", counts, len(d.countries))
	}
	return nil
}

// RegionsVersion and CountriesVersion identify the vendored data; they change
// whenever the content changes and are used as the HTTP ETag.
func (d *Dataset) RegionsVersion() string   { return d.regionsHeader.Version }
func (d *Dataset) CountriesVersion() string { return d.countriesHeader.Version }

// Provinces returns provinces in code order, active only unless includeInactive.
func (d *Dataset) Provinces(includeInactive bool) []Province {
	out := make([]Province, 0, len(d.provinces))
	for _, p := range d.provinces {
		if p.IsActive || includeInactive {
			out = append(out, p)
		}
	}
	return out
}

// Province returns the province with the given code (active or not).
func (d *Dataset) Province(code string) (Province, bool) {
	i, ok := d.provinceByCode[strings.TrimSpace(code)]
	if !ok {
		return Province{}, false
	}
	return d.provinces[i], true
}

// SearchRegencies filters kabupaten/kota. provinceCode "" means all provinces
// (callers check that a non-empty code exists). q matches a case-insensitive
// substring of the name, or the exact code.
func (d *Dataset) SearchRegencies(q, provinceCode string, includeInactive bool) []Regency {
	q = Normalize(q)
	var idx []int
	if provinceCode != "" {
		idx = d.regenciesByProvince[provinceCode]
	}
	out := make([]Regency, 0)
	match := func(r Regency) {
		if !r.IsActive && !includeInactive {
			return
		}
		if q != "" && !strings.Contains(Normalize(r.Name), q) && r.Code != q {
			return
		}
		out = append(out, r)
	}
	if provinceCode != "" {
		for _, i := range idx {
			match(d.regencies[i])
		}
		return out
	}
	for _, r := range d.regencies {
		match(r)
	}
	return out
}

// SearchCountries filters countries by a case-insensitive substring of the
// English or Indonesian name, or an exact alpha-2/alpha-3/numeric code.
func (d *Dataset) SearchCountries(q string, includeInactive bool) []Country {
	q = Normalize(q)
	out := make([]Country, 0, len(d.countries))
	for _, c := range d.countries {
		if !c.IsActive && !includeInactive {
			continue
		}
		if q != "" &&
			!strings.Contains(Normalize(c.NameEN), q) &&
			!strings.Contains(Normalize(c.NameID), q) &&
			strings.ToLower(c.Alpha2) != q && strings.ToLower(c.Alpha3) != q && c.Numeric != q {
			continue
		}
		out = append(out, c)
	}
	return out
}

// ResolveProvince matches an active province by name or code, leniently on
// case and whitespace.
func (d *Dataset) ResolveProvince(value string) (Province, bool) {
	if i, ok := d.provinceByName[Normalize(value)]; ok {
		return d.provinces[i], true
	}
	if p, ok := d.Province(value); ok && p.IsActive {
		return p, true
	}
	return Province{}, false
}

// ResolveRegency matches active kabupaten/kota by name (lenient on case and
// whitespace) or by code. Several matches are possible by name; when
// provinceCode is set only that province's regency is returned. ok is false
// when nothing matches at all; inProvince is false when the name exists but
// not in the given province.
func (d *Dataset) ResolveRegency(value, provinceCode string) (r Regency, ok bool, inProvince bool) {
	var candidates []Regency
	for _, i := range d.regencyByName[Normalize(value)] {
		candidates = append(candidates, d.regencies[i])
	}
	if len(candidates) == 0 {
		code := strings.TrimSpace(value)
		for _, reg := range d.regencies {
			if reg.Code == code && reg.IsActive {
				candidates = append(candidates, reg)
			}
		}
	}
	if len(candidates) == 0 {
		return Regency{}, false, false
	}
	if provinceCode == "" {
		return candidates[0], true, true
	}
	for _, c := range candidates {
		if c.ProvinceCode == provinceCode {
			return c, true, true
		}
	}
	return candidates[0], true, false
}

// ResolveCountry matches an active country by alpha-2, alpha-3, English or
// Indonesian name, leniently on case and whitespace. name is the display name
// to store: the matched name's language, English for a code.
func (d *Dataset) ResolveCountry(value string) (c Country, name string, ok bool) {
	m, ok := d.countryByKey[Normalize(value)]
	if !ok {
		return Country{}, "", false
	}
	return d.countries[m.index], m.name, true
}
