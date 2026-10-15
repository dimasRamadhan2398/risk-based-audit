package models

import "time"

// KafkaOutboxEvent is a business event waiting to be published to Kafka. Rows
// are written by the request that changed the data (after its commit) and
// removed by the relay (pkg/kafka Relay) once the broker has acknowledged them.
// A row that survives a restart is published on the next start.
type KafkaOutboxEvent struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	Topic      string    `gorm:"type:varchar(255);not null" json:"topic"`
	MessageKey string    `gorm:"type:varchar(255);not null" json:"key"`
	EventType  string    `gorm:"type:varchar(100);not null" json:"eventType"`
	Payload    string    `gorm:"type:text;not null" json:"payload"`
	Attempts   int       `gorm:"not null;default:0" json:"attempts"`
	LastError  string    `gorm:"type:text" json:"lastError"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
}

func (KafkaOutboxEvent) TableName() string { return "kafka_outbox_events" }
