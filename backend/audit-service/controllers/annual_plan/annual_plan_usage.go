package annualplan

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"audit-service/controllers/crud"
	"audit-service/models"
	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/response"
	form "audit-service/services/activity_plan_form"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// annualActivityView is a PKAT row with the RKA that holds it.
type annualActivityView struct {
	models.AnnualAuditActivity
	ActivityPlanRef *form.ActivityPlanRef `json:"activityPlanRef"`
}

// annualPlanView is an annual plan as returned by GET /annual-audit-plans and
// GET /annual-audit-plans/:id: the stored plan plus isUsed (a live RKA or
// audit activity points at it) and activities[].activityPlanRef. Row ids are
// always set (form.EnsureRowIDs), so the PKAT form can send them back.
type annualPlanView struct {
	models.AuditAnnual
	Activities []annualActivityView `json:"activities"`
	IsUsed     bool                 `json:"isUsed"`
}

func decoratePlans(db *gorm.DB, plans []models.AuditAnnual) ([]annualPlanView, error) {
	ids := make([]uuid.UUID, len(plans))
	for i := range plans {
		ids[i] = plans[i].ID
	}
	usage, err := form.PlanUsage(db, ids)
	if err != nil {
		return nil, err
	}
	out := make([]annualPlanView, len(plans))
	for i := range plans {
		form.EnsureRowIDs(&plans[i])
		acts := make([]annualActivityView, len(plans[i].Activities))
		for j, a := range plans[i].Activities {
			acts[j] = annualActivityView{AnnualAuditActivity: a, ActivityPlanRef: usage.Refs[a.ID]}
		}
		out[i] = annualPlanView{AuditAnnual: plans[i], Activities: acts, IsUsed: usage.Used[plans[i].ID]}
	}
	return out, nil
}

// DecorateList is the crud.ItemsDecorator of GET /annual-audit-plans.
func DecorateList(db *gorm.DB) crud.ItemsDecorator {
	return func(c *gin.Context, items interface{}) (interface{}, error) {
		plans, ok := items.(*[]models.AuditAnnual)
		if !ok {
			return nil, fmt.Errorf("annualplan: unexpected items %T", items)
		}
		return decoratePlans(db, *plans)
	}
}

// GetByID is GET /annual-audit-plans/:id with the usage fields.
func GetByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid AuditAnnual ID")
			return
		}
		var plan models.AuditAnnual
		if err := db.First(&plan, "id = ?", id).Error; err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				response.NotFound(c, "AuditAnnual not found")
				return
			}
			response.InternalServerError(c, "Failed to fetch AuditAnnual")
			return
		}
		views, err := decoratePlans(db, []models.AuditAnnual{plan})
		if err != nil {
			response.InternalServerError(c, "Failed to fetch AuditAnnual")
			return
		}
		response.OK(c, "AuditAnnual fetched successfully", views[0])
	}
}

// Delete is DELETE /annual-audit-plans/:id. A plan in use (a live RKA or
// audit activity points at it) is not deleted: 409 ANNUAL_PLAN_IN_USE, and the
// client offers deactivation (PUT {isActive:false}) instead. The plan row is
// locked so a concurrent RKA save cannot slip in between check and delete.
func Delete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid AuditAnnual ID")
			return
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			var plan models.AuditAnnual
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&plan, "id = ?", id).Error; err != nil {
				return err
			}
			usage, err := form.PlanUsage(tx, []uuid.UUID{id})
			if err != nil {
				return err
			}
			if usage.Used[id] {
				return form.Fail(form.CodeAnnualPlanInUse, http.StatusConflict,
					"this annual audit plan (PKAT) is used by an activity plan or audit activity and cannot be deleted; deactivate it instead",
					form.Details{})
			}
			return tx.Delete(&models.AuditAnnual{}, "id = ?", id).Error
		})
		if err != nil {
			var appErr *apperrors.AppError
			switch {
			case stderrors.Is(err, gorm.ErrRecordNotFound):
				response.NotFound(c, "AuditAnnual not found")
			case stderrors.As(err, &appErr):
				response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
			default:
				response.InternalServerError(c, "Failed to delete AuditAnnual")
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "AuditAnnual deleted successfully",
		})
	}
}

// RefreshSnapshotsAfterUpdate is the AfterHook of PUT /annual-audit-plans/:id:
// in the same transaction, the RKA rows linked to the plan's rows get the new
// PKAT values in their snapshot columns, and the plan's RKAs its year.
func RefreshSnapshotsAfterUpdate(tx *gorm.DB, entity interface{}) error {
	plan, ok := entity.(*models.AuditAnnual)
	if !ok {
		return fmt.Errorf("annualplan: unexpected entity %T", entity)
	}
	return form.RefreshSnapshots(tx, plan.ID)
}
