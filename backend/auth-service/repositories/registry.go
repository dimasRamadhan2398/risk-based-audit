package repositories

import (
	"auth-service/pkg/config"
	pkgKafka "auth-service/pkg/kafka"
	pkgRedis "auth-service/pkg/redis"

	"gorm.io/gorm"
)

type Registry struct {
	db             *gorm.DB
	kafka          *pkgKafka.Producer
	redis          *pkgRedis.Client
	redisRepo      *pkgRedis.Repository
	config         *config.Config
	eventPublisher pkgKafka.IEventPublisher
	cacheRepo      ICacheRepository
}

type IRepositoryRegistry interface {
	GetUserRepository() IUserRepository
	GetRoleRepository() IRoleRepository
	GetPermissionRepository() IPermissionRepository
	GetRefreshTokenRepository() IRefreshTokenRepository
	GetSystemAlertLogRepository() ISysAlertLogRepository
	GetMFASetupRepository() IMFASetupRepository
	GetTrustedDeviceRepository() ITrustedDeviceRepository
	GetConfidentialityAgreementRepository() IConfidentialityAgreementRepository
	GetCacheRepository() ICacheRepository
	GetKafkaProducer() *pkgKafka.Producer
	GetConfig() *config.Config
	GetEventPublisher() pkgKafka.IEventPublisher
}

func NewRegistry(db *gorm.DB) IRepositoryRegistry {
	return &Registry{db: db}
}

func NewRegistryWithRedis(db *gorm.DB, redisClient *pkgRedis.Client, redisRepo *pkgRedis.Repository) IRepositoryRegistry {
	return &Registry{
		db:        db,
		redis:     redisClient,
		redisRepo: redisRepo,
	}
}

func NewRegistryWithKafka(db *gorm.DB, kafkaProducer *pkgKafka.Producer) IRepositoryRegistry {
	return &Registry{
		db:    db,
		kafka: kafkaProducer,
	}
}

func NewRegistryWithEventPublisher(db *gorm.DB, eventPublisher pkgKafka.IEventPublisher) IRepositoryRegistry {
	return &Registry{
		db:             db,
		eventPublisher: eventPublisher,
	}
}

// NewRegistryWithAll creates a registry with all dependencies
func NewRegistryWithAll(
	db *gorm.DB,
	redisClient *pkgRedis.Client,
	redisRepo *pkgRedis.Repository,
	kafkaProducer *pkgKafka.Producer,
	eventPublisher pkgKafka.IEventPublisher,
	cacheRepo ICacheRepository,
	cfg *config.Config,
) IRepositoryRegistry {
	return &Registry{
		db:             db,
		kafka:          kafkaProducer,
		redis:          redisClient,
		redisRepo:      redisRepo,
		config:         cfg,
		eventPublisher: eventPublisher,
		cacheRepo:      cacheRepo,
	}
}

func (r *Registry) GetUserRepository() IUserRepository {
	return NewUserRepository(r.db)
}

func (r *Registry) GetRoleRepository() IRoleRepository {
	return NewRoleRepository(r.db)
}

func (r *Registry) GetPermissionRepository() IPermissionRepository {
	return NewPermissionRepository(r.db)
}

func (r *Registry) GetRefreshTokenRepository() IRefreshTokenRepository {
	return NewRefreshTokenRepository(r.db)
}

func (r *Registry) GetSystemAlertLogRepository() ISysAlertLogRepository {
	return NewSysAlertLogRepository(r.db)
}

func (r *Registry) GetMFASetupRepository() IMFASetupRepository {
	return NewMFASetupRepository(r.db)
}

func (r *Registry) GetTrustedDeviceRepository() ITrustedDeviceRepository {
	return NewTrustedDeviceRepository(r.db)
}

func (r *Registry) GetConfidentialityAgreementRepository() IConfidentialityAgreementRepository {
	return NewConfidentialityAgreementRepository(r.db)
}

func (r *Registry) GetCacheRepository() ICacheRepository {
	if r.cacheRepo != nil {
		return r.cacheRepo
	}
	if r.redisRepo != nil && r.redis != nil {
		return NewCacheRepository(r.redisRepo, r.redis)
	}
	return nil
}

func (r *Registry) GetKafkaProducer() *pkgKafka.Producer {
	return r.kafka
}

func (r *Registry) GetConfig() *config.Config {
	return r.config
}

func (r *Registry) GetEventPublisher() pkgKafka.IEventPublisher {
	return r.eventPublisher
}
