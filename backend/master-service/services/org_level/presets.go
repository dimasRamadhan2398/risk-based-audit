package org_level

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"master-service/models"
	"master-service/pkg/orgmode"
	levelRepo "master-service/repositories/org_level"

	"github.com/google/uuid"
)

// Preset names accepted by `master seed-org-levels --preset` and
// onboard-tenant.sh --org-preset (docs/org-structure-model-plan.md 4.2).
const (
	PresetSimple    = "simple"
	PresetCorporate = "corporate"
	PresetBank      = "bank"
	PresetCustom    = "custom"
)

// LevelStep is the gap between preset level numbers (10, 20, 30, ...), so a
// level can later be inserted between two others without renumbering.
const LevelStep = 10

// MaxNameLength is the org_levels.name column size.
const MaxNameLength = 50

type presetLevel struct {
	name, key string
}

var presets = map[string][]presetLevel{
	PresetSimple: {{"Departemen", models.OrgLevelKeyDepartment}},
	PresetCorporate: {
		{"Divisi", models.OrgLevelKeyDivision},
		{"Departemen", models.OrgLevelKeyDepartment},
	},
	PresetBank: {
		{"Direktorat", models.OrgLevelKeyDirectorate},
		{"Divisi", models.OrgLevelKeyDivision},
		{"Departemen", models.OrgLevelKeyDepartment},
		{"Seksi", models.OrgLevelKeySection},
	},
}

// ParseLevelNames splits a --levels value ("A,B,C"), trimming blanks.
func ParseLevelNames(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if n := strings.TrimSpace(p); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// BuildPreset returns the levels of a preset, numbered 10, 20, 30, ...
// names, when given, override the stored names (required for custom); a preset
// level keeps its name_key only when its name is the preset default.
func BuildPreset(preset string, names []string) ([]models.OrgLevel, error) {
	preset = strings.ToLower(strings.TrimSpace(preset))
	var levels []presetLevel
	switch preset {
	case PresetCustom:
		if len(names) == 0 {
			return nil, fmt.Errorf("preset custom needs level names (--levels \"A,B,C\")")
		}
		for _, n := range names {
			levels = append(levels, presetLevel{name: n})
		}
	case PresetSimple, PresetCorporate, PresetBank:
		base := presets[preset]
		levels = append(levels, base...)
		if len(names) > 0 {
			if len(names) != len(base) {
				return nil, fmt.Errorf("preset %s has %d level(s), got %d name(s)", preset, len(base), len(names))
			}
			for i, n := range names {
				key := ""
				if strings.EqualFold(n, base[i].name) {
					key = base[i].key
				}
				levels[i] = presetLevel{name: n, key: key}
			}
		}
	default:
		return nil, fmt.Errorf("unknown preset %q (want simple, corporate, bank or custom)", preset)
	}
	if len(levels) > orgmode.MaxLevels {
		return nil, fmt.Errorf("at most %d levels, got %d", orgmode.MaxLevels, len(levels))
	}
	seen := map[string]bool{}
	out := make([]models.OrgLevel, 0, len(levels))
	for i, l := range levels {
		name := strings.TrimSpace(l.name)
		if name == "" || utf8.RuneCountInString(name) > MaxNameLength {
			return nil, fmt.Errorf("level name %q must be 1-%d characters", l.name, MaxNameLength)
		}
		if seen[strings.ToLower(name)] {
			return nil, fmt.Errorf("duplicate level name %q", name)
		}
		seen[strings.ToLower(name)] = true
		var key *string
		if l.key != "" {
			k := l.key
			key = &k
		}
		out = append(out, models.OrgLevel{
			ID:       uuid.New(),
			Level:    (i + 1) * LevelStep,
			Name:     name,
			NameKey:  key,
			IsActive: true,
		})
	}
	return out, nil
}

// SeedResult reports what SeedPreset did.
type SeedResult struct {
	Created  int
	Existing int64 // live levels already present (seeding skipped)
}

// SeedPreset seeds the preset when org_levels has no live row. It never
// touches existing levels, so admin edits survive a re-run (idempotent).
func SeedPreset(repo levelRepo.Repository, preset string, names []string) (SeedResult, error) {
	levels, err := BuildPreset(preset, names)
	if err != nil {
		return SeedResult{}, err
	}
	var res SeedResult
	err = repo.Transaction(func(tx levelRepo.Repository) error {
		if _, err := tx.LockLive(); err != nil {
			return err
		}
		n, err := tx.CountLive()
		if err != nil {
			return err
		}
		if n > 0 {
			res.Existing = n
			return nil
		}
		for i := range levels {
			if err := tx.Create(&levels[i]); err != nil {
				return err
			}
			res.Created++
		}
		return nil
	})
	return res, err
}
