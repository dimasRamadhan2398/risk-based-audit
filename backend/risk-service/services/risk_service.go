package services

import (
	"context"
	"time"

	"risk-service/models"
	"risk-service/pkg/masterclient"
	"risk-service/repositories"

	"github.com/google/uuid"
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

type RiskAssessmentRes struct {
	ID           string `json:"id"`
	Year         int    `json:"year"`
	ImpactQ1     int    `json:"impact_q1"`
	ImpactQ2     int    `json:"impact_q2"`
	ImpactQ3     int    `json:"impact_q3"`
	ImpactQ4     int    `json:"impact_q4"`
	LikelihoodQ1 int    `json:"likelihood_q1"`
	LikelihoodQ2 int    `json:"likelihood_q2"`
	LikelihoodQ3 int    `json:"likelihood_q3"`
	LikelihoodQ4 int    `json:"likelihood_q4"`
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
type RiskRequest struct {
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
		})
	}
	return res
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

	if len(req.Assessments) > 0 {
		for _, astReq := range req.Assessments {
			ast := models.RiskAssessment{
				ID:             uuid.New(),
				RiskRegisterID: regID,
				Year:           astReq.Year,
				ImpactQ1:       astReq.ImpactQ1,
				ImpactQ2:       astReq.ImpactQ2,
				ImpactQ3:       astReq.ImpactQ3,
				ImpactQ4:       astReq.ImpactQ4,
				LikelihoodQ1:   astReq.LikelihoodQ1,
				LikelihoodQ2:   astReq.LikelihoodQ2,
				LikelihoodQ3:   astReq.LikelihoodQ3,
				LikelihoodQ4:   astReq.LikelihoodQ4,
			}
			if err := s.repo.CreateAssessment(&ast); err != nil {
				return nil, err
			}
		}
	} else {
		ast := models.RiskAssessment{
			ID:             uuid.New(),
			RiskRegisterID: regID,
			Year:           time.Now().Year(),
			ImpactQ1:       req.Impact,
			ImpactQ2:       req.Impact,
			ImpactQ3:       req.Impact,
			ImpactQ4:       req.Impact,
			LikelihoodQ1:   req.Likelihood,
			LikelihoodQ2:   req.Likelihood,
			LikelihoodQ3:   req.Likelihood,
			LikelihoodQ4:   req.Likelihood,
		}
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

func (s *riskService) Update(ctx context.Context, id uuid.UUID, req *RiskRequest) (*RiskResponse, error) {
	register, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Validate the location before writing anything.
	loc, err := s.locationForRequest(ctx, req, &register.Profile)
	if err != nil {
		return nil, err
	}

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

	if err := s.repo.SaveProfile(&register.Profile); err != nil {
		return nil, err
	}

	if len(req.Assessments) > 0 {
		for _, astReq := range req.Assessments {
			existing, err := s.repo.FindAssessmentByYear(id, astReq.Year)
			if err != nil {
				// Create new assessment
				newAst := models.RiskAssessment{
					ID:             uuid.New(),
					RiskRegisterID: id,
					Year:           astReq.Year,
					ImpactQ1:       astReq.ImpactQ1,
					ImpactQ2:       astReq.ImpactQ2,
					ImpactQ3:       astReq.ImpactQ3,
					ImpactQ4:       astReq.ImpactQ4,
					LikelihoodQ1:   astReq.LikelihoodQ1,
					LikelihoodQ2:   astReq.LikelihoodQ2,
					LikelihoodQ3:   astReq.LikelihoodQ3,
					LikelihoodQ4:   astReq.LikelihoodQ4,
				}
				s.repo.CreateAssessment(&newAst)
			} else {
				existing.ImpactQ1 = astReq.ImpactQ1
				existing.ImpactQ2 = astReq.ImpactQ2
				existing.ImpactQ3 = astReq.ImpactQ3
				existing.ImpactQ4 = astReq.ImpactQ4
				existing.LikelihoodQ1 = astReq.LikelihoodQ1
				existing.LikelihoodQ2 = astReq.LikelihoodQ2
				existing.LikelihoodQ3 = astReq.LikelihoodQ3
				existing.LikelihoodQ4 = astReq.LikelihoodQ4
				s.repo.SaveAssessment(existing)
			}
		}
	}

	updatedAssessments, err := s.repo.FindAssessmentsByRegisterID(id)
	if err != nil {
		return nil, err
	}

	locationID, branch := s.responseBranch(ctx, register.Profile, loc)
	return &RiskResponse{
		ID:          id.String(),
		LocationID:  locationID,
		Name:        req.Name,
		Impact:      req.Impact,
		Likelihood:  req.Likelihood,
		Severity:    req.Severity,
		Category:    req.Category,
		Branch:      branch,
		Description: req.Description,
		Assessments: toAssessmentResponses(updatedAssessments),
	}, nil
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
