package ic_assessment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Assignments (PIC per principle or PoF) and evidence of the QUANTITATIVE_POF
// method (docs/coso-pof-assessment-api.md sections 5 and 6).

// scopeOf resolves a code to its principle and PoF codes and scope type
func scopeOf(code string, pofs []models.ICAssessmentPoF, els []models.ICAssessmentElement) (scopeType, principle, pof string, el *models.ICAssessmentElement, ok bool) {
	code = strings.TrimSpace(code)
	if p, found := PrincipleByCode(strings.ToUpper(code)); found {
		return models.ICScopePrinciple, p.Code, "", nil, true
	}
	for i := range pofs {
		if pofs[i].PofCode == code {
			return models.ICScopePoF, pofs[i].PrincipleCode, pofs[i].PofCode, nil, true
		}
	}
	for i := range els {
		if els[i].ElementCode == code {
			for j := range pofs {
				if pofs[j].PofCode == els[i].PofCode {
					return models.ICScopeElement, pofs[j].PrincipleCode, pofs[j].PofCode, &els[i], true
				}
			}
		}
	}
	return "", "", "", nil, false
}

// AssignmentRequest is the body of PUT /:id/pof-assignments
type AssignmentRequest struct {
	ScopeType      string `json:"scope_type"`
	ScopeCode      string `json:"scope_code"`
	AssigneeUserID string `json:"assignee_user_id"`
	AssigneeName   string `json:"assignee_name"`
	RoleLabel      string `json:"role_label"`
	DueDate        string `json:"due_date"`
	Note           string `json:"note"`
}

// ListAssignments returns the assignments of an assessment (assignment-only
// callers see their own)
func (s *Service) ListAssignments(ctx context.Context, src Sources, actor Actor, id uuid.UUID, mine bool) ([]models.ICAssessmentAssignment, error) {
	_, ac, err := s.loadQuant(ctx, src, actor, id)
	if err != nil {
		return nil, err
	}
	q := s.DB.Where("assessment_id = ?", id)
	if mine || ac.restricted {
		q = q.Where("assignee_user_id = ?", actor.UserID)
	}
	out := []models.ICAssessmentAssignment{}
	if err := q.Order("scope_code, created_at").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// PutAssignment assigns a person to a principle or PoF (upsert by scope and person)
func (s *Service) PutAssignment(ctx context.Context, actor Actor, id uuid.UUID, req AssignmentRequest) (*models.ICAssessmentAssignment, error) {
	if !actor.IsAssessor() {
		return nil, forbiddenf("Only assessors can assign a person in charge")
	}
	fields := FieldErrors{}
	scopeType := strings.ToUpper(strings.TrimSpace(req.ScopeType))
	if scopeType != models.ICScopePrinciple && scopeType != models.ICScopePoF {
		fields["scope_type"] = "must be PRINCIPLE or POF"
	}
	code := strings.TrimSpace(req.ScopeCode)
	if scopeType == models.ICScopePrinciple {
		code = strings.ToUpper(code)
	}
	user := strings.TrimSpace(req.AssigneeUserID)
	if user == "" || utf8.RuneCountInString(user) > 64 {
		fields["assignee_user_id"] = "is required (the user id, at most 64 characters)"
	}
	name := strings.TrimSpace(req.AssigneeName)
	if utf8.RuneCountInString(name) > 200 {
		fields["assignee_name"] = "must be at most 200 characters"
	}
	label := strings.ToUpper(strings.TrimSpace(req.RoleLabel))
	if label != "" && !OwnerLabelRe.MatchString(label) {
		fields["role_label"] = "use A-Z, 0-9 and _ (2-40 characters), for example HEAD_OF_IT"
	}
	var due *time.Time
	if d := strings.TrimSpace(req.DueDate); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			fields["due_date"] = "must be YYYY-MM-DD"
		} else {
			due = &t
		}
	}
	note := strings.TrimSpace(req.Note)
	checkText(fields, "note", &note)
	if len(fields) > 0 {
		return nil, fields
	}
	now := s.now()
	var out models.ICAssessmentAssignment
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if a.Method != MethodQuantitativePoF {
			return methodMismatch(a.Method, "assignments exist only for the QUANTITATIVE_POF method")
		}
		if !isEditable(a.Status) {
			return conflictf("INVALID_STATUS", "Assignments can only change while the assessment is DRAFT or REOPENED (current status: %s)", a.Status)
		}
		if err := s.mustBeOpen(tx, a.Year, a.Quarter); err != nil {
			return err
		}
		pofs, els, err := loadPoFRows(tx, id)
		if err != nil {
			return err
		}
		st, _, _, _, ok := scopeOf(code, pofs, els)
		if !ok || st != scopeType {
			return FieldErrors{"scope_code": "is not a " + strings.ToLower(scopeType) + " code of this assessment"}
		}
		row := models.ICAssessmentAssignment{AssessmentID: id, ScopeType: scopeType, ScopeCode: code, AssigneeUserID: user,
			AssigneeName: name, RoleLabel: label, DueDate: due, Note: note, AssignedByUserID: actor.UserID, AssignedByName: actor.Name,
			CreatedAt: now, UpdatedAt: now}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "assessment_id"}, {Name: "scope_type"}, {Name: "scope_code"}, {Name: "assignee_user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"assignee_name", "role_label", "due_date", "note", "assigned_by_user_id", "assigned_by_name", "updated_at"}),
		}).Create(&row).Error; err != nil {
			return err
		}
		if err := tx.Where("assessment_id = ? AND scope_type = ? AND scope_code = ? AND assignee_user_id = ?", id, scopeType, code, user).Take(&out).Error; err != nil {
			return err
		}
		return writeEvent(tx, actor, &id, ActionPofAssign, a.Status, a.Status, "", nil,
			map[string]interface{}{"scope_type": scopeType, "scope_code": code, "assignee_user_id": user, "assignee_name": name, "role_label": label}, now)
	})
	if err != nil {
		var fe FieldErrors
		if errors.As(err, &fe) {
			return nil, fe
		}
		return nil, err
	}
	return &out, nil
}

