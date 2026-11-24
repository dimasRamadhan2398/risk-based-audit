package main

import (
	"fmt"
	"strings"

	"audit-service/models"

	"gorm.io/gorm"
)

// CSUL "Strategic Plan Audit Internal" (Renstra Audit Internal 5 tahun).
//
// The plan is stored one row per KPI in strategic_plans; the rows of one
// Strategic Objective (SO) share the SO code (SO-IA04 has three KPIs). Year 1..5
// of the template are the real years csulYears. The vision, mission and goal
// (Visi/Misi/Tujuan) live in master-service (SeedVisionMissionGoals).

var csulYears = []string{"2026", "2027", "2028", "2029", "2030"}

const csulGoalID = "G-001"

type csulKPI struct {
	No         int
	SOCode     string
	SO         string
	KPI        string
	Unit       string
	HibHig     string
	Targets    [5]string // Year 1..5; "" = not set
	TargetNote map[int]string
	Initiative string
	Years      [5]string // "√" carried out, "-" not
}

const (
	yes = "√"
	no  = "-"
)

func csulKPIs() []csulKPI {
	all5 := [5]string{yes, yes, yes, yes, yes}
	return []csulKPI{
		{1, "SO-IA01", "Memastikan kepatuhan terhadap kebijakan dan regulasi perusahaan dan institusi terkait",
			"Tingkat kepatuhan perusahaan terhadap kebijakan dan regulasi yang berlaku", "%", "HIG",
			[5]string{"80.0", "82.0", "84.0", "85.0", "86.0"}, nil,
			"Program sosialisasi kepatuhan perusahaan", [5]string{yes, yes, yes, no, no}},
		{2, "SO-IA02", "Melakukan mitigasi terhadap risiko-risiko prioritas operasional dan keuangan",
			"Tercapainya tingkat risiko residual sesuai risk appetite statement (RAS) perusahaan", "Score", "HIG",
			[5]string{"0.80", "0.83", "0.85", "0.87", "0.90"}, nil,
			"Melakukan evaluasi secara berkelanjutan terhadap ketepatan dan kelengkapan kontrol risiko", all5},
		{3, "SO-IA03", "Memastikan tingkat efektivitas kontrol risiko prioritas perusahaan",
			"Tingkat efektivitas kontrol risiko prioritas perusahaan", "%", "HIG",
			[5]string{"56", "60", "64", "68", "72"}, nil,
			"Melakukan evaluasi secara berkelanjutan terhadap kedisiplinan implementasi kontrol risiko", all5},
		{4, "SO-IA04", "Meningkatkan efisiensi dan efektivitas pelaksanaan audit internal",
			"Tingkat efisiensi biaya audit internal", "%", "HIB",
			[5]string{"62.7", "62.5", "62.3", "62.1", "62.0"}, nil,
			"Program efisiensi biaya audit internal", all5},
		{5, "SO-IA04", "Meningkatkan efisiensi dan efektivitas pelaksanaan audit internal",
			"Realisasi rencana audit", "%", "HIG",
			[5]string{"100", "100", "100", "100", "100"}, nil,
			"Program pelaksanaan audit internal sesuai rencana", all5},
		{6, "SO-IA04", "Meningkatkan efisiensi dan efektivitas pelaksanaan audit internal",
			"Tingkat realisasi rencana tindak lanjut hasil audit", "%", "HIG",
			[5]string{"100", "100", "100", "100", "100"}, nil,
			"Program pemantauan pelaksanaan rekomendasi rencana tindak lanjut dari hasil audit internal", all5},
		{7, "SO-IA05", "Melakukan asesmen ulang terhadap risiko-risiko prioritas perusahaan",
			"Ketepatan risiko-risiko prioritas", "%", "HIG",
			[5]string{"70", "75", "80", "85", "88"}, nil,
			"Program evaluasi risiko-risiko prioritas perusahaan berdasarkan pencapaian target KPI perusahaan", all5},
		{8, "SO-IA06", "Menjamin kualitas asurans audit internal (penilaian mandiri secara periodik oleh Audit Internal) dalam rangka QAIP",
			"Terlaksananya penilaian mandiri kualitas fungsi Audit Internal secara periodik oleh Audit Internal dalam rangka QAIP", "%", "HIG",
			[5]string{"100", "100", "100", "100", "100"}, nil,
			"Program penilaian mandiri kualitas fungsi Audit Internal secara periodik", [5]string{yes, yes, no, no, no}},
		{9, "SO-IA07", "Menjamin kualitas asurans Internal Audit hasil penilaian eksternal (SAIV atau QAR) sesuai GIAS",
			"Tingkat kesesuaian (conformance) audit internal terhadap GIAS", "%", "HIG",
			[5]string{"", "", "", "75.0", ""}, map[int]string{3: "PC"},
			"Program sAIV atau External Quality Assurance Review", [5]string{no, no, no, yes, no}},
		{10, "SO-IA08", "Melaksanakan peran strategic partner",
			"Jumlah layanan konsultansi yang dilaksanakan", "Amount", "HIG",
			[5]string{"2", "3", "4", "5", "6"}, nil,
			"Program layanan konsultansi kepada manajemen perusahaan", all5},
		{11, "SO-IA09", "Mengembangkan kompetensi internal auditor",
			"Jumlah auditor internal yang memiliki sertifikasi auditor internal", "Amount", "HIG",
			[5]string{"5", "7", "10", "13", "15"}, nil,
			"Program pendidikan dan pelatihan auditor internal", all5},
	}
}

