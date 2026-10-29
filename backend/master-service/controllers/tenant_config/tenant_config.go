package tenant_config

import (
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	tenantConfigSvc "master-service/services/tenant_config"

	"github.com/gin-gonic/gin"
)

// TenantConfigControllerInterface serves GET /tenant-config (both modes).
type TenantConfigControllerInterface interface {
	Get(ctx *gin.Context)
}

type TenantConfigController struct {
	*base.BaseController
	svc tenantConfigSvc.ServiceInterface
}

func NewTenantConfigController(svc tenantConfigSvc.ServiceInterface, validator *validations.Validator) TenantConfigControllerInterface {
	return &TenantConfigController{BaseController: base.NewBaseController(validator), svc: svc}
}

// Get answers {"org_structure":{"mode":"legacy"}} or, on flexible tenants,
// {"org_structure":{"mode":"flexible","max_levels":6,"levels":[...]}} inside
// the usual {success,message,data} envelope.
func (c *TenantConfigController) Get(ctx *gin.Context) {
	cfg, err := c.svc.Get(ctx.Request.Context())
	if err != nil {
		c.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Tenant configuration fetched successfully", cfg)
}
