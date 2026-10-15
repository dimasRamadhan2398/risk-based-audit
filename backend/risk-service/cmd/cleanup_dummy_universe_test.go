package main

import (
	"testing"

	"github.com/google/uuid"
)

// fixtureUniverse mirrors the shape of standard_audit_universe on the prod
// database before the boot-time clean-up ran: the 8 legacy nodes, their 6
// department children under PT A, and master-synced / user-added rows that must
// survive any clean-up.
func fixtureUniverse() (nodes []universeNode, legacyIDs map[string]uuid.UUID, realIDs map[string]uuid.UUID) {
	legacyIDs = map[string]uuid.UUID{}
	realIDs = map[string]uuid.UUID{}
	add := func(into map[string]uuid.UUID, name string, parent *uuid.UUID) uuid.UUID {
		id := uuid.New()
		into[name] = id
		nodes = append(nodes, universeNode{ID: id, Name: name, ParentID: parent})
		return id
	}
	ptrTo := func(id uuid.UUID) *uuid.UUID { return &id }

	subs := add(realIDs, "Subsidiaries/Business Unit", nil)
	branches := add(realIDs, "Branches/Regional Offices", nil)
	products := add(realIDs, "Products/Services", nil)

	ptA := add(legacyIDs, "PT A", ptrTo(subs))
	add(legacyIDs, "PT B", ptrTo(subs))
	add(legacyIDs, "PT C", ptrTo(subs))
	add(legacyIDs, "Jakarta", ptrTo(branches))
	add(legacyIDs, "Surabaya", ptrTo(branches))
	add(legacyIDs, "Product/Service A", ptrTo(products))
	add(legacyIDs, "Product/Service B", ptrTo(products))
	add(legacyIDs, "Product/Service C", ptrTo(products))
	for _, dept := range []string{"Finance", "HR", "IT", "Operations", "Procurement", "Sales & Marketing"} {
		add(legacyIDs, dept, ptrTo(ptA))
	}

	// A master-synced company that is not on the legacy list.
	add(realIDs, "Head Office", ptrTo(branches))
	return nodes, legacyIDs, realIDs
}

func idSet(nodes []universeNode) map[uuid.UUID]bool {
	out := make(map[uuid.UUID]bool, len(nodes))
	for _, n := range nodes {
		out[n.ID] = true
	}
	return out
}

func TestPlanLegacyDummyCleanup_RemovesExactlyTheFourteenLegacyRows(t *testing.T) {
	nodes, legacy, real := fixtureUniverse()

	plan := planLegacyDummyCleanup(nodes, legacyDummyNames)

	if got := len(plan.Named); got != 8 {
		t.Fatalf("named nodes: want 8, got %d", got)
	}
	if got := len(plan.Children); got != 6 {
		t.Fatalf("children: want 6, got %d", got)
	}
	planned := idSet(append(append([]universeNode{}, plan.Named...), plan.Children...))
	for name, id := range legacy {
		if !planned[id] {
			t.Errorf("legacy row %q was not planned for removal", name)
		}
	}
	for name, id := range real {
		if planned[id] {
			t.Errorf("real row %q must not be planned for removal", name)
		}
	}
}

func TestPlanLegacyDummyCleanup_NoLegacyRowsIsNoOp(t *testing.T) {
	nodes, _, _ := fixtureUniverse()
	// Keep only the real rows: a database that was never seeded with dummies.
	var real []universeNode
	for _, n := range nodes {
		if !contains(legacyDummyNames, n.Name) {
			real = append(real, n)
		}
	}

	plan := planLegacyDummyCleanup(real, legacyDummyNames)

	if len(plan.Named) != 0 || len(plan.Children) != 0 {
		t.Fatalf("expected empty plan, got named=%d children=%d", len(plan.Named), len(plan.Children))
	}
}

func TestPlanLegacyDummyCleanup_ChildrenAreOneLevelOnly(t *testing.T) {
	// A grandchild of a legacy node is not removed, matching the original
	// one-level delete. Only direct children are in scope.
	root := uuid.New()
	legacyNode := uuid.New()
	child := uuid.New()
	grandchild := uuid.New()
	nodes := []universeNode{
		{ID: root, Name: "Subsidiaries/Business Unit"},
		{ID: legacyNode, Name: "PT A", ParentID: &root},
		{ID: child, Name: "Finance", ParentID: &legacyNode},
		{ID: grandchild, Name: "Payroll", ParentID: &child},
	}

	plan := planLegacyDummyCleanup(nodes, legacyDummyNames)

	planned := idSet(append(append([]universeNode{}, plan.Named...), plan.Children...))
	if planned[grandchild] {
		t.Fatal("grandchild of a legacy node must not be planned by the one-level clean-up")
	}
	if !planned[child] {
		t.Fatal("direct child of a legacy node must be planned")
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
