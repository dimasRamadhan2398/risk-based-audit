package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"risk-service/models"
	"risk-service/pkg/masterclient"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Demo RCM rows tagged with COSO 2013 principles (decision Q12), for the
// internal control assessment walkthrough.
//
// Runs only from the full `./risk seed` (onboard-tenant.sh --with-demo-data
// and local dev). Never from `risk up`, `seed --catalog-only` or serve, so it
// never reaches prod (deploy-prod.sh refuses `risk seed`).
//
// Idempotent: a demo row is inserted only when no live row with its
// control_code and year exists, and existing rows are never modified. Rows
// link to the seeded demo risks by risk event where that risk exists.
//
// The department text uses the master-service demo department names
// (seeders.SeedDepartments: "Finance & Accounting", "Information
// Technology"). department_id is filled when master-service answers at seed
// time; otherwise it stays NULL and `./risk backfill-rcm-department --apply`
// links the rows on that demo stack.

type demoRCM struct {
	controlCode, riskCode, riskEvent, description, controlType, owner, department string
	ratings                                                                       [5]int
	component                                                                     string
	principles                                                                    []string
	notes                                                                         string
}

var demoCosoRCMs = []demoRCM{
	{"CTL-FA-01", "FIN-001", "Target pendapatan dan laba tidak tercapai",
		"Monthly management review of revenue and margin against budget, with documented variance follow-up.",
		"Detective", "Finance Manager", "Finance & Accounting", [5]int{4, 4, 4, 4, 3}, "CA", []string{"P10", "P12", "P16"},
		"Demo: review is documented; automated variance alerts planned."},
	{"CTL-FA-02", "FIN-002", "Target efisiensi biaya operasional dan umum tidak tercapai",
		"Budget owners approve every cost centre overrun above 5% before payment.",
		"Preventive", "Head of Finance", "Finance & Accounting", [5]int{4, 3, 3, 3, 2}, "CA", []string{"P10", "P12"},
		"Demo: approval matrix in place; evidence of approval incomplete."},
	{"CTL-FA-03", "FIN-004", "Terjadinya fraud",
		"Annual fraud risk assessment and whistleblowing channel reviewed by the audit committee.",
		"Detective", "Head of Internal Audit", "Finance & Accounting", [5]int{5, 4, 4, 4, 3}, "RA", []string{"P01", "P08", "P14"},
		"Demo: fraud risk assessment approved by the audit committee."},
	{"CTL-IT-01", "TEC-003", "Ancaman terhadap Cyber Security dan perlindungan data pribadi",
		"MFA on all privileged accounts and weekly vulnerability scanning with tracked remediation.",
		"Preventive", "IT Security Lead", "Information Technology", [5]int{5, 4, 4, 5, 4}, "CA", []string{"P11", "P13"},
		"Demo: scans automated; remediation SLA tracked."},
	{"CTL-IT-02", "TEC-005", "Implementasi teknologi dan digitalisasi tidak berhasil",
		"Change advisory board approves production changes; post-implementation reviews logged.",
		"Preventive", "IT Manager", "Information Technology", [5]int{3, 3, 2, 3, 3}, "MA", []string{"P11", "P16", "P17"},
		"Demo: post-implementation reviews are not always completed."},
}

