package services

import (
	"errors"
	"fmt"
	"log"

	"risk-service/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedCosoCatalogue inserts the COSO 2013 components and principles that are
// missing (docs/coso-2013-component-assessment-plan.md 4.1.1).
//
// It runs in `risk up`, so prod, existing tenants and --empty-data tenants
// all get it. Reference data only, safe on every database:
//   - a row is looked up by (framework, code) and created only when absent;
//   - an existing row is never updated or deleted (an edited name_en or an
//     is_active=false stays as it is);
//   - the insert is ON CONFLICT DO NOTHING, so a concurrent `risk up` cannot
//     fail on the unique (framework, code) index.
func SeedCosoCatalogue(db *gorm.DB) error {
	var inserted int
	for _, comp := range models.Coso2013Catalogue {
		var existing models.CosoComponent
		err := db.Where("framework = ? AND code = ?", models.CosoFramework2013, comp.Code).First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			row := models.CosoComponent{
				Framework: models.CosoFramework2013,
				Code:      comp.Code,
				SortOrder: comp.SortOrder,
				NameEn:    comp.NameEn,
			}
			res := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
			if res.Error != nil {
				return fmt.Errorf("seed coso component %s: %w", comp.Code, res.Error)
			}
			inserted += int(res.RowsAffected)
		case err != nil:
			return fmt.Errorf("check coso component %s: %w", comp.Code, err)
		}

		for _, p := range comp.Principles {
			var existingP models.CosoPrinciple
			err := db.Where("framework = ? AND code = ?", models.CosoFramework2013, p.Code).First(&existingP).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				row := models.CosoPrinciple{
					Framework:     models.CosoFramework2013,
					Code:          p.Code,
					Number:        p.Number,
					ComponentCode: comp.Code,
					NameEn:        p.NameEn,
					IsActive:      true,
				}
				res := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
				if res.Error != nil {
					return fmt.Errorf("seed coso principle %s: %w", p.Code, res.Error)
				}
				inserted += int(res.RowsAffected)
			case err != nil:
				return fmt.Errorf("check coso principle %s: %w", p.Code, err)
			}
		}
	}
	if inserted > 0 {
		log.Printf("seeded %d missing COSO 2013 catalogue rows", inserted)
	}
	return nil
}
