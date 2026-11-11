package repositories

import (
	"fmt"
	"sort"
	"strings"

	"risk-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IRiskRepository interface {
	FindAll() ([]models.RiskRegister, error)
	FindPage(f RiskFilter) ([]models.RiskRegister, int64, error)
	DistinctCategories(fiscalYear int) ([]string, error)
	FindByID(id uuid.UUID) (*models.RiskRegister, error)
	CreateProfile(profile *models.RiskProfile) error
	CreateRegister(register *models.RiskRegister) error
	CreateAssessment(assessment *models.RiskAssessment) error
	SaveRegister(register *models.RiskRegister) error
	SaveProfile(profile *models.RiskProfile) error
	SaveAssessment(assessment *models.RiskAssessment) error
	DeleteRegister(register *models.RiskRegister) error
	DeleteProfile(id uuid.UUID) error
	FindAssessmentByYear(regID uuid.UUID, year int) (*models.RiskAssessment, error)
	FindAssessmentsByRegisterID(regID uuid.UUID) ([]models.RiskAssessment, error)
}

type riskRepository struct {
	db *gorm.DB
}

func NewRiskRepository(db *gorm.DB) IRiskRepository {
	return &riskRepository{db: db}
}

func (r *riskRepository) FindAll() ([]models.RiskRegister, error) {
	var registers []models.RiskRegister
	err := r.db.Preload("Profile").Preload("Assessments").Find(&registers).Error
	return registers, err
}