func seedDemoRCMCoso(db *gorm.DB) error {
	if err := seedCosoCatalogue(db); err != nil {
		return err
	}
	var principles []models.CosoPrinciple
	if err := db.Where("framework = ?", models.CosoFramework2013).Find(&principles).Error; err != nil {
		return fmt.Errorf("load coso principles: %w", err)
	}
	principleID := map[string]uuid.UUID{}
	for _, p := range principles {
		principleID[p.Code] = p.ID
	}

	deptIDs := demoDepartmentIDs()
	year := 2026
	inserted := 0
	for _, d := range demoCosoRCMs {
		var n int64
		if err := db.Model(&models.RiskControlMatrix{}).
			Where("control_code = ? AND year = ?", d.controlCode, year).Count(&n).Error; err != nil {
			return fmt.Errorf("check demo rcm %s: %w", d.controlCode, err)
		}
		if n > 0 {
			continue
		}
		comp := d.component
		row := models.RiskControlMatrix{
			ID:                           uuid.New(),
			RiskCode:                     d.riskCode,
			RiskEvent:                    d.riskEvent,
			ControlCode:                  d.controlCode,
			ControlDescription:           d.description,
			ControlType:                  d.controlType,
			ControlOwner:                 d.owner,
			Department:                   d.department,
			Year:                         year,
			DesignEffectivenessWeight:    20,
			DesignEffectivenessRating:    d.ratings[0],
			OperatingEffectivenessWeight: 20,
			OperatingEffectivenessRating: d.ratings[1],
			CoverageCompletenessWeight:   20,
			CoverageCompletenessRating:   d.ratings[2],
			TimelinessWeight:             20,
			TimelinessRating:             d.ratings[3],
			AutomationMonitoringWeight:   20,
			AutomationMonitoringRating:   d.ratings[4],
			InherentRisk:                 20,
			ResidualRisk:                 8,
			Notes:                        d.notes,
			CosoComponentCode:            &comp,
		}
		if id, ok := deptIDs[normalizeDepartmentName(d.department)]; ok {
			row.DepartmentID = &id
		}
		var reg models.RiskRegister
		err := db.Where("risk_event = ?", d.riskEvent).Order("created_at").First(&reg).Error
		switch {
		case err == nil:
			row.RiskID = &reg.ID
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return fmt.Errorf("find demo risk %q: %w", d.riskEvent, err)
		}
		row.CalculateWeightedScore()

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Omit(clause.Associations).Create(&row).Error; err != nil {
				return err
			}
			for _, code := range d.principles {
				pid, ok := principleID[code]
				if !ok {
					return fmt.Errorf("principle %s missing from the catalogue", code)
				}
				tag := models.RcmCosoPrinciple{RcmID: row.ID, PrincipleID: pid, CreatedAt: time.Now()}
				if err := tx.Omit(clause.Associations).Create(&tag).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("seed demo rcm %s: %w", d.controlCode, err)
		}
		inserted++
	}
	if inserted > 0 {
		log.Printf("seeded %d demo COSO-tagged RCM rows (%d with a department_id)", inserted, countLinked(deptIDs))
	}
	return nil
}

func countLinked(ids map[string]uuid.UUID) int {
	n := 0
	for _, d := range demoCosoRCMs {
		if _, ok := ids[normalizeDepartmentName(d.department)]; ok {
			n++
		}
	}
	return n
}

// demoDepartmentIDs maps normalized active master department names to ids,
// best-effort: an empty map when master-service cannot be read (seeding goes
// on with department_id NULL).
func demoDepartmentIDs() map[string]uuid.UUID {
	out := map[string]uuid.UUID{}
	if cfg == nil {
		return out
	}
	auth, _, err := masterAuthorization(cfg.JWT.Secret, time.Now())
	if err != nil {
		log.Printf("demo RCM: no token for master-service, department_id left NULL: %v", err)
		return out
	}
	client := masterclient.NewDepartmentClient(cfg.MasterService.BaseURL(), 5*time.Second, time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	depts, err := client.ListDepartments(masterclient.WithAuthorization(ctx, auth))
	if err != nil {
		log.Printf("demo RCM: master-service departments unavailable, department_id left NULL (run backfill-rcm-department --apply later): %v", err)
		return out
	}
	dup := map[string]bool{}
	for _, d := range depts {
		if !d.IsActive {
			continue
		}
		key := normalizeDepartmentName(d.Name)
		if _, seen := out[key]; seen {
			dup[key] = true
		}
		out[key] = d.ID
	}
	for key := range dup {
		delete(out, key)
	}
	return out
}
