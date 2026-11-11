package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TableColumn struct {
	Name       string `json:"name"`
	DataType   string `json:"dataType"`
	IsNullable bool   `json:"isNullable"`
	IsPrimary  bool   `json:"isPrimary"`
}

type SchemaTable struct {
	TableName   string        `json:"tableName"`
	RowCount    int64         `json:"rowCount"`
	ColumnCount int           `json:"columnCount"`
	Columns     []TableColumn `json:"columns"`
	Description string        `json:"description,omitempty"`
}

type CustomAnomalyRule struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Field        string  `json:"field"`
	Operator     string  `json:"operator"` // ">", "<", ">=", "<=", "==", "!=", "outside"
	Threshold    float64 `json:"threshold"`
	ThresholdStr string  `json:"thresholdStr,omitempty"`
	Severity     string  `json:"severity"` // "Critical", "High", "Medium", "Low"
	Category     string  `json:"category"`
	Description  string  `json:"description,omitempty"`
	IsActive     bool    `json:"isActive"`
}

type TableDataMapping struct {
	TableName        string              `json:"tableName"`
	TargetScope      string              `json:"targetScope"`            // risk_management, audit_features, qar_features, data_analytics
	TargetModule     string              `json:"targetModule"`           // e.g. "Risk Profile / Heatmap", "AI: Anomaly Detection"
	AuditField       string              `json:"auditField"`             // mapped field (e.g. "transaction_amount", "user_id")
	SyncInterval     string              `json:"syncInterval"`           // "Hourly", "Real-time", "Daily"
	AnomalyRules     []string            `json:"anomalyRules,omitempty"` // e.g. ["Threshold Breach (>100M)", "Weekend Transaction"]
	CustomRules      []CustomAnomalyRule `json:"customRules,omitempty"`  // User-defined anomaly rules for AI model training
	IsActive         bool                `json:"isActive"`
	// Column that marks a source row as deleted (sent by the Data Source settings page); "" = none.
	SoftDeleteColumn string              `json:"softDeleteColumn,omitempty"`
}

type DataSourceConnection struct {
	ID           uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string             `gorm:"type:varchar(255);not null" json:"name"`
	Type         string             `gorm:"type:varchar(50);not null;default:'postgres'" json:"type"`
	Host         string             `gorm:"type:varchar(255);not null" json:"host"`
	Port         int                `gorm:"type:int;not null;default:5432" json:"port"`
	Database     string             `gorm:"type:varchar(255);not null" json:"database"`
	Environment  string             `gorm:"type:varchar(50);default:'Production'" json:"environment"`
	Status       string             `gorm:"type:varchar(50);default:'Connected'" json:"status"`
	SyncSchedule string             `gorm:"type:varchar(100);default:'Hourly'" json:"syncSchedule"`
	LastSync     string             `gorm:"type:varchar(100)" json:"lastSync"`
	LastError    string             `gorm:"type:text" json:"lastError,omitempty"`
	SSL          bool               `gorm:"default:true" json:"ssl"`
	Username     string             `gorm:"type:varchar(255)" json:"username,omitempty"`
	Password     string             `gorm:"type:varchar(255)" json:"password,omitempty"`
	Scopes       []string           `gorm:"serializer:json" json:"scopes,omitempty"`
	DataMappings []TableDataMapping `gorm:"serializer:json" json:"dataMappings,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	DeletedAt    gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (DataSourceConnection) TableName() string {
	return "data_source_connections"
}

type DataSourceActivityLog struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ConnectionID   uuid.UUID      `gorm:"type:uuid;index" json:"connection_id"`
	ConnectionName string         `gorm:"type:varchar(255)" json:"connName"`
	Type           string         `gorm:"type:varchar(50)" json:"type"`
	Event          string         `gorm:"type:varchar(255)" json:"event"`
	Status         string         `gorm:"type:varchar(50)" json:"status"`
	Records        int            `gorm:"type:int" json:"records"`
	Duration       string         `gorm:"type:varchar(50)" json:"duration"`
	Timestamp      string         `gorm:"type:varchar(100)" json:"timestamp"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DataSourceActivityLog) TableName() string {
	return "data_source_activity_logs"
}
