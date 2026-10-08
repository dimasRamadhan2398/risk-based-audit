package services

import (
	"errors"

	"risk-service/models"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory IRiskRepository for service tests.
type fakeRepo struct {
	registers    map[uuid.UUID]*models.RiskRegister
	assessments  map[uuid.UUID][]models.RiskAssessment
	order        []uuid.UUID
	lastProfile  models.RiskProfile
	profileSaves int
	writeCount   int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		registers:   map[uuid.UUID]*models.RiskRegister{},
		assessments: map[uuid.UUID][]models.RiskAssessment{},
	}
}

// addRisk seeds a register + profile directly, bypassing write counters.
func (r *fakeRepo) addRisk(name string, profile models.RiskProfile) uuid.UUID {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	reg := &models.RiskRegister{ID: uuid.New(), ProfileID: profile.ID, RiskEvent: name, Profile: profile}
	r.registers[reg.ID] = reg
	r.order = append(r.order, reg.ID)
	return reg.ID
}

func (r *fakeRepo) writes() int { return r.writeCount }

func (r *fakeRepo) FindAll() ([]models.RiskRegister, error) {
	out := make([]models.RiskRegister, 0, len(r.order))
	for _, id := range r.order {
		reg := *r.registers[id]
		reg.Assessments = r.assessments[id]
		out = append(out, reg)
	}
	return out, nil
}

func (r *fakeRepo) FindByID(id uuid.UUID) (*models.RiskRegister, error) {
	reg, ok := r.registers[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	cp := *reg
	return &cp, nil
}

func (r *fakeRepo) CreateProfile(p *models.RiskProfile) error {
	r.writeCount++
	r.lastProfile = *p
	return nil
}

func (r *fakeRepo) CreateRegister(reg *models.RiskRegister) error {
	r.writeCount++
	cp := *reg
	cp.Profile = r.lastProfile
	r.registers[reg.ID] = &cp
	r.order = append(r.order, reg.ID)
	return nil
}

func (r *fakeRepo) CreateAssessment(a *models.RiskAssessment) error {
	r.writeCount++
	r.assessments[a.RiskRegisterID] = append(r.assessments[a.RiskRegisterID], *a)
	return nil
}

func (r *fakeRepo) SaveRegister(reg *models.RiskRegister) error {
	r.writeCount++
	cp := *reg
	r.registers[reg.ID] = &cp
	return nil
}

func (r *fakeRepo) SaveProfile(p *models.RiskProfile) error {
	r.writeCount++
	r.profileSaves++
	r.lastProfile = *p
	for _, reg := range r.registers {
		if reg.ProfileID == p.ID {
			reg.Profile = *p
		}
	}
	return nil
}

func (r *fakeRepo) SaveAssessment(a *models.RiskAssessment) error {
	r.writeCount++
	return nil
}

func (r *fakeRepo) DeleteRegister(reg *models.RiskRegister) error {
	r.writeCount++
	delete(r.registers, reg.ID)
	return nil
}

func (r *fakeRepo) DeleteProfile(uuid.UUID) error {
	r.writeCount++
	return nil
}

func (r *fakeRepo) FindAssessmentByYear(regID uuid.UUID, year int) (*models.RiskAssessment, error) {
	for _, a := range r.assessments[regID] {
		if a.Year == year {
			cp := a
			return &cp, nil
		}
	}
	return nil, errors.New("record not found")
}

func (r *fakeRepo) FindAssessmentsByRegisterID(regID uuid.UUID) ([]models.RiskAssessment, error) {
	return r.assessments[regID], nil
}
