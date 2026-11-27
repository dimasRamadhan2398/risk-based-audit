package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LetterMember is one team member of a letter. EmployeeID (master-service
// employee UUID) and EmployeeCode (employees.employee_code, the value stored in
// users.employee_id) are optional: letters saved before they existed carry
// only Name, and the member is then matched to a user by name (see
// services/me_agenda). The members list is a JSON column, so the extra keys
// need no migration.
type LetterMember struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	EmployeeID   string `json:"employeeId,omitempty"`
	EmployeeCode string `json:"employeeCode,omitempty"`
}

type AssignmentLetter struct {
	ID                 uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LetterNumber       string              `gorm:"type:varchar(100);uniqueIndex:idx_assignment_letters_letter_number,where:deleted_at IS NULL;not null" json:"letterNumber"`
	Status             string              `gorm:"type:varchar(50);default:'Draft'" json:"status"`
	AuditTitle         string              `gorm:"type:varchar(255)" json:"auditTitle"`
	Leader             string              `gorm:"type:varchar(200)" json:"leader"`
	// LeaderEmployeeID / LeaderEmployeeCode identify the leader the same way
	// LetterMember does; empty on letters saved before they existed.
	LeaderEmployeeID   string              `gorm:"type:varchar(64)" json:"leaderEmployeeId,omitempty"`
	LeaderEmployeeCode string              `gorm:"type:varchar(64)" json:"leaderEmployeeCode,omitempty"`
	Category           string              `gorm:"type:varchar(100)" json:"category"`
	AuditYear          string              `gorm:"type:varchar(10)" json:"auditYear"`
	AuditTeam          string              `gorm:"type:varchar(100)" json:"auditTeam"`
	StartPeriod        string              `gorm:"type:varchar(100)" json:"startPeriod"`
	FinishPeriod       string              `gorm:"type:varchar(100)" json:"finishPeriod"`
	WorkingUnit        string              `gorm:"type:varchar(255)" json:"workingUnit"`
	// WorkingUnitDepartmentID is the audited unit as a master-service
	// department (soft reference). WorkingUnit stays as its display text
	// ("Name (CODE)"); writes validate the id against the master and rewrite
	// WorkingUnit from it (controllers/assignment_letter.ResolveWorkingUnit).
	// NULL on letters saved before it existed until backfilled with
	// `./audit backfill-department-ids`.
	WorkingUnitDepartmentID *uuid.UUID `gorm:"type:uuid;index" json:"workingUnitDepartmentId"`
	CompanyID          *uuid.UUID          `gorm:"type:uuid;index" json:"companyId,omitempty"`
	CompanyName        string              `gorm:"type:varchar(255)" json:"companyName,omitempty"`
	ExecutionPeriod    string              `gorm:"type:varchar(255)" json:"executionPeriod"`
	AuditPurpose       string              `gorm:"type:text" json:"auditPurpose"`
	LetterDate         *time.Time          `json:"letterDate"`
	CAESignature       string              `gorm:"type:text" json:"caeSignature"`
	MembersList        []LetterMember      `gorm:"serializer:json" json:"membersList"`
	PurposeList        []string            `gorm:"serializer:json" json:"purposeList"`
	ScopeList          []string            `gorm:"serializer:json" json:"scopeList"`
	CcList             []string            `gorm:"serializer:json" json:"ccList"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	DeletedAt          gorm.DeletedAt      `gorm:"index" json:"-"`
}

func (AssignmentLetter) TableName() string {
	return "assignment_letters"
}

func (a *AssignmentLetter) UnmarshalJSON(data []byte) error {
	type Alias AssignmentLetter
	aux := struct {
		LetterDate *string `json:"letterDate"`
		CompanyID  *string `json:"companyId"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.CompanyID != nil {
		cidStr := strings.TrimSpace(*aux.CompanyID)
		if cidStr == "" || cidStr == "null" {
			a.CompanyID = nil
		} else if parsedID, err := uuid.Parse(cidStr); err == nil {
			a.CompanyID = &parsedID
		} else {
			a.CompanyID = nil
		}
	}

	if aux.LetterDate != nil {
		dateStr := strings.TrimSpace(*aux.LetterDate)
		if dateStr == "" || dateStr == "null" {
			a.LetterDate = nil
		} else {
			formats := []string{
				time.RFC3339,
				"2006-01-02T15:04:05Z07:00",
				"2006-01-02T15:04:05",
				"2006-01-02 15:04:05",
				"2006-01-02",
				"02/01/2006",
			}
			var parsed bool
			for _, f := range formats {
				if t, err := time.Parse(f, dateStr); err == nil {
					a.LetterDate = &t
					parsed = true
					break
				}
			}
			if !parsed && len(dateStr) >= 10 {
				if t, err := time.Parse("2006-01-02", dateStr[:10]); err == nil {
					a.LetterDate = &t
					parsed = true
				}
			}
			if !parsed {
				a.LetterDate = nil
			}
		}
	} else {
		a.LetterDate = nil
	}

	return nil
}

