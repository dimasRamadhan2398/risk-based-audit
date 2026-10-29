package org_level

import (
	"context"
	"errors"
	"testing"

	"master-service/internal/orgtest"
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/kafka"
	levelRepo "master-service/repositories/org_level"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestPresetsUseGappedNumbering(t *testing.T) {
	cases := map[string][]int{
		PresetSimple:    {10},
		PresetCorporate: {10, 20},
		PresetBank:      {10, 20, 30, 40},
	}
	for preset, want := range cases {
		levels, err := BuildPreset(preset, nil)
		if err != nil {
			t.Fatalf("%s: %v", preset, err)
		}
		if len(levels) != len(want) {
			t.Fatalf("%s: %d levels", preset, len(levels))
		}
		for i, l := range levels {
			if l.Level != want[i] || l.NameKey == nil || !l.IsActive {
				t.Fatalf("%s level %d: %+v", preset, i, l)
			}
		}
	}
	bank, _ := BuildPreset(PresetBank, nil)
	if bank[0].Name != "Direktorat" || *bank[3].NameKey != models.OrgLevelKeySection {
		t.Fatalf("bank preset: %+v", bank)
	}

	custom, err := BuildPreset(PresetCustom, ParseLevelNames("Group, Unit ,Team"))
	if err != nil || len(custom) != 3 || custom[2].Level != 30 || custom[2].Name != "Team" || custom[0].NameKey != nil {
		t.Fatalf("custom: %+v %v", custom, err)
	}
	// Override keeps name_key only when the name is the preset default.
	over, err := BuildPreset(PresetCorporate, []string{"Division", "Departemen"})
	if err != nil || over[0].NameKey != nil || over[1].NameKey == nil {
		t.Fatalf("override: %+v %v", over, err)
	}
	for _, bad := range []struct {
		preset string
		names  []string
	}{
		{PresetCustom, nil},
		{PresetBank, []string{"A", "B"}},
		{"flat", nil},
		{PresetCustom, []string{"A", "a"}},
		{PresetCustom, []string{"1", "2", "3", "4", "5", "6", "7"}},
	} {
		if _, err := BuildPreset(bad.preset, bad.names); err == nil {
			t.Fatalf("BuildPreset(%s, %v) should fail", bad.preset, bad.names)
		}
	}
}

func TestSeedPresetIsIdempotentAndKeepsEdits(t *testing.T) {
	db := orgtest.DB(t)
	repo := levelRepo.New(db)
	res, err := SeedPreset(repo, PresetBank, nil)
	if err != nil || res.Created != 4 {
		t.Fatalf("first seed: %+v %v", res, err)
	}
	// Admin renames a level; a re-run must not undo it or add rows.
	if err := db.Model(&models.OrgLevel{}).Where("level = 20").Updates(map[string]any{"name": "Biro", "name_key": nil}).Error; err != nil {
		t.Fatal(err)
	}
	res, err = SeedPreset(repo, PresetSimple, nil)
	if err != nil || res.Created != 0 || res.Existing != 4 {
		t.Fatalf("second seed: %+v %v", res, err)
	}
	var l models.OrgLevel
	db.Where("level = 20").First(&l)
	if l.Name != "Biro" {
		t.Fatalf("edit overwritten: %+v", l)
	}
}

type env struct {
	db      *gorm.DB
	svc     *Service
	levels  []models.OrgLevel
	company uuid.UUID
}

func newEnv(t *testing.T) *env {
	db := orgtest.DB(t)
	return &env{db: db, svc: NewService(db, kafka.NewDomainPublisher(db, "test")), levels: orgtest.BankLevels(t, db), company: orgtest.Company(t, db)}
}

func (e *env) dept(t *testing.T, code string, level int, parent *uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if err := e.db.Exec(`INSERT INTO departments (id, department_code, department_name, level, company_id, is_active, parent_department_id)
		VALUES (?, ?, ?, ?, ?, 1, ?)`, id, code, "Unit "+code, level, e.company, parent).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

// keep returns the current set as LevelInputs, with fn applied.
func (e *env) keep(fn func(i int, in *LevelInput)) []LevelInput {
	var out []LevelInput
	for i, l := range e.levels {
		id := l.ID
		in := LevelInput{ID: &id, Level: l.Level, Name: l.Name}
		if fn != nil {
			fn(i, &in)
		}
		out = append(out, in)
	}
	return out
}

func hasViolation(pv *Preview, code string) bool {
	for _, v := range pv.Violations {
		if v.Code == code {
			return true
		}
	}
	return false
}

func TestPreviewIsADryRun(t *testing.T) {
	e := newEnv(t)
	dir := e.dept(t, "DIR", 10, nil)
	div := e.dept(t, "DIV", 20, &dir)
	orgtest.Employee(t, e.db, e.company, div)
	orgtest.Employee(t, e.db, e.company, div)

	pv, err := e.svc.Preview(context.Background(), ChangeRequest{Levels: e.keep(func(i int, in *LevelInput) {
		if i == 1 {
			in.Name = "Biro" // rename Divisi
		}
	})})
	if err != nil {
		t.Fatal(err)
	}
	if !pv.CanApply || len(pv.Diff.Renamed) != 1 || pv.Impact.DepartmentCount != 1 || pv.Impact.EmployeeCount != 2 {
		t.Fatalf("preview: %+v", pv)
	}
	if pv.Impact.LettersModified || len(pv.Impact.UnitUsageIDs) != 1 || pv.Impact.UnitUsageIDs[0] != div {
		t.Fatalf("impact: %+v", pv.Impact)
	}
	if pv.BaseVersion != Version(e.levels) {
		t.Fatal("base_version mismatch")
	}
	// Nothing was written.
	var l models.OrgLevel
	e.db.Where("level = 20").First(&l)
	if l.Name != "Divisi" || len(orgtest.Outbox(t, e.db)) != 0 {
		t.Fatal("preview wrote data")
	}
	var logs int64
	e.db.Model(&models.OrgChangeLog{}).Count(&logs)
	if logs != 0 {
		t.Fatal("preview wrote a change log")
	}
}

func TestPreviewViolations(t *testing.T) {
	e := newEnv(t)
	dir := e.dept(t, "DIR", 10, nil)
	e.dept(t, "DIV", 20, &dir)

	// Inverting two used levels: Direktorat 10 -> 25, Divisi 20 stays: child <= parent.
	pv, _ := e.svc.Preview(context.Background(), ChangeRequest{Levels: e.keep(func(i int, in *LevelInput) {
		if i == 0 {
			in.Level = 25
		}
	})})
	if pv.CanApply || !hasViolation(pv, VLevelOrderConflict) {
		t.Fatalf("inversion not reported: %+v", pv.Violations)
	}
	// Deleting a level in use.
	pv, _ = e.svc.Preview(context.Background(), ChangeRequest{Levels: e.keep(nil)[1:], DeleteIDs: []uuid.UUID{e.levels[0].ID}})
	if pv.CanApply || !hasViolation(pv, VLevelInUse) {
		t.Fatalf("in-use delete not reported: %+v", pv.Violations)
	}
	// Seven levels, duplicates, out of range, missing level.
	levels := e.keep(nil)
	levels = append(levels, LevelInput{Level: 50, Name: "Unit"}, LevelInput{Level: 60, Name: "Tim"}, LevelInput{Level: 1000, Name: "seksi"})
	pv, _ = e.svc.Preview(context.Background(), ChangeRequest{Levels: levels})
	for _, code := range []string{VTooManyLevels, VLevelOutOfRange, VDuplicateName} {
		if !hasViolation(pv, code) {
			t.Fatalf("missing %s: %+v", code, pv.Violations)
		}
	}
	pv, _ = e.svc.Preview(context.Background(), ChangeRequest{Levels: e.keep(nil)[:3]})
	if !hasViolation(pv, VLevelMissing) {
		t.Fatalf("missing level not reported: %+v", pv.Violations)
	}
	pv, _ = e.svc.Preview(context.Background(), ChangeRequest{Levels: e.keep(func(i int, in *LevelInput) {
		if i == 1 {
			in.Level = 10
		}
	})})
	if !hasViolation(pv, VDuplicateLevel) {
		t.Fatalf("duplicate number not reported: %+v", pv.Violations)
	}
}

func TestApplyRenumberMovesDepartmentsAndLogs(t *testing.T) {
	e := newEnv(t)
	dir := e.dept(t, "DIR", 10, nil)
	div := e.dept(t, "DIV", 20, &dir)
	dep := e.dept(t, "DEP", 30, &div)
	e.dept(t, "SEC", 40, &dep)

	req := ChangeRequest{Levels: e.keep(func(i int, in *LevelInput) {
		in.Level = (i + 1) * 100
		if i == 0 {
			in.Name = "Direktorat Utama" // rename + renumber: key is cleared
		}
	})}
	pv, _ := e.svc.Preview(context.Background(), req)
	res, err := e.svc.Apply(context.Background(), ApplyRequest{ChangeRequest: req, BaseVersion: pv.BaseVersion},
		Actor{UserID: "u1", Username: "admin", Roles: []string{"ADMIN"}, IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(res.Levels) != 4 || res.Levels[0].Level != 100 || res.Levels[3].Level != 400 || res.Levels[0].NameKey != nil {
		t.Fatalf("levels after apply: %+v", res.Levels)
	}
	if res.Levels[1].NameKey == nil || *res.Levels[1].NameKey != models.OrgLevelKeyDivision {
		t.Fatal("renumber alone must keep the name key")
	}
	var levels []int
	e.db.Model(&models.Department{}).Order("level").Pluck("level", &levels)
	if len(levels) != 4 || levels[0] != 100 || levels[3] != 400 {
		t.Fatalf("departments not moved: %v", levels)
	}
	if res.EventsQueued != 4 || len(orgtest.Outbox(t, e.db)) != 4 {
		t.Fatalf("events: %d queued", res.EventsQueued)
	}
	var log models.OrgChangeLog
	if err := e.db.First(&log).Error; err != nil {
		t.Fatalf("change log missing: %v", err)
	}
	if log.Action != models.OrgChangeLogActionApply || log.ActorUsername != "admin" || log.Before == "" || log.After == "" || log.Impact == "" || log.BaseVersion != pv.BaseVersion {
		t.Fatalf("change log: %+v", log)
	}
	// Same base_version again: someone changed the levels -> 409.
	_, err = e.svc.Apply(context.Background(), ApplyRequest{ChangeRequest: req, BaseVersion: pv.BaseVersion}, Actor{})
	if appErr, ok := apperrors.As(err); !ok || appErr.Code != apperrors.CodeOrgLevelsVersionConflict {
		t.Fatalf("want version conflict, got %v", err)
	}
}

func TestApplySwapsNamesAndInsertsBetween(t *testing.T) {
	e := newEnv(t)
	req := ChangeRequest{Levels: append(e.keep(func(i int, in *LevelInput) {
		switch i {
		case 0:
			in.Name = "Divisi" // swap with level 20
		case 1:
			in.Name = "Direktorat"
		}
	}), LevelInput{Level: 25, Name: "Biro"})}
	pv, _ := e.svc.Preview(context.Background(), req)
	if !pv.CanApply {
		t.Fatalf("violations: %+v", pv.Violations)
	}
	res, err := e.svc.Apply(context.Background(), ApplyRequest{ChangeRequest: req, BaseVersion: pv.BaseVersion}, Actor{})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(res.Levels) != 5 || res.Levels[0].Name != "Divisi" || res.Levels[2].Level != 25 {
		t.Fatalf("levels: %+v", res.Levels)
	}
}

func TestApplyRejectsViolationsWithoutWriting(t *testing.T) {
	e := newEnv(t)
	dir := e.dept(t, "DIR", 10, nil)
	e.dept(t, "DIV", 20, &dir)
	req := ChangeRequest{Levels: e.keep(func(i int, in *LevelInput) {
		if i == 0 {
			in.Level = 25
		}
	})}
	_, err := e.svc.Apply(context.Background(), ApplyRequest{ChangeRequest: req, BaseVersion: Version(e.levels)}, Actor{})
	var verr *ViolationsError
	if !errors.As(err, &verr) || verr.StatusCode != 422 || len(verr.Violations) == 0 {
		t.Fatalf("want 422 with violations, got %v", err)
	}
	var l models.OrgLevel
	e.db.Where("id = ?", e.levels[0].ID).First(&l)
	if l.Level != 10 {
		t.Fatal("rejected apply changed data")
	}
}

type failingPublisher struct{}

func (failingPublisher) PublishDepartmentOrgTx(context.Context, *gorm.DB, string, *models.Department) error {
	return errors.New("outbox down")
}

func TestApplyRollsBackWhenAnyStepFails(t *testing.T) {
	e := newEnv(t)
	e.svc = NewService(e.db, failingPublisher{})
	e.dept(t, "DIR", 10, nil)
	req := ChangeRequest{Levels: e.keep(func(i int, in *LevelInput) { in.Level = (i + 1) * 100 })}
	_, err := e.svc.Apply(context.Background(), ApplyRequest{ChangeRequest: req, BaseVersion: Version(e.levels)}, Actor{})
	if err == nil {
		t.Fatal("apply should fail")
	}
	var levels []int
	e.db.Model(&models.OrgLevel{}).Order("level").Pluck("level", &levels)
	var deptLevel int
	e.db.Model(&models.Department{}).Pluck("level", &deptLevel)
	var logs int64
	e.db.Model(&models.OrgChangeLog{}).Count(&logs)
	if levels[0] != 10 || deptLevel != 10 || logs != 0 {
		t.Fatalf("not rolled back: levels=%v dept=%d logs=%d", levels, deptLevel, logs)
	}
}
