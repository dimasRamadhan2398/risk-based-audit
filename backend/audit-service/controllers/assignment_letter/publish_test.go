package assignmentletter

import "testing"

func TestCanPublish(t *testing.T) {
	cases := []struct {
		roles []string
		want  bool
	}{
		{[]string{"ADMIN"}, true},
		{[]string{"admin"}, true},
		// The CAE: seeded as EXECUTIVE; CHIEF_AUDIT_EXECUTIVE when an admin creates it
		{[]string{"EXECUTIVE"}, true},
		{[]string{"executive"}, true},
		{[]string{"CHIEF_AUDIT_EXECUTIVE"}, true},
		{[]string{"Chief Audit Executive"}, true},
		{[]string{"CAE"}, true},
		{[]string{"AUDIT_MANAGER"}, true},
		{[]string{"Audit Manager"}, true},
		{[]string{"AUDITOR", "AUDIT_MANAGER"}, true},
		// Team leaders and auditors draft letters but do not publish them
		{[]string{"AUDITOR"}, false},
		{[]string{"LEAD_AUDITOR"}, false},
		{[]string{"TEAM_LEADER"}, false},
		{[]string{"MANAGER"}, false},
		{[]string{"VIEWER"}, false},
		{[]string{"AUDITEE"}, false},
		{[]string{"DEPARTMENT_HEAD"}, false},
		{nil, false},
	}
	for _, tc := range cases {
		if got := CanPublish(tc.roles); got != tc.want {
			t.Errorf("CanPublish(%q) = %v, want %v", tc.roles, got, tc.want)
		}
	}
}
