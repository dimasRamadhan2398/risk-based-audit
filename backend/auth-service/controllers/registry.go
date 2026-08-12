package controllers

import (
	authCtrl "auth-service/controllers/auth"
	confidentialityCtrl "auth-service/controllers/confidentiality"
	mfaCtrl "auth-service/controllers/mfa"
	roleCtrl "auth-service/controllers/role"
	trustedDevicesCtrl "auth-service/controllers/trusted-devices"
	userCtrl "auth-service/controllers/user"
	"auth-service/pkg/validations"
	"auth-service/services"
)

type Registry struct {
	service   services.IServiceRegistry
	validator *validations.Validator
}

type IControllerRegistry interface {
	GetAuth() authCtrl.AuthControllerInterface
	GetUser() userCtrl.UserControllerInterface
	GetRole() roleCtrl.RoleControllerInterface
	GetMfa() mfaCtrl.MfaControllerInterface
	GetTrustedDevices() trustedDevicesCtrl.TrustedDevicesControllerInterface
	GetConfidentiality() confidentialityCtrl.ConfidentialityControllerInterface
}

func NewControllerRegistry(service services.IServiceRegistry, validator *validations.Validator) IControllerRegistry {
	return &Registry{
		service:   service,
		validator: validator,
	}
}

func (r *Registry) GetAuth() authCtrl.AuthControllerInterface {
	return authCtrl.NewAuthController(r.validator, r.service.GetAuthService())
}

func (r *Registry) GetUser() userCtrl.UserControllerInterface {
	return userCtrl.NewUserController(r.validator, r.service.GetUserService())
}

func (r *Registry) GetRole() roleCtrl.RoleControllerInterface {
	return roleCtrl.NewRoleController(r.validator, r.service.GetRoleService())
}

func (r *Registry) GetMfa() mfaCtrl.MfaControllerInterface {
	return mfaCtrl.NewMfaController(r.validator, r.service.GetMfaService())
}

func (r *Registry) GetTrustedDevices() trustedDevicesCtrl.TrustedDevicesControllerInterface {
	return trustedDevicesCtrl.NewTrustedDevicesController(r.validator, r.service.GetTrustedDevicesService())
}

func (r *Registry) GetConfidentiality() confidentialityCtrl.ConfidentialityControllerInterface {
	return confidentialityCtrl.NewConfidentialityController(r.validator, r.service.GetConfidentialityService())
}
