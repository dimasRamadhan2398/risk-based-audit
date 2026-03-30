package trusteddevices

import (
	"auth-service/clients"
	"auth-service/controllers"

	"github.com/gin-gonic/gin"
)

type TrustedDevicesRoute struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
	client     clients.IClientRegistry
}

type ITrustedDevicesRoute interface {
	Run()
}

func NewTrustedDevicesRoute(controller controllers.IControllerRegistry, group *gin.RouterGroup, client clients.IClientRegistry) ITrustedDevicesRoute {
	return &TrustedDevicesRoute{
		controller: controller,
		group:      group,
		client:     client,
	}
}

// Run implements ITrustedDevicesRoute.
func (t *TrustedDevicesRoute) Run() {
	devices := t.group.Group("/devices")
	{
		devices.GET("", t.controller.GetTrustedDevices().GetTrustedDevices)
		devices.POST("/enroll", t.controller.GetTrustedDevices().GenerateEnrollmentQR)
		devices.POST("/enroll/verify/:token", t.controller.GetTrustedDevices().VerifyEnrollmentToken)
		devices.DELETE("/:device_id", t.controller.GetTrustedDevices().UnenrollTrustedDevice)
	}
}