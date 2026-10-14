// Package annualplan holds request handling specific to /annual-audit-plans
// that the generic crud handlers do not cover: assigning the plan code
// (AuditAnnual.Code, PKAT-{year}-{seq}).
package annualplan

import (
	"fmt"
	"net/http"

	"audit-service/models"
	"audit-service/pkg/activitycode"
	apperrors "audit-service/pkg/errors"

	"gorm.io/gorm"
)

// AssignCodeOnCreate is a crud.CreateHook for AuditAnnual. The plan code is
// generated from the plan year inside the create transaction; a code sent by
// the client is ignored. Imported plans (uploaded_annual_plans) are a separate
// table and keep the code of the document.
func AssignCodeOnCreate(tx *gorm.DB, entity interface{}) error {
	plan, ok := entity.(*models.AuditAnnual)
	if !ok {
		return fmt.Errorf("annualplan: unexpected entity %T", entity)
	}
	if plan.Year < 1000 || plan.Year > 9999 {
		return apperrors.New("VALIDATION_ERROR", "year must be a four-digit plan year", http.StatusBadRequest)
	}
	code, err := activitycode.NextPlan(tx, plan.Year)
	if err != nil {
		return err
	}
	plan.Code = code
	return nil
}
