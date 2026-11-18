package models

import (
	"strings"
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
	Severity     string  `json:"severity"` // "High", "Moderate to High", "Moderate", "Low to Moderate", "Low"
	Category     string  `json:"category"`
	Description  string  `json:"description,omitempty"`
	IsActive     bool    `json:"isActive"`
}

type TableDataMapping struct {
	TableName        string              `json:"tableName"`
	TargetScope      string              `json:"targetScope"`              // risk_management, audit_features, qar_features, data_analytics
	TargetModule     string              `json:"targetModule"`             // e.g. "Risk Profile / Heatmap", "AI: Anomaly Detection"
	CanonicalEntity  string              `json:"canonicalEntity,omitempty"` // generic, transactions, loans, gl_entries, ai_training_pool, accounts, customers
	AuditField       string              `json:"auditField"`               // mapped PK / key field
	WatermarkColumn  string              `json:"watermarkColumn,omitempty"`
	SoftDeleteColumn string              `json:"softDeleteColumn,omitempty"` // "" = none
	SyncInterval     string              `json:"syncInterval"`               // "Every 15 mins", "Hourly", "Daily", "Manual Only"
	AnomalyRules     []string            `json:"anomalyRules,omitempty"`     // e.g. ["Threshold Breach (>100M)", "Weekend Transaction"]
	CustomRules      []CustomAnomalyRule `json:"customRules,omitempty"`      // User-defined anomaly rules for AI model training
	IsActive         bool                `json:"isActive"`
}

// DataSourceConnection is an external source. SSLMode is a libpq sslmode (disable, require,
// verify-ca, verify-full) and SSL mirrors it (SSL = SSLMode != disable). SSLRootCert is an optional
// PEM CA bundle for the verify modes; a CA certificate is public, so it is stored unencrypted.
// Rows created before ssl_mode existed are backfilled by cmd/migrate.go.
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
	LastSyncedAt *time.Time         `json:"lastSyncedAt,omitempty"`
	NextSyncAt   *time.Time         `json:"nextSyncAt,omitempty"`
	LastError    string             `gorm:"type:text" json:"lastError,omitempty"`
	SSL          bool               `json:"ssl"` // no gorm default: default:true made GORM insert true for an explicit false (ssl_mode=disable)
	SSLMode      string             `gorm:"type:varchar(20)" json:"sslMode"`
	SSLRootCert  string             `gorm:"type:text" json:"sslRootCert,omitempty"`
	Username     string             `gorm:"type:varchar(255)" json:"username,omitempty"`
	Password     string             `gorm:"type:varchar(255)" json:"password,omitempty"`
	Scopes       []string           `gorm:"serializer:json" json:"scopes,omitempty"`
	DataMappings []TableDataMapping `gorm:"serializer:json" json:"dataMappings,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	DeletedAt    gorm.DeletedAt     `gorm:"index" json:"-"`
}

const (
	SyncScheduleEvery15Mins = "Every 15 mins"
	SyncScheduleHourly      = "Hourly"
	SyncScheduleDaily       = "Daily"
	SyncScheduleManualOnly  = "Manual Only"
)

// NormalizeSyncSchedule maps incoming/legacy sync schedule values to the supported presets.
// Legacy "Real-time (CDC)" is retired and normalized to "Every 15 mins".
func NormalizeSyncSchedule(schedule string) string {
	switch strings.ToLower(strings.TrimSpace(schedule)) {
	case "", "hourly", "1h":
		return SyncScheduleHourly
	case "every 15 mins", "every_15_mins", "15m", "15 mins", "real-time (cdc)", "realtime (cdc)", "real-time", "realtime":
		return SyncScheduleEvery15Mins
	case "daily", "24h", "1d":
		return SyncScheduleDaily
	case "manual only", "manual_only", "manual":
		return SyncScheduleManualOnly
	default:
		return SyncScheduleManualOnly
	}
}

// SyncScheduleInterval returns the polling duration for a sync schedule.
// Manual Only returns false so the scheduler skips it.
func SyncScheduleInterval(schedule string) (time.Duration, bool) {
	switch NormalizeSyncSchedule(schedule) {
	case SyncScheduleEvery15Mins:
		return 15 * time.Minute, true
	case SyncScheduleHourly:
		return 1 * time.Hour, true
	case SyncScheduleDaily:
		return 24 * time.Hour, true
	default:
		return 0, false
	}
}

// ComputeNextSyncAt returns the next scheduled sync timestamp relative to from,
// or nil when the schedule is Manual Only.
func ComputeNextSyncAt(schedule string, from time.Time) *time.Time {
	interval, ok := SyncScheduleInterval(schedule)
	if !ok {
		return nil
	}
	next := from.UTC().Add(interval)
	return &next
}

// ResolveCanonicalEntity deterministically selects the Lakehouse canonical entity
// for a source table and target module so users do not have to pick it manually.
func ResolveCanonicalEntity(tableName, targetModule, current string) string {
	tbl := strings.ToLower(strings.TrimSpace(tableName))
	mod := strings.ToLower(strings.TrimSpace(targetModule))

	switch {
	case strings.Contains(tbl, "trx") || strings.Contains(tbl, "transact"):
		return "transactions"
	case strings.Contains(tbl, "loan") || strings.Contains(tbl, "credit") || strings.Contains(tbl, "kredit"):
		return "loans"
	case strings.Contains(tbl, "gl") || strings.Contains(tbl, "ledger") || strings.Contains(tbl, "journal"):
		return "gl_entries"
	case strings.Contains(tbl, "acct") || strings.Contains(tbl, "account"):
		return "accounts"
	case strings.Contains(tbl, "cust") || strings.Contains(tbl, "customer") || strings.Contains(tbl, "cif"):
		return "customers"
	case strings.Contains(mod, "anomaly"):
		return "ai_training_pool"
	case current != "" && current != "generic":
		return current
	default:
		return "generic"
	}
}

// ApplyAutoMappingDefaults fills missing canonicalEntity and syncInterval values
// on each TableDataMapping so Data Hub always receives complete pipeline metadata.
func ApplyAutoMappingDefaults(mappings []TableDataMapping, defaultSchedule string) []TableDataMapping {
	normalizedSchedule := NormalizeSyncSchedule(defaultSchedule)
	for i := range mappings {
		mappings[i].CanonicalEntity = ResolveCanonicalEntity(mappings[i].TableName, mappings[i].TargetModule, mappings[i].CanonicalEntity)
		if strings.TrimSpace(mappings[i].SyncInterval) == "" {
			mappings[i].SyncInterval = normalizedSchedule
		} else {
			mappings[i].SyncInterval = NormalizeSyncSchedule(mappings[i].SyncInterval)
		}
	}
	return mappings
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
