package repositories

import (
	"sync"

	"auth-service/pkg/activitylog"
	"auth-service/pkg/config"
	pkgKafka "auth-service/pkg/kafka"
	"auth-service/pkg/logger"
	pkgRedis "auth-service/pkg/redis"
	"auth-service/pkg/secretcrypt"

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
	activity       activitylog.Recorder

	mfaCipherOnce sync.Once
	mfaCipher     *secretcrypt.Cipher
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
	GetActivityRecorder() activitylog.Recorder
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
	activity activitylog.Recorder,
) IRepositoryRegistry {
	return &Registry{
		db:             db,
		kafka:          kafkaProducer,
		redis:          redisClient,
		redisRepo:      redisRepo,
		config:         cfg,
		eventPublisher: eventPublisher,
		cacheRepo:      cacheRepo,
		activity:       activity,
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
	return NewMFASetupRepositoryWithCipher(r.db, r.getMFACipher())
}

// getMFACipher builds the MFA secret cipher from config once. serve validates
// the settings before the registry is used, so an error here means the
// registry was built without that check; it is logged and secrets are then
// written as plaintext while encrypted rows fail closed on read.
func (r *Registry) getMFACipher() *secretcrypt.Cipher {
	r.mfaCipherOnce.Do(func() {
		if r.config == nil {
			return
		}
		c, err := r.config.MFACipher()
		if err != nil {
			logger.Error("MFA secret encryption is misconfigured", logger.LogField("error", err.Error()))
			return
		}
		r.mfaCipher = c
	})
	return r.mfaCipher
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

// GetActivityRecorder returns the activity log recorder (a no-op when none was wired).
func (r *Registry) GetActivityRecorder() activitylog.Recorder {
	if r.activity == nil {
		return activitylog.NopRecorder{}
	}
	return r.activity
}
