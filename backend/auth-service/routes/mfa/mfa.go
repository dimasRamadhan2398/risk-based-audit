package mfa

import (
	"auth-service/clients"
	"auth-service/controllers"

	"github.com/gin-gonic/gin"
)

type MfaRoute struct {
	controller controllers.IControllerRegistry
	group *gin.RouterGroup
	client clients.IClientRegistry
}

type IMfaRoute interface {
	Run()
}

func NewMfaRoute(controller controllers.IControllerRegistry, group *gin.RouterGroup, client clients.IClientRegistry) IMfaRoute {
	return &MfaRoute{
		controller: controller,
		group: group,
		client: client,
	}
}

func (m *MfaRoute) Run() {
	group := m.group.Group("/mfa")
	group.POST("/verify", m.controller.GetMfa().VerifyMFA)
	group.POST("/enroll", m.controller.GetMfa().EnrollMFA)
	group.POST("/disable", m.controller.GetMfa().UnenrollMFA)
	group.POST("/email-code", m.controller.GetMfa().SendEmailCode)
	group.POST("/verify-email-code", m.controller.GetMfa().VerifyEmailCode)
	group.GET("/status", m.controller.GetMfa().GetMFAStatus)
}