func (r *riskRepository) FindByID(id uuid.UUID) (*models.RiskRegister, error) {
	var register models.RiskRegister
	err := r.db.Preload("Profile").First(&register, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &register, nil
}

func (r *riskRepository) CreateProfile(profile *models.RiskProfile) error {
	return r.db.Create(profile).Error
}

func (r *riskRepository) CreateRegister(register *models.RiskRegister) error {
	return r.db.Create(register).Error
}

func (r *riskRepository) CreateAssessment(assessment *models.RiskAssessment) error {
	return r.db.Create(assessment).Error
}

func (r *riskRepository) SaveRegister(register *models.RiskRegister) error {
	return r.db.Save(register).Error
}

func (r *riskRepository) SaveProfile(profile *models.RiskProfile) error {
	return r.db.Save(profile).Error
}

func (r *riskRepository) SaveAssessment(assessment *models.RiskAssessment) error {
	return r.db.Save(assessment).Error
}

func (r *riskRepository) DeleteRegister(register *models.RiskRegister) error {
	return r.db.Delete(register).Error
}

func (r *riskRepository) DeleteProfile(id uuid.UUID) error {
	return r.db.Delete(&models.RiskProfile{}, "id = ?", id).Error
}

func (r *riskRepository) FindAssessmentByYear(regID uuid.UUID, year int) (*models.RiskAssessment, error) {
	var assessment models.RiskAssessment
	err := r.db.First(&assessment, "risk_register_id = ? AND year = ?", regID, year).Error
	if err != nil {
		return nil, err
	}
	return &assessment, nil
}

func (r *riskRepository) FindAssessmentsByRegisterID(regID uuid.UUID) ([]models.RiskAssessment, error) {
	var assessments []models.RiskAssessment
	err := r.db.Find(&assessments, "risk_register_id = ?", regID).Error
	return assessments, err
}

// Sort keys accepted by RiskFilter.SortBy (the API allowlist).
const (
	SortSeverity   = "severity"
	SortScore      = "score" // impact x likelihood of Period in FiscalYear
	SortImpact     = "impact"
	SortLikelihood = "likelihood"
	SortName       = "name"
	SortCategory   = "category"
)

// RiskFilter selects and orders Corporate Risk Profile rows. Zero values mean
// "not set". With FiscalYear set, only risks assessed in that year match and
// name/category/description/location are read from that year's snapshot
// (falling back to the latest version where the snapshot is empty), exactly
// as the frontend shows them; without it the latest version is used.
type RiskFilter struct {
	Category   string // exact, case-insensitive, trimmed
	LocationID string // a UUID, or Unassigned
	Unassigned bool   // risks with no location_id
	FiscalYear int
	Period     string // "q1".."q4"; only used by SortScore
	Search     string // substring of name/category/description
	SortBy     string // one of the Sort* constants; "" = SortSeverity
	SortDesc   *bool  // nil = default for the key (desc for scores, asc otherwise)

	Paged    bool
	Page     int // 1-based
	PageSize int
}

// likeEscaper escapes LIKE wildcards; paired with ESCAPE '\' below.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// riskExprs are the SQL expressions for the filterable columns. They are
// constants chosen by whether a fiscal year join is present, never built from
// request input.
type riskExprs struct{ name, category, description, location string }

func exprsFor(year int) riskExprs {
	if year > 0 {
		return riskExprs{
			name:        "COALESCE(NULLIF(a.risk_event, ''), risk_register.risk_event)",
			category:    "COALESCE(NULLIF(a.category, ''), p.category)",
			description: "COALESCE(NULLIF(a.description, ''), p.description, '')",
			location:    "COALESCE(a.location_id, p.location_id)",
		}
	}
	return riskExprs{
		name:        "risk_register.risk_event",
		category:    "p.category",
		description: "COALESCE(p.description, '')",
		location:    "p.location_id",
	}
}

func (r *riskRepository) scoped(year int) *gorm.DB {
	q := r.db.Model(&models.RiskRegister{}).
		Joins("JOIN risk_profile p ON p.id = risk_register.profile_id AND p.deleted_at IS NULL")
	if year > 0 {
		q = q.Joins("JOIN risk_assessment a ON a.risk_register_id = risk_register.id AND a.year = ?", year)
	}
	return q
}

func (r *riskRepository) filtered(f RiskFilter) *gorm.DB {
	e := exprsFor(f.FiscalYear)
	q := r.scoped(f.FiscalYear)
	if c := strings.TrimSpace(f.Category); c != "" {
		q = q.Where("LOWER(TRIM("+e.category+")) = ?", strings.ToLower(c))
	}
	if f.Unassigned {
		q = q.Where(e.location + " IS NULL")
	} else if f.LocationID != "" {
		q = q.Where(e.location+" = ?", f.LocationID)
	}
	if s := strings.ToLower(strings.TrimSpace(f.Search)); s != "" {
		pat := "%" + likeEscaper.Replace(s) + "%"
		q = q.Where(fmt.Sprintf(
			"(LOWER(%s) LIKE ? ESCAPE '\\' OR LOWER(%s) LIKE ? ESCAPE '\\' OR LOWER(%s) LIKE ? ESCAPE '\\')",
			e.name, e.category, e.description), pat, pat, pat)
	}
	return q
}

// orderClause is built only from allowlisted constants.
func orderClause(f RiskFilter) string {
	desc := func(def bool) string {
		d := def
		if f.SortDesc != nil {
			d = *f.SortDesc
		}
		if d {
			return " DESC"
		}
		return " ASC"
	}
	e := exprsFor(f.FiscalYear)
	var primary string
	switch f.SortBy {
	case SortName:
		primary = "LOWER(" + e.name + ")" + desc(false)
	case SortCategory:
		primary = "LOWER(" + e.category + ")" + desc(false)
	case SortImpact:
		primary = "risk_register.inherent_impact" + desc(true)
	case SortLikelihood:
		primary = "risk_register.inherent_likelihood" + desc(true)
	case SortScore:
		if f.FiscalYear > 0 {
			q := map[string]string{"q1": "1", "q2": "2", "q3": "3", "q4": "4"}[f.Period]
			if q == "" {
				q = "1"
			}
			primary = "(a.impact_q" + q + " * a.likelihood_q" + q + ")" + desc(true)
			break
		}
		fallthrough
	default:
		primary = "risk_register.inherent_score" + desc(true)
	}
	// Name and id tie-breakers keep page boundaries stable.
	return primary + ", LOWER(" + e.name + ") ASC, risk_register.id ASC"
}

func (r *riskRepository) FindPage(f RiskFilter) ([]models.RiskRegister, int64, error) {
	var total int64
	if f.Paged {
		if err := r.filtered(f).Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}
	q := r.filtered(f).Select("risk_register.*").
		Preload("Profile").Preload("Assessments").
		Order(orderClause(f))
	if f.Paged {
		q = q.Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize)
	}
	var registers []models.RiskRegister
	if err := q.Find(&registers).Error; err != nil {
		return nil, 0, err
	}
	if !f.Paged {
		total = int64(len(registers))
	}
	return registers, total, nil
}

// DistinctCategories lists the trimmed, non-empty categories (one spelling per
// case-insensitive value), optionally as read in one fiscal year.
func (r *riskRepository) DistinctCategories(fiscalYear int) ([]string, error) {
	expr := exprsFor(fiscalYear).category
	var raw []string
	err := r.scoped(fiscalYear).Distinct().
		Where("TRIM("+expr+") <> ''").
		Pluck("TRIM("+expr+")", &raw).Error
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(raw))
	for _, c := range raw {
		k := strings.ToLower(c)
		if !seen[k] {
			seen[k] = true
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i]) < strings.ToLower(out[j]) })
	return out, nil
}
