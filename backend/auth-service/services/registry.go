package services

import (
	"auth-service/pkg/config"
	"auth-service/repositories"

	authServices "auth-service/services/auth"
	mfaServices "auth-service/services/mfa"
	trustedDevicesServices "auth-service/services/trusted-devices"
	userServices "auth-service/services/user"
)

type Registry struct {
	repository repositories.IRepositoryRegistry
	config     *config.Config
}

// NewAuthService implements IServiceRegistry.
func (r *Registry) GetAuthService() authServices.AuthServiceInterface {
	return authServices.NewAuthService(r.repository.GetUserRepository(), r.repository.GetCacheRepository().GetClient(), r.repository.GetConfig(), r.repository.GetKafkaProducer())
}

// GetMfaService implements IServiceRegistry.
func (r *Registry) GetMfaService() mfaServices.MfaServiceInterface {
	return mfaServices.NewMfaService(r.repository.GetMFASetupRepository(), r.repository.GetUserRepository(), r.repository.GetEventPublisher(), r.repository.GetCacheRepository().GetClient())
}

// GetUserService implements IServiceRegistry.
func (r *Registry) GetUserService() userServices.UserServiceInterface {
	return userServices.NewUserService(r.repository.GetUserRepository(), r.repository.GetKafkaProducer())
}

// GetTrustedDevicesService implements IServiceRegistry.
func (r *Registry) GetTrustedDevicesService() trustedDevicesServices.TrustedDevicesServiceInterface {
	return trustedDevicesServices.NewTrustedDevicesService(r.repository.GetTrustedDeviceRepository(), r.repository.GetUserRepository(), r.repository.GetEventPublisher(), r.repository.GetCacheRepository().GetClient())
}

type IServiceRegistry interface {
	GetAuthService() authServices.AuthServiceInterface
	GetUserService() userServices.UserServiceInterface
	GetMfaService() mfaServices.MfaServiceInterface
	GetTrustedDevicesService() trustedDevicesServices.TrustedDevicesServiceInterface
}
func NewServiceRegistry(repository repositories.IRepositoryRegistry) IServiceRegistry {
	return &Registry{repository: repository}
}
