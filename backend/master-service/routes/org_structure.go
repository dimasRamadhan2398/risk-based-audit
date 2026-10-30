package routes

import (
	"context"
	"net/http"

	orgLevelCtrl "master-service/controllers/org_level"
	tenantConfigCtrl "master-service/controllers/tenant_config"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// OrgRouteOptions configures the organisation-structure routes
// (docs/org-structure-model-plan.md section 6).
type OrgRouteOptions struct {
	// Flexible is the effective mode (orgmode.Resolve). False = legacy.
	Flexible bool
	// StepUpStore consumes step-up token ids once (the tenant's Redis). nil
	// makes apply answer 503.
	StepUpStore middleware.JTIStore
	// Resync, when set (flexible), replaces the legacy department resync so
	// the queued events carry the org fields.
	Resync func(ctx context.Context) (int, error)
}

// OrgLevelRoles may change org levels (decision D8). No aliases.
var OrgLevelRoles = []string{"ADMIN", "EXECUTIVE"}

// registerOrgStructureRoutes adds /tenant-config (both modes) and /org-levels
// (flexible; 404 ORG_STRUCTURE_NOT_ENABLED on legacy without touching the DB).
// api must already carry Authenticate().
func registerOrgStructureRoutes(api *gin.RouterGroup, tenant tenantConfigCtrl.TenantConfigControllerInterface,
	levels orgLevelCtrl.OrgLevelControllerInterface, auth *middleware.AuthMiddleware, opts OrgRouteOptions) {

	api.GET("/tenant-config", tenant.Get)

	if !opts.Flexible || levels == nil {
		notEnabled := func(c *gin.Context) { base.RespondError(c, apperrors.ErrOrgStructureNotEnabled) }
		api.Any("/org-levels", notEnabled)
		api.Any("/org-levels/*rest", notEnabled)
		return
	}

	orgLevels := api.Group("/org-levels", middleware.RequireFlexible(opts.Flexible))
	{
		orgLevels.GET("", levels.List)
		changes := orgLevels.Group("/changes", auth.RequireRoles(OrgLevelRoles...))
		{
			changes.GET("", levels.ChangeLog)
			changes.POST("/preview", levels.Preview)
			changes.POST("/apply", auth.RequireStepUp(middleware.StepUpPurposeOrgLevels, opts.StepUpStore), levels.Apply)
		}
	}
}

// resyncOrgDepartments is the flexible replacement of resyncDepartmentEvents.
func resyncOrgDepartments(resync func(ctx context.Context) (int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		queued, err := resync(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "queued": queued, "message": "Failed to queue department events"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"success": true, "queued": queued})
	}
}