// JoinAuditPurposes joins the non-blank entries of a letter's purposeList
// into one line, the form used for a single "audit purpose" value.
func JoinAuditPurposes(purposes []string) string {
	parts := make([]string, 0, len(purposes))
	for _, p := range purposes {
		if p = strings.TrimSpace(p); p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "; ")
}

// EffectiveAuditPurpose is the letter's audit purpose: the auditPurpose
// column when set, otherwise its purposeList. The Assignment Letter form only
// edits purposeList ("Audit Purpose" in the UI), so letters created there
// store an empty audit_purpose.
func (a *AssignmentLetter) EffectiveAuditPurpose() string {
	if p := strings.TrimSpace(a.AuditPurpose); p != "" {
		return p
	}
	return JoinAuditPurposes(a.PurposeList)
}

// AfterFind fills an empty AuditPurpose from PurposeList, so every read
// (list, detail, Working Paper sync, reports) returns the letter's purpose.
// Nothing is written back.
func (a *AssignmentLetter) AfterFind(tx *gorm.DB) error {
	if strings.TrimSpace(a.AuditPurpose) == "" {
		a.AuditPurpose = JoinAuditPurposes(a.PurposeList)
	}
	return nil
}

// FindAssignmentLetter looks a letter up by letter number, or by id when ref
// is a UUID. Working papers and fieldwork store either form.
func FindAssignmentLetter(tx *gorm.DB, ref string) (*AssignmentLetter, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, false
	}
	var letter AssignmentLetter
	if err := tx.Where("letter_number = ?", ref).First(&letter).Error; err == nil {
		return &letter, true
	}
	if _, err := uuid.Parse(ref); err == nil {
		if err := tx.Where("id = ?", ref).First(&letter).Error; err == nil {
			return &letter, true
		}
	}
	return nil, false
}

var letterYearRe = regexp.MustCompile(`\b((?:19|20|21)\d{2})\b`)

func normalizeLetterYear(year string) string {
	trimmed := strings.TrimSpace(year)
	if m := letterYearRe.FindString(trimmed); m != "" {
		return m
	}
	if len(trimmed) >= 4 {
		candidate := trimmed[:4]
		if _, err := strconv.Atoi(candidate); err == nil {
			return candidate
		}
	}
	return fmt.Sprintf("%d", time.Now().Year())
}

func normalizeLetterTeam(team string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(team))
	trimmed = strings.ReplaceAll(trimmed, "/", "-")
	if trimmed == "" {
		return "SKAI"
	}
	return trimmed
}

// parseLetterSequence extracts the sequence integer from a letter number only
// when it belongs to the target audit team and year (ST-{seq}/{TEAM}/{YEAR}).
func parseLetterSequence(letterNumber, targetTeam, targetYear string) (int, bool) {
	parts := strings.Split(strings.TrimSpace(letterNumber), "/")
	if len(parts) != 3 {
		return 0, false
	}
	seqPart := strings.ToUpper(strings.TrimSpace(parts[0]))
	teamPart := strings.ToUpper(strings.TrimSpace(parts[1]))
	yearPart := strings.TrimSpace(parts[2])

	if !strings.HasPrefix(seqPart, "ST-") || teamPart != targetTeam || yearPart != targetYear {
		return 0, false
	}
	numStr := strings.TrimPrefix(seqPart, "ST-")
	if numStr == "" {
		return 0, false
	}
	seq, err := strconv.Atoi(numStr)
	if err != nil || seq <= 0 {
		return 0, false
	}
	return seq, true
}

