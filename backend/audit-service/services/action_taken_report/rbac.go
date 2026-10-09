package action_taken_report

import (
	"sort"
	"strings"

	"audit-service/models"
)

// Permission keys. They match the auth-service permission names (seeded in
// auth-service/pkg/database/seeders) so the frontend can check the same keys.
//
// The JWT only carries role names, so audit-service maps roles to these keys
// itself (RolePermissions); the role/permission rows in auth-service do not
// change what this service allows.
const (
	// PermView: may open ATRs assigned to them (as PIC). Every authenticated
	// user effectively has it for their own ATRs.
	PermView = "view_action_taken_report"
	// PermViewAll: may list and open every ATR
	PermViewAll = "view_all_action_taken_report"
	// PermAssign: may set the PIC and due date
	PermAssign = "assign_action_taken_report"
	// PermUpdate: may fill the action plan, upload evidence and submit, but only
	// on ATRs where they are the PIC. Being assigned as PIC is what grants these
	// actions; the key documents which roles are expected to be PICs.
	PermUpdate = "update_action_taken_report"
	// PermReview: may approve or reject a submitted ATR
	PermReview = "review_action_taken_report"
	// PermCancel: may cancel an open ATR
	PermCancel = "cancel_action_taken_report"
	// PermDelete: may delete an ATR
	PermDelete = "delete_action_taken_report"
)

// AllPermissions lists every ATR permission key
var AllPermissions = []string{PermView, PermViewAll, PermAssign, PermUpdate, PermReview, PermCancel, PermDelete}

// Role names as they appear in the JWT "roles" claim (compared
// case-insensitively). AUDIT_MANAGER and CHIEF_AUDIT_EXECUTIVE are not seeded;
// they apply once an admin creates roles with these names.
const (
	RoleAdmin        = "ADMIN"
	RoleCAE          = "CHIEF_AUDIT_EXECUTIVE"
	RoleAuditManager = "AUDIT_MANAGER"
	RoleAuditor      = "AUDITOR"
	RoleExecutive    = "EXECUTIVE"
	RoleDeptHead     = "DEPARTMENT_HEAD"
	RoleAuditee      = "AUDITEE"
	RoleViewer       = "VIEWER"
)

// RolePermissions is the ATR role matrix. Roles not listed (e.g. VIEWER) have
// no ATR permission and only see ATRs they are the PIC of.
var RolePermissions = map[string][]string{
	RoleAdmin:        AllPermissions,
	RoleCAE:          {PermView, PermViewAll, PermReview, PermCancel},
	RoleAuditManager: {PermView, PermViewAll, PermAssign, PermReview, PermCancel},
	RoleAuditor:      {PermView, PermViewAll, PermAssign, PermReview, PermUpdate},
	RoleExecutive:    {PermView, PermViewAll},
	RoleDeptHead:     {PermView, PermUpdate},
	RoleAuditee:      {PermView, PermUpdate},
}

// roleAliases maps other spellings of the same role, as admins may create
// them, onto the canonical names above
var roleAliases = map[string]string{
	"CAE":                    RoleCAE,
	"HEAD_OF_INTERNAL_AUDIT": RoleCAE,
	"HEAD_OF_SKAI":           RoleCAE,
	"KEPALA_SPI":             RoleCAE,
	"MANAGER_AUDIT":          RoleAuditManager,
	"AUDIT_STAFF":            RoleAuditor,
	"ADMINISTRATOR":          RoleAdmin,
	"SUPERADMIN":             RoleAdmin,
	"SUPER_ADMIN":            RoleAdmin,
}

// CanonicalRole upper-cases a role name, turns spaces/hyphens into
// underscores and resolves aliases ("Audit Manager" -> AUDIT_MANAGER)
func CanonicalRole(role string) string {
	r := strings.ToUpper(strings.TrimSpace(role))
	r = strings.Join(strings.FieldsFunc(r, func(c rune) bool {
		return c == ' ' || c == '-' || c == '_' || c == '\t'
	}), "_")
	if a, ok := roleAliases[r]; ok {
		return a
	}
	return r
}