// csulStrategicPlans converts the template into strategic_plans rows.
func csulStrategicPlans() []models.StrategicPlan {
	kpis := csulKPIs()
	plans := make([]models.StrategicPlan, 0, len(kpis))
	for _, k := range kpis {
		targets := map[string]string{}
		years := map[string]string{}
		notes := map[string]string{}
		for i, y := range csulYears {
			targets[y] = k.Targets[i]
			years[y] = k.Years[i]
			if n, ok := k.TargetNote[i]; ok {
				notes[y] = n
			}
		}
		plans = append(plans, models.StrategicPlan{
			Code:               k.SOCode,
			GoalID:             csulGoalID,
			StrategicObjective: k.SO,
			KPI:                k.KPI,
			Unit:               k.Unit,
			HibHig:             k.HibHig,
			PeriodType:         "Yearly",
			SelectedPeriod:     csulYears[0],
			YearStart:          2026,
			YearEnd:            2030,
			KPITargets:         targets,
			KPIActuals:         map[string]string{},
			Target:             k.Targets[0],
			Initiative:         k.Initiative,
			InitiativeYears:    years,
			TargetNotes:        notes,
		})
	}
	return plans
}

// seedStrategicPlansCSUL upserts the CSUL plan keyed by (code, kpi), so it can
// be re-run safely. With replaceGeneric, rows of the generic demo seed that use
// the codes SO-IA01..SO-IA09 but carry a different KPI are soft-deleted so the
// two sets do not show side by side.
func seedStrategicPlansCSUL(db *gorm.DB, replaceGeneric bool) error {
	plans := csulStrategicPlans()
	keep := map[string]bool{}
	codes := map[string]bool{}
	for _, p := range plans {
		keep[p.Code+"|"+p.KPI] = true
		codes[p.Code] = true
	}

	if replaceGeneric {
		var rows []models.StrategicPlan
		codeList := make([]string, 0, len(codes))
		for c := range codes {
			codeList = append(codeList, c)
		}
		if err := db.Where("code IN ?", codeList).Find(&rows).Error; err != nil {
			return err
		}
		for _, r := range rows {
			if !keep[r.Code+"|"+r.KPI] {
				if err := db.Delete(&models.StrategicPlan{}, "id = ?", r.ID).Error; err != nil {
					return err
				}
			}
		}
	}

	for i := range plans {
		var existing models.StrategicPlan
		err := db.Unscoped().Where("code = ? AND kpi = ?", plans[i].Code, plans[i].KPI).First(&existing).Error
		switch {
		case err == gorm.ErrRecordNotFound:
			if err := db.Create(&plans[i]).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			plans[i].ID = existing.ID
			plans[i].CreatedAt = existing.CreatedAt
			plans[i].KPIActuals = existing.KPIActuals // keep realisation entered by users
			plans[i].Actual = existing.Actual
			plans[i].Calculation = existing.Calculation
			plans[i].Status = existing.Status
			plans[i].Category = existing.Category
			if err := db.Unscoped().Save(&plans[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// csulSummary is a one-line count used in log output.
func csulSummary(plans []models.StrategicPlan) string {
	so := map[string]bool{}
	init := 0
	for _, p := range plans {
		so[p.Code] = true
		if strings.TrimSpace(p.Initiative) != "" {
			init++
		}
	}
	return fmt.Sprintf("%d SO, %d KPI, %d initiative", len(so), len(plans), init)
}
