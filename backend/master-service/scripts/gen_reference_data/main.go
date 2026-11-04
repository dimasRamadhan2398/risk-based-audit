// Command gen_reference_data turns the raw upstream sources into the vendored
// files embedded by pkg/reference:
//
//	pkg/reference/data/regions.json   (provinces + kabupaten/kota, Kemendagri codes)
//	pkg/reference/data/countries.json (ISO 3166-1 countries, CLDR names en/id)
//
// The inputs are read as plain text/JSON only; nothing in them is executed.
// Counts are checked and the output is re-validated with reference.Load before
// anything is written, so a bad source never replaces good data.
//
// Usage (from backend/master-service):
//
//	go run ./scripts/gen_reference_data \
//	  -wilayah /path/to/wilayah.sql -wilayah-commit <sha> \
//	  -cldr-en /path/to/main/en/territories.json \
//	  -cldr-id /path/to/main/id/territories.json \
//	  -cldr-codemappings /path/to/supplemental/codeMappings.json \
//	  -cldr-version 49.0.0 \
//	  [-generated-at 2026-10-31T00:00:00Z] [-out pkg/reference/data]
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"master-service/pkg/reference"
)

const generatorName = "backend/master-service/scripts/gen_reference_data"

func main() {
	wilayah := flag.String("wilayah", "", "path to cahyadsn/wilayah db/wilayah.sql")
	wilayahCommit := flag.String("wilayah-commit", "", "upstream commit the wilayah.sql was taken from")
	cldrEN := flag.String("cldr-en", "", "path to cldr-localenames-full main/en/territories.json")
	cldrID := flag.String("cldr-id", "", "path to cldr-localenames-full main/id/territories.json")
	cldrCM := flag.String("cldr-codemappings", "", "path to cldr-core supplemental/codeMappings.json")
	cldrVersion := flag.String("cldr-version", "", "CLDR release, e.g. 49.0.0")
	generatedAt := flag.String("generated-at", "", "RFC3339 timestamp to record (default: now, UTC)")
	out := flag.String("out", "pkg/reference/data", "output directory")
	flag.Parse()

	for name, v := range map[string]string{
		"-wilayah": *wilayah, "-wilayah-commit": *wilayahCommit, "-cldr-en": *cldrEN,
		"-cldr-id": *cldrID, "-cldr-codemappings": *cldrCM, "-cldr-version": *cldrVersion,
	} {
		if strings.TrimSpace(v) == "" {
			fatalf("%s is required", name)
		}
	}
	if *generatedAt == "" {
		*generatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	regions, err := buildRegions(*wilayah, *wilayahCommit, *generatedAt)
	if err != nil {
		fatalf("regions: %v", err)
	}
	countries, err := buildCountries(*cldrEN, *cldrID, *cldrCM, *cldrVersion, *generatedAt)
	if err != nil {
		fatalf("countries: %v", err)
	}

	rb, err := marshal(regions)
	if err != nil {
		fatalf("marshal regions: %v", err)
	}
	cb, err := marshal(countries)
	if err != nil {
		fatalf("marshal countries: %v", err)
	}
	if _, err := reference.Load(rb, cb); err != nil {
		fatalf("generated data does not validate: %v", err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fatalf("%v", err)
	}
	if err := os.WriteFile(filepath.Join(*out, "regions.json"), rb, 0o644); err != nil {
		fatalf("%v", err)
	}
	if err := os.WriteFile(filepath.Join(*out, "countries.json"), cb, 0o644); err != nil {
		fatalf("%v", err)
	}
	fmt.Printf("regions   %s: %d provinces, %d regencies (%d kabupaten, %d kota)\n",
		regions.Version, len(regions.Provinces), len(regions.Regencies), regions.Counts["kabupaten"], regions.Counts["kota"])
	fmt.Printf("countries %s: %d countries\n", countries.Version, len(countries.Countries))
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "gen_reference_data: "+format+"\n", args...)
	os.Exit(1)
}

func marshal(v interface{}) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func fileSHA256(path string) (string, []byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), b, nil
}

// contentVersion is label + the first 12 hex digits of the data's hash, so the
// version (and the ETag) changes exactly when the served content changes.
func contentVersion(label string, data ...interface{}) (string, error) {
	h := sha256.New()
	for _, d := range data {
		b, err := json.Marshal(d)
		if err != nil {
			return "", err
		}
		h.Write(b)
	}
	return label + "-" + hex.EncodeToString(h.Sum(nil))[:12], nil
}

// --- regions -----------------------------------------------------------------