// DeleteAssignment removes an assignment
func (s *Service) DeleteAssignment(ctx context.Context, actor Actor, id, assignmentID uuid.UUID) error {
	if !actor.IsAssessor() {
		return forbiddenf("Only assessors can change assignments")
	}
	now := s.now()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if a.Method != MethodQuantitativePoF {
			return methodMismatch(a.Method, "assignments exist only for the QUANTITATIVE_POF method")
		}
		if !isEditable(a.Status) {
			return conflictf("INVALID_STATUS", "Assignments can only change while the assessment is DRAFT or REOPENED (current status: %s)", a.Status)
		}
		if err := s.mustBeOpen(tx, a.Year, a.Quarter); err != nil {
			return err
		}
		var row models.ICAssessmentAssignment
		if err := tx.Where("id = ? AND assessment_id = ?", assignmentID, id).Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrorsNotFound("Assignment not found")
			}
			return err
		}
		if err := tx.Delete(&row).Error; err != nil {
			return err
		}
		return writeEvent(tx, actor, &id, ActionPofAssign, a.Status, a.Status, "removed",
			map[string]interface{}{"scope_type": row.ScopeType, "scope_code": row.ScopeCode, "assignee_user_id": row.AssigneeUserID}, nil, now)
	})
}

// ---------------------------------------------------------------------------
// Evidence
// ---------------------------------------------------------------------------

// EvidenceInput is one evidence item to add. Content is empty for a reference.
type EvidenceInput struct {
	ScopeCode   string
	Title       string
	DocNumber   string
	Note        string
	FileName    string
	ContentType string
	Content     []byte
	Source      string
	ImportID    *uuid.UUID
}

// ListPoFEvidence lists the evidence metadata (optionally of one scope)
func (s *Service) ListPoFEvidence(ctx context.Context, src Sources, actor Actor, id uuid.UUID, scopeCode string) ([]models.ICPoFEvidence, error) {
	a, ac, err := s.loadQuant(ctx, src, actor, id)
	if err != nil {
		return nil, err
	}
	_ = a
	q := s.DB.Where("assessment_id = ?", id)
	if sc := strings.TrimSpace(scopeCode); sc != "" {
		q = q.Where("scope_code = ?", sc)
	}
	rows := []models.ICPoFEvidence{}
	if err := q.Order("created_at").Find(&rows).Error; err != nil {
		return nil, err
	}
	if !ac.restricted {
		return rows, nil
	}
	pofs, els, err := loadPoFRows(s.DB, id)
	if err != nil {
		return nil, err
	}
	out := []models.ICPoFEvidence{}
	for _, r := range rows {
		_, pr, pf, _, ok := scopeOf(r.ScopeCode, pofs, els)
		if ok && ac.allows(pr, pf) {
			out = append(out, r)
		}
	}
	return out, nil
}

