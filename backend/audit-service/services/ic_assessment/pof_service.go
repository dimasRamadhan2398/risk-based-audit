package ic_assessment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/riskclient"
	atrsvc "audit-service/services/action_taken_report"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QUANTITATIVE_POF service: creation snapshot, reading the scored tree, writing
// values (docs/coso-pof-assessment-api.md sections 3-5).

const maxTextLen = 4000

// ComponentNames are the English component names
var ComponentNames = map[string]string{
	"CE": "Control Environment", "RA": "Risk Assessment", "CA": "Control Activities",
	"IC": "Information & Communication", "MA": "Monitoring Activities",
}

// ---------------------------------------------------------------------------
// Creation
// ---------------------------------------------------------------------------

// pofCreateSetup carries what Create needs to build a QUANTITATIVE_POF assessment
type pofCreateSetup struct {
	Catalogue  *riskclient.PofCatalogue
	Thresholds Thresholds
	Owners     map[string]string
}

// preparePoFCreate reads and validates the catalogue, the thresholds and the
// owner overrides before the creating transaction starts
func (s *Service) preparePoFCreate(ctx context.Context, src Sources) (*pofCreateSetup, error) {
	cat, err := src.PofCatalogue(ctx, "")
	if err != nil {
		return nil, unavailablef("Could not read the COSO PoF catalogue from risk-service (run `risk up` so the catalogue exists)")
	}
	if issues := ValidateCatalogue(cat); len(issues) > 0 {
		return nil, invalidList("The COSO PoF catalogue of risk-service breaks a weight invariant", issues)
	}
	th, _, err := LoadPofThresholds(s.DB)
	if err != nil {
		return nil, err
	}
	owners, err := ownerOverrideMap(s.DB)
	if err != nil {
		return nil, err
	}
	return &pofCreateSetup{Catalogue: cat, Thresholds: th, Owners: owners}, nil
}

func (p *pofCreateSetup) applyHeader(a *models.ICAssessment) {
	y, pt := p.Thresholds.Yes, p.Thresholds.Partial
	a.CatalogVersion = p.Catalogue.Version
	a.PofYesThreshold, a.PofPartialThreshold = &y, &pt
}

func (p *pofCreateSetup) insertRows(tx *gorm.DB, id uuid.UUID, now time.Time) (int, int, error) {
	pofs, els := MaterializeRows(id, p.Catalogue, p.Owners, now)
	if err := tx.CreateInBatches(&pofs, 100).Error; err != nil {
		return 0, 0, err
	}
	if err := tx.CreateInBatches(&els, 100).Error; err != nil {
		return 0, 0, err
	}
	return len(pofs), len(els), nil
}

// ---------------------------------------------------------------------------
// Loading
// ---------------------------------------------------------------------------

func thresholdsOf(a *models.ICAssessment) Thresholds {
	th := DefaultThresholds()
	if a.PofYesThreshold != nil && a.PofPartialThreshold != nil {
		cand := Thresholds{Yes: *a.PofYesThreshold, Partial: *a.PofPartialThreshold}
		if cand.Validate() == nil {
			th = cand
		}
	}
	return th
}

func loadPoFRows(db *gorm.DB, id uuid.UUID) ([]models.ICAssessmentPoF, []models.ICAssessmentElement, error) {
	var pofs []models.ICAssessmentPoF
	if err := db.Where("assessment_id = ?", id).Order("sort_order, pof_code").Find(&pofs).Error; err != nil {
		return nil, nil, err
	}
	var els []models.ICAssessmentElement
	if err := db.Where("assessment_id = ?", id).Order("sort_order, element_code").Find(&els).Error; err != nil {
		return nil, nil, err
	}
	return pofs, els, nil
}