// legacyLetterNumberUniqueConstraint reports whether the database still has a
// non-partial unique index/constraint on assignment_letters.letter_number that
// covers soft-deleted rows. On migrated databases (partial index WHERE
// deleted_at IS NULL), it returns false so soft-deleted rows do not hijack or
// skip sequential numbers.
func legacyLetterNumberUniqueConstraint(tx *gorm.DB) bool {
	if tx == nil || tx.Dialector == nil {
		return false
	}
	db := tx.Session(&gorm.Session{NewDB: true})
	switch db.Dialector.Name() {
	case "postgres":
		var count int64
		err := db.Raw(`
			SELECT COUNT(*)
			FROM pg_index ix
			JOIN pg_class tbl ON tbl.oid = ix.indrelid
			JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
			WHERE tbl.relname = 'assignment_letters'
			  AND ns.nspname = current_schema()
			  AND ix.indisunique = true
			  AND ix.indpred IS NULL
			  AND EXISTS (
			    SELECT 1 FROM pg_attribute a
			    WHERE a.attrelid = tbl.oid
			      AND a.attname = 'letter_number'
			      AND a.attnum = ANY(ix.indkey)
			  )
		`).Scan(&count).Error
		return err == nil && count > 0
	case "sqlite":
		type sqliteIndex struct {
			Name    string
			Unique  int
			Partial int
		}
		var indexes []sqliteIndex
		if err := db.Raw("PRAGMA index_list('assignment_letters')").Scan(&indexes).Error; err != nil {
			return false
		}
		type sqliteIndexCol struct {
			Name string
		}
		for _, idx := range indexes {
			if idx.Unique == 0 || idx.Partial != 0 {
				continue
			}
			var cols []sqliteIndexCol
			quoted := `"` + strings.ReplaceAll(idx.Name, `"`, `""`) + `"`
			if err := db.Raw("PRAGMA index_info(" + quoted + ")").Scan(&cols).Error; err != nil {
				continue
			}
			for _, col := range cols {
				if strings.EqualFold(col.Name, "letter_number") {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

// NextLetterNumber generates the next sequential letter number for the given
// team and year (ST-{NNN}/{TEAM}/{YEAR}), starting at 001 for each team/year
// so different audit teams never share sequence counters and sequence outliers
// never cause jumps.
func NextLetterNumber(tx *gorm.DB, team, year string) string {
	includeDeleted := legacyLetterNumberUniqueConstraint(tx)
	return nextLetterNumberWithScope(tx, team, year, includeDeleted)
}

func nextLetterNumberWithScope(tx *gorm.DB, team, year string, includeDeleted bool) string {
	cleanYear := normalizeLetterYear(year)
	targetTeam := normalizeLetterTeam(team)

	query := tx.Session(&gorm.Session{NewDB: true}).Model(&AssignmentLetter{})
	if includeDeleted {
		query = query.Unscoped()
	}

	var numbers []string
	query.Pluck("letter_number", &numbers)

	used := make(map[int]bool, len(numbers))
	for _, numStr := range numbers {
		if seq, ok := parseLetterSequence(numStr, targetTeam, cleanYear); ok {
			used[seq] = true
		}
	}

	nextSeq := 1
	for used[nextSeq] {
		nextSeq++
	}
	return fmt.Sprintf("ST-%03d/%s/%s", nextSeq, targetTeam, cleanYear)
}

func (a *AssignmentLetter) BeforeCreate(tx *gorm.DB) error {
	targetTeam := normalizeLetterTeam(a.AuditTeam)
	cleanYear := normalizeLetterYear(a.AuditYear)

	if tx != nil && tx.Dialector != nil && tx.Dialector.Name() == "postgres" {
		_ = tx.Session(&gorm.Session{NewDB: true}).Exec("SELECT pg_advisory_xact_lock(hashtext(?))", "assignment_letter:"+targetTeam+":"+cleanYear).Error
	}

	includeDeleted := legacyLetterNumberUniqueConstraint(tx)
	scopeDB := tx.Session(&gorm.Session{NewDB: true}).Model(&AssignmentLetter{})
	if includeDeleted {
		scopeDB = scopeDB.Unscoped()
	}

	a.LetterNumber = strings.TrimSpace(a.LetterNumber)
	if a.LetterNumber == "" {
		a.LetterNumber = nextLetterNumberWithScope(tx, targetTeam, cleanYear, includeDeleted)
		return nil
	}

	// Only replace a client-supplied number if it actually collides in the active (or legacy global) scope.
	var count int64
	if err := scopeDB.Where("UPPER(TRIM(letter_number)) = ?", strings.ToUpper(a.LetterNumber)).Count(&count).Error; err == nil && count > 0 {
		a.LetterNumber = nextLetterNumberWithScope(tx, targetTeam, cleanYear, includeDeleted)
	}

	return nil
}

