package routes

import (
	"auth-service/clients"
	"auth-service/controllers"
	"auth-service/pkg/middleware"
	"auth-service/pkg/validations"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

type Registry struct {
	controller    controllers.IControllerRegistry
	group         *gin.RouterGroup
	authMiddleware *middleware.AuthMiddleware
	client        clients.IClientRegistry
}

type IRegistry interface {
	Serve()
	SetController(controller controllers.IControllerRegistry)
}

func NewRouteRegistry(
	group *gin.RouterGroup,
	authMiddleware *middleware.AuthMiddleware,
) IRegistry {
	return &Registry{
		group:          group,
		authMiddleware: authMiddleware,
	}
}

// NewControllerRegistry is a wrapper that creates the controller registry
func NewControllerRegistry(service services.IServiceRegistry, validator *validations.Validator) controllers.IControllerRegistry {
	return controllers.NewControllerRegistry(service, validator)
}

func (r *Registry) SetController(controller controllers.IControllerRegistry) {
	r.controller = controller
}

func (r *Registry) Serve() {
	r.auth()
	r.users()
	r.mfa()
	r.trustedDevices()
}

// auth registers auth routes
func (r *Registry) auth() {
	auth := r.group.Group("/auth")
	{
		auth.POST("/login", r.controller.GetAuth().Login)
		auth.POST("/verify-mfa-login", r.controller.GetAuth().VerifyMFALogin)
		auth.POST("/register", r.controller.GetAuth().Register)

		// Protected routes
		protected := auth.Group("")
		protected.Use(r.authMiddleware.Authenticate())
		{
			protected.POST("/logout", r.controller.GetAuth().Logout)
			protected.POST("/change-password", r.controller.GetAuth().ChangePassword)
			protected.GET("/me", r.controller.GetAuth().Me)
		}
	}
}

// mfa registers MFA routes
func (r *Registry) mfa() {
	mfa := r.group.Group("/mfa")
	mfa.Use(r.authMiddleware.Authenticate())
	{
		mfa.GET("/status", r.controller.GetMfa().GetMFAStatus)
		mfa.POST("/enroll", r.controller.GetMfa().EnrollMFA)
		mfa.POST("/verify", r.controller.GetMfa().VerifyMFA)
		mfa.POST("/disable", r.controller.GetMfa().UnenrollMFA)
		mfa.POST("/email-code", r.controller.GetMfa().SendEmailCode)
		mfa.POST("/verify-email-code", r.controller.GetMfa().VerifyEmailCode)
	}
}

// trustedDevices registers trusted devices routes
func (r *Registry) trustedDevices() {
	devices := r.group.Group("/devices")
	devices.Use(r.authMiddleware.Authenticate())
	{
		devices.GET("", r.controller.GetTrustedDevices().GetTrustedDevices)
		devices.POST("/enroll", r.controller.GetTrustedDevices().GenerateEnrollmentQR)
		devices.POST("/enroll/verify/:token", r.controller.GetTrustedDevices().VerifyEnrollmentToken)
		devices.DELETE("/:device_id", r.controller.GetTrustedDevices().UnenrollTrustedDevice)
	}
}

// users registers user management routes
func (r *Registry) users() {
	users := r.group.Group("/users")
	users.Use(r.authMiddleware.Authenticate())
	{
		users.GET("", r.controller.GetUser().ListUsers)
		users.POST("", r.controller.GetUser().CreateUser)
		users.GET("/:id", r.controller.GetUser().GetUser)
		users.PUT("/:id", r.controller.GetUser().UpdateUser)
		users.DELETE("/:id", r.controller.GetUser().DeleteUser)
	}
}
