package services

import (
	"risk-service/models"
	"risk-service/repositories"

	"github.com/google/uuid"
)

var branchToUUID = map[string]string{
	"Head Office":     "00000000-0000-0000-0000-000000000001",
	"Jakarta Branch":   "00000000-0000-0000-0000-000000000002",
	"Surabaya Branch":  "00000000-0000-0000-0000-000000000003",
	"Bandung Branch":   "00000000-0000-0000-0000-000000000004",
	"Bali Branch":      "00000000-0000-0000-0000-000000000005",
}

var uuidToBranch = map[string]string{
	"00000000-0000-0000-0000-000000000001": "Head Office",
	"00000000-0000-0000-0000-000000000002": "Jakarta Branch",
	"00000000-0000-0000-0000-000000000003": "Surabaya Branch",
	"00000000-0000-0000-0000-000000000004": "Bandung Branch",
	"00000000-0000-0000-0000-000000000005": "Bali Branch",
}

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

type RiskResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Impact      int                 `json:"impact"`
	Likelihood  int                 `json:"likelihood"`
	Severity    int                 `json:"severity"`
	Category    string              `json:"category"`
	Branch      string              `json:"branch"`
	Description string              `json:"description"`
	Assessments []RiskAssessmentRes `json:"assessments"`
}

type RiskRequest struct {
	Name        string              `json:"name"`
	Impact      int                 `json:"impact"`
	Likelihood  int                 `json:"likelihood"`
	Severity    int                 `json:"severity"`
	Category    string              `json:"category"`
	Branch      string              `json:"branch"`
	Description string              `json:"description"`
	Assessments []RiskAssessmentReq `json:"assessments"`
}

type IRiskService interface {
	GetAll() ([]RiskResponse, error)
	Create(req *RiskRequest) (*RiskResponse, error)
	Update(id uuid.UUID, req *RiskRequest) (*RiskResponse, error)
	Delete(id uuid.UUID) error
}

type riskService struct {
	repo repositories.IRiskRepository
}

func NewRiskService(repo repositories.IRiskRepository) IRiskService {
	return &riskService{repo: repo}
}

func (s *riskService) GetAll() ([]RiskResponse, error) {
	registers, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	data := make([]RiskResponse, 0)
	for _, reg := range registers {
		branch := uuidToBranch[reg.Profile.DepartmentID.String()]
		if branch == "" {
			branch = "Head Office"
		}

		assessmentsRes := make([]RiskAssessmentRes, 0)
		for _, ast := range reg.Assessments {
			assessmentsRes = append(assessmentsRes, RiskAssessmentRes{
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

		data = append(data, RiskResponse{
			ID:          reg.ID.String(),
			Name:        reg.RiskEvent,
			Impact:      reg.InherentImpact,
			Likelihood:  reg.InherentLikelihood,
			Severity:    reg.InherentScore,
			Category:    reg.Profile.Category,
			Branch:      branch,
			Description: reg.Profile.Description,
			Assessments: assessmentsRes,
		})
	}

	return data, nil
}

func (s *riskService) Create(req *RiskRequest) (*RiskResponse, error) {
	pID := uuid.New()
	deptID := uuid.Nil
	if val, ok := branchToUUID[req.Branch]; ok {
		deptID = uuid.MustParse(val)
	}

	profile := models.RiskProfile{
		ID:           pID,
		DepartmentID: deptID,
		OwnerID:      uuid.Nil,
		Category:     req.Category,
		Description:  req.Description,
	}

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
			Year:           2026,
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

	assessmentsRes := make([]RiskAssessmentRes, 0)
	for _, ast := range createdAssessments {
		assessmentsRes = append(assessmentsRes, RiskAssessmentRes{
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

	return &RiskResponse{
		ID:          regID.String(),
		Name:        req.Name,
		Impact:      req.Impact,
		Likelihood:  req.Likelihood,
		Severity:    req.Severity,
		Category:    req.Category,
		Branch:      req.Branch,
		Description: req.Description,
		Assessments: assessmentsRes,
	}, nil
}

func (s *riskService) Update(id uuid.UUID, req *RiskRequest) (*RiskResponse, error) {
	register, err := s.repo.FindByID(id)
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
	if val, ok := branchToUUID[req.Branch]; ok {
		register.Profile.DepartmentID = uuid.MustParse(val)
	}

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

	assessmentsRes := make([]RiskAssessmentRes, 0)
	for _, ast := range updatedAssessments {
		assessmentsRes = append(assessmentsRes, RiskAssessmentRes{
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

	return &RiskResponse{
		ID:          id.String(),
		Name:        req.Name,
		Impact:      req.Impact,
		Likelihood:  req.Likelihood,
		Severity:    req.Severity,
		Category:    req.Category,
		Branch:      req.Branch,
		Description: req.Description,
		Assessments: assessmentsRes,
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
