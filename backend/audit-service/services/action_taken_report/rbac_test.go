package action_taken_report

import (
	"reflect"
	"testing"

	"audit-service/models"
)

func TestCanonicalRole(t *testing.T) {
	cases := map[string]string{
		"ADMIN": RoleAdmin, "admin": RoleAdmin, "Super Admin": RoleAdmin, "administrator": RoleAdmin,
		"CHIEF_AUDIT_EXECUTIVE": RoleCAE, "Chief Audit Executive": RoleCAE, "cae": RoleCAE,
		"EXECUTIVE": RoleCAE, "Kepala SPI": RoleCAE, "Head of SKAI": RoleCAE,
		"AUDIT_MANAGER": RoleAuditManager, "Audit Manager": RoleAuditManager, "audit-manager": RoleAuditManager,
		"AuditManager": RoleAuditManager, "Manager Audit": RoleAuditManager,
		"AUDITOR": RoleAuditor, "audit_staff": RoleAuditor, "Lead Auditor": RoleAuditor,
		"AUDITEE": RoleAuditee, "Department Head": RoleDeptHead, "viewer": RoleViewer,
		// Not an audit manager on the server
		"MANAGER":       "MANAGER",
		"Finance Staff": "FINANCE_STAFF",
	}
	for in, want := range cases {
		if got := CanonicalRole(in); got != want {
			t.Errorf("CanonicalRole(%q) = %q, want %q", in, got, want)
		}
	}
}

// The role matrix agreed with the product owner
func TestRoleMatrix(t *testing.T) {
	cases := []struct {
		roles []string
		want  []string
	}{
		{[]string{"ADMIN"}, []string{PermAssign, PermCancel, PermDelete, PermReview, PermUpdate, PermView, PermViewAll}},
		{[]string{"CHIEF_AUDIT_EXECUTIVE"}, []string{PermCancel, PermReview, PermView, PermViewAll}},
		{[]string{"EXECUTIVE"}, []string{PermCancel, PermReview, PermView, PermViewAll}},
		{[]string{"AUDIT_MANAGER"}, []string{PermAssign, PermCancel, PermReview, PermView, PermViewAll}},
		{[]string{"AUDITOR"}, []string{PermAssign, PermReview, PermUpdate, PermView, PermViewAll}},
		{[]string{"AUDITEE"}, []string{PermUpdate, PermView}},
		{[]string{"DEPARTMENT_HEAD"}, []string{PermUpdate, PermView}},
		{[]string{"VIEWER"}, []string{PermView}},
		{[]string{"MANAGER"}, []string{PermView}},
		{nil, []string{PermView}},
		{[]string{"AUDITEE", "AUDITOR"}, []string{PermAssign, PermReview, PermUpdate, PermView, PermViewAll}},
	}
	for _, tc := range cases {
		if got := PermissionsFor(tc.roles).List(); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("PermissionsFor(%v) = %v, want %v", tc.roles, got, tc.want)
		}
	}
}

func TestActorRules(t *testing.T) {
	pic := "11111111-1111-1111-1111-111111111111"
	atr := &models.ActionTakenReport{PICUserID: pic, Status: models.ATRStatusPendingReview, ActionPlan: "x"}

	auditorPIC := NewActor(pic, "a", []string{"AUDITOR"})
	if auditorPIC.CanReview(atr) {
		t.Fatal("an auditor may not review an ATR they are the PIC of")
	}
	adminPIC := NewActor(pic, "a", []string{"ADMIN"})
	if !adminPIC.CanReview(atr) {
		t.Fatal("admin may review even as PIC")
	}
	other := NewActor("22222222-2222-2222-2222-222222222222", "b", []string{"AUDITEE"})
	if other.CanView(atr) || other.CanActAsPIC(atr) {
		t.Fatal("an auditee who is not the PIC sees/acts on the ATR")
	}
	// Empty ids never match
	noID := NewActor("", "c", []string{"AUDITEE"})
	if noID.IsPIC(&models.ActionTakenReport{}) {
		t.Fatal("empty user id matched an unassigned ATR")
	}
	// Case-insensitive id compare
	upper := NewActor("11111111-1111-1111-1111-111111111111", "d", nil)
	if !upper.IsPIC(&models.ActionTakenReport{PICUserID: "11111111-1111-1111-1111-111111111111"}) {
		t.Fatal("PIC not recognised")
	}
}
