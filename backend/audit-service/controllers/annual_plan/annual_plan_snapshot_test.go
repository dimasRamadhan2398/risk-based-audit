package annualplan_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"audit-service/models"
)

const (
	snapDeptID   = "3a2d190a-7fb3-4c54-ad9f-787c5014ecf6"
	snapUnivID   = "6cf62026-3096-4b36-9750-9f03aa2d6618"
	snapOtherDep = "7b3e301b-8ec4-4d65-b0a0-898d6125e7a0"
)

// The PKAT form does not send riskName, riskLevel or department (see
// frontend/utils/annualPlanPayload.ts), so a create or an update must fill them
// from the lookup, and keep them when a row is saved again.
func TestSnapshotsFilledOnCreateAndKeptOnUpdate(t *testing.T) {
	r, db := activityRouter(t)
	if err := db.Create(&models.AnnualSnapshotRef{Kind: models.SnapshotRefDepartment, RefID: snapDeptID, Name: "Information Technology"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.AnnualSnapshotRef{Kind: models.SnapshotRefEntity, RefID: snapUnivID, Name: "Pengadaan TI"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.AnnualSnapshotRef{Kind: models.SnapshotRefYearScore, RefID: snapUnivID, Year: 2026, RiskLevel: "High"}).Error; err != nil {
		t.Fatal(err)
	}

	// Body as the frontend sends it: ids, no snapshot columns.
	post := `{"year":2026,"title":"Rencana Audit 2026","activities":[{"itemNumber":"1","name":"Audit TI","category":"Assurance","departmentId":"` + snapDeptID + `","auditUniverseId":"` + snapUnivID + `","timelineText":"Q2","auditorCount":2,"totalMandays":20,"supervisorName":"Budi"}]}`
	w := doJSON(t, r, http.MethodPost, "/annual-audit-plans", post)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST: %d %s", w.Code, w.Body.String())
	}
	id, _ := dataOf(t, w)["id"].(string)

	row := firstRow(t, getPlan(t, r, id))
	if row["department"] != "Information Technology" || row["riskName"] != "Pengadaan TI" || row["riskLevel"] != "High" {
		t.Fatalf("POST did not fill snapshots: department=%v riskName=%v riskLevel=%v", row["department"], row["riskName"], row["riskLevel"])
	}

	// The lookup goes away (for example the event sync has not caught up). A
	// save that keeps the same ids must still keep the snapshots.
	if err := db.Where("1 = 1").Delete(&models.AnnualSnapshotRef{}).Error; err != nil {
		t.Fatal(err)
	}
	put := rowBody(snapDeptID, snapUnivID, row["id"].(string))
	w = doJSON(t, r, http.MethodPut, "/annual-audit-plans/"+id, `{"activities":[`+put+`]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT: %d %s", w.Code, w.Body.String())
	}
	row = firstRow(t, getPlan(t, r, id))
	if row["department"] != "Information Technology" || row["riskName"] != "Pengadaan TI" || row["riskLevel"] != "High" {
		t.Fatalf("PUT blanked a kept snapshot: department=%v riskName=%v riskLevel=%v", row["department"], row["riskName"], row["riskLevel"])
	}

	// The row now points at another department the lookup does not know. Its
	// old name must not stay with the new id.
	put = rowBody(snapOtherDep, snapUnivID, row["id"].(string))
	w = doJSON(t, r, http.MethodPut, "/annual-audit-plans/"+id, `{"activities":[`+put+`]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT: %d %s", w.Code, w.Body.String())
	}
	row = firstRow(t, getPlan(t, r, id))
	if row["department"] != "" {
		t.Fatalf("old department name kept for a new departmentId: %v", row["department"])
	}
	if row["riskName"] != "Pengadaan TI" {
		t.Fatalf("riskName of an unchanged universe id lost: %v", row["riskName"])
	}
}

func rowBody(deptID, univID, rowID string) string {
	b, _ := json.Marshal(map[string]interface{}{
		"id":              rowID,
		"itemNumber":      "1",
		"name":            "Audit TI",
		"category":        "Assurance",
		"departmentId":    deptID,
		"auditUniverseId": univID,
		"timelineText":    "Q2",
		"auditorCount":    2,
		"totalMandays":    20,
		"supervisorName":  "Budi",
	})
	return string(b)
}

func firstRow(t *testing.T, plan map[string]interface{}) map[string]interface{} {
	t.Helper()
	acts, _ := plan["activities"].([]interface{})
	if len(acts) != 1 {
		t.Fatalf("want 1 activity, got %d", len(acts))
	}
	return acts[0].(map[string]interface{})
}