// One data row of wilayah.sql: ('32.73','Kota Bandung'), with a quote inside
// a name written as two single quotes (SQL escaping).
var (
	rowRe      = regexp.MustCompile(`^\('([0-9.]+)','((?:[^']|'')*)'\)[,;]\s*$`)
	codePrefix = regexp.MustCompile(`^\('([0-9]{2}(?:\.[0-9]{2})?)',`)
)

func buildRegions(path, commit, generatedAt string) (*reference.RegionsFile, error) {
	sum, raw, err := fileSHA256(path)
	if err != nil {
		return nil, err
	}

	provinces := []reference.Province{}
	regencies := []reference.Regency{}
	seen := map[string]bool{}
	kab, kota := 0, 0

	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for lineNo := 1; sc.Scan(); lineNo++ {
		line := strings.TrimRight(sc.Text(), "\r")
		// Only rows whose code has one or two segments are wanted. Anything that
		// starts like such a row but does not parse is a hard error, so a format
		// change upstream cannot silently drop rows.
		pm := codePrefix.FindStringSubmatch(line)
		if pm == nil {
			continue
		}
		m := rowRe.FindStringSubmatch(line)
		if m == nil {
			return nil, fmt.Errorf("line %d: unparseable row %q", lineNo, line)
		}
		code := m[1]
		name := strings.Join(strings.Fields(strings.ReplaceAll(m[2], "''", "'")), " ")
		if name == "" {
			return nil, fmt.Errorf("line %d: empty name for %s", lineNo, code)
		}
		if seen[code] {
			return nil, fmt.Errorf("line %d: duplicate code %s", lineNo, code)
		}
		seen[code] = true

		if !strings.Contains(code, ".") {
			provinces = append(provinces, reference.Province{Code: code, Name: name, IsActive: true, ReplacedBy: []string{}})
			continue
		}
		typ, err := regencyType(code, name)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if typ == reference.TypeKabupaten {
			kab++
		} else {
			kota++
		}
		regencies = append(regencies, reference.Regency{
			Code: code, ProvinceCode: code[:2], Type: typ, Name: name, IsActive: true, ReplacedBy: []string{},
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	if len(provinces) != reference.ExpectedProvinces || kab != reference.ExpectedKabupaten || kota != reference.ExpectedKota {
		return nil, fmt.Errorf("counts provinces=%d kabupaten=%d kota=%d, want %d/%d/%d",
			len(provinces), kab, kota, reference.ExpectedProvinces, reference.ExpectedKabupaten, reference.ExpectedKota)
	}
	sort.Slice(provinces, func(i, j int) bool { return provinces[i].Code < provinces[j].Code })
	sort.Slice(regencies, func(i, j int) bool { return regencies[i].Code < regencies[j].Code })

	short := commit
	if len(short) > 7 {
		short = short[:7]
	}
	version, err := contentVersion("kemendagri-2025-"+short, provinces, regencies)
	if err != nil {
		return nil, err
	}
	return &reference.RegionsFile{
		Header: reference.Header{
			Dataset:    "id-regions",
			Version:    version,
			LegalBasis: "Kepmendagri 300.2.2-2138/2025 jo. 300.2.2-2430/2025",
			Sources: []reference.Source{{
				Name:    "cahyadsn/wilayah db/wilayah.sql",
				URL:     "https://github.com/cahyadsn/wilayah/blob/" + commit + "/db/wilayah.sql",
				Commit:  commit,
				License: "MIT",
				SHA256:  sum,
			}},
			GeneratedAt: generatedAt,
			Generator:   generatorName,
		},
		Counts: map[string]int{
			"provinces": len(provinces), "regencies": len(regencies), "kabupaten": kab, "kota": kota,
		},
		Provinces: provinces,
		Regencies: regencies,
	}, nil
}

// regencyType derives KABUPATEN/KOTA from the official name and cross-checks
// it with the Kemendagri numbering (01-69 kabupaten, 71-99 kota). DKI's
// "Kabupaten Administrasi"/"Kota Administrasi" keep their names.
func regencyType(code, name string) (string, error) {
	var byName string
	switch {
	case strings.HasPrefix(name, "Kabupaten "):
		byName = reference.TypeKabupaten
	case strings.HasPrefix(name, "Kota "):
		byName = reference.TypeKota
	default:
		return "", fmt.Errorf("%s %q: name starts with neither Kabupaten nor Kota", code, name)
	}
	byCode := reference.TypeKabupaten
	if code[3:] >= "71" {
		byCode = reference.TypeKota
	}
	if byName != byCode {
		return "", fmt.Errorf("%s %q: name says %s, code says %s", code, name, byName, byCode)
	}
	return byName, nil
}

// --- countries ---------------------------------------------------------------

// Excluded even though CLDR lists alpha-3 and numeric codes for them.
var (
	deprecatedCodes = map[string]bool{
		"AN": true, "BU": true, "CS": true, "DD": true, "FX": true, "NT": true,
		"SU": true, "TP": true, "YD": true, "YU": true, "ZR": true,
	}
	nonCountryCodes = map[string]bool{"EU": true}
	// Indonesian display names that differ from CLDR on purpose.
	nameIDOverrides = map[string]string{"PS": "Palestina"}
)

func userAssigned(code string) bool {
	if code == "AA" || code == "ZZ" || strings.HasPrefix(code, "X") {
		return true
	}
	return code[0] == 'Q' && code[1] >= 'M' && code[1] <= 'Z'
}

type cldrTerritories struct {
	Main map[string]struct {
		LocaleDisplayNames struct {
			Territories map[string]string `json:"territories"`
		} `json:"localeDisplayNames"`
	} `json:"main"`
}

type cldrCodeMappings struct {
	Supplemental struct {
		CodeMappings map[string]struct {
			Numeric string `json:"_numeric"`
			Alpha3  string `json:"_alpha3"`
		} `json:"codeMappings"`
	} `json:"supplemental"`
}

func readTerritories(path, locale string) (map[string]string, string, error) {
	sum, raw, err := fileSHA256(path)
	if err != nil {
		return nil, "", err
	}
	var t cldrTerritories
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, "", fmt.Errorf("%s: %w", path, err)
	}
	m, ok := t.Main[locale]
	if !ok || len(m.LocaleDisplayNames.Territories) == 0 {
		return nil, "", fmt.Errorf("%s: no territories for locale %q", path, locale)
	}
	return m.LocaleDisplayNames.Territories, sum, nil
}

func buildCountries(enPath, idPath, cmPath, cldrVersion, generatedAt string) (*reference.CountriesFile, error) {
	en, enSum, err := readTerritories(enPath, "en")
	if err != nil {
		return nil, err
	}
	id, idSum, err := readTerritories(idPath, "id")
	if err != nil {
		return nil, err
	}
	cmSum, raw, err := fileSHA256(cmPath)
	if err != nil {
		return nil, err
	}
	var cm cldrCodeMappings
	if err := json.Unmarshal(raw, &cm); err != nil {
		return nil, fmt.Errorf("%s: %w", cmPath, err)
	}

	countries := []reference.Country{}
	for code, m := range cm.Supplemental.CodeMappings {
		if len(code) != 2 || m.Alpha3 == "" || m.Numeric == "" {
			continue
		}
		if userAssigned(code) || deprecatedCodes[code] || nonCountryCodes[code] {
			continue
		}
		nameEN := strings.TrimSpace(en[code])
		nameID := strings.TrimSpace(id[code])
		if o, ok := nameIDOverrides[code]; ok {
			nameID = o
		}
		if nameEN == "" || nameID == "" {
			return nil, fmt.Errorf("%s: missing en or id name", code)
		}
		countries = append(countries, reference.Country{
			Alpha2: code, Alpha3: m.Alpha3, Numeric: m.Numeric, NameEN: nameEN, NameID: nameID, IsActive: true,
		})
	}
	if len(countries) != reference.ExpectedCountries {
		return nil, fmt.Errorf("got %d countries, want %d", len(countries), reference.ExpectedCountries)
	}
	sort.Slice(countries, func(i, j int) bool { return countries[i].Alpha2 < countries[j].Alpha2 })

	version, err := contentVersion("cldr-"+cldrVersion, countries)
	if err != nil {
		return nil, err
	}
	base := "https://github.com/unicode-org/cldr-json/blob/" + cldrVersion + "/cldr-json/"
	license := "Unicode-3.0"
	return &reference.CountriesFile{
		Header: reference.Header{
			Dataset:    "countries",
			Version:    version,
			LegalBasis: "Unicode CLDR " + cldrVersion + " (ISO 3166-1 codes)",
			Sources: []reference.Source{
				{Name: "cldr-localenames-full main/en/territories.json", URL: base + "cldr-localenames-full/main/en/territories.json", Version: cldrVersion, License: license, SHA256: enSum},
				{Name: "cldr-localenames-full main/id/territories.json", URL: base + "cldr-localenames-full/main/id/territories.json", Version: cldrVersion, License: license, SHA256: idSum},
				{Name: "cldr-core supplemental/codeMappings.json", URL: base + "cldr-core/supplemental/codeMappings.json", Version: cldrVersion, License: license, SHA256: cmSum},
			},
			GeneratedAt: generatedAt,
			Generator:   generatorName,
		},
		Counts:    map[string]int{"countries": len(countries)},
		Countries: countries,
	}, nil
}
