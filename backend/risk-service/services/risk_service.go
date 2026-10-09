package services

import (
	"context"
	"errors"
	"time"

	"risk-service/models"
	"risk-service/pkg/masterclient"
	"risk-service/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Branch resolution against the Location master lives in risk_location.go.

type RiskAssessmentReq struct {
	Year         int `json:"year"`
	ImpactQ1     int `json:"impact_q1"`
	ImpactQ2     int `json:"impact_q2"`
	ImpactQ3     int `json:"impact_q3"`
	ImpactQ4     int `json:"impact_q4"`
	LikelihoodQ1 int `json:"likelihood_q1"`
	LikelihoodQ2 int `json:"likelihood_q2"`
	LikelihoodQ3 int `json:"likelihood_q3"`
	LikelihoodQ4 int `json:"likelihood_q4"`
}

// RiskAssessmentRes is one year of a risk's CRP. Name/Category/Description and
// LocationID/Branch are how the risk read in that year (Branch is the location
// name stored with the snapshot); they are empty/null only on a snapshot that
// has not been backfilled yet.
type RiskAssessmentRes struct {
	ID           string  `json:"id"`
	Year         int     `json:"year"`
	ImpactQ1     int     `json:"impact_q1"`
	ImpactQ2     int     `json:"impact_q2"`
	ImpactQ3     int     `json:"impact_q3"`
	ImpactQ4     int     `json:"impact_q4"`
	LikelihoodQ1 int     `json:"likelihood_q1"`
	LikelihoodQ2 int     `json:"likelihood_q2"`
	LikelihoodQ3 int     `json:"likelihood_q3"`
	LikelihoodQ4 int     `json:"likelihood_q4"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	LocationID   *string `json:"location_id"`
	Branch       *string `json:"branch"`
}

// RiskResponse is one Corporate Risk Profile item.
//
// LocationID / Branch are the Location master row the risk is linked to and its
// current master name; both are null when the risk is not linked to a
// registered location. While master-service is unreachable a stored
// location_id is still returned but Branch is null.
type RiskResponse struct {
	ID          string              `json:"id"`
	LocationID  *string             `json:"location_id"`
	Name        string              `json:"name"`
	Impact      int                 `json:"impact"`
	Likelihood  int                 `json:"likelihood"`
	Severity    int                 `json:"severity"`
	Category    string              `json:"category"`
	Branch      *string             `json:"branch"`
	Description string              `json:"description"`
	Assessments []RiskAssessmentRes `json:"assessments"`
}

// RiskRequest is the create/update body. LocationID (a Location master UUID)
// is preferred; Branch is only for older clients and must match a registered
// location's name. Sending neither leaves an existing risk's location as is.
//
// Year is the CRP year being edited (0 = the current year). An update writes
// that year's assessment only; assessments for other years in the body are
// ignored so a stale client copy cannot rewrite history.
type RiskRequest struct {
	Year        int                 `json:"year"`
	Name        string              `json:"name"`
	LocationID  string              `json:"location_id"`
	Impact      int                 `json:"impact"`
	Likelihood  int                 `json:"likelihood"`
	Severity    int                 `json:"severity"`
	Category    string              `json:"category"`
	Branch      string              `json:"branch"`
	Description string              `json:"description"`
	Assessments []RiskAssessmentReq `json:"assessments"`
}

type IRiskService interface {
	GetAll(ctx context.Context) ([]RiskResponse, error)
	Create(ctx context.Context, req *RiskRequest) (*RiskResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *RiskRequest) (*RiskResponse, error)
	Delete(id uuid.UUID) error
}

type riskService struct {
	repo      repositories.IRiskRepository
	locations masterclient.LocationSource
}

func NewRiskService(repo repositories.IRiskRepository, locations masterclient.LocationSource) IRiskService {
	return &riskService{repo: repo, locations: locations}
}

func toAssessmentResponses(assessments []models.RiskAssessment) []RiskAssessmentRes {
	res := make([]RiskAssessmentRes, 0, len(assessments))
	for _, ast := range assessments {
		res = append(res, RiskAssessmentRes{
			ID:           ast.ID.String(),
			Year:         ast.Year,
			ImpactQ1:     ast.ImpactQ1,
			ImpactQ2:     ast.ImpactQ2,
			ImpactQ3:     ast.ImpactQ3,
			ImpactQ4:     ast.ImpactQ4,
			LikelihoodQ1: ast.LikelihoodQ1,
			LikelihoodQ2: ast.LikelihoodQ2,
			LikelihoodQ3: ast.LikelihoodQ3,
			LikelihoodQ4: ast.LikelihoodQ4,
			Name:         ast.RiskEvent,
			Category:     ast.Category,
			Description:  ast.Description,
			LocationID:   uuidStrPtr(ast.LocationID),
			Branch:       nonEmptyStrPtr(ast.LocationName),
		})
	}
	return res
}

func uuidStrPtr(id *uuid.UUID) *string {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return strPtr(id.String())
}

func nonEmptyStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// setQuarters copies a request's quarterly scores onto an assessment row.
func setQuarters(ast *models.RiskAssessment, q RiskAssessmentReq) {
	ast.ImpactQ1, ast.ImpactQ2, ast.ImpactQ3, ast.ImpactQ4 = q.ImpactQ1, q.ImpactQ2, q.ImpactQ3, q.ImpactQ4
	ast.LikelihoodQ1, ast.LikelihoodQ2, ast.LikelihoodQ3, ast.LikelihoodQ4 = q.LikelihoodQ1, q.LikelihoodQ2, q.LikelihoodQ3, q.LikelihoodQ4
}

// flatQuarters is the same impact/likelihood for all four quarters.
func flatQuarters(year, impact, likelihood int) RiskAssessmentReq {
	return RiskAssessmentReq{
		Year:     year,
		ImpactQ1: impact, ImpactQ2: impact, ImpactQ3: impact, ImpactQ4: impact,
		LikelihoodQ1: likelihood, LikelihoodQ2: likelihood, LikelihoodQ3: likelihood, LikelihoodQ4: likelihood,
	}
}

// setSnapshot records how the risk reads in the assessment's year: the
// request's name/category/description, and its location when one was
// validated (nil loc keeps the row's location).
func setSnapshot(ast *models.RiskAssessment, req *RiskRequest, loc *masterclient.Location) {
	now := time.Now()
	ast.RiskEvent = req.Name
	ast.Category = req.Category
	ast.Description = req.Description
	if loc != nil {
		id := loc.ID
		ast.LocationID = &id
		ast.LocationName = loc.Name
	}
	ast.SnapshotAt = &now
}

func (s *riskService) GetAll(ctx context.Context) ([]RiskResponse, error) {
	registers, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	// One master snapshot for the whole list; nil if master is unreachable.
	idx := s.locationIndexOrNil(ctx)

	data := make([]RiskResponse, 0, len(registers))
	for _, reg := range registers {
		locationID, branch := resolveBranch(reg.Profile, idx)
		data = append(data, RiskResponse{
			ID:          reg.ID.String(),
			LocationID:  locationID,
			Name:        reg.RiskEvent,
			Impact:      reg.InherentImpact,
			Likelihood:  reg.InherentLikelihood,
			Severity:    reg.InherentScore,
			Category:    reg.Profile.Category,
			Branch:      branch,
			Description: reg.Profile.Description,
			Assessments: toAssessmentResponses(reg.Assessments),
		})
	}

	return data, nil
}

func (s *riskService) Create(ctx context.Context, req *RiskRequest) (*RiskResponse, error) {
	// Validate the location before writing anything.
	loc, err := s.locationForRequest(ctx, req, nil)
	if err != nil {
		return nil, err
	}

	pID := uuid.New()
	profile := models.RiskProfile{
		ID:           pID,
		DepartmentID: uuid.Nil,
		OwnerID:      uuid.Nil,
		Category:     req.Category,
		Description:  req.Description,
	}
	applyLocation(&profile, loc)

	if err := s.repo.CreateProfile(&profile); err != nil {
		return nil, err
	}

	regID := uuid.New()
	register := models.RiskRegister{
		ID:                   regID,
		ProfileID:            pID,
		RiskSource:           models.RiskSourceDirect,
		RiskEvent:            req.Name,
		InherentLikelihood:   req.Likelihood,
		InherentImpact:       req.Impact,
		InherentScore:        req.Severity,
		ControlEffectiveness: 0,
		ResidualScore:        req.Severity,
		FinalRiskLevel:       models.RiskFinalLevelHigh,
		Status:               models.RiskRegisterStatusApproved,
	}

	if err := s.repo.CreateRegister(&register); err != nil {
		return nil, err
	}

	astReqs := req.Assessments
	if len(astReqs) == 0 {
		astReqs = []RiskAssessmentReq{flatQuarters(time.Now().Year(), req.Impact, req.Likelihood)}
	}
	for _, astReq := range astReqs {
		ast := models.RiskAssessment{ID: uuid.New(), RiskRegisterID: regID, Year: astReq.Year}
		setQuarters(&ast, astReq)
		setSnapshot(&ast, req, loc)
		if err := s.repo.CreateAssessment(&ast); err != nil {
			return nil, err
		}
	}

	createdAssessments, err := s.repo.FindAssessmentsByRegisterID(regID)
	if err != nil {
		return nil, err
	}

	locationID, branch := s.responseBranch(ctx, profile, loc)
	return &RiskResponse{
		ID:          regID.String(),
		LocationID:  locationID,
		Name:        req.Name,
		Impact:      req.Impact,
		Likelihood:  req.Likelihood,
		Severity:    req.Severity,
		Category:    req.Category,
		Branch:      branch,
		Description: req.Description,
		Assessments: toAssessmentResponses(createdAssessments),
	}, nil
}

// responseBranch is the location/branch to echo after a write: the location
// just validated, else whatever the saved profile resolves to.
func (s *riskService) responseBranch(ctx context.Context, p models.RiskProfile, loc *masterclient.Location) (*string, *string) {
	if loc != nil {
		return strPtr(loc.ID.String()), strPtr(loc.Name)
	}
	_, legacy := legacySentinelBranch[p.DepartmentID]
	if p.LocationID == nil && p.LocationName == "" && !legacy {
		return nil, nil // unlinked; no need to ask master-service
	}
	return resolveBranch(p, s.locationIndexOrNil(ctx))
}

// Update edits the risk as it reads in one CRP year (req.Year, default the
// current year): that year's assessment row gets the scores and the
// name/category/description/location snapshot. risk_register/risk_profile —
// the latest version, used to start a new year — move only when the edited
// year is the most recent year the risk is assessed in, so editing last
// year's CRP never rewrites this year's, and vice versa.
func (s *riskService) Update(ctx context.Context, id uuid.UUID, req *RiskRequest) (*RiskResponse, error) {
	register, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	year := req.Year
	if year == 0 {
		year = time.Now().Year()
	}

	ast, err := s.repo.FindAssessmentByYear(id, year)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	isNew := ast == nil
	if isNew {
		// A year the risk was not assessed in yet starts from the latest version.
		ast = &models.RiskAssessment{
			ID:             uuid.New(),
			RiskRegisterID: id,
			Year:           year,
			LocationID:     register.Profile.LocationID,
			LocationName:   register.Profile.LocationName,
		}
		setQuarters(ast, flatQuarters(year, req.Impact, req.Likelihood))
	}

	// Validate the location before writing anything. "Unchanged" means
	// unchanged for this year, so compare against the year's snapshot.
	current := register.Profile
	if ast.LocationID != nil {
		current.LocationID = ast.LocationID
	}
	loc, err := s.locationForRequest(ctx, req, &current)
	if err != nil {
		return nil, err
	}

	for _, astReq := range req.Assessments {
		if astReq.Year == year {
			setQuarters(ast, astReq)
			break
		}
	}
	setSnapshot(ast, req, loc)

	if isNew {
		err = s.repo.CreateAssessment(ast)
	} else {
		err = s.repo.SaveAssessment(ast)
	}
	if err != nil {
		return nil, err
	}

	assessments, err := s.repo.FindAssessmentsByRegisterID(id)
	if err != nil {
		return nil, err
	}

	var masterLoc *masterclient.Location
	if year >= latestYear(assessments) {
		register.RiskEvent = req.Name
		register.InherentLikelihood = req.Likelihood
		register.InherentImpact = req.Impact
		register.InherentScore = req.Severity
		register.ResidualScore = req.Severity

		if err := s.repo.SaveRegister(register); err != nil {
			return nil, err
		}

		register.Profile.Category = req.Category
		register.Profile.Description = req.Description
		applyLocation(&register.Profile, loc)
		masterLoc = loc

		if err := s.repo.SaveProfile(&register.Profile); err != nil {
			return nil, err
		}
	}

	locationID, branch := s.responseBranch(ctx, register.Profile, masterLoc)
	return &RiskResponse{
		ID:          id.String(),
		LocationID:  locationID,
		Name:        register.RiskEvent,
		Impact:      register.InherentImpact,
		Likelihood:  register.InherentLikelihood,
		Severity:    register.InherentScore,
		Category:    register.Profile.Category,
		Branch:      branch,
		Description: register.Profile.Description,
		Assessments: toAssessmentResponses(assessments),
	}, nil
}

func latestYear(assessments []models.RiskAssessment) int {
	latest := 0
	for _, a := range assessments {
		if a.Year > latest {
			latest = a.Year
		}
	}
	return latest
}

func (s *riskService) Delete(id uuid.UUID) error {
	register, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteRegister(register); err != nil {
		return err
	}

	return s.repo.DeleteProfile(register.ProfileID)
}
