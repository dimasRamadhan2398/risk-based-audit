package auth

import (
	"auth-service/clients"
	"auth-service/controllers"

	"github.com/gin-gonic/gin"
)

type AuthRoute struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
	client     clients.IClientRegistry
}

type IAuthRoute interface {
	Run()
}

func NewAuthRoute(controller controllers.IControllerRegistry, group *gin.RouterGroup, client clients.IClientRegistry) IAuthRoute {
	return &AuthRoute{
		controller: controller,
		group:      group,
		client:     client,
	}
}

// Run implements IAuthRoute.
func (a *AuthRoute) Run() {
	group := a.group.Group("/auth")
	group.POST("/login", a.controller.GetAuth().Login)
	group.POST("/register", a.controller.GetAuth().Register)
	group.POST("/logout", a.controller.GetAuth().Logout)
	group.POST("/change-password", a.controller.GetAuth().ChangePassword)
	group.GET("/me", a.controller.GetAuth().Me)
}
