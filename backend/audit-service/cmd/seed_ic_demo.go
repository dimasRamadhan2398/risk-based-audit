package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"audit-service/models"
	"audit-service/pkg/database"
	"audit-service/pkg/logger"
	"audit-service/pkg/masterclient"
	"audit-service/pkg/servicetoken"
	icsvc "audit-service/services/ic_assessment"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	seedICDemoYear       int
	seedICDemoDepartment string
)

// seed-ic-demo creates ONE approved COSO internal control assessment for the
// demo walkthrough. It is a separate command, not part of "audit seed" or
// "audit migrate up": "audit seed" also runs on some prod compose starts, and
// demo assessments must never reach a real tenant. It belongs in the
// --with-demo-data path of onboarding, after "master seed" (it needs the
// master departments) and after "risk seed" (tagged RCM rows give the CA
// signal data).
func init() {
	cmd := &cobra.Command{
		Use:   "seed-ic-demo",
		Short: "Demo data only: create one approved internal control (COSO 2013) assessment",
		Long: `Creates one APPROVED internal control assessment (department level, annual)
for the demo walkthrough, through the normal workflow: create, propose (the
signals are read from this stack's master-, auth- and risk-service with a
self-signed service token, or ` + servicetoken.OverrideEnv + `), principle and
component ratings, submit by a demo auditor, approve by a demo CAE.

Demo data only. Never run it on a production tenant.

Idempotent: when the department already has an assessment for the year,
nothing is written. When master-service has no matching department the
command logs a warning and exits 0, so onboarding does not fail.

Usage: audit seed-ic-demo [--year 2026] [--department DEPT-001]`,
		RunE: runSeedICDemo,
	}
	cmd.Flags().IntVar(&seedICDemoYear, "year", 0, "Assessment year (default: current year)")
	cmd.Flags().StringVar(&seedICDemoDepartment, "department", "DEPT-001", "Master department code (or name) to assess")
	rootCmd.AddCommand(cmd)
}

func runSeedICDemo(cmd *cobra.Command, args []string) error {
	initLogger()
	defer logger.Sync()

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}
	if err := db.AutoMigrate(models.ICAssessmentModels()...); err != nil {
		return err
	}
	if err := migrateICAssessment(db); err != nil {
		return err
	}
	authorization, source, err := servicetoken.Authorization(os.Getenv(servicetoken.OverrideEnv), cfg.JWT.Secret, time.Now())
	if err != nil {
		return err
	}
	logger.Info("IC demo seed: reading sources", logger.LogField("token", source))
	year := seedICDemoYear
	if year == 0 {
		year = time.Now().Year()
	}
	created, msg, err := seedICDemoAssessment(context.Background(), db, icsvc.NewHTTPSources(authorization), authorization, year, seedICDemoDepartment)
	if err != nil {
		logger.Error("IC demo seed failed", logger.LogField("error", err))
		return err
	}
	if created {
		logger.Info("IC demo seed: " + msg)
	} else {
		logger.Warn("IC demo seed skipped: " + msg)
	}
	return nil
}

// seedICDemoAssessment runs the demo assessment through the workflow.
// created is false (with the reason) when it skipped.
func seedICDemoAssessment(ctx context.Context, db *gorm.DB, src icsvc.Sources, authorization string, year int, deptRef string) (bool, string, error) {
	depts, err := masterclient.GetClient().ListDepartments(ctx, authorization)
	if err != nil {
		return false, "master-service departments not readable (" + err.Error() + "); run after master seed", nil
	}
	ref := strings.TrimSpace(deptRef)
	var deptID, deptName string
	for _, d := range depts {
		if strings.EqualFold(strings.TrimSpace(d.DepartmentCode), ref) || strings.EqualFold(strings.TrimSpace(d.DepartmentName), ref) {
			deptID, deptName = d.ID, d.DepartmentName
			break
		}
	}
	if deptID == "" {
		return false, fmt.Sprintf("no master department %q; run after master seed", ref), nil
	}
	return seedICDemoForDepartment(ctx, db, src, deptID, deptName, year)
}

func strp(s string) *string { return &s }
func intp(n int) *int       { return &n }

