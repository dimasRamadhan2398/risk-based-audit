package services

import (
	"auth-service/pkg/config"
	"auth-service/repositories"

	authServices "auth-service/services/auth"
	confidentialityServices "auth-service/services/confidentiality"
	emailServices "auth-service/services/email"
	mfaServices "auth-service/services/mfa"
	roleServices "auth-service/services/role"
	trustedDevicesServices "auth-service/services/trusted-devices"
	userServices "auth-service/services/user"
)

type Registry struct {
	repository repositories.IRepositoryRegistry
	config     *config.Config
}

// NewAuthService implements IServiceRegistry.
func (r *Registry) GetAuthService() authServices.AuthServiceInterface {
	return authServices.NewAuthService(
		r.repository.GetUserRepository(),
		r.repository.GetMFASetupRepository(),
		r.repository.GetTrustedDeviceRepository(),
		r.repository.GetCacheRepository().GetClient(),
		r.repository.GetConfig(),
		r.repository.GetKafkaProducer(),
	)
}

// GetEmailService implements IServiceRegistry.
func (r *Registry) GetEmailService() emailServices.EmailServiceInterface {
	return emailServices.NewEmailService(&r.repository.GetConfig().SMTP, r.repository.GetConfig().App.AppName)
}

// GetMfaService implements IServiceRegistry.
func (r *Registry) GetMfaService() mfaServices.MfaServiceInterface {
	return mfaServices.NewMfaService(r.repository.GetMFASetupRepository(), r.repository.GetUserRepository(), r.repository.GetEventPublisher(), r.repository.GetCacheRepository().GetClient(), r.GetEmailService())
}

// GetUserService implements IServiceRegistry.
func (r *Registry) GetUserService() userServices.UserServiceInterface {
	return userServices.NewUserService(r.repository.GetUserRepository(), r.repository.GetKafkaProducer())
}

// GetRoleService implements IServiceRegistry.
func (r *Registry) GetRoleService() roleServices.IRoleService {
	var redisClient = r.repository.GetCacheRepository()
	var rawClient = redisClient.GetClient()
	return roleServices.NewRoleService(
		r.repository.GetRoleRepository(),
		r.repository.GetPermissionRepository(),
		r.repository.GetEventPublisher(),
		rawClient,
	)
}

// GetTrustedDevicesService implements IServiceRegistry.
func (r *Registry) GetTrustedDevicesService() trustedDevicesServices.TrustedDevicesServiceInterface {
	return trustedDevicesServices.NewTrustedDevicesService(r.repository.GetTrustedDeviceRepository(), r.repository.GetUserRepository(), r.repository.GetEventPublisher(), r.repository.GetCacheRepository().GetClient())
}

// GetConfidentialityService implements IServiceRegistry.
func (r *Registry) GetConfidentialityService() confidentialityServices.ConfidentialityServiceInterface {
	return confidentialityServices.NewConfidentialityService(r.repository.GetConfidentialityAgreementRepository())
}

type IServiceRegistry interface {
	GetAuthService() authServices.AuthServiceInterface
	GetUserService() userServices.UserServiceInterface
	GetRoleService() roleServices.IRoleService
	GetMfaService() mfaServices.MfaServiceInterface
	GetTrustedDevicesService() trustedDevicesServices.TrustedDevicesServiceInterface
	GetEmailService() emailServices.EmailServiceInterface
	GetConfidentialityService() confidentialityServices.ConfidentialityServiceInterface
}

func NewServiceRegistry(repository repositories.IRepositoryRegistry) IServiceRegistry {
	return &Registry{repository: repository}
}
