package global

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ===== SUBSCRIPTION PLAN =====

type PlanType string

const (
    PlanStarter    PlanType = "STARTER"     
    PlanPro        PlanType = "PRO"         
    PlanEnterprise PlanType = "ENTERPRISE" 
)

type SubscriptionPlan struct {
    ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    Code             string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
    Name             string         `gorm:"type:varchar(100);not null" json:"name"`
    PlanType         PlanType       `gorm:"type:varchar(20);not null" json:"plan_type"`
    MaxCompanies     int            `gorm:"default:1" json:"max_companies"`      // -1 = unlimited
    MaxUsers         int            `gorm:"default:10" json:"max_users"`         // -1 = unlimited
    MaxStorageGB     int            `gorm:"default:5" json:"max_storage_gb"`
    PriceMonthly     float64        `gorm:"type:decimal(12,2)" json:"price_monthly"`
    PriceAnnually    float64        `gorm:"type:decimal(12,2)" json:"price_annually"`
    Features         string 		`gorm:"type:text" json:"features"`          // feature flags
    IsActive         bool           `gorm:"default:true" json:"is_active"`
    CreatedAt        time.Time      `json:"created_at"`
    UpdatedAt        time.Time      `json:"updated_at"`
}


// ===== TENANT (Your Direct Client) =====

type TenantStatus string

const (
    TenantStatusActive    TenantStatus = "ACTIVE"
    TenantStatusSuspended TenantStatus = "SUSPENDED"
    TenantStatusTrial     TenantStatus = "TRIAL"
    TenantStatusExpired   TenantStatus = "EXPIRED"
)

type Tenant struct {
    ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    Code           string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`     // e.g. "CLIENTA"
    Name           string         `gorm:"type:varchar(150);not null" json:"name"`                // client company name
    AdminEmail     string         `gorm:"type:varchar(100);not null" json:"admin_email"`         // primary contact
    Phone          string         `gorm:"type:varchar(20)" json:"phone"`
    Status         TenantStatus   `gorm:"type:varchar(20);default:'TRIAL'" json:"status"`
    TrialEndsAt    *time.Time     `json:"trial_ends_at,omitempty"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
    DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}


// ===== TENANT SUBSCRIPTION =====

type TenantSubscription struct {
    ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

    TenantID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
    Tenant         Tenant         `gorm:"foreignKey:TenantID" json:"tenant"`

    PlanID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"plan_id"`
    Plan           SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan"`

    StartsAt       time.Time      `json:"starts_at"`
    ExpiresAt      time.Time      `json:"expires_at"`
    IsActive       bool           `gorm:"default:true" json:"is_active"`
    BillingCycle   string         `gorm:"type:varchar(20);default:'MONTHLY'" json:"billing_cycle"` // MONTHLY | ANNUAL
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
}


// ===== TENANT DATABASE CONFIG =====
// Tells your app where to connect for each tenant

type TenantDatabaseConfig struct {
    ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

    TenantID       uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"tenant_id"`
    Tenant         Tenant         `gorm:"foreignKey:TenantID" json:"tenant"`

    DBHost         string         `gorm:"type:varchar(255);not null" json:"db_host"`
    DBPort         int            `gorm:"default:5432" json:"db_port"`
    DBName         string         `gorm:"type:varchar(100);not null" json:"db_name"`
    DBUser         string         `gorm:"type:varchar(100);not null" json:"db_user"`
    DBPassword     string         `gorm:"type:varchar(255);not null" json:"-"`    // encrypted at rest
    MaxConnections int            `gorm:"default:10" json:"max_connections"`
    SSLMode        string         `gorm:"type:varchar(20);default:'require'" json:"ssl_mode"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
}