func calcFor(db *gorm.DB, a *models.ICAssessment) (*TreeCalc, []models.ICAssessmentPoF, []models.ICAssessmentElement, error) {
	pofs, els, err := loadPoFRows(db, a.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	return Calculate(pofs, els, thresholdsOf(a)), pofs, els, nil
}

// quantSubmitIssues syncs the derived rows (in memory) and lists what blocks a submit
func quantSubmitIssues(db *gorm.DB, a *models.ICAssessment) ([]string, error) {
	tree, _, _, err := calcFor(db, a)
	if err != nil {
		return nil, err
	}
	SyncDerivedRows(a, tree)
	issues := ValidateQuantForSubmit(a, tree)
	if issues == nil {
		issues = []string{}
	}
	return issues, nil
}

// updateQuantConclusion handles PUT /:id of a QUANTITATIVE_POF assessment: only
// the conclusion fields can change (principle and component rows are derived)
func (s *Service) updateQuantConclusion(actor Actor, id uuid.UUID, req UpdateRequest) (*Detail, error) {
	if len(req.Components) > 0 || len(req.Principles) > 0 {
		return nil, methodMismatch(MethodQuantitativePoF, "principle and component rows are derived from the PoF scores; send only final_conclusion, conclusion_text and conclusion_justification")
	}
	now := s.now()
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if !isEditable(a.Status) {
			return conflictf("INVALID_STATUS", "Only a DRAFT or REOPENED assessment can be changed (current status: %s)", a.Status)
		}
		if err := checkVersion(a, req.Version); err != nil {
			return err
		}
		if err := s.mustBeOpen(tx, a.Year, a.Quarter); err != nil {
			return err
		}
		tree, _, _, err := calcFor(tx, a)
		if err != nil {
			return err
		}
		SyncDerivedRows(a, tree)
		before := map[string]interface{}{"final_conclusion": a.FinalConclusion, "conclusion_text": a.ConclusionText,
			"conclusion_justification": a.ConclusionJustification}
		if req.FinalConclusion != nil {
			a.FinalConclusion = upper(req.FinalConclusion)
		}
		if req.ConclusionText != nil {
			a.ConclusionText = strings.TrimSpace(*req.ConclusionText)
		}
		if req.ConclusionJustification != nil {
			a.ConclusionJustification = strings.TrimSpace(*req.ConclusionJustification)
		}
		if err := ValidateFinalConclusion(a, a.Components, a.Principles, false); err != nil {
			return invalidf("%s", err.Error())
		}
		after := map[string]interface{}{}
		for k, v := range map[string]string{"final_conclusion": a.FinalConclusion, "conclusion_text": a.ConclusionText,
			"conclusion_justification": a.ConclusionJustification} {
			if before[k] != v {
				after[k] = v
			} else {
				delete(before, k)
			}
		}
		if len(after) == 0 {
			return nil
		}
		if err := bump(tx, a, req.Version, []string{StatusDraft, StatusReopened}, map[string]interface{}{
			"final_conclusion": a.FinalConclusion, "conclusion_text": a.ConclusionText,
			"conclusion_justification": a.ConclusionJustification, "derived_conclusion": a.DerivedConclusion,
		}, now); err != nil {
			return err
		}
		return writeEvent(tx, actor, &a.ID, ActionUpdate, a.Status, a.Status, "", before, after, now)
	})
	if err != nil {
		return nil, err
	}
	return s.finish(actor, id)
}

// ---------------------------------------------------------------------------
// Access
// ---------------------------------------------------------------------------

// scopeAccess says which principles / PoFs the caller may see and write
type scopeAccess struct {
	restricted bool
	principles map[string]bool
	pofs       map[string]bool
}

func (a scopeAccess) allows(principle, pof string) bool {
	if !a.restricted {
		return true
	}
	return a.principles[principle] || a.pofs[pof]
}

func assignedScopes(db *gorm.DB, assessmentID uuid.UUID, userID string) (scopeAccess, error) {
	ac := scopeAccess{restricted: true, principles: map[string]bool{}, pofs: map[string]bool{}}
	if strings.TrimSpace(userID) == "" {
		return ac, nil
	}
	var rows []models.ICAssessmentAssignment
	if err := db.Where("assessment_id = ? AND assignee_user_id = ?", assessmentID, userID).Find(&rows).Error; err != nil {
		return ac, err
	}
	for _, r := range rows {
		if r.ScopeType == models.ICScopePrinciple {
			ac.principles[r.ScopeCode] = true
		} else {
			ac.pofs[r.ScopeCode] = true
		}
	}
	return ac, nil
}

// loadQuant loads a QUANTITATIVE_POF assessment for reading and returns the
// caller's scope. 409 METHOD_MISMATCH for a qualitative assessment.
func (s *Service) loadQuant(ctx context.Context, src Sources, actor Actor, id uuid.UUID) (*models.ICAssessment, scopeAccess, error) {
	a, err := loadFull(s.DB, id)
	if err != nil {
		return nil, scopeAccess{}, err
	}
	if !actor.CanRead() {
		return nil, scopeAccess{}, forbiddenf("No role on this account may read internal control assessments")
	}
	full := scopeAccess{}
	if actor.IsScopedReader() {
		cerr := s.checkRead(ctx, src, actor, a)
		if cerr != nil {
			ac, err := assignedScopes(s.DB, id, actor.UserID)
			if err != nil {
				return nil, scopeAccess{}, err
			}
			if len(ac.principles) == 0 && len(ac.pofs) == 0 {
				return nil, scopeAccess{}, cerr
			}
			full = ac
		}
	}
	if a.Method != MethodQuantitativePoF {
		return nil, scopeAccess{}, methodMismatch(a.Method, "this endpoint exists only for the QUANTITATIVE_POF method")
	}
	return a, full, nil
}

