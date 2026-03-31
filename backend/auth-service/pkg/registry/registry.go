package registry

import (
	"auth-service/pkg/config"
	kafkapkg "auth-service/pkg/kafka"
	"auth-service/pkg/redis"
	"go.uber.org/zap"
)

// KafkaRegistry holds both producer and consumer for convenience
type KafkaRegistry struct {
	Producer *kafkapkg.Producer
	Consumer *kafkapkg.Consumer
}

// RedisRegistry holds Redis client and repository
type RedisRegistry struct {
	Client     *redis.Client
	Repository *redis.Repository
}

// InfraRegistry holds all infrastructure dependencies
type InfraRegistry struct {
	Redis  *RedisRegistry
	Kafka  *KafkaRegistry
	Logger *zap.Logger
}

// NewInfraRegistry creates a new infrastructure registry
func NewInfraRegistry(cfg *config.Config, logger *zap.Logger) (*InfraRegistry, error) {
	registry := &InfraRegistry{
		Logger: logger,
	}

	// Initialize Redis
	redisClient, err := redis.NewRedisConnection(&cfg.Redis)
	if err != nil {
		return nil, err
	}

	registry.Redis = &RedisRegistry{
		Client:     redisClient,
		Repository: redis.NewRepository(redisClient),
	}

	// Initialize Kafka Producer
	kafkaProducer, err := kafkapkg.NewProducer(kafkapkg.Config{
		Brokers:     cfg.Kafka.Brokers,
		Topic:       cfg.Kafka.Topic,
		ServiceName: cfg.Kafka.ServiceName,
		Enabled:     cfg.Kafka.Enabled,
	}, logger)
	if err != nil {
		return nil, err
	}

	// Initialize Kafka Consumer
	kafkaConsumer, err := kafkapkg.NewConsumer(kafkapkg.ConsumerConfig{
		Brokers:  cfg.KafkaConsumer.Brokers,
		GroupID:  cfg.KafkaConsumer.GroupID,
		Topics:   cfg.KafkaConsumer.Topics,
		Enabled:  cfg.KafkaConsumer.Enabled,
		Version:  cfg.KafkaConsumer.Version,
		Assignor: cfg.KafkaConsumer.Assignor,
	}, logger)
	if err != nil {
		return nil, err
	}

	registry.Kafka = &KafkaRegistry{
		Producer: kafkaProducer,
		Consumer: kafkaConsumer,
	}

	return registry, nil
}

// Close gracefully shuts down all infrastructure connections
func (r *InfraRegistry) Close() error {
	// Close Kafka producer
	if r.Kafka != nil && r.Kafka.Producer != nil {
		if err := r.Kafka.Producer.Close(); err != nil {
			r.Logger.Error("Failed to close Kafka producer", zap.Error(err))
		}
	}

	// Close Kafka consumer
	if r.Kafka != nil && r.Kafka.Consumer != nil {
		if err := r.Kafka.Consumer.Stop(); err != nil {
			r.Logger.Error("Failed to stop Kafka consumer", zap.Error(err))
		}
	}

	// Close Redis connection
	if r.Redis != nil && r.Redis.Client != nil {
		if err := r.Redis.Client.Close(); err != nil {
			r.Logger.Error("Failed to close Redis connection", zap.Error(err))
		}
	}

	return nil
}
