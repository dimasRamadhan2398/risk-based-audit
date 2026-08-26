package seeders

import (
	"fmt"
	"time"

	"master-service/models"

	"gorm.io/gorm"
)

type Seeder struct {
	DB *gorm.DB
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{DB: db}
}

func (s *Seeder) RunAll() error {
	seeders := []struct {
		name string
		fn   func() error
	}{
		{"Locations", s.SeedLocations},
		{"Companies", s.SeedCompanies},
		{"JobRoles", s.SeedJobRoles},
		{"Likelihoods", s.SeedLikelihoods},
		{"Impacts", s.SeedImpacts},
		{"RiskLevels", s.SeedRiskLevels},
		{"RiskCategories", s.SeedRiskCategories},
		{"RiskMatrixCells", s.SeedRiskMatrixCells},
		{"BusinessUnits", s.SeedBusinessUnits},
		{"Departments", s.SeedDepartments},
		{"Employees", s.SeedEmployees},
		{"RiskRegisters", s.SeedRiskRegisters},
		{"AuditPeriods", s.SeedAuditPeriods},
		{"AnnualAuditPlans", s.SeedAnnualAuditPlans},
		{"AuditScopes", s.SeedAuditScopes},
		{"Controls", s.SeedControls},
		{"ControlAssessments", s.SeedControlAssessments},
		{"AuditFindings", s.SeedAuditFindings},
		{"AuditIssues", s.SeedAuditIssues},
		{"AuditRecommendations", s.SeedAuditRecommendations},
		{"MitigationActions", s.SeedMitigationActions},
		{"RiskCauses", s.SeedRiskCauses},
		{"RiskEffects", s.SeedRiskEffects},
		{"RiskIndicators", s.SeedRiskIndicators},
		{"RiskIndicatorLogs", s.SeedRiskIndicatorLogs},
		{"AuditUniverse", s.SeedAuditUniverse},
		{"AuditWorkpapers", s.SeedAuditWorkpapers},
		{"VisionMissionGoals", s.SeedVisionMissionGoals},
		{"QAReports", s.SeedQAReports},
		{"ConsultingServices", s.SeedConsultingServices},
	}

	for _, seeder := range seeders {
		fmt.Printf("Running seeder: %s...\n", seeder.name)
		if err := seeder.fn(); err != nil {
			return fmt.Errorf("failed to seed %s: %w", seeder.name, err)
		}
		fmt.Printf("Seeder %s completed successfully\n", seeder.name)
	}

	return nil
}

func (s *Seeder) SeedLocations() error {
	seeds := []models.Location{
		{
			Name:       "AIFL Headquarters",
			Address:    "Wisma AIFL, Jl. Jend. Sudirman Kav. 52-53",
			City:       "Jakarta Selatan",
			Province:   "DKI Jakarta",
			PostalCode: "12190",
			Country:    "Indonesia",
			Latitude:   float64Ptr(-6.2088),
			Longitude:  float64Ptr(106.8166),
			IsActive:   true,
		},
		{
			Name:       "Surabaya Branch",
			Address:    "Graha Pena, Jl. Ahmad Yani No. 88",
			City:       "Surabaya",
			Province:   "Jawa Timur",
			PostalCode: "60234",
			Country:    "Indonesia",
			Latitude:   float64Ptr(-7.2575),
			Longitude:  float64Ptr(112.7521),
			IsActive:   true,
		},
		{
			Name:       "Bandung Branch",
			Address:    "Jl. Braga No. 99",
			City:       "Bandung",
			Province:   "Jawa Barat",
			PostalCode: "40111",
			Country:    "Indonesia",
			Latitude:   float64Ptr(-6.9175),
			Longitude:  float64Ptr(107.6191),
			IsActive:   true,
		},
	}

	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.Location{Name: seeds[i].Name}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedCompanies() error {
	hqLocation, err := getByName[models.Location](s.DB, "name", "AIFL Headquarters")
	if err != nil {
		return err
	}
	sbyLocation, err := getByName[models.Location](s.DB, "name", "Surabaya Branch")
	if err != nil {
		return err
	}

	seeds := []models.Company{
		{
			CompanyCode:   "AIFL-HQ",
			CompanyName:   "PT AIFL Indonesia",
			LegalName:     "PT AIFL Indonesia Tbk",
			TaxID:         "01.234.567.8-901.000",
			CompanyType:   models.CompanyTypeHolding,
			LocationID:    &hqLocation.ID,
			Phone:         "+62-21-1234-5678",
			Email:         "info@aifl.co.id",
			Website:       "https://www.aifl.co.id",
			IsActive:      true,
			EstablishedAt: timePtr(time.Date(2015, 1, 15, 0, 0, 0, 0, time.UTC)),
		},
		{
			CompanyCode:   "AIFL-FIN",
			CompanyName:   "AIFL Financial Services",
			LegalName:     "PT AIFL Financial Services",
			TaxID:         "02.345.678.9-012.000",
			CompanyType:   models.CompanyTypeSubsidiary,
			LocationID:    &hqLocation.ID,
			Phone:         "+62-21-2345-6789",
			Email:         "info@aifl-finance.co.id",
			Website:       "https://www.aifl-finance.co.id",
			IsActive:      true,
			EstablishedAt: timePtr(time.Date(2018, 6, 1, 0, 0, 0, 0, time.UTC)),
		},
		{
			CompanyCode:   "AIFL-SBY",
			CompanyName:   "AIFL Surabaya",
			LegalName:     "PT AIFL Indonesia Kantor Surabaya",
			TaxID:         "03.456.789.0-123.000",
			CompanyType:   models.CompanyTypeBranch,
			LocationID:    &sbyLocation.ID,
			Phone:         "+62-31-4567-8901",
			Email:         "surabaya@aifl.co.id",
			IsActive:      true,
			EstablishedAt: timePtr(time.Date(2020, 3, 15, 0, 0, 0, 0, time.UTC)),
		},
	}

	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.Company{CompanyCode: seeds[i].CompanyCode}).Error; err != nil {
			return err
		}
	}

	holding, err := getByName[models.Company](s.DB, "company_code", "AIFL-HQ")
	if err != nil {
		return err
	}

	if err := s.DB.Model(&models.Company{}).Where("company_code = ?", "AIFL-FIN").Update("parent_id", holding.ID).Error; err != nil {
		return err
	}
	if err := s.DB.Model(&models.Company{}).Where("company_code = ?", "AIFL-SBY").Update("parent_id", holding.ID).Error; err != nil {
		return err
	}
	return nil
}