// AddPoFEvidence attaches a file or reference to a PoF or element. The caller
// (HTTP layer) has already validated and virus-scanned the file.
func (s *Service) AddPoFEvidence(ctx context.Context, actor Actor, id uuid.UUID, in EvidenceInput) (*models.ICPoFEvidence, error) {
	fields := FieldErrors{}
	title := strings.TrimSpace(in.Title)
	doc := strings.TrimSpace(in.DocNumber)
	note := strings.TrimSpace(in.Note)
	if utf8.RuneCountInString(title) > 255 {
		fields["title"] = "must be at most 255 characters"
	}
	if utf8.RuneCountInString(doc) > 100 {
		fields["doc_number"] = "must be at most 100 characters"
	}
	checkText(fields, "note", &note)
	if len(in.Content) == 0 && title == "" && doc == "" {
		fields["title"] = "a reference needs a title or a doc_number"
	}
	if strings.TrimSpace(in.ScopeCode) == "" {
		fields["scope_code"] = "is required (a PoF or element code)"
	}
	if len(fields) > 0 {
		return nil, fields
	}
	now := s.now()
	var out *models.ICPoFEvidence
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if a.Method != MethodQuantitativePoF {
			return methodMismatch(a.Method, "PoF evidence exists only for the QUANTITATIVE_POF method")
		}
		locked, err := IsPeriodLocked(tx, a.Year, a.Quarter)
		if err != nil {
			return err
		}
		pofs, els, err := loadPoFRows(tx, id)
		if err != nil {
			return err
		}
		st, pr, pf, el, ok := scopeOf(in.ScopeCode, pofs, els)
		if !ok || st == models.ICScopePrinciple {
			return FieldErrors{"scope_code": "is not a PoF or element code of this assessment"}
		}
		ac := scopeAccess{}
		if !actor.IsAssessor() {
			if ac, err = assignedScopes(tx, id, actor.UserID); err != nil {
				return err
			}
		}
		if err := s.canEdit(a, locked && in.Source != "IMPORT", actor, ac, pr, pf, el); err != nil {
			return err
		}
		row := models.ICPoFEvidence{AssessmentID: id, ScopeType: st, ScopeCode: strings.TrimSpace(in.ScopeCode), Kind: "REFERENCE",
			Title: title, DocNumber: doc, Note: note, Source: "MANUAL", UploadedByUserID: actor.UserID, UploadedByName: actor.Name, CreatedAt: now}
		if in.Source == "IMPORT" {
			row.Source, row.ImportID = "IMPORT", in.ImportID
		}
		if len(in.Content) > 0 {
			sum := sha256.Sum256(in.Content)
			row.Kind, row.FileName, row.ContentType, row.SizeBytes = "FILE", truncate(in.FileName, 255), truncate(in.ContentType, 100), int64(len(in.Content))
			row.Sha256, row.Content = hex.EncodeToString(sum[:]), in.Content
			if row.Title == "" {
				row.Title = row.FileName
			}
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		out = &row
		return writeEvent(tx, actor, &id, ActionPofEvidence, a.Status, a.Status, "", nil,
			map[string]interface{}{"scope_code": row.ScopeCode, "kind": row.Kind, "title": row.Title, "sha256": row.Sha256}, now)
	})
	if err != nil {
		var fe FieldErrors
		if errors.As(err, &fe) {
			return nil, fe
		}
		return nil, err
	}
	return out, nil
}

// GetPoFEvidenceFile returns one evidence row with its bytes (read access applies)
func (s *Service) GetPoFEvidenceFile(ctx context.Context, src Sources, actor Actor, id, evidenceID uuid.UUID) (*models.ICPoFEvidence, error) {
	_, ac, err := s.loadQuant(ctx, src, actor, id)
	if err != nil {
		return nil, err
	}
	var row models.ICPoFEvidence
	if err := s.DB.Where("id = ? AND assessment_id = ?", evidenceID, id).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrorsNotFound("Evidence not found")
		}
		return nil, err
	}
	if ac.restricted {
		pofs, els, err := loadPoFRows(s.DB, id)
		if err != nil {
			return nil, err
		}
		_, pr, pf, _, ok := scopeOf(row.ScopeCode, pofs, els)
		if !ok || !ac.allows(pr, pf) {
			return nil, apperrorsNotFound("Evidence not found")
		}
	}
	if row.Kind != "FILE" || len(row.Content) == 0 {
		return nil, apperrorsNotFound("This evidence item has no file")
	}
	return &row, nil
}

// DeletePoFEvidence removes an evidence item (soft delete)
func (s *Service) DeletePoFEvidence(ctx context.Context, actor Actor, id, evidenceID uuid.UUID) error {
	now := s.now()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if a.Method != MethodQuantitativePoF {
			return methodMismatch(a.Method, "PoF evidence exists only for the QUANTITATIVE_POF method")
		}
		var row models.ICPoFEvidence
		if err := tx.Where("id = ? AND assessment_id = ?", evidenceID, id).Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrorsNotFound("Evidence not found")
			}
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
		_, pr, pf, el, _ := scopeOf(row.ScopeCode, pofs, els)
		ac := scopeAccess{}
		if !actor.IsAssessor() {
			if ac, err = assignedScopes(tx, id, actor.UserID); err != nil {
				return err
			}
		}
		if err := s.canEdit(a, locked, actor, ac, pr, pf, el); err != nil {
			return err
		}
		if err := tx.Delete(&row).Error; err != nil {
			return err
		}
		return writeEvent(tx, actor, &id, ActionPofEvidence, a.Status, a.Status, "removed",
			map[string]interface{}{"scope_code": row.ScopeCode, "kind": row.Kind, "title": row.Title, "sha256": row.Sha256}, nil, now)
	})
}
