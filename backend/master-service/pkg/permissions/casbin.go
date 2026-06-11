package permissions

import (
	"sync"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type CasbinEnforcer struct {
	enforcer *casbin.SyncedEnforcer
	mu 	   	  sync.RWMutex
}

func NewCasbinEnforcer(db *gorm.DB) (*CasbinEnforcer, error) {
    // store policies in DB instead of CSV (for runtime updates)
    adapter, err := gormadapter.NewAdapterByDB(db)
    if err != nil {
        return nil, err
    }

    enforcer, err := casbin.NewEnforcer("config/rbac_model.conf", adapter)
    if err != nil {
        return nil, err
    }

    // load policies from DB
    if err := enforcer.LoadPolicy(); err != nil {
        return nil, err
    }

    return &CasbinEnforcer{enforcer: &casbin.SyncedEnforcer{Enforcer: enforcer}}, nil
}

func (c *CasbinEnforcer) Enforce(userID, path, method string) (bool, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.enforcer.Enforce(userID, path, method)
}

// AssignRole assigns a role to a user
func (c *CasbinEnforcer) AssignRole(userID, role string) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    _, err := c.enforcer.AddRoleForUser(userID, role)
    return err
}

// RevokeRole removes a role from a user
func (c *CasbinEnforcer) RevokeRole(userID, role string) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    _, err := c.enforcer.DeleteRoleForUser(userID, role)
    return err
}

// GetRoles returns all roles for a user
func (c *CasbinEnforcer) GetRoles(userID string) ([]string, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.enforcer.GetRolesForUser(userID)
}

func (c *CasbinEnforcer) LoadPolicy() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.enforcer.LoadPolicy()
}