func (s *Seeder) SeedJobRoles() error {
	seeds := []models.JobRole{
		{JobRoleCode: "DIR-001", JobRoleName: "Finance Director", JobRoleDescription: "Executive finance lead", JobPositionType: models.PositionTypeDirector, IsActive: true},
		{JobRoleCode: "MGR-001", JobRoleName: "Finance Manager", JobRoleDescription: "Finance function manager", JobPositionType: models.PositionTypeManager, IsActive: true},
		{JobRoleCode: "MGR-002", JobRoleName: "IT Manager", JobRoleDescription: "IT function manager", JobPositionType: models.PositionTypeManager, IsActive: true},
		{JobRoleCode: "AUD-003", JobRoleName: "Audit Manager", JobRoleDescription: "Internal audit manager", JobPositionType: models.PositionTypeManager, IsActive: true},
		{JobRoleCode: "AUD-001", JobRoleName: "Internal Auditor", JobRoleDescription: "Internal auditor", JobPositionType: models.PositionTypeStaff, IsActive: true},
		{JobRoleCode: "STF-001", JobRoleName: "Staff Accountant", JobRoleDescription: "Accounting staff", JobPositionType: models.PositionTypeStaff, IsActive: true},
	}

	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.JobRole{JobRoleCode: seeds[i].JobRoleCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedLikelihoods() error {
	seeds := []models.Likelihood{
		{Code: "LIK-1", Name: "Rare", Score: 1, Description: "May occur only in exceptional circumstances", IsActive: true},
		{Code: "LIK-2", Name: "Unlikely", Score: 2, Description: "Could occur at some time", IsActive: true},
		{Code: "LIK-3", Name: "Possible", Score: 3, Description: "Might occur at some time", IsActive: true},
		{Code: "LIK-4", Name: "Likely", Score: 4, Description: "Will probably occur", IsActive: true},
		{Code: "LIK-5", Name: "Almost Certain", Score: 5, Description: "Expected to occur in most circumstances", IsActive: true},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.Likelihood{Code: seeds[i].Code}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedImpacts() error {
	seeds := []models.Impact{
		{Code: "IMP-1", Name: "Insignificant", Score: 1, Description: "No significant impact", IsActive: true},
		{Code: "IMP-2", Name: "Minor", Score: 2, Description: "Minor impact", IsActive: true},
		{Code: "IMP-3", Name: "Moderate", Score: 3, Description: "Moderate impact", IsActive: true},
		{Code: "IMP-4", Name: "Major", Score: 4, Description: "Significant impact", IsActive: true},
		{Code: "IMP-5", Name: "Catastrophic", Score: 5, Description: "Extreme impact", IsActive: true},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.Impact{Code: seeds[i].Code}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedRiskLevels() error {
	seeds := []models.RiskLevel{
		{RiskCode: "LOW", RiskName: "Low Risk", RiskDescription: "Acceptable risk level", MinScore: 1, MaxScore: 6, Color: "green", IsActive: true},
		{RiskCode: "MEDIUM", RiskName: "Medium Risk", RiskDescription: "Manage with additional controls", MinScore: 7, MaxScore: 12, Color: "yellow", IsActive: true},
		{RiskCode: "HIGH", RiskName: "High Risk", RiskDescription: "Requires management attention", MinScore: 13, MaxScore: 18, Color: "orange", IsActive: true},
		{RiskCode: "CRITICAL", RiskName: "Critical Risk", RiskDescription: "Requires immediate action", MinScore: 19, MaxScore: 25, Color: "red", IsActive: true},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.RiskLevel{RiskCode: seeds[i].RiskCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedRiskCategories() error {
	seeds := []models.RiskCategory{
		{Code: "RC-STR", Name: "Strategic Risk", Description: "Business strategy and market risks", IsActive: true},
		{Code: "RC-OPR", Name: "Operational Risk", Description: "Process and operations risks", IsActive: true},
		{Code: "RC-FIN", Name: "Financial Risk", Description: "Financial and treasury risks", IsActive: true},
		{Code: "RC-COM", Name: "Compliance Risk", Description: "Regulatory and policy compliance risks", IsActive: true},
		{Code: "RC-IT", Name: "Information Technology Risk", Description: "IT and cybersecurity risks", IsActive: true},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.RiskCategory{Code: seeds[i].Code}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedRiskMatrixCells() error {
	var likelihoods []models.Likelihood
	if err := s.DB.Order("score asc").Find(&likelihoods).Error; err != nil {
		return err
	}
	var impacts []models.Impact
	if err := s.DB.Order("score asc").Find(&impacts).Error; err != nil {
		return err
	}

	lowRisk, err := getByName[models.RiskLevel](s.DB, "risk_code", "LOW")
	if err != nil {
		return err
	}
	mediumRisk, err := getByName[models.RiskLevel](s.DB, "risk_code", "MEDIUM")
	if err != nil {
		return err
	}
	highRisk, err := getByName[models.RiskLevel](s.DB, "risk_code", "HIGH")
	if err != nil {
		return err
	}
	criticalRisk, err := getByName[models.RiskLevel](s.DB, "risk_code", "CRITICAL")
	if err != nil {
		return err
	}

	for _, likelihood := range likelihoods {
		for _, impact := range impacts {
			score := likelihood.Score * impact.Score
			riskLevelID := lowRisk.ID
			if score >= 19 {
				riskLevelID = criticalRisk.ID
			} else if score >= 13 {
				riskLevelID = highRisk.ID
			} else if score >= 7 {
				riskLevelID = mediumRisk.ID
			}

			cell := models.RiskMatrixCell{
				LikelihoodID: likelihood.ID,
				ImpactID:     impact.ID,
				Score:        score,
				RiskLevelID:  riskLevelID,
			}
			if err := s.DB.FirstOrCreate(&cell, models.RiskMatrixCell{
				LikelihoodID: likelihood.ID,
				ImpactID:     impact.ID,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Seeder) SeedBusinessUnits() error {
	company, err := getByName[models.Company](s.DB, "company_code", "AIFL-HQ")
	if err != nil {
		return err
	}

	seeds := []models.BusinessUnit{
		{
			BusinessUnitCode:        "BU-CORP",
			BusinessUnitName:        "Corporate Services",
			BusinessUnitDescription: "Corporate support and governance functions",
			CostCenterCode:          "CC-1000",
			CompanyID:               company.ID,
			IsActive:                true,
		},
		{
			BusinessUnitCode:        "BU-OPS",
			BusinessUnitName:        "Operations",
			BusinessUnitDescription: "Core operational and delivery functions",
			CostCenterCode:          "CC-2000",
			CompanyID:               company.ID,
			IsActive:                true,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.BusinessUnit{BusinessUnitCode: seeds[i].BusinessUnitCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedDepartments() error {
	company, err := getByName[models.Company](s.DB, "company_code", "AIFL-HQ")
	if err != nil {
		return err
	}
	corpBU, err := getByName[models.BusinessUnit](s.DB, "business_unit_code", "BU-CORP")
	if err != nil {
		return err
	}
	opsBU, err := getByName[models.BusinessUnit](s.DB, "business_unit_code", "BU-OPS")
	if err != nil {
		return err
	}

	seeds := []models.Department{
		{
			DepartmentCode:        "DEPT-001",
			DepartmentName:        "Finance & Accounting",
			DepartmentDescription: "Handles financial accounting and reporting",
			Level:                 1,
			CompanyID:             company.ID,
			BusinessUnitID:        corpBU.ID,
			IsActive:              true,
		},
		{
			DepartmentCode:        "DEPT-002",
			DepartmentName:        "Information Technology",
			DepartmentDescription: "Manages infrastructure and applications",
			Level:                 1,
			CompanyID:             company.ID,
			BusinessUnitID:        corpBU.ID,
			IsActive:              true,
		},
		{
			DepartmentCode:        "DEPT-003",
			DepartmentName:        "Internal Audit",
			DepartmentDescription: "Performs independent audit assurance",
			Level:                 1,
			CompanyID:             company.ID,
			BusinessUnitID:        corpBU.ID,
			IsActive:              true,
		},
		{
			DepartmentCode:        "DEPT-004",
			DepartmentName:        "Operations",
			DepartmentDescription: "Runs core operational activities",
			Level:                 1,
			CompanyID:             company.ID,
			BusinessUnitID:        opsBU.ID,
			IsActive:              true,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.Department{DepartmentCode: seeds[i].DepartmentCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedEmployees() error {
	company, err := getByName[models.Company](s.DB, "company_code", "AIFL-HQ")
	if err != nil {
		return err
	}
	hqLocation, err := getByName[models.Location](s.DB, "name", "AIFL Headquarters")
	if err != nil {
		return err
	}

	finDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-001")
	if err != nil {
		return err
	}
	itDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-002")
	if err != nil {
		return err
	}
	auditDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-003")
	if err != nil {
		return err
	}
	opsDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-004")
	if err != nil {
		return err
	}

	dirFinance, err := getByName[models.JobRole](s.DB, "job_role_code", "DIR-001")
	if err != nil {
		return err
	}
	mgrFinance, err := getByName[models.JobRole](s.DB, "job_role_code", "MGR-001")
	if err != nil {
		return err
	}
	mgrIT, err := getByName[models.JobRole](s.DB, "job_role_code", "MGR-002")
	if err != nil {
		return err
	}
	auditMgr, err := getByName[models.JobRole](s.DB, "job_role_code", "AUD-003")
	if err != nil {
		return err
	}
	auditorRole, err := getByName[models.JobRole](s.DB, "job_role_code", "AUD-001")
	if err != nil {
		return err
	}

	joinDate := time.Date(2018, 1, 15, 0, 0, 0, 0, time.UTC)
	employees := []models.Employee{
		{
			EmployeeCode:      "EMP-0001",
			FullName:          "Ahmad Wijaya",
			Email:             "ahmad.wijaya@aifl.co.id",
			Phone:             "+6281234567801",
			CompanyID:         company.ID,
			DepartmentID:      finDept.ID,
			JobRoleID:         dirFinance.ID,
			LevelGrade:        8,
			WorkLocationID:    &hqLocation.ID,
			ResidenceAddress:  "Jl. Sudirman No. 10, Jakarta",
			ResidenceCity:     "Jakarta",
			ResidenceProvince: "DKI Jakarta",
			ResidencePostal:   "10220",
			IsActive:          true,
			JoinDate:          joinDate,
		},
		{
			EmployeeCode:      "EMP-0002",
			FullName:          "Citra Dewi",
			Email:             "citra.dewi@aifl.co.id",
			Phone:             "+6281234567802",
			CompanyID:         company.ID,
			DepartmentID:      finDept.ID,
			JobRoleID:         mgrFinance.ID,
			LevelGrade:        6,
			WorkLocationID:    &hqLocation.ID,
			ResidenceAddress:  "Jl. Thamrin No. 15, Jakarta",
			ResidenceCity:     "Jakarta",
			ResidenceProvince: "DKI Jakarta",
			ResidencePostal:   "10350",
			IsActive:          true,
			JoinDate:          joinDate,
		},
		{
			EmployeeCode:      "EMP-0003",
			FullName:          "Dian Pratama",
			Email:             "dian.pratama@aifl.co.id",
			Phone:             "+6281234567803",
			CompanyID:         company.ID,
			DepartmentID:      itDept.ID,
			JobRoleID:         mgrIT.ID,
			LevelGrade:        6,
			WorkLocationID:    &hqLocation.ID,
			ResidenceAddress:  "Jl. Rasuna Said, Jakarta",
			ResidenceCity:     "Jakarta",
			ResidenceProvince: "DKI Jakarta",
			ResidencePostal:   "12940",
			IsActive:          true,
			JoinDate:          joinDate,
		},
		{
			EmployeeCode:      "EMP-0004",
			FullName:          "Fitri Handayani",
			Email:             "fitri.handayani@aifl.co.id",
			Phone:             "+6281234567804",
			CompanyID:         company.ID,
			DepartmentID:      auditDept.ID,
			JobRoleID:         auditMgr.ID,
			LevelGrade:        6,
			WorkLocationID:    &hqLocation.ID,
			ResidenceAddress:  "Jl. Menteng, Jakarta",
			ResidenceCity:     "Jakarta",
			ResidenceProvince: "DKI Jakarta",
			ResidencePostal:   "10410",
			IsActive:          true,
			JoinDate:          joinDate,
		},
		{
			EmployeeCode:      "EMP-0005",
			FullName:          "Hendra Wijaya",
			Email:             "hendra.wijaya@aifl.co.id",
			Phone:             "+6281234567805",
			CompanyID:         company.ID,
			DepartmentID:      auditDept.ID,
			JobRoleID:         auditorRole.ID,
			LevelGrade:        4,
			WorkLocationID:    &hqLocation.ID,
			ResidenceAddress:  "Jl. Kelapa Gading, Jakarta",
			ResidenceCity:     "Jakarta",
			ResidenceProvince: "DKI Jakarta",
			ResidencePostal:   "14240",
			IsActive:          true,
			JoinDate:          joinDate,
		},
		{
			EmployeeCode:      "EMP-0006",
			FullName:          "Budi Santoso",
			Email:             "budi.santoso@aifl.co.id",
			Phone:             "+6281234567806",
			CompanyID:         company.ID,
			DepartmentID:      opsDept.ID,
			JobRoleID:         mgrFinance.ID,
			LevelGrade:        5,
			WorkLocationID:    &hqLocation.ID,
			ResidenceAddress:  "Jl. Gatot Subroto No. 25, Jakarta",
			ResidenceCity:     "Jakarta",
			ResidenceProvince: "DKI Jakarta",
			ResidencePostal:   "10250",
			IsActive:          true,
			JoinDate:          joinDate,
		},
	}

	for i := range employees {
		if err := s.DB.FirstOrCreate(&employees[i], models.Employee{EmployeeCode: employees[i].EmployeeCode}).Error; err != nil {
			return err
		}
	}

	emp1, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0001")
	if err != nil {
		return err
	}
	emp2, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0002")
	if err != nil {
		return err
	}
	emp3, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0003")
	if err != nil {
		return err
	}
	emp4, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0004")
	if err != nil {
		return err
	}
	emp5, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0005")
	if err != nil {
		return err
	}

	if err := s.DB.Model(&models.Employee{}).Where("employee_code = ?", "EMP-0002").Update("manager_id", emp1.ID).Error; err != nil {
		return err
	}
	if err := s.DB.Model(&models.Employee{}).Where("employee_code = ?", "EMP-0003").Update("manager_id", emp1.ID).Error; err != nil {
		return err
	}
	if err := s.DB.Model(&models.Employee{}).Where("employee_code = ?", "EMP-0004").Update("manager_id", emp1.ID).Error; err != nil {
		return err
	}
	if err := s.DB.Model(&models.Employee{}).Where("employee_code = ?", "EMP-0005").Update("manager_id", emp4.ID).Error; err != nil {
		return err
	}

	if err := s.DB.Model(&models.Department{}).Where("department_code = ?", "DEPT-001").Update("pic_id", emp2.ID).Error; err != nil {
		return err
	}
	if err := s.DB.Model(&models.Department{}).Where("department_code = ?", "DEPT-002").Update("pic_id", emp3.ID).Error; err != nil {
		return err
	}
	if err := s.DB.Model(&models.Department{}).Where("department_code = ?", "DEPT-003").Update("pic_id", emp4.ID).Error; err != nil {
		return err
	}
	if err := s.DB.Model(&models.Department{}).Where("department_code = ?", "DEPT-004").Update("pic_id", emp5.ID).Error; err != nil {
		return err
	}
	return nil
}

func (s *Seeder) SeedRiskRegisters() error {
	itDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-002")
	if err != nil {
		return err
	}
	finDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-001")
	if err != nil {
		return err
	}
	itCat, err := getByName[models.RiskCategory](s.DB, "code", "RC-IT")
	if err != nil {
		return err
	}
	finCat, err := getByName[models.RiskCategory](s.DB, "code", "RC-FIN")
	if err != nil {
		return err
	}
	itOwner, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0003")
	if err != nil {
		return err
	}
	finOwner, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0002")
	if err != nil {
		return err
	}
	l3, err := getByName[models.Likelihood](s.DB, "code", "LIK-3")
	if err != nil {
		return err
	}
	l4, err := getByName[models.Likelihood](s.DB, "code", "LIK-4")
	if err != nil {
		return err
	}
	i3, err := getByName[models.Impact](s.DB, "code", "IMP-3")
	if err != nil {
		return err
	}
	i4, err := getByName[models.Impact](s.DB, "code", "IMP-4")
	if err != nil {
		return err
	}
	mediumRisk, err := getByName[models.RiskLevel](s.DB, "risk_code", "MEDIUM")
	if err != nil {
		return err
	}
	highRisk, err := getByName[models.RiskLevel](s.DB, "risk_code", "HIGH")
	if err != nil {
		return err
	}

	assessmentDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	reviewDate := time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)

	seeds := []models.RiskRegister{
		{
			Code:                 "RR-2026-001",
			Name:                 "Data Breach Risk",
			Description:          "Unauthorized access to sensitive customer data",
			DepartmentID:         itDept.ID,
			RiskCategoryID:       itCat.ID,
			RiskOwnerID:          itOwner.ID,
			InherentLikelihoodID: l4.ID,
			InherentImpactID:     i4.ID,
			InherentScore:        16,
			InherentRiskLevelID:  highRisk.ID,
			ResidualLikelihoodID: l3.ID,
			ResidualImpactID:     i4.ID,
			ResidualScore:        12,
			ResidualRiskLevelID:  mediumRisk.ID,
			AssessmentDate:       assessmentDate,
			NextReviewDate:       reviewDate,
			IsActive:             true,
		},
		{
			Code:                 "RR-2026-002",
			Name:                 "Financial Reporting Error",
			Description:          "Material misstatement in financial reporting process",
			DepartmentID:         finDept.ID,
			RiskCategoryID:       finCat.ID,
			RiskOwnerID:          finOwner.ID,
			InherentLikelihoodID: l3.ID,
			InherentImpactID:     i4.ID,
			InherentScore:        12,
			InherentRiskLevelID:  mediumRisk.ID,
			ResidualLikelihoodID: l3.ID,
			ResidualImpactID:     i3.ID,
			ResidualScore:        9,
			ResidualRiskLevelID:  mediumRisk.ID,
			AssessmentDate:       assessmentDate,
			NextReviewDate:       reviewDate,
			IsActive:             true,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.RiskRegister{Code: seeds[i].Code}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedAuditPeriods() error {
	company, err := getByName[models.Company](s.DB, "company_code", "AIFL-HQ")
	if err != nil {
		return err
	}

	seeds := []models.AuditPeriod{
		{
			PeriodCode: "2026",
			PeriodName: "Annual 2026",
			Year:       2026,
			StartDate:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:     models.AuditPeriodStatusActive,
			CompanyID:  company.ID,
			IsActive:   true,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.AuditPeriod{PeriodCode: seeds[i].PeriodCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedAnnualAuditPlans() error {
	period, err := getByName[models.AuditPeriod](s.DB, "period_code", "2026")
	if err != nil {
		return err
	}
	finDept, _ := getByName[models.Department](s.DB, "department_code", "DEPT-001")
	itDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-002")
	if err != nil {
		return err
	}
	auditDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-003")
	if err != nil {
		return err
	}
	opsDept, _ := getByName[models.Department](s.DB, "department_code", "DEPT-004")

	requester, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0004")
	if err != nil {
		return err
	}
	approver, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0001")
	if err != nil {
		return err
	}
	risk1, err := getByName[models.RiskRegister](s.DB, "code", "RR-2026-001")
	if err != nil {
		return err
	}

	approvedAt := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	seeds := []models.AnnualAuditPlan{
		{
			PlanCode:         "AAP-2026-001",
			PlanTitle:        "Program Kerja Pemeriksaan Tahunan 2026",
			Description:      "Risk-based annual audit plan over operational and IT controls",
			AuditPeriodID:    period.ID,
			RiskRegisterID:   &risk1.ID,
			DepartmentID:     itDept.ID,
			Priority:         models.AuditPlanPriorityHigh,
			Status:           models.AuditPlanStatusApproved,
			EstimatedDays:    77,
			PlannedStartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			PlannedEndDate:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			ApprovedByID:     &approver.ID,
			ApprovedAt:       &approvedAt,
			ApprovalNotes:    "Approved in annual audit committee meeting.",
			RequestedByID:    requester.ID,
			RevisionNumber:   1,
			Activities: []models.AnnualAuditPlanActivity{
				{
					ItemNumber:         "1.1",
					Category:           models.CategoryAssurance,
					GroupTitle:         "Audit Kantor Pusat & Unit Kerja",
					Title:              "Audit Siklus Pendapatan & Piutang",
					InvolvedDepartments: []models.Department{finDept, opsDept},
					TimelineText:       "Jan - Feb",
					AuditorCount:       3,
					TotalMandays:       15,
					SupervisorName:     "Budi Santoso",
					NotesObjective:     "Fokus pada validitas Sales Revenue dan piutang tertagih",
				},
				{
					ItemNumber:         "1.2",
					Category:           models.CategoryAssurance,
					GroupTitle:         "Audit Kantor Pusat & Unit Kerja",
					Title:              "Audit Pengadaan & COGS",
					InvolvedDepartments: []models.Department{finDept, opsDept},
					TimelineText:       "Mar - Apr",
					AuditorCount:       3,
					TotalMandays:       20,
					SupervisorName:     "Budi Santoso",
					NotesObjective:     "Review efisiensi biaya bahan baku & vendor management",
				},
				{
					ItemNumber:         "2.2",
					Category:           models.CategoryAssurance,
					GroupTitle:         "Audit Khusus/Investigasi",
					Title:              "Audit TI (Keamanan Data)",
					InvolvedDepartments: []models.Department{itDept},
					TimelineText:       "Sep",
					AuditorCount:       2,
					TotalMandays:       14,
					SupervisorName:     "Eksternal Ahli",
					NotesObjective:     "Audit kerentanan sistem keuangan & cyber security",
				},
				{
					ItemNumber:         "1.1",
					Category:           models.CategoryConsulting,
					GroupTitle:         "Layanan Konsultansi",
					Title:              "Review Implementasi ERP Baru",
					InvolvedDepartments: []models.Department{itDept, finDept},
					TimelineText:       "Jul - Sep",
					AuditorCount:       2,
					TotalMandays:       30,
					SupervisorName:     "Head of Audit",
					NotesObjective:     "Memberikan masukan kontrol sistem pada modul Keuangan",
				},
			},
		},
		{
			PlanCode:         "AAP-2026-002",
			PlanTitle:        "Internal Audit Quality & Compliance Plan",
			Description:      "Quality assessment of internal audit execution practices",
			AuditPeriodID:    period.ID,
			DepartmentID:     auditDept.ID,
			Priority:         models.AuditPlanPriorityMedium,
			Status:           models.AuditPlanStatusSubmitted,
			EstimatedDays:    12,
			PlannedStartDate: time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
			PlannedEndDate:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
			RequestedByID:    requester.ID,
			RevisionNumber:   1,
			Activities: []models.AnnualAuditPlanActivity{
				{
					ItemNumber:         "1.1",
					Category:           models.CategoryOther,
					GroupTitle:         "Kegiatan Lainnya",
					Title:              "Penyusunan Laporan Tahunan Audit",
					InvolvedDepartments: []models.Department{auditDept},
					TimelineText:       "Des",
					AuditorCount:       3,
					TotalMandays:       10,
					SupervisorName:     "Head of Audit",
					NotesObjective:     "Rekapitulasi temuan dan efektivitas internal control",
				},
			},
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.AnnualAuditPlan{PlanCode: seeds[i].PlanCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedAuditScopes() error {
	plan, err := getByName[models.AnnualAuditPlan](s.DB, "plan_code", "AAP-2026-001")
	if err != nil {
		return err
	}
	itDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-002")
	if err != nil {
		return err
	}
	risk, err := getByName[models.RiskRegister](s.DB, "code", "RR-2026-001")
	if err != nil {
		return err
	}

	seeds := []models.AuditScope{
		{
			ScopeCode:      "SCP-2026-001",
			ScopeName:      "Access Management Controls",
			ScopeType:      models.AuditScopeTypeSystem,
			Description:    "Identity and access lifecycle management scope",
			AuditPlanID:    plan.ID,
			InScope:        "User provisioning, access reviews, privileged access",
			OutOfScope:     "Legacy non-production systems",
			DepartmentID:   &itDept.ID,
			RiskRegisterID: &risk.ID,
			Objectives:     "Assess design and operating effectiveness of IAM controls",
			IsActive:       true,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.AuditScope{ScopeCode: seeds[i].ScopeCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedControls() error {
	risk, err := getByName[models.RiskRegister](s.DB, "code", "RR-2026-001")
	if err != nil {
		return err
	}
	itOwner, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0003")
	if err != nil {
		return err
	}
	itDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-002")
	if err != nil {
		return err
	}

	seeds := []models.Control{
		{
			ControlCode:     "CTL-2026-001",
			ControlName:     "Quarterly User Access Review",
			Description:     "Quarterly review and certification of user access rights",
			ControlType:     models.ControlTypeDetective,
			ControlCategory: models.ControlCategoryITApplication,
			RiskRegisterID:  risk.ID,
			OwnerID:         itOwner.ID,
			DepartmentID:    itDept.ID,
			Frequency:       "Quarterly",
			Documentation:   "IT-SOP-ACCESS-001",
			IsKeyControl:    true,
			IsActive:        true,
		},
		{
			ControlCode:     "CTL-2026-002",
			ControlName:     "Maker Checker for Journal Entry",
			Description:     "Dual-control approval before posting manual journals",
			ControlType:     models.ControlTypePreventive,
			ControlCategory: models.ControlCategoryManual,
			RiskRegisterID:  risk.ID,
			OwnerID:         itOwner.ID,
			DepartmentID:    itDept.ID,
			Frequency:       "Daily",
			Documentation:   "FIN-SOP-JE-007",
			IsKeyControl:    true,
			IsActive:        true,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.Control{ControlCode: seeds[i].ControlCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedControlAssessments() error {
	control, err := getByName[models.Control](s.DB, "control_code", "CTL-2026-001")
	if err != nil {
		return err
	}
	plan, err := getByName[models.AnnualAuditPlan](s.DB, "plan_code", "AAP-2026-001")
	if err != nil {
		return err
	}
	scope, err := getByName[models.AuditScope](s.DB, "scope_code", "SCP-2026-001")
	if err != nil {
		return err
	}
	auditor, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0005")
	if err != nil {
		return err
	}

	assessmentDate := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC)
	seeds := []models.ControlAssessment{
		{
			AssessmentCode:   "CAS-2026-001",
			ControlID:        control.ID,
			AuditPlanID:      plan.ID,
			AuditScopeID:     &scope.ID,
			AssessmentStatus: models.AssessmentStatusCompleted,
			TestingStatus:    models.TestingStatusPassed,
			Effectiveness:    models.ControlEffectivenessEffective,
			TestMethod:       "Inspection and reperformance",
			TestPeriodStart:  time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			TestPeriodEnd:    time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
			SampleSize:       25,
			ExceptionsFound:  1,
			Finding:          "Control generally operating effectively with minor exception.",
			RootCause:        "Late recertification by one business unit.",
			Recommendation:   "Automate reminder escalation workflow.",
			AuditorID:        auditor.ID,
			AssessmentDate:   &assessmentDate,
			CompletedAt:      &completedAt,
			IsActive:         true,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.ControlAssessment{AssessmentCode: seeds[i].AssessmentCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedAuditFindings() error {
	plan, err := getByName[models.AnnualAuditPlan](s.DB, "plan_code", "AAP-2026-001")
	if err != nil {
		return err
	}
	scope, err := getByName[models.AuditScope](s.DB, "scope_code", "SCP-2026-001")
	if err != nil {
		return err
	}
	assessment, err := getByName[models.ControlAssessment](s.DB, "assessment_code", "CAS-2026-001")
	if err != nil {
		return err
	}
	risk, err := getByName[models.RiskRegister](s.DB, "code", "RR-2026-001")
	if err != nil {
		return err
	}
	dept, err := getByName[models.Department](s.DB, "department_code", "DEPT-002")
	if err != nil {
		return err
	}
	owner, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0003")
	if err != nil {
		return err
	}
	auditor, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0005")
	if err != nil {
		return err
	}
	dueDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	seeds := []models.AuditFinding{
		{
			FindingCode:          "FND-2026-001",
			AuditPlanID:          plan.ID,
			AuditScopeID:         &scope.ID,
			ControlAssessmentID:  &assessment.ID,
			FindingType:          models.FindingTypeControlFailure,
			Severity:             models.FindingSeverityMedium,
			Title:                "Delayed Quarterly Access Recertification",
			Description:          "One business unit completed user-access recertification after due date.",
			Facts:                "1 of 25 sampled recertifications was completed 14 days late.",
			Criterion:            "Quarterly recertifications must be completed by period end.",
			Condition:            "Recertification evidence was posted late.",
			Cause:                "Manual follow-up and limited escalation.",
			Effect:               "Increased risk of unauthorized access persistence.",
			RiskRegisterID:       &risk.ID,
			DepartmentID:         dept.ID,
			OwnerID:              owner.ID,
			AuditorID:            auditor.ID,
			Status:               models.FindingStatusOpen,
			AgreedWithManagement: true,
			IssueDate:            time.Date(2026, 2, 26, 0, 0, 0, 0, time.UTC),
			DueDate:              &dueDate,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.AuditFinding{FindingCode: seeds[i].FindingCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedAuditIssues() error {
	finding, err := getByName[models.AuditFinding](s.DB, "finding_code", "FND-2026-001")
	if err != nil {
		return err
	}
	plan, err := getByName[models.AnnualAuditPlan](s.DB, "plan_code", "AAP-2026-001")
	if err != nil {
		return err
	}
	risk, err := getByName[models.RiskRegister](s.DB, "code", "RR-2026-001")
	if err != nil {
		return err
	}
	control, err := getByName[models.Control](s.DB, "control_code", "CTL-2026-001")
	if err != nil {
		return err
	}
	dept, err := getByName[models.Department](s.DB, "department_code", "DEPT-002")
	if err != nil {
		return err
	}
	owner, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0003")
	if err != nil {
		return err
	}
	auditor, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0005")
	if err != nil {
		return err
	}
	dueDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	seeds := []models.AuditIssue{
		{
			IssueCode:           "ISS-2026-001",
			FindingID:           &finding.ID,
			AuditPlanID:         plan.ID,
			Category:            models.IssueCategoryControlDeficiency,
			Priority:            models.IssuePriorityMedium,
			Title:               "Escalation gap in access recertification process",
			Description:         "Escalation does not trigger early enough for overdue certifications.",
			CurrentState:        "Manual reminder only sent near due date.",
			ExpectedState:       "Automated, tiered escalation should trigger before due date.",
			Impact:              "Delayed recertifications increase control exposure.",
			RiskRegisterID:      &risk.ID,
			ControlID:           &control.ID,
			DepartmentID:        dept.ID,
			OwnerID:             owner.ID,
			AuditorID:           auditor.ID,
			Status:              models.IssueStatusInRemediation,
			RootCause:           "Lack of SLA-based workflow automation.",
			IssueDate:           time.Date(2026, 2, 27, 0, 0, 0, 0, time.UTC),
			DueDate:             &dueDate,
			RemediationPlan:     "Implement escalation workflow in IAM tooling with owner dashboards.",
			RemediationEvidence: "Change request CR-SEC-144 initiated.",
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.AuditIssue{IssueCode: seeds[i].IssueCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedAuditRecommendations() error {
	finding, err := getByName[models.AuditFinding](s.DB, "finding_code", "FND-2026-001")
	if err != nil {
		return err
	}
	responsible, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0003")
	if err != nil {
		return err
	}
	verifier, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0004")
	if err != nil {
		return err
	}
	dueDate := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	seeds := []models.AuditRecommendation{
		{
			RecommendationCode: "REC-2026-001",
			FindingID:          finding.ID,
			Recommendation:     "Automate access-recertification reminders and escalation notifications.",
			Priority:           models.RecommendationPriorityShortTerm,
			Benefits:           "Reduces overdue recertifications and improves accountability.",
			ManagementResponse: models.RecommendationResponseAgree,
			ManagementComment:  "IT agrees and will deliver in Q2.",
			ActionPlan:         "Enhance IAM scheduler and manager escalation matrix.",
			ResponsibleID:      responsible.ID,
			DueDate:            &dueDate,
			Status:             models.RecommendationStatusInProgress,
			Progress:           40,
			VerifiedByID:       &verifier.ID,
			CostEstimate:       float64Ptr(15000),
			ActualCost:         float64Ptr(6000),
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.AuditRecommendation{RecommendationCode: seeds[i].RecommendationCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedMitigationActions() error {
	risk, err := getByName[models.RiskRegister](s.DB, "code", "RR-2026-001")
	if err != nil {
		return err
	}
	control, err := getByName[models.Control](s.DB, "control_code", "CTL-2026-001")
	if err != nil {
		return err
	}
	issue, err := getByName[models.AuditIssue](s.DB, "issue_code", "ISS-2026-001")
	if err != nil {
		return err
	}
	owner, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0003")
	if err != nil {
		return err
	}
	dept, err := getByName[models.Department](s.DB, "department_code", "DEPT-002")
	if err != nil {
		return err
	}
	verifier, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0004")
	if err != nil {
		return err
	}

	seeds := []models.MitigationAction{
		{
			ActionCode:         "MIT-2026-001",
			RiskRegisterID:     risk.ID,
			ControlID:          &control.ID,
			IssueID:            &issue.ID,
			ActionTitle:        "Implement SLA-based recertification escalations",
			Description:        "Configure escalating reminders and dashboard tracking for recertification.",
			ActionType:         "enhance_control",
			Status:             models.MitigationStatusInProgress,
			Progress:           models.MitigationProgressInitiated,
			CompletionPercent:  35,
			OwnerID:            owner.ID,
			DepartmentID:       dept.ID,
			BudgetEstimate:     float64Ptr(15000),
			ActualCost:         float64Ptr(4500),
			PlannedStartDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			PlannedEndDate:     time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
			ActualStartDate:    timePtr(time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)),
			VerifiedByID:       &verifier.ID,
			VerificationResult: "Pending completion and evidence review.",
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.MitigationAction{ActionCode: seeds[i].ActionCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedRiskCauses() error {
	risk, err := getByName[models.RiskRegister](s.DB, "code", "RR-2026-001")
	if err != nil {
		return err
	}
	seeds := []models.RiskCause{
		{
			CauseCode:      "RCS-2026-001",
			CauseName:      "Weak periodic access review governance",
			Description:    "Inconsistent ownership and follow-up for access recertification deadlines.",
			Category:       models.RiskCauseCategoryProcess,
			RiskRegisterID: risk.ID,
			IsActive:       true,
		},
		{
			CauseCode:      "RCS-2026-002",
			CauseName:      "Limited automation in IAM reminders",
			Description:    "Manual reminders increase likelihood of delayed responses.",
			Category:       models.RiskCauseCategoryTechnology,
			RiskRegisterID: risk.ID,
			IsActive:       true,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.RiskCause{CauseCode: seeds[i].CauseCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedRiskEffects() error {
	risk, err := getByName[models.RiskRegister](s.DB, "code", "RR-2026-001")
	if err != nil {
		return err
	}
	seeds := []models.RiskEffect{
		{
			EffectCode:     "REF-2026-001",
			EffectName:     "Unauthorized user access may persist",
			Description:    "Expired or inappropriate access can remain active longer than intended.",
			Category:       models.RiskEffectCategoryInformation,
			RiskRegisterID: risk.ID,
			IsActive:       true,
		},
		{
			EffectCode:     "REF-2026-002",
			EffectName:     "Regulatory exposure from control failure",
			Description:    "Weak access controls may create compliance observations.",
			Category:       models.RiskEffectCategoryCompliance,
			RiskRegisterID: risk.ID,
			IsActive:       true,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.RiskEffect{EffectCode: seeds[i].EffectCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedRiskIndicators() error {
	risk, err := getByName[models.RiskRegister](s.DB, "code", "RR-2026-001")
	if err != nil {
		return err
	}
	owner, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0003")
	if err != nil {
		return err
	}

	seeds := []models.RiskIndicator{
		{
			IndicatorCode:  "KRI-2026-001",
			IndicatorName:  "Overdue Access Recertifications",
			Description:    "Percentage of user access recertifications completed after due date",
			RiskRegisterID: risk.ID,
			Metric:         "Overdue Recertification Rate",
			Unit:           "%",
			Frequency:      "monthly",
			ThresholdMin:   float64Ptr(0),
			ThresholdMax:   float64Ptr(5),
			ToleranceLevel: float64Ptr(8),
			CurrentValue:   float64Ptr(6.5),
			LastUpdatedAt:  timePtr(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)),
			Trend:          models.IndicatorDirectionIncreasing,
			TrendComment:   "Slight increase due to delayed certification in one unit",
			Status:         models.IndicatorStatusAlert,
			DataSource:     "IAM Report",
			DataSourceURL:  "https://internal.aifl/reporting/iam/kri-overdue",
			OwnerID:        owner.ID,
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.RiskIndicator{IndicatorCode: seeds[i].IndicatorCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedRiskIndicatorLogs() error {
	indicator, err := getByName[models.RiskIndicator](s.DB, "indicator_code", "KRI-2026-001")
	if err != nil {
		return err
	}
	seeds := []models.RiskIndicatorLog{
		{
			IndicatorID: indicator.ID,
			Value:       4.2,
			RecordedAt:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			Note:        "Within threshold",
		},
		{
			IndicatorID: indicator.ID,
			Value:       6.5,
			RecordedAt:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Note:        "Above threshold due to one delayed certification",
		},
	}

	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.RiskIndicatorLog{
			IndicatorID: seeds[i].IndicatorID,
			RecordedAt:  seeds[i].RecordedAt,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedAuditUniverse() error {
	itDept, err := getByName[models.Department](s.DB, "department_code", "DEPT-002")
	if err != nil {
		return err
	}
	root := models.AuditUniverse{
		EntityCode:      "AUV-2026-001",
		EntityName:      "Enterprise IT Services",
		EntityType:      models.UniverseEntityTypeSystem,
		Description:     "Core enterprise systems and shared IT services",
		BusinessOwner:   "Dian Pratama",
		RiskRating:      "HIGH",
		LastAuditYear:   intPtr(2024),
		LastAuditDate:   timePtr(time.Date(2024, 8, 30, 0, 0, 0, 0, time.UTC)),
		LastAuditResult: "Significant",
		AuditFrequency:  1,
		IsMandatory:     true,
		IsHighPriority:  true,
		DepartmentID:    &itDept.ID,
		Status:          models.UniverseStatusActive,
	}
	if err := s.DB.FirstOrCreate(&root, models.AuditUniverse{EntityCode: root.EntityCode}).Error; err != nil {
		return err
	}

	child := models.AuditUniverse{
		EntityCode:     "AUV-2026-002",
		EntityName:     "Identity and Access Management",
		EntityType:     models.UniverseEntityTypeProcess,
		Description:    "User lifecycle and privileged access management",
		BusinessOwner:  "Dian Pratama",
		ParentID:       &root.ID,
		RiskRating:     "HIGH",
		AuditFrequency: 1,
		IsMandatory:    true,
		IsHighPriority: true,
		DepartmentID:   &itDept.ID,
		Status:         models.UniverseStatusUnderAudit,
	}
	if err := s.DB.FirstOrCreate(&child, models.AuditUniverse{EntityCode: child.EntityCode}).Error; err != nil {
		return err
	}
	return nil
}

func (s *Seeder) SeedAuditWorkpapers() error {
	plan, err := getByName[models.AnnualAuditPlan](s.DB, "plan_code", "AAP-2026-001")
	if err != nil {
		return err
	}
	scope, err := getByName[models.AuditScope](s.DB, "scope_code", "SCP-2026-001")
	if err != nil {
		return err
	}
	assessment, err := getByName[models.ControlAssessment](s.DB, "assessment_code", "CAS-2026-001")
	if err != nil {
		return err
	}
	finding, err := getByName[models.AuditFinding](s.DB, "finding_code", "FND-2026-001")
	if err != nil {
		return err
	}
	auditor, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0005")
	if err != nil {
		return err
	}
	reviewer, err := getByName[models.Employee](s.DB, "employee_code", "EMP-0004")
	if err != nil {
		return err
	}
	reviewedAt := time.Date(2026, 2, 26, 0, 0, 0, 0, time.UTC)

	seeds := []models.AuditWorkpaper{
		{
			WorkpaperCode:       "WP-2026-001",
			Title:               "Access Recertification Sample Testing",
			WorkpaperType:       models.WorkpaperTypeControlTesting,
			Description:         "Sample selection, test procedure, and result of access recertification controls",
			AuditPlanID:         plan.ID,
			AuditScopeID:        &scope.ID,
			ControlAssessmentID: &assessment.ID,
			AuditFindingID:      &finding.ID,
			Content:             "Selected 25 users across critical applications. 1 delayed recertification identified.",
			AttachmentURL:       "https://internal.aifl/documents/workpapers/WP-2026-001",
			Status:              models.WorkpaperStatusInReview,
			Version:             1,
			AuditorID:           auditor.ID,
			ReviewerID:          &reviewer.ID,
			ReviewedAt:          &reviewedAt,
			ReviewNotes:         "Include root-cause follow-up evidence in final pack.",
		},
	}
	for i := range seeds {
		if err := s.DB.FirstOrCreate(&seeds[i], models.AuditWorkpaper{WorkpaperCode: seeds[i].WorkpaperCode}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedVisionMissionGoals() error {
	// Get company AIFL-HQ
	company, err := getByName[models.Company](s.DB, "company_code", "AIFL-HQ")
	if err != nil {
		return err
	}

	// Seeder 1: Published VMG (Tahun Aktif 2026-2031)
	seeds := []models.VisionMissionGoals{
		{
			CompanyID:     company.ID,
			Period:        "2026 - 2031",
			EffectiveDate: timePtr(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
			Vision:        "Menjadi unit audit internal yang terpercaya, profesional, dan memberikan nilai tambah dalam mendukung tata kelola perusahaan yang baik.",
			Mission:       "1. Melaksanakan audit berbasis risiko secara independen dan objektif\n2. Memberikan rekomendasi yang konstruktif untuk perbaikan proses bisnis\n3. Meningkatkan kompetensi dan profesionalisme auditor internal\n4. Mendukung penerapan Good Corporate Governance (GCG)",
			Version:       "v1.0",
			Status:        models.VmgStatusPublished,
			Notes:         "Published VMG untuk periode 2026-2031",
			CreatedBy:     "Ahmad Wijaya",
			ModifiedBy:    "Ahmad Wijaya",
			Goals: []models.VmgGoal{
				{
					GoalCode:           "G-001",
					GoalName:           "Peningkatan Efektivitas Audit Berbasis Risiko",
					GoalDescription:    "Meningkatkan cakupan dan ketepatan audit berbasis risiko",
					StrategicObjective: "SO-001: Risk-Based Audit Excellence",
					KPI:                "Persentase Rencana Audit Terealisasi",
					Target:             "100%",
					Unit:               "%",
					BaselineYear:       "2025",
					BaselineValue:      "90%",
				},
				{
					GoalCode:           "G-002",
					GoalName:           "Peningkatan Kualitas Rekomendasi Audit",
					GoalDescription:    "Meningkatkan rasio tindak lanjut rekomendasi audit oleh auditee",
					StrategicObjective: "SO-002: Action Plan Follow-up",
					KPI:                "Rasio Penyelesaian Temuan",
					Target:             "95%",
					Unit:               "%",
					BaselineYear:       "2025",
					BaselineValue:      "85%",
				},
				{
					GoalCode:           "G-003",
					GoalName:           "Peningkatan Kompetensi Auditor",
					GoalDescription:    "Meningkatkan jumlah sertifikasi profesional yang dimiliki auditor",
					StrategicObjective: "SO-003: Auditor Professionalism",
					KPI:                "Persentase Auditor Bersertifikat (CIA/CISA/QIA)",
					Target:             "80%",
					Unit:               "%",
					BaselineYear:       "2025",
					BaselineValue:      "60%",
				},
				{
					GoalCode:           "G-004",
					GoalName:           "Memperkuat Peran Consulting",
					GoalDescription:    "Meningkatkan kontribusi kegiatan consulting/advisory bagi operasional bisnis",
					StrategicObjective: "SO-004: Value-Added Consulting",
					KPI:                "Indeks Kepuasan Auditee terhadap Advisory",
					Target:             "90%",
					Unit:               "%",
					BaselineYear:       "2025",
					BaselineValue:      "80%",
				},
			},
		},
	}

	// Seeder 2: Published VMG dari periode sebelumnya
	effectiveDate := timePtr(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
	seeds = append(seeds, models.VisionMissionGoals{
		CompanyID:     company.ID,
		Period:        "2021 - 2025",
		EffectiveDate: effectiveDate,
		Vision:        "Menjadi perusahaan investasi terkemuka di Indonesia dengan pertumbuhan berkelanjutan.",
		Mission:       "1. Memberikan nilai tambah bagi shareholder\n2. Mengembangkan SDM yang kompeten\n3. Meningkatkan efisiensi operasional melalui digitalisasi",
		Version:       "v1.0",
		Status:        models.VmgStatusPublished,
		Notes:         "VMG periode 2021-2025 yang sudah selesai",
		CreatedBy:     "Budi Santoso",
		ModifiedBy:    "Budi Santoso",
		Goals: []models.VmgGoal{
			{
				GoalCode:           "G-001",
				GoalName:           "Handal dalam asurans dan consulting ",
				GoalDescription:    "Mengelola resiko untuk mencapai kinerja terbaik",
				StrategicObjective: "SO-001: Sales Growth",
				KPI:                "Total Penjualan",
				Target:             "Rp 1 Triliun",
				Unit:               "Rp",
				BaselineYear:       "2020",
				BaselineValue:      "Rp 750 Miliar",
			},
			{
				GoalCode:           "G-002",
				GoalName:           "Intelligent Risk Based Audit System",
				GoalDescription:    "Implementasi sistem digital untuk seluruh proses bisnis audit",
				StrategicObjective: "SO-002: Technology",
				KPI:                "Prosentase Digitalisasi",
				Target:             "80%",
				Unit:               "%",
				BaselineYear:       "2020",
				BaselineValue:      "30%",
			},
		},
	})

	// Seeder 3: In Review VMG untuk subsidiary
	subsidiary, err := getByName[models.Company](s.DB, "company_code", "AIFL-FIN")
	if err == nil {
		seeds = append(seeds, models.VisionMissionGoals{
			CompanyID:     subsidiary.ID,
			Period:        "2026 - 2030",
			EffectiveDate: timePtr(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)),
			Vision:        "Menjadi perusahaan jasa keuangan terpercaya yang memberikan layanan innovative dan sustainable.",
			Mission:       "1. Menyediakan produk keuangan yang sesuai kebutuhanNasabah\n2. Menerapkan risk management yang prudent\n3. Berkomitmen pada good corporate governance",
			Version:       "v1.0",
			Status:        models.VmgStatusInReview,
			Notes:         "Sedang dalam proses review oleh parent company",
			CreatedBy:     "Citra Dewi",
			ModifiedBy:    "Citra Dewi",
			Goals: []models.VmgGoal{
				{
					GoalCode:           "G-001",
					GoalName:           "AUM Growth",
					GoalDescription:    "Meningkatkan Assets Under Management",
					StrategicObjective: "SO-001: Asset Growth",
					KPI:                "Total AUM",
					Target:             "Rp 5 Triliun",
					Unit:               "Rp",
					BaselineYear:       "2025",
					BaselineValue:      "Rp 3 Triliun",
				},
				{
					GoalCode:           "G-002",
					GoalName:           "Risk Management Excellence",
					GoalDescription:    "Meningkatkan kualitas risk management framework",
					StrategicObjective: "SO-002: Risk Management",
					KPI:                "Risk Maturity Level",
					Target:             "Level 4",
					Unit:               "Level",
					BaselineYear:       "2025",
					BaselineValue:      "Level 3",
				},
				{
					GoalCode:           "G-003",
					GoalName:           "Compliance Rate",
					GoalDescription:    "Memastikan kepatuhan terhadap seluruh regulasi yang berlaku",
					StrategicObjective: "SO-003: Compliance",
					KPI:                "Compliance Index",
					Target:             "98%",
					Unit:               "%",
					BaselineYear:       "2025",
					BaselineValue:      "95%",
				},
				{
					GoalCode:           "G-004",
					GoalName:           "Employee Development",
					GoalDescription:    "Mengembangkan kapabilitas dan kesejahteraan karyawan",
					StrategicObjective: "SO-004: Human Capital",
					KPI:                "Training Hours per Employee",
					Target:             "40",
					Unit:               "Hours",
					BaselineYear:       "2025",
					BaselineValue:      "25 Hours",
				},
			},
		})
	}

	// Insert all VMG records
	for i := range seeds {
		// Check if already exists by Period and CompanyID
		var existing models.VisionMissionGoals
		err := s.DB.Where("period = ? AND company_id = ?", seeds[i].Period, seeds[i].CompanyID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			// Create new record
			if err := s.DB.Create(&seeds[i]).Error; err != nil {
				return fmt.Errorf("failed to seed VisionMissionGoals: %w", err)
			}
			fmt.Printf("  Seeded VisionMissionGoals: %s (%s)\n", seeds[i].Period, seeds[i].Status)
		} else if err != nil {
			return fmt.Errorf("failed to check existing VMG: %w", err)
		} else {
			// update to published just in case
			s.DB.Model(&existing).Update("status", seeds[i].Status)
			fmt.Printf("  Updated VisionMissionGoals: %s (status to %s)\n", seeds[i].Period, seeds[i].Status)
		}
	}

	return nil
}

func (s *Seeder) SeedQAReports() error {
	seeds := []models.QAReport{
		{
			Type:              "Regular Self Assessment (RSA)",
			Period:            "Q3 2025",
			ReportName:        "Operational Efficiency Q3",
			Result:            "8.7/10",
			Status:            "Completed",
			ConductedBy:       "Internal Audit Team A",
			AssessmentTitle:   "RSA - Audit 2025 Q3",
			InternalEvaluator: "Internal Audit Team A",
			Attachment: &models.QAReportAttachment{
				Name:       "Report_RSA_Q3_2025.pdf",
				Size:       "1.5 MB",
				UploadedAt: "2025-10-01",
				FilePath:   "",
			},
		},
		{
			Type:            "Self Assessment w/ Independent Validation (SAIV)",
			Period:          "Cycle 2025",
			ReportName:      "Self Assessment GIAS '22-24",
			Result:          "Generally Conformed",
			Status:          "Completed",
			ConductedBy:     "PT Independent Consultant X",
			AssessmentTitle: "SAIV - Cycle 2025",
			Validator:       "PT Independent Consultant X",
			Attachment: &models.QAReportAttachment{
				Name:       "Certificate_SAIV_2025.pdf",
				Size:       "2.1 MB",
				UploadedAt: "2025-11-15",
				FilePath:   "",
			},
		},
		{
			Type:            "Quality Assurance Review (QAR)",
			Period:          "Year 2025",
			ReportName:      "External QAR (IPPF 2027)",
			Result:          "G/C*",
			Status:          "Completed",
			ConductedBy:     "Deloitte Independent Consultant",
			AssessmentTitle: "QAR - Year 2025",
			Validator:       "Deloitte Independent Consultant",
			Attachment: &models.QAReportAttachment{
				Name:       "Report_External_QAR_2025.pdf",
				Size:       "4.2 MB",
				UploadedAt: "2025-12-20",
				FilePath:   "",
			},
		},
		{
			Type:            "Regular Self Assessment (RSA)",
			Period:          "Q2 2025",
			ReportName:      "Operational Efficiency Q2",
			Result:          "8.3/10",
			Status:          "Completed",
			AssessmentTitle: "RSA - Audit 2025 Q2",
		},
		{
			Type:            "Regular Self Assessment (RSA)",
			Period:          "Q1 2025",
			ReportName:      "Operational Efficiency Q1",
			Result:          "6.9/10",
			Status:          "Completed",
			AssessmentTitle: "RSA - Audit 2025 Q1",
		},
		{
			Type:            "BUMN IACM Assessment",
			Period:          "Year 2025",
			ReportName:      "BUMN IACM Assessment 2025",
			Result:          "4",
			Status:          "Completed",
			ConductedBy:     "BPKP / Kementerian BUMN",
			AssessmentTitle: "BUMN IACM Assessment 2025",
		},
	}

	for i := range seeds {
		var existing models.QAReport
		err := s.DB.Where("report_name = ? AND period = ?", seeds[i].ReportName, seeds[i].Period).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := s.DB.Create(&seeds[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Seeder) SeedConsultingServices() error {
	seeds := []models.ConsultingService{
		{
			Title:          "Review & Assessment SOP Procurement & Purchasing",
			Category:       "Policy & SOP Review",
			RequestorDept:  "Procurement",
			Period:         "Q1 2026",
			ConsultantName: "Budi Hartanto",
			Status:         "Completed",
			Notes:          "Rekomendasi perbaikan kontrol persetujuan pembelian di atas Rp 100jt telah disetujui.",
			Attachment: &models.ConsultingAttachment{
				Name:       "SOP_Procurement_Review_Report.pdf",
				Size:       "1.2 MB",
				UploadedAt: "2026-03-10",
				FilePath:   "",
			},
		},
		{
			Title:          "IT Security Governance Assessment & Training",
			Category:       "IT Advisory",
			RequestorDept:  "IT",
			Period:         "Q2 2026",
			ConsultantName: "Wahyu Hidayat",
			Status:         "In Progress",
			Notes:          "Melakukan review celah keamanan dan melatih staf IT tentang standar ISO 27001.",
			Attachment: &models.ConsultingAttachment{
				Name:       "IT_Security_Audit_Scope.pdf",
				Size:       "850 KB",
				UploadedAt: "2026-05-18",
				FilePath:   "",
			},
		},
		{
			Title:          "Corporate Governance & Compliance Assessment",
			Category:       "Governance Assessment",
			RequestorDept:  "Legal",
			Period:         "Q3 2026",
			ConsultantName: "Carolina Wijaya",
			Status:         "Planned",
			Notes:          "Asesmen keselarasan kebijakan internal perusahaan terhadap regulasi POJK terbaru.",
		},
	}

	for i := range seeds {
		var existing models.ConsultingService
		err := s.DB.Where("title = ? AND period = ?", seeds[i].Title, seeds[i].Period).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := s.DB.Create(&seeds[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func getByName[T any](db *gorm.DB, column string, value interface{}) (T, error) {
	var out T
	if err := db.Where(column+" = ?", value).First(&out).Error; err != nil {
		return out, err
	}
	return out, nil
}

func float64Ptr(v float64) *float64  { return &v }
func intPtr(v int) *int              { return &v }
func timePtr(t time.Time) *time.Time { return &t }