// Perms is the set of ATR permissions a caller holds
type Perms map[string]bool

// PermissionsFor returns the ATR permissions of a caller with the given roles
func PermissionsFor(roles []string) Perms {
	p := Perms{PermView: true}
	for _, role := range roles {
		for _, perm := range RolePermissions[CanonicalRole(role)] {
			p[perm] = true
		}
	}
	return p
}

// IsAdmin reports whether roles include ADMIN (any spelling)
func IsAdmin(roles []string) bool {
	for _, r := range roles {
		if CanonicalRole(r) == RoleAdmin {
			return true
		}
	}
	return false
}

// List returns the held permission keys, sorted
func (p Perms) List() []string {
	out := make([]string, 0, len(p))
	for k, v := range p {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// Actor is the authenticated caller
type Actor struct {
	UserID   string
	Username string
	Roles    []string
	Perms    Perms
	Admin    bool
}

// NewActor builds the caller from the JWT claims
func NewActor(userID, username string, roles []string) Actor {
	return Actor{UserID: userID, Username: username, Roles: roles, Perms: PermissionsFor(roles), Admin: IsAdmin(roles)}
}

// DisplayName is what is recorded as uploaded_by / reviewed_by
func (a Actor) DisplayName() string {
	if strings.TrimSpace(a.Username) != "" {
		return a.Username
	}
	return a.UserID
}

// IsPIC reports whether the caller is the ATR's assigned PIC
func (a Actor) IsPIC(r *models.ActionTakenReport) bool {
	return a.UserID != "" && r.PICUserID != "" && strings.EqualFold(a.UserID, r.PICUserID)
}

// CanView reports whether the caller may see the ATR
func (a Actor) CanView(r *models.ActionTakenReport) bool {
	return a.Perms[PermViewAll] || a.IsPIC(r)
}

// CanActAsPIC reports whether the caller may fill the action plan, upload
// evidence and submit: the assigned PIC, or an admin
func (a Actor) CanActAsPIC(r *models.ActionTakenReport) bool {
	return a.Admin || a.IsPIC(r)
}

// CanReview reports whether the caller may review: the review permission,
// and (segregation of duties) not the ATR's own PIC unless admin
func (a Actor) CanReview(r *models.ActionTakenReport) bool {
	return a.Perms[PermReview] && (a.Admin || !a.IsPIC(r))
}

// Allowed action names returned per ATR in allowed_actions
const (
	ActionAssign           = "assign"
	ActionUpdateActionPlan = "update_action_plan"
	ActionUploadEvidence   = "upload_evidence"
	ActionSubmit           = "submit"
	ActionReview           = "review"
	ActionCancel           = "cancel"
	ActionDelete           = "delete"
)

// AllowedActions lists what the caller can do with the ATR right now (role
// and state both checked), for the UI to show the matching buttons. The
// endpoints enforce the same rules.
func (a Actor) AllowedActions(r *models.ActionTakenReport) []string {
	out := []string{}
	status := r.Status
	open := !models.IsATRClosed(status)
	editable := status == models.ATRStatusPlanned || status == models.ATRStatusInProgress
	if a.Perms[PermAssign] && open {
		out = append(out, ActionAssign)
	}
	if a.CanActAsPIC(r) && editable {
		out = append(out, ActionUpdateActionPlan, ActionUploadEvidence)
	}
	if a.CanActAsPIC(r) && status == models.ATRStatusInProgress && strings.TrimSpace(r.ActionPlan) != "" {
		out = append(out, ActionSubmit)
	}
	if a.CanReview(r) && status == models.ATRStatusPendingReview {
		out = append(out, ActionReview)
	}
	if a.Perms[PermCancel] && open {
		out = append(out, ActionCancel)
	}
	if a.Perms[PermDelete] {
		out = append(out, ActionDelete)
	}
	return out
}