func seedICDemoForDepartment(ctx context.Context, db *gorm.DB, src icsvc.Sources, deptID, deptName string, year int) (bool, string, error) {
	var n int64
	if err := db.Model(&models.ICAssessment{}).Where("department_id = ? AND year = ?", deptID, year).Count(&n).Error; err != nil {
		return false, "", err
	}
	if n > 0 {
		return false, fmt.Sprintf("%s already has an assessment for %d", deptName, year), nil
	}
	s := icsvc.New(db)
	auditor := icsvc.Actor{UserID: "demo-auditor", Name: "Demo Auditor", Roles: []string{"AUDITOR"}}
	cae := icsvc.Actor{UserID: "demo-cae", Name: "Demo Chief Audit Executive", Roles: []string{"EXECUTIVE"}}

	d, err := s.Create(ctx, src, auditor, icsvc.CreateRequest{DepartmentID: deptID, Year: year, Quarter: 0})
	if err != nil {
		return false, "", fmt.Errorf("create: %w", err)
	}
	if d, err = s.Propose(ctx, src, auditor, d.ID, icsvc.ProposeRequest{Version: d.Version}); err != nil {
		return false, "", fmt.Errorf("propose: %w", err)
	}

	// Demo profile: RA and MA partially effective (one deficiency each), the
	// rest present and functioning
	partial := map[string]string{
		"P07": "Demo data: two key risks were not reassessed in Q2 after the system change.",
		"P17": "Demo data: three follow-up actions were closed after their due date.",
	}
	var principles []icsvc.PrincipleInput
	for _, p := range icsvc.Principles {
		in := icsvc.PrincipleInput{PrincipleCode: p.Code, Status: strp(icsvc.PrincipleYes), Severity: strp(icsvc.SeverityNone)}
		if r, ok := partial[p.Code]; ok {
			in.Status, in.Severity, in.Rationale = strp(icsvc.PrinciplePartial), strp(icsvc.SeverityDeficiency), strp(r)
		}
		principles = append(principles, in)
	}
	final := map[string][3]interface{}{
		icsvc.ComponentCE: {4, icsvc.PrincipleYes, icsvc.SeverityNone},
		icsvc.ComponentRA: {3, icsvc.PrinciplePartial, icsvc.SeverityDeficiency},
		icsvc.ComponentCA: {4, icsvc.PrincipleYes, icsvc.SeverityNone},
		icsvc.ComponentIC: {4, icsvc.PrincipleYes, icsvc.SeverityNone},
		icsvc.ComponentMA: {3, icsvc.PrinciplePartial, icsvc.SeverityDeficiency},
	}
	var components []icsvc.ComponentInput
	for _, c := range d.Components {
		f := final[c.ComponentCode]
		in := icsvc.ComponentInput{ComponentCode: c.ComponentCode, FinalRating: intp(f[0].(int)),
			FinalStatus: strp(f[1].(string)), FinalSeverity: strp(f[2].(string)),
			Conclusion: strp("Demo data: component conclusion for the walkthrough.")}
		if c.ProposedRating == nil || *c.ProposedRating != f[0].(int) {
			in.OverrideReason = strp("Demo data: rating set by auditor judgement after walkthrough interviews.")
		}
		components = append(components, in)
	}
	if d, err = s.Update(ctx, auditor, d.ID, icsvc.UpdateRequest{
		Version: d.Version, Principles: principles, Components: components,
		FinalConclusion:         strp(icsvc.ConclusionEffective),
		ConclusionText:          strp("Demo data: internal control is effective; two deficiencies are being remediated."),
		ConclusionJustification: strp("Demo data: the deficiencies in RA and MA are not significant, alone or together."),
	}); err != nil {
		return false, "", fmt.Errorf("update: %w", err)
	}
	if d, err = s.Submit(ctx, auditor, d.ID, icsvc.TransitionRequest{Version: d.Version}); err != nil {
		return false, "", fmt.Errorf("submit: %w", err)
	}
	ok := true
	if d, err = s.Approve(ctx, cae, d.ID, icsvc.ApproveRequest{Version: d.Version, OperateTogether: &ok,
		OperateTogetherNote: "Demo data", ApprovalNote: "Demo data"}); err != nil {
		return false, "", fmt.Errorf("approve: %w", err)
	}
	return true, fmt.Sprintf("approved assessment %s for %s %d", d.ID, deptName, year), nil
}
