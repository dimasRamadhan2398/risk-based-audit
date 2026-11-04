package reference

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"

	apperrors "master-service/pkg/errors"
)

func mustDefault(t *testing.T) *Dataset {
	t.Helper()
	d, err := Default()
	if err != nil {
		t.Fatalf("embedded data invalid: %v", err)
	}
	return d
}

func TestEmbeddedCounts(t *testing.T) {
	d := mustDefault(t)
	if n := len(d.Provinces(false)); n != 38 {
		t.Errorf("provinces = %d, want 38", n)
	}
	kab, kota := 0, 0
	for _, r := range d.SearchRegencies("", "", false) {
		switch r.Type {
		case TypeKabupaten:
			kab++
		case TypeKota:
			kota++
		}
	}
	if kab != 416 || kota != 98 {
		t.Errorf("kabupaten/kota = %d/%d, want 416/98", kab, kota)
	}
	if n := len(d.SearchCountries("", false)); n != 249 {
		t.Errorf("countries = %d, want 249", n)
	}
	if d.RegionsVersion() == "" || d.CountriesVersion() == "" {
		t.Error("empty dataset version")
	}
}

func TestEmbeddedCleaning(t *testing.T) {
	d := mustDefault(t)
	cases := map[string]string{
		"31.72": "Kota Administrasi Jakarta Utara", // trailing space in source
		"31.01": "Kabupaten Administrasi Kepulauan Seribu",
		"92.03": "Kabupaten Fak Fak", // source spelling kept
		"32.73": "Kota Bandung",
		"32.04": "Kabupaten Bandung",
	}
	for code, want := range cases {
		r, ok, _ := d.ResolveRegency(code, "")
		if !ok || r.Name != want {
			t.Errorf("%s = %q (ok=%v), want %q", code, r.Name, ok, want)
		}
	}
	if r, _, _ := d.ResolveRegency("31.72", ""); r.Type != TypeKota {
		t.Errorf("31.72 type = %s, want KOTA", r.Type)
	}
	if r, _, _ := d.ResolveRegency("31.01", ""); r.Type != TypeKabupaten {
		t.Errorf("31.01 type = %s, want KABUPATEN", r.Type)
	}
	if p, ok := d.Province("31"); !ok || p.Name != "Daerah Khusus Ibukota Jakarta" {
		t.Errorf("31 = %q", p.Name)
	}
	ps, _, ok := d.ResolveCountry("PS")
	if !ok || ps.NameID != "Palestina" {
		t.Errorf("PS name_id = %q, want Palestina", ps.NameID)
	}
	if _, _, ok := d.ResolveCountry("XK"); ok {
		t.Error("Kosovo (XK) must not be in the list")
	}
	for _, code := range []string{"EU", "AN", "YU", "ZZ", "QO", "AA"} {
		if _, _, ok := d.ResolveCountry(code); ok {
			t.Errorf("%s must be excluded", code)
		}
	}
}

// TestNoCodeEverDisappears fails when a code shipped in an earlier dataset
// version is missing now. Retire codes with is_active=false and replaced_by;
// never delete them, stored references would dangle.
func TestNoCodeEverDisappears(t *testing.T) {
	d := mustDefault(t)
	f, err := os.Open("testdata/known_codes.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	known := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		known[line] = true
		parts := strings.Fields(line)
		var found bool
		switch parts[0] {
		case "province":
			_, found = d.Province(parts[1])
		case "regency":
			_, found = d.regencyIndex(parts[1])
		case "country":
			_, found = d.countryByAlpha2[parts[1]]
		default:
			t.Fatalf("bad snapshot line %q", line)
		}
		if !found {
			t.Errorf("%s disappeared from the dataset", line)
		}
	}
	// New codes must be appended to the snapshot so they are protected too.
	for _, p := range d.Provinces(true) {
		if !known["province "+p.Code] {
			t.Errorf("province %s missing from testdata/known_codes.txt", p.Code)
		}
	}
	for _, r := range d.SearchRegencies("", "", true) {
		if !known["regency "+r.Code] {
			t.Errorf("regency %s missing from testdata/known_codes.txt", r.Code)
		}
	}
	for _, c := range d.SearchCountries("", true) {
		if !known["country "+c.Alpha2] {
			t.Errorf("country %s missing from testdata/known_codes.txt", c.Alpha2)
		}
	}
}