// canEdit reports whether the caller may write an item now (state, period,
// assignment and independence). The error explains a refusal.
func (s *Service) canEdit(a *models.ICAssessment, locked bool, actor Actor, ac scopeAccess, principle, pof string, el *models.ICAssessmentElement) error {
	if !isEditable(a.Status) {
		return conflictf("INVALID_STATUS", "Values can only be changed while the assessment is DRAFT or REOPENED (current status: %s)", a.Status)
	}
	if locked {
		return conflictf("PERIOD_LOCKED", "The period %s is locked", periodLabel(a.Year, a.Quarter))
	}
	if !actor.IsAssessor() && !ac.allows(principle, pof) {
		return forbiddenf("You are not an assessor and %s is not assigned to you", pof)
	}
	if !actor.IsAssessor() && !ac.restricted {
		return forbiddenf("Only assessors or the assigned person can change values")
	}
	if el != nil && el.RequiresIndependentAssessor && atrsvc.HasCanonicalRole(actor.Roles, atrsvc.RoleCAE) {
		return forbiddenf("Element %s evaluates the internal audit function and requires an independent assessor: the chief audit executive may not fill it in. Assign it to a QAIP or external assessor", el.ElementCode)
	}
	return nil
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

// AssigneeDTO is one assignment as shown on a principle or PoF
type AssigneeDTO struct {
	ID        uuid.UUID `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	RoleLabel string    `json:"role_label"`
	DueDate   *string   `json:"due_date"`
}

// ElementDTO is one element of a PoF
type ElementDTO struct {
	Code                        string     `json:"code"`
	Type                        string     `json:"type"`
	DescriptionEN               string     `json:"description_en"`
	DescriptionID               string     `json:"description_id"`
	Weight                      float64    `json:"weight"`
	DefaultMethods              []string   `json:"default_methods"`
	Methods                     []string   `json:"methods"`
	ScaleType                   string     `json:"scale_type"`
	RequiresIndependentAssessor bool       `json:"requires_independent_assessor"`
	Achievement                 *float64   `json:"achievement"`
	Score                       *float64   `json:"score"`
	Explanation                 string     `json:"explanation"`
	Recommendation              string     `json:"recommendation"`
	NeedsExplanation            bool       `json:"needs_explanation"`
	EvidenceCount               int        `json:"evidence_count"`
	CanEdit                     bool       `json:"can_edit"`
	UpdatedByName               string     `json:"updated_by_name"`
	UpdatedAt                   *time.Time `json:"updated_at"`
}

// PoFDTO is one point of focus
type PoFDTO struct {
	Code              string        `json:"code"`
	PrincipleCode     string        `json:"principle_code"`
	NameEN            string        `json:"name_en"`
	NameID            string        `json:"name_id"`
	Weight            float64       `json:"weight"`
	Basis             string        `json:"basis"`
	DirectAchievement *float64      `json:"direct_achievement"`
	Achievement       *float64      `json:"achievement"`
	Score             *float64      `json:"score"`
	Status            string        `json:"status"`
	AssessedWeight    float64       `json:"assessed_weight"`
	AssessedScore     float64       `json:"assessed_score"`
	Explanation       string        `json:"explanation"`
	Recommendation    string        `json:"recommendation"`
	Methods           []string      `json:"methods"`
	NeedsExplanation  bool          `json:"needs_explanation"`
	EvidenceCount     int           `json:"evidence_count"`
	DefaultOwnerRoles []string      `json:"default_owner_roles"`
	Assignees         []AssigneeDTO `json:"assignees"`
	Progress          struct {
		ElementsTotal    int `json:"elements_total"`
		ElementsAssessed int `json:"elements_assessed"`
	} `json:"progress"`
	UpdatedByName string       `json:"updated_by_name"`
	UpdatedAt     *time.Time   `json:"updated_at"`
	Elements      []ElementDTO `json:"elements,omitempty"`
	// Evidence and Assignments are filled by GET /:id/pof/:code only
	Evidence    []models.ICPoFEvidence          `json:"evidence,omitempty"`
	Assignments []models.ICAssessmentAssignment `json:"assignments,omitempty"`
}

// PrincipleDTO is one principle with its PoFs
type PrincipleDTO struct {
	Code          string `json:"code"`
	Number        int    `json:"number"`
	NameEN        string `json:"name_en"`
	ComponentCode string `json:"component_code"`
	ScoreBlock
	DefaultOwnerRoles []string      `json:"default_owner_roles"`
	Assignees         []AssigneeDTO `json:"assignees"`
	Pofs              []PoFDTO      `json:"pofs"`
}

// ComponentDTO is one component
type ComponentDTO struct {
	Code   string `json:"code"`
	NameEN string `json:"name_en"`
	ScoreBlock
	Principles []PrincipleDTO `json:"principles"`
}

// TreeDTO is GET /:id/pof
type TreeDTO struct {
	AssessmentID      uuid.UUID      `json:"assessment_id"`
	Method            string         `json:"method"`
	CatalogVersion    string         `json:"catalog_version"`
	Status            string         `json:"status"`
	Version           int            `json:"version"`
	Editable          bool           `json:"editable"`
	Filtered          bool           `json:"filtered"`
	Thresholds        Thresholds     `json:"thresholds"`
	Total             ScoreBlock     `json:"total"`
	DerivedConclusion string         `json:"derived_conclusion"`
	Components        []ComponentDTO `json:"components"`
}

// PoFQuery are the filters of GET /:id/pof
type PoFQuery struct {
	ExpandElements bool
	Component      string
	Principle      string
	Status         string // PoF status filter
	AssigneeMe     bool
	Needs          string // explanation | evidence
}

type dtoCtx struct {
	a        *models.ICAssessment
	actor    Actor
	ac       scopeAccess
	locked   bool
	evidence map[string]int
	assigns  map[string][]models.ICAssessmentAssignment
	svc      *Service
}

func (c *dtoCtx) assignees(scopeType, code string) []AssigneeDTO {
	out := []AssigneeDTO{}
	for _, as := range c.assigns[scopeType+"/"+code] {
		d := AssigneeDTO{ID: as.ID, UserID: as.AssigneeUserID, Name: as.AssigneeName, RoleLabel: as.RoleLabel}
		if as.DueDate != nil {
			s := as.DueDate.Format("2006-01-02")
			d.DueDate = &s
		}
		out = append(out, d)
	}
	return out
}

func tptr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func (c *dtoCtx) pofDTO(pc *PoFCalc, withElements bool) PoFDTO {
	r := pc.Row
	d := PoFDTO{Code: r.PofCode, PrincipleCode: r.PrincipleCode, NameEN: r.NameEn, NameID: r.NameID, Weight: round4(r.Weight),
		Basis: pc.Basis, DirectAchievement: r.Achievement, Status: pc.Status,
		AssessedWeight: round4(pc.AssessedWeight), AssessedScore: round4(pc.AssessedScore),
		Explanation: r.Explanation, Recommendation: r.Recommendation, Methods: SplitMethods(r.Methods),
		EvidenceCount: c.evidence["POF/"+r.PofCode], DefaultOwnerRoles: SplitMethods(r.DefaultOwnerRole),
		Assignees: c.assignees(models.ICScopePoF, r.PofCode), UpdatedByName: r.UpdatedByName, UpdatedAt: tptr(r.UpdatedAt)}
	if pc.Complete {
		a, s := round4(*pc.Achievement), round4(*pc.Score)
		d.Achievement, d.Score = &a, &s
	}
	d.Progress.ElementsTotal, d.Progress.ElementsAssessed = len(pc.Elements), pc.ElementsDone
	switch pc.Basis {
	case BasisPoF:
		d.NeedsExplanation = ExplanationNeeded(r.Achievement, r.Explanation)
	case BasisElements:
		for _, e := range pc.Elements {
			if ExplanationNeeded(e.Achievement, e.Explanation) {
				d.NeedsExplanation = true
			}
		}
	}
	if withElements {
		d.Elements = make([]ElementDTO, 0, len(pc.Elements))
		for _, e := range pc.Elements {
			ed := ElementDTO{Code: e.ElementCode, Type: e.Type, DescriptionEN: e.DescriptionEn, DescriptionID: e.DescriptionID,
				Weight: round4(e.Weight), DefaultMethods: SplitMethods(e.DefaultMethods), Methods: SplitMethods(e.Methods),
				ScaleType: e.ScaleType, RequiresIndependentAssessor: e.RequiresIndependentAssessor, Achievement: e.Achievement,
				Explanation: e.Explanation, Recommendation: e.Recommendation,
				NeedsExplanation: ExplanationNeeded(e.Achievement, e.Explanation), EvidenceCount: c.evidence["ELEMENT/"+e.ElementCode],
				UpdatedByName: e.UpdatedByName, UpdatedAt: tptr(e.UpdatedAt)}
			if e.Achievement != nil {
				sc := round4(e.Weight * *e.Achievement)
				ed.Score = &sc
			}
			ed.CanEdit = c.svc.canEdit(c.a, c.locked, c.actor, c.ac, r.PrincipleCode, r.PofCode, e) == nil
			d.Elements = append(d.Elements, ed)
		}
	}
	return d
}

func (s *Service) newDTOCtx(a *models.ICAssessment, actor Actor, ac scopeAccess) (*dtoCtx, error) {
	locked, err := IsPeriodLocked(s.DB, a.Year, a.Quarter)
	if err != nil {
		return nil, err
	}
	c := &dtoCtx{a: a, actor: actor, ac: ac, locked: locked, evidence: map[string]int{}, assigns: map[string][]models.ICAssessmentAssignment{}, svc: s}
	type cnt struct {
		ScopeType string
		ScopeCode string
		N         int
	}
	var counts []cnt
	if err := s.DB.Model(&models.ICPoFEvidence{}).Select("scope_type, scope_code, count(*) as n").
		Where("assessment_id = ?", a.ID).Group("scope_type, scope_code").Scan(&counts).Error; err != nil {
		return nil, err
	}
	for _, k := range counts {
		c.evidence[k.ScopeType+"/"+k.ScopeCode] = k.N
	}
	var as []models.ICAssessmentAssignment
	if err := s.DB.Where("assessment_id = ?", a.ID).Order("created_at").Find(&as).Error; err != nil {
		return nil, err
	}
	for _, x := range as {
		k := x.ScopeType + "/" + x.ScopeCode
		c.assigns[k] = append(c.assigns[k], x)
	}
	return c, nil
}

func matchStatus(filter, status string) bool {
	return filter == "" || strings.EqualFold(filter, status)
}

// GetTree returns the scored tree (GET /:id/pof)
func (s *Service) GetTree(ctx context.Context, src Sources, actor Actor, id uuid.UUID, q PoFQuery) (*TreeDTO, error) {
	a, ac, err := s.loadQuant(ctx, src, actor, id)
	if err != nil {
		return nil, err
	}
	tree, _, _, err := calcFor(s.DB, a)
	if err != nil {
		return nil, err
	}
	SyncDerivedRows(a, tree)
	dc, err := s.newDTOCtx(a, actor, ac)
	if err != nil {
		return nil, err
	}
	mine := map[string]bool{}
	if q.AssigneeMe {
		mine = map[string]bool{}
		for k, list := range dc.assigns {
			for _, as := range list {
				if SameUser(as.AssigneeUserID, actor.UserID) {
					mine[k] = true
				}
			}
		}
	}
	out := &TreeDTO{AssessmentID: a.ID, Method: a.Method, CatalogVersion: a.CatalogVersion, Status: a.Status, Version: a.Version,
		Editable: isEditable(a.Status) && !dc.locked, Thresholds: tree.Thresholds, Total: tree.Total, DerivedConclusion: a.DerivedConclusion,
		Components: []ComponentDTO{}}
	filtered := q.Component != "" || q.Principle != "" || q.Status != "" || q.AssigneeMe || q.Needs != ""
	out.Filtered = filtered
	for _, cc := range tree.Components {
		if q.Component != "" && !strings.EqualFold(q.Component, cc.Code) {
			continue
		}
		cd := ComponentDTO{Code: cc.Code, NameEN: ComponentNames[cc.Code], ScoreBlock: cc.ScoreBlock, Principles: []PrincipleDTO{}}
		for _, pr := range cc.Principles {
			if q.Principle != "" && !strings.EqualFold(q.Principle, pr.Principle.Code) {
				continue
			}
			pd := PrincipleDTO{Code: pr.Principle.Code, Number: pr.Principle.Number, NameEN: pr.Principle.NameEN, ComponentCode: cc.Code,
				ScoreBlock: pr.ScoreBlock, Assignees: dc.assignees(models.ICScopePrinciple, pr.Principle.Code), Pofs: []PoFDTO{}}
			if len(pr.Pofs) > 0 {
				pd.DefaultOwnerRoles = SplitMethods(pr.Pofs[0].Row.DefaultOwnerRole)
			}
			if pd.DefaultOwnerRoles == nil {
				pd.DefaultOwnerRoles = []string{}
			}
			for _, pc := range pr.Pofs {
				if !ac.allows(pr.Principle.Code, pc.Row.PofCode) {
					continue
				}
				if !matchStatus(q.Status, pc.Status) {
					continue
				}
				if q.AssigneeMe && !mine[models.ICScopePoF+"/"+pc.Row.PofCode] && !mine[models.ICScopePrinciple+"/"+pr.Principle.Code] {
					continue
				}
				dto := dc.pofDTO(pc, q.ExpandElements)
				switch q.Needs {
				case "explanation":
					if !dto.NeedsExplanation {
						continue
					}
				case "evidence":
					if dto.EvidenceCount > 0 || pc.Basis == BasisNone {
						continue
					}
				}
				pd.Pofs = append(pd.Pofs, dto)
			}
			if ac.restricted && len(pd.Pofs) == 0 {
				continue
			}
			if filtered && len(pd.Pofs) == 0 {
				continue
			}
			cd.Principles = append(cd.Principles, pd)
		}
		if (filtered || ac.restricted) && len(cd.Principles) == 0 {
			continue
		}
		out.Components = append(out.Components, cd)
	}
	return out, nil
}

// GetPoF returns one PoF with all its elements, evidence and assignments
func (s *Service) GetPoF(ctx context.Context, src Sources, actor Actor, id uuid.UUID, code string) (*PoFDTO, error) {
	a, ac, err := s.loadQuant(ctx, src, actor, id)
	if err != nil {
		return nil, err
	}
	tree, _, _, err := calcFor(s.DB, a)
	if err != nil {
		return nil, err
	}
	pc := tree.PoF(code)
	if pc == nil || !ac.allows(pc.Row.PrincipleCode, code) {
		return nil, apperrorsNotFound("Point of focus not found")
	}
	dc, err := s.newDTOCtx(a, actor, ac)
	if err != nil {
		return nil, err
	}
	dto := dc.pofDTO(pc, true)
	var ev []models.ICPoFEvidence
	if err := s.DB.Where("assessment_id = ? AND (scope_code = ? OR scope_code LIKE ?)", id, code, code+".%").Order("created_at").Find(&ev).Error; err != nil {
		return nil, err
	}
	dto.Evidence = ev
	if dto.Evidence == nil {
		dto.Evidence = []models.ICPoFEvidence{}
	}
	var as []models.ICAssessmentAssignment
	if err := s.DB.Where("assessment_id = ? AND ((scope_type = ? AND scope_code = ?) OR (scope_type = ? AND scope_code = ?))",
		id, models.ICScopePoF, code, models.ICScopePrinciple, pc.Row.PrincipleCode).Find(&as).Error; err != nil {
		return nil, err
	}
	dto.Assignments = as
	return &dto, nil
}

// Summary is GET /:id/pof-summary
type Summary2 struct {
	Total      ScoreBlock `json:"total"`
	Components []struct {
		Code   string `json:"code"`
		NameEN string `json:"name_en"`
		ScoreBlock
	} `json:"components"`
	Principles []struct {
		Code          string `json:"code"`
		ComponentCode string `json:"component_code"`
		NameEN        string `json:"name_en"`
		ScoreBlock
	} `json:"principles"`
	Progress       SummaryProgress `json:"progress"`
	NeedsAttention []string        `json:"needs_attention"`
}

// SummaryProgress counts PoFs by state
type SummaryProgress struct {
	PofsTotal          int            `json:"pofs_total"`
	PofsNotAssessed    int            `json:"pofs_not_assessed"`
	PofsInProgress     int            `json:"pofs_in_progress"`
	PofsAssessed       int            `json:"pofs_assessed"`
	ByStatus           map[string]int `json:"by_status"`
	ElementsTotal      int            `json:"elements_total"`
	ElementsAssessed   int            `json:"elements_assessed"`
	MissingExplanation int            `json:"missing_explanation"`
}

// GetPoFSummary returns the Lampiran-style roll-up and the progress counts
func (s *Service) GetPoFSummary(ctx context.Context, src Sources, actor Actor, id uuid.UUID) (*Summary2, error) {
	a, ac, err := s.loadQuant(ctx, src, actor, id)
	if err != nil {
		return nil, err
	}
	if ac.restricted {
		return nil, forbiddenf("The summary covers the whole assessment; it is not available for an assignment-only view")
	}
	tree, _, _, err := calcFor(s.DB, a)
	if err != nil {
		return nil, err
	}
	out := &Summary2{Total: tree.Total, NeedsAttention: []string{}}
	out.Progress.ByStatus = map[string]int{PrincipleYes: 0, PrinciplePartial: 0, PrincipleNo: 0}
	for _, cc := range tree.Components {
		out.Components = append(out.Components, struct {
			Code   string `json:"code"`
			NameEN string `json:"name_en"`
			ScoreBlock
		}{cc.Code, ComponentNames[cc.Code], cc.ScoreBlock})
		for _, pr := range cc.Principles {
			out.Principles = append(out.Principles, struct {
				Code          string `json:"code"`
				ComponentCode string `json:"component_code"`
				NameEN        string `json:"name_en"`
				ScoreBlock
			}{pr.Principle.Code, cc.Code, pr.Principle.NameEN, pr.ScoreBlock})
		}
	}
	for _, pc := range tree.AllPoFs() {
		out.Progress.PofsTotal++
		out.Progress.ElementsTotal += len(pc.Elements)
		out.Progress.ElementsAssessed += pc.ElementsDone
		missing := false
		switch pc.Basis {
		case BasisPoF:
			missing = ExplanationNeeded(pc.Row.Achievement, pc.Row.Explanation)
		case BasisElements:
			for _, e := range pc.Elements {
				if ExplanationNeeded(e.Achievement, e.Explanation) {
					missing = true
					out.Progress.MissingExplanation++
				}
			}
		}
		if pc.Basis == BasisPoF && missing {
			out.Progress.MissingExplanation++
		}
		switch pc.Status {
		case NodeNotAssessed:
			out.Progress.PofsNotAssessed++
		case NodeInProgress:
			out.Progress.PofsInProgress++
		default:
			out.Progress.PofsAssessed++
			out.Progress.ByStatus[pc.Status]++
		}
		if pc.Status == PrincipleNo || pc.Status == PrinciplePartial || missing {
			out.NeedsAttention = append(out.NeedsAttention, pc.Row.PofCode)
		}
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Writing values
// ---------------------------------------------------------------------------

// PoFValueInput sets the PoF-level value; omitted fields keep their value
type PoFValueInput struct {
	PofCode        string        `json:"pof_code"`
	Achievement    optionalFloat `json:"achievement"`
	Explanation    *string       `json:"explanation"`
	Recommendation *string       `json:"recommendation"`
	Methods        *[]string     `json:"methods"`
}

// ElementValueInput sets one element's value
type ElementValueInput struct {
	ElementCode    string        `json:"element_code"`
	Achievement    optionalFloat `json:"achievement"`
	Explanation    *string       `json:"explanation"`
	Recommendation *string       `json:"recommendation"`
	Methods        *[]string     `json:"methods"`
}

// ValuesRequest is the body of PUT /:id/pof-values
type ValuesRequest struct {
	Pofs     []PoFValueInput     `json:"pofs"`
	Elements []ElementValueInput `json:"elements"`
}

// MaxValueItems caps one PUT pof-values request
const MaxValueItems = 500

// ValuesResult is the answer of PUT /:id/pof-values
type ValuesResult struct {
	Version  int        `json:"version"`
	Total    ScoreBlock `json:"total"`
	Affected []struct {
		Pof       PoFDTO     `json:"pof"`
		Principle ScoreBlock `json:"principle"`
		Component ScoreBlock `json:"component"`
	} `json:"affected"`
	DerivedConclusion string `json:"derived_conclusion"`
}

func checkText(fields FieldErrors, key string, p *string) {
	if p == nil {
		return
	}
	if utf8.RuneCountInString(*p) > maxTextLen {
		fields[key] = fmt.Sprintf("must be at most %d characters", maxTextLen)
	}
	if strings.ContainsRune(*p, 0) {
		fields[key] = "must not contain NUL characters"
	}
}

func floatEq(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return round(*a, 6) == round(*b, 6)
}

func valMap(achievement *float64, explanation, recommendation, methods string) map[string]interface{} {
	m := map[string]interface{}{"explanation": explanation, "recommendation": recommendation, "methods": methods}
	if achievement != nil {
		m["achievement"] = *achievement
	} else {
		m["achievement"] = nil
	}
	return m
}

// SaveValues validates and applies PoF and element values atomically
func (s *Service) SaveValues(ctx context.Context, src Sources, actor Actor, id uuid.UUID, req ValuesRequest) (*ValuesResult, error) {
	var result *ValuesResult
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		r, err := s.saveValuesTx(tx, ctx, src, actor, id, req, MaxValueItems)
		result = r
		return err
	})
	if err != nil {
		var fe FieldErrors
		if errors.As(err, &fe) {
			return nil, fe
		}
		return nil, err
	}
	return result, nil
}

// saveValuesTx is SaveValues inside a transaction (also used by the import)
func (s *Service) saveValuesTx(tx *gorm.DB, ctx context.Context, src Sources, actor Actor, id uuid.UUID, req ValuesRequest, maxItems int) (*ValuesResult, error) {
	if len(req.Pofs)+len(req.Elements) == 0 {
		return nil, FieldErrors{"body": "send at least one entry in pofs or elements"}
	}
	if len(req.Pofs)+len(req.Elements) > maxItems {
		return nil, FieldErrors{"body": fmt.Sprintf("at most %d entries per request", maxItems)}
	}
	now := s.now()
	var result *ValuesResult
	err := func() error {
		a, ac, err := s.loadQuantTx(tx, ctx, src, actor, id)
		if err != nil {
			return err
		}
		locked, err := IsPeriodLocked(tx, a.Year, a.Quarter)
		if err != nil {
			return err
		}
		pofs, els, err := loadPoFRows(tx, id)
		if err != nil {
			return err
		}
		pofIdx := map[string]int{}
		for i := range pofs {
			pofIdx[pofs[i].PofCode] = i
		}
		elIdx := map[string]int{}
		for i := range els {
			elIdx[els[i].ElementCode] = i
		}
		fields := FieldErrors{}
		before, after := map[string]interface{}{}, map[string]interface{}{}
		changedPoF, changedEl := map[int]bool{}, map[int]bool{}
		seen := map[string]bool{}
		var forbidden error

		apply := func(key, code string, ach optionalFloat, expl, rec *string, methods *[]string,
			curAch **float64, curExpl, curRec, curMeth *string, scale string, pofLevel bool) bool {
			changed := false
			checkText(fields, key+".explanation", expl)
			checkText(fields, key+".recommendation", rec)
			newMeth := *curMeth
			if methods != nil {
				m, err := NormalizeMethods(*methods)
				if err != nil {
					fields[key+".methods"] = err.Error()
				} else {
					newMeth = m
				}
			}
			var newAch *float64 = *curAch
			if ach.Set {
				if ach.Value == nil {
					newAch = nil
				} else {
					if err := ValidateAchievement(*ach.Value, scale, pofLevel); err != nil {
						fields[key+".achievement"] = err.Error()
					} else {
						v := round(*ach.Value, 6)
						newAch = &v
					}
				}
			}
			if _, bad := fields[key+".achievement"]; bad {
				return false
			}
			newExpl, newRec := *curExpl, *curRec
			if expl != nil {
				newExpl = strings.TrimSpace(*expl)
			}
			if rec != nil {
				newRec = strings.TrimSpace(*rec)
			}
			if !floatEq(*curAch, newAch) || newExpl != *curExpl || newRec != *curRec || newMeth != *curMeth {
				changed = true
				before[code] = valMap(*curAch, *curExpl, *curRec, *curMeth)
				after[code] = valMap(newAch, newExpl, newRec, newMeth)
				*curAch, *curExpl, *curRec, *curMeth = newAch, newExpl, newRec, newMeth
			}
			return changed
		}

		for i, in := range req.Elements {
			key := fmt.Sprintf("elements[%d]", i)
			code := strings.TrimSpace(in.ElementCode)
			ei, ok := elIdx[code]
			switch {
			case code == "":
				fields[key+".element_code"] = "is required"
				continue
			case !ok:
				fields[key+".element_code"] = "is not an element of this assessment"
				continue
			case seen[code]:
				fields[key+".element_code"] = "appears more than once"
				continue
			}
			seen[code] = true
			e := &els[ei]
			p := &pofs[pofIdx[e.PofCode]]
			if ferr := s.canEdit(a, locked, actor, ac, p.PrincipleCode, p.PofCode, e); ferr != nil && forbidden == nil {
				forbidden = ferr
				continue
			}
			if apply(key, code, in.Achievement, in.Explanation, in.Recommendation, in.Methods,
				&e.Achievement, &e.Explanation, &e.Recommendation, &e.Methods, e.ScaleType, false) {
				changedEl[ei] = true
				e.UpdatedByUserID, e.UpdatedByName = actor.UserID, actor.Name
			}
		}
		for i, in := range req.Pofs {
			key := fmt.Sprintf("pofs[%d]", i)
			code := strings.TrimSpace(in.PofCode)
			pi, ok := pofIdx[code]
			switch {
			case code == "":
				fields[key+".pof_code"] = "is required"
				continue
			case !ok:
				fields[key+".pof_code"] = "is not a point of focus of this assessment"
				continue
			case seen["P/"+code]:
				fields[key+".pof_code"] = "appears more than once"
				continue
			}
			seen["P/"+code] = true
			p := &pofs[pi]
			if ferr := s.canEdit(a, locked, actor, ac, p.PrincipleCode, p.PofCode, nil); ferr != nil && forbidden == nil {
				forbidden = ferr
				continue
			}
			if apply(key, "pof:"+code, in.Achievement, in.Explanation, in.Recommendation, in.Methods,
				&p.Achievement, &p.Explanation, &p.Recommendation, &p.Methods, ScaleQuestionnaire, true) {
				changedPoF[pi] = true
				p.UpdatedByUserID, p.UpdatedByName = actor.UserID, actor.Name
			}
		}
		if len(fields) > 0 {
			return fields
		}
		if forbidden != nil {
			return forbidden
		}
		// Elements win: a PoF-level value cannot be set while elements have values
		for i, in := range req.Pofs {
			pi := pofIdx[strings.TrimSpace(in.PofCode)]
			if !in.Achievement.Set || in.Achievement.Value == nil {
				continue
			}
			for _, e := range els {
				if e.PofCode == pofs[pi].PofCode && e.Achievement != nil {
					return conflictf("POF_HAS_ELEMENT_SCORES", "%s: pofs[%d] cannot take a PoF-level achievement while its elements have values (the elements decide); clear the elements first", pofs[pi].PofCode, i)
				}
			}
		}
		if len(changedPoF)+len(changedEl) == 0 {
			tree := Calculate(pofs, els, thresholdsOf(a))
			result = s.valuesResult(a, tree, nil, nil)
			return nil
		}
		for i := range changedEl {
			e := els[i]
			if err := tx.Model(&models.ICAssessmentElement{}).Where("id = ?", e.ID).Updates(map[string]interface{}{
				"achievement": e.Achievement, "explanation": e.Explanation, "recommendation": e.Recommendation, "methods": e.Methods,
				"updated_by_user_id": actor.UserID, "updated_by_name": actor.Name, "updated_at": now,
			}).Error; err != nil {
				return err
			}
			els[i].UpdatedAt = now
		}
		for i := range changedPoF {
			p := pofs[i]
			if err := tx.Model(&models.ICAssessmentPoF{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
				"achievement": p.Achievement, "explanation": p.Explanation, "recommendation": p.Recommendation, "methods": p.Methods,
				"updated_by_user_id": actor.UserID, "updated_by_name": actor.Name, "updated_at": now,
			}).Error; err != nil {
				return err
			}
			pofs[i].UpdatedAt = now
		}
		tree := Calculate(pofs, els, thresholdsOf(a))
		if err := persistDerived(tx, a, tree, now); err != nil {
			return err
		}
		if err := writeEvent(tx, actor, &a.ID, ActionPofUpdate, a.Status, a.Status, "", before, after, now); err != nil {
			return err
		}
		touched := map[string]bool{}
		for i := range changedEl {
			touched[els[i].PofCode] = true
		}
		for i := range changedPoF {
			touched[pofs[i].PofCode] = true
		}
		var codes []string
		for c := range touched {
			codes = append(codes, c)
		}
		result = s.valuesResult(a, tree, codes, &dtoCtx{a: a, actor: actor, ac: ac, locked: locked, evidence: map[string]int{}, assigns: map[string][]models.ICAssessmentAssignment{}, svc: s})
		return nil
	}()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) valuesResult(a *models.ICAssessment, tree *TreeCalc, codes []string, dc *dtoCtx) *ValuesResult {
	out := &ValuesResult{Version: a.Version, Total: tree.Total, DerivedConclusion: a.DerivedConclusion}
	sortStrings(codes)
	for _, code := range codes {
		pc := tree.PoF(code)
		if pc == nil || dc == nil {
			continue
		}
		pr := tree.Principle(pc.Row.PrincipleCode)
		comp := tree.Component(pc.Row.ComponentCode)
		item := struct {
			Pof       PoFDTO     `json:"pof"`
			Principle ScoreBlock `json:"principle"`
			Component ScoreBlock `json:"component"`
		}{Pof: dc.pofDTO(pc, false), Principle: pr.ScoreBlock, Component: comp.ScoreBlock}
		out.Affected = append(out.Affected, item)
	}
	return out
}

// persistDerived syncs and saves the principle and component rows and the
// header's derived conclusion (no version bump: value writes are per row)
func persistDerived(tx *gorm.DB, a *models.ICAssessment, tree *TreeCalc, now time.Time) error {
	SyncDerivedRows(a, tree)
	for _, p := range a.Principles {
		if err := tx.Model(&models.ICAssessmentPrinciple{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
			"status": p.Status, "severity": p.Severity, "rationale": p.Rationale, "updated_at": now,
		}).Error; err != nil {
			return err
		}
	}
	if err := saveComponents(tx, a.Components, now); err != nil {
		return err
	}
	return tx.Model(&models.ICAssessment{}).Where("id = ?", a.ID).Updates(map[string]interface{}{
		"derived_conclusion": a.DerivedConclusion, "updated_at": now,
	}).Error
}

// loadQuantTx is loadQuant inside a transaction (write path)
func (s *Service) loadQuantTx(tx *gorm.DB, ctx context.Context, src Sources, actor Actor, id uuid.UUID) (*models.ICAssessment, scopeAccess, error) {
	a, err := loadFull(tx, id)
	if err != nil {
		return nil, scopeAccess{}, err
	}
	if a.Method != MethodQuantitativePoF {
		return nil, scopeAccess{}, methodMismatch(a.Method, "this endpoint exists only for the QUANTITATIVE_POF method")
	}
	ac := scopeAccess{}
	if !actor.IsAssessor() {
		ac, err = assignedScopes(tx, id, actor.UserID)
		if err != nil {
			return nil, scopeAccess{}, err
		}
		if len(ac.principles) == 0 && len(ac.pofs) == 0 {
			return nil, scopeAccess{}, forbiddenf("Only assessors or the assigned person can change values")
		}
	}
	return a, ac, nil
}

func apperrorsNotFound(msg string) error {
	return apperrors.New("NOT_FOUND", msg, 404)
}

func sortStrings(s []string) { sort.Strings(s) }

// optionalFloat tells an omitted JSON field from an explicit null and a number
type optionalFloat struct {
	Set   bool
	Value *float64
}

// UnmarshalJSON implements json.Unmarshaler
func (o *optionalFloat) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return fmt.Errorf("must be a number or null")
	}
	o.Value = &f
	return nil
}
