// Package tenant_config serves GET /tenant-config: what the frontend needs to
// know about this tenant's feature set (docs/org-structure-model-plan.md 5.2).
package tenant_config

import (
	"context"

	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/orgmode"
	levelRepo "master-service/repositories/org_level"

	"gorm.io/gorm"
)

// Level is one active org level as the frontend sees it.
type Level struct {
	ID      string  `json:"id"`
	Level   int     `json:"level"`
	Name    string  `json:"name"`
	NameKey *string `json:"name_key"`
}

// OrgStructure is the org_structure block. Legacy carries only the mode.
type OrgStructure struct {
	Mode      orgmode.Mode `json:"mode"`
	MaxLevels int          `json:"max_levels,omitempty"`
	// Pointer so a flexible tenant with no active level still sends [].
	Levels *[]Level `json:"levels,omitempty"`
}

// TenantConfig is the response body (inside the usual data envelope).
type TenantConfig struct {
	OrgStructure OrgStructure `json:"org_structure"`
}

// ServiceInterface is the tenant config service.
type ServiceInterface interface {
	Get(ctx context.Context) (*TenantConfig, error)
}

type service struct {
	db   *gorm.DB
	mode orgmode.Mode
}

// NewService builds the service for the resolved mode. In legacy mode it never
// touches the database.
func NewService(db *gorm.DB, mode orgmode.Mode) ServiceInterface {
	return &service{db: db, mode: mode}
}

func (s *service) Get(ctx context.Context) (*TenantConfig, error) {
	if !s.mode.IsFlexible() {
		return &TenantConfig{OrgStructure: OrgStructure{Mode: orgmode.Legacy}}, nil
	}
	levels, err := levelRepo.New(s.db.WithContext(ctx)).ListLive()
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load organisation levels.")
	}
	list := []Level{}
	for _, l := range levels {
		if l.IsActive {
			list = append(list, toLevel(l))
		}
	}
	return &TenantConfig{OrgStructure: OrgStructure{Mode: orgmode.Flexible, MaxLevels: orgmode.MaxLevels, Levels: &list}}, nil
}

func toLevel(l models.OrgLevel) Level {
	return Level{ID: l.ID.String(), Level: l.Level, Name: l.Name, NameKey: l.NameKey}
}