func (d *Dataset) regencyIndex(code string) (int, bool) {
	for i, r := range d.regencies {
		if r.Code == code {
			return i, true
		}
	}
	return 0, false
}

// mutate decodes the embedded files, applies fn and re-encodes them.
func mutate(t *testing.T, fn func(*RegionsFile, *CountriesFile)) ([]byte, []byte) {
	t.Helper()
	var rf RegionsFile
	var cf CountriesFile
	if err := json.Unmarshal(regionsJSON, &rf); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(countriesJSON, &cf); err != nil {
		t.Fatal(err)
	}
	fn(&rf, &cf)
	rb, _ := json.Marshal(rf)
	cb, _ := json.Marshal(cf)
	return rb, cb
}

func TestLoadRejectsInvalidData(t *testing.T) {
	cases := map[string]func(*RegionsFile, *CountriesFile){
		"duplicate province code": func(r *RegionsFile, _ *CountriesFile) { r.Provinces[1].Code = r.Provinces[0].Code },
		"duplicate regency code":  func(r *RegionsFile, _ *CountriesFile) { r.Regencies[1].Code = r.Regencies[0].Code },
		"regency prefix mismatch": func(r *RegionsFile, _ *CountriesFile) { r.Regencies[0].ProvinceCode = "12" },
		"unknown province": func(r *RegionsFile, _ *CountriesFile) {
			r.Regencies[0].Code, r.Regencies[0].ProvinceCode = "99.01", "99"
		},
		"province missing (count)": func(r *RegionsFile, _ *CountriesFile) {
			r.Provinces = r.Provinces[:len(r.Provinces)-1]
			r.Counts["provinces"]--
			// drop that province's regencies so only the count check can fail
			last := r.Provinces[len(r.Provinces)-1].Code
			kept := r.Regencies[:0]
			for _, g := range r.Regencies {
				if g.ProvinceCode <= last {
					kept = append(kept, g)
				}
			}
			r.Regencies = kept
			r.Counts["regencies"] = len(kept)
		},
		"regency count":         func(r *RegionsFile, _ *CountriesFile) { r.Regencies[0].IsActive = false },
		"bad type":              func(r *RegionsFile, _ *CountriesFile) { r.Regencies[0].Type = "KECAMATAN" },
		"type/name mismatch":    func(r *RegionsFile, _ *CountriesFile) { r.Regencies[0].Type = TypeKota },
		"untrimmed name":        func(r *RegionsFile, _ *CountriesFile) { r.Provinces[0].Name += " " },
		"missing version":       func(r *RegionsFile, _ *CountriesFile) { r.Version = "" },
		"header count mismatch": func(r *RegionsFile, _ *CountriesFile) { r.Counts["regencies"] = 1 },
		"bad replaced_by":       func(r *RegionsFile, _ *CountriesFile) { r.Regencies[0].ReplacedBy = []string{"00.00"} },
		"duplicate alpha2":      func(_ *RegionsFile, c *CountriesFile) { c.Countries[1].Alpha2 = c.Countries[0].Alpha2 },
		"bad numeric":           func(_ *RegionsFile, c *CountriesFile) { c.Countries[0].Numeric = "20" },
		"country count": func(_ *RegionsFile, c *CountriesFile) {
			c.Countries = c.Countries[1:]
			c.Counts["countries"]--
		},
		"empty country name": func(_ *RegionsFile, c *CountriesFile) { c.Countries[0].NameID = "" },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			rb, cb := mutate(t, fn)
			if _, err := Load(rb, cb); err == nil {
				t.Fatal("Load accepted invalid data")
			}
		})
	}

	t.Run("unknown field", func(t *testing.T) {
		rb := []byte(strings.Replace(string(regionsJSON), `"dataset"`, `"surprise": 1, "dataset"`, 1))
		if _, err := Load(rb, countriesJSON); err == nil {
			t.Fatal("Load accepted an unknown field")
		}
	})
	t.Run("unchanged data loads", func(t *testing.T) {
		rb, cb := mutate(t, func(*RegionsFile, *CountriesFile) {})
		if _, err := Load(rb, cb); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResolveLenient(t *testing.T) {
	d := mustDefault(t)
	if p, ok := d.ResolveProvince("  jawa   BARAT "); !ok || p.Code != "32" {
		t.Errorf("province lenient match failed: %+v %v", p, ok)
	}
	if _, ok := d.ResolveProvince("DKI Jakarta"); ok {
		t.Error("aliases are not accepted")
	}
	if c, name, ok := d.ResolveCountry("jepang"); !ok || c.Alpha2 != "JP" || name != "Jepang" {
		t.Errorf("Indonesian name: %+v %q %v", c, name, ok)
	}
	if c, name, ok := d.ResolveCountry("idn"); !ok || c.Alpha2 != "ID" || name != "Indonesia" {
		t.Errorf("alpha3: %+v %q %v", c, name, ok)
	}
	if got := d.SearchRegencies("BANDUNG", "", false); len(got) < 3 {
		t.Errorf("search bandung = %d results", len(got))
	}
	if got := d.SearchRegencies("bandung", "32", false); len(got) != 3 {
		// Kabupaten Bandung, Kabupaten Bandung Barat, Kota Bandung
		t.Errorf("search bandung in 32 = %d, want 3", len(got))
	}
	if got := d.SearchCountries("indo", false); len(got) != 1 || got[0].Alpha2 != "ID" {
		t.Errorf("search indo = %+v", got)
	}
}

func TestCheckRegion(t *testing.T) {
	d := mustDefault(t)
	str := func(s string) *string { return &s }

	t.Run("canonicalises", func(t *testing.T) {
		p, c := "jawa barat", " kota  bandung"
		fields := map[string]string{}
		d.CheckRegion(NameField{Key: "p", Value: &p}, NameField{Key: "c", Value: &c}, fields)
		if len(fields) != 0 || p != "Jawa Barat" || c != "Kota Bandung" {
			t.Fatalf("p=%q c=%q fields=%v", p, c, fields)
		}
	})
	t.Run("city in other province", func(t *testing.T) {
		p, c := "Jawa Timur", "Kota Bandung"
		fields := map[string]string{}
		d.CheckRegion(NameField{Key: "p", Value: &p}, NameField{Key: "c", Value: &c}, fields)
		if fields["c"] != apperrors.FieldMismatch {
			t.Fatalf("fields=%v", fields)
		}
	})
	t.Run("unknown names", func(t *testing.T) {
		p, c := "Jawa Barat", "Bandung"
		fields := map[string]string{}
		d.CheckRegion(NameField{Key: "p", Value: &p}, NameField{Key: "c", Value: &c}, fields)
		if fields["c"] != apperrors.FieldInvalidOption {
			t.Fatalf("fields=%v", fields)
		}
	})
	t.Run("legacy unchanged accepted", func(t *testing.T) {
		p, c := "DKI Jakarta ", "Jakarta"
		fields := map[string]string{}
		d.CheckRegion(NameField{Key: "p", Value: &p, Stored: str("DKI Jakarta")},
			NameField{Key: "c", Value: &c, Stored: str("Jakarta")}, fields)
		if len(fields) != 0 || p != "DKI Jakarta" || c != "Jakarta" {
			t.Fatalf("p=%q c=%q fields=%v", p, c, fields)
		}
	})
}
