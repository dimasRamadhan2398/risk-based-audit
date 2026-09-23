package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"risk-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScoreGuideline struct {
	Score int    `json:"score"`
	Desc  string `json:"desc"`
}

func seedRiskFactorsAndUniverse(db *gorm.DB) error {
	log.Println("Seeding Standard Risk Factors...")
	
	// Check if already seeded
	var count int64
	db.Model(&models.StandardRiskFactor{}).Count(&count)
	if count > 0 {
		log.Println("Standard Risk Factors already seeded.")
	} else {
		factors := getStandardRiskFactors()
		for _, f := range factors {
			if err := db.Create(&f).Error; err != nil {
				return fmt.Errorf("failed to seed standard risk factor %s: %w", f.Name, err)
			}
		}
		log.Println("Successfully seeded 24 Standard Risk Factors.")
	}

	log.Println("Seeding Standard Audit Universe...")
	db.Model(&models.StandardAuditUniverse{}).Count(&count)
	if count > 0 {
		log.Println("Standard Audit Universe already seeded.")
	} else {
		if err := seedStandardAuditUniverse(db); err != nil {
			return err
		}
	}

	log.Println("Seeding Corporate Risk Factors...")
	db.Model(&models.CorporateRiskFactor{}).Count(&count)
	if count > 0 {
		log.Println("Corporate Risk Factors already seeded.")
	} else {
		if err := seedCorporateRiskFactors(db); err != nil {
			return err
		}
	}

	log.Println("Seeding Corporate Audit Universe & 2026 establishment...")
	db.Model(&models.CorporateAuditUniverse{}).Count(&count)
	if count > 0 {
		log.Println("Corporate Audit Universe already seeded.")
	} else {
		if err := seedCorporateAuditUniverseAndScores(db); err != nil {
			return err
		}
	}

	return nil
}

func getStandardRiskFactors() []models.StandardRiskFactor {
	now := time.Now()
	
	createGuidelines := func(g5, g4, g3, g2, g1 string) string {
		guidelines := []ScoreGuideline{
			{Score: 5, Desc: g5},
			{Score: 4, Desc: g4},
			{Score: 3, Desc: g3},
			{Score: 2, Desc: g2},
			{Score: 1, Desc: g1},
		}
		b, _ := json.Marshal(guidelines)
		return string(b)
	}

	return []models.StandardRiskFactor{
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000001"),
			Name:        "Financial Materiality",
			Description: "Size of financial exposure",
			ScoreGuidelines: createGuidelines(
				"High – High revenue, expenditure, assets, or transaction values; major contributor to enterprise financial results",
				"Medium to High – Financial value affecting several business units or major operations",
				"Medium - Financial exposure with limited enterprise-wide impact",
				"Low to Medium - Value operations with limited financial significance",
				"Low - Financial activities or administrative support functions",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000002"),
			Name:        "Strategic Importance",
			Description: "Contribution to organizational goals and objectives",
			ScoreGuidelines: createGuidelines(
				"High - Directly supports enterprise strategy, Board priorities, or mission-critical objectives",
				"Medium to High - Supports major strategic programs or key business capabilities",
				"Medium - Contributes indirectly to strategic objectives",
				"Low to Medium - Primarily supports internal operations with limited strategic influence",
				"Low - Administrative or routine activities with minimal strategic impact",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000003"),
			Name:        "Inherent Risk",
			Description: "Natural level of risk before controls",
			ScoreGuidelines: createGuidelines(
				"High – High value transactions, complex operations, significant fraud potential, heavy regulation, critical technology dependence",
				"Medium to High - Complexity with several inherent risk drivers",
				"Medium - Operational and financial exposure",
				"Low to Medium - Complexity and financial impact",
				"Low - Routine, repetitive, and standardized activities with minimal exposure",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000004"),
			Name:        "Internal Control Effectiveness",
			Description: "Effectiveness of existing controls",
			ScoreGuidelines: createGuidelines(
				"High - Controls are ineffective or absent; frequent incidents; risk remains well above tolerance.",
				"Medium to High - Controls only partially effective; several unresolved issues; elevated exposure.",
				"Medium - Controls generally effective with some improvement opportunities.",
				"Low to Medium - Strong controls with only minor residual exposure.",
				"Low - Highly effective controls and continuous monitoring; residual risk is minimal.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000005"),
			Name:        "Regulatory & Compliance Impact",
			Description: "Exposure to regulatory violations",
			ScoreGuidelines: createGuidelines(
				"High - Highly regulated activity, frequent regulatory examinations, severe penalties for non-compliance",
				"Medium to High - Significant compliance obligations with regular reporting requirements",
				"Medium - Moderate regulatory requirements and periodic oversight",
				"Low to Medium - Limited compliance obligations with low enforcement risk",
				"Low - Minimal regulatory exposure beyond standard business requirements",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000006"),
			Name:        "Operational Complexity",
			Description: "Difficulty of business processes",
			ScoreGuidelines: createGuidelines(
				"High - Enterprise-wide operations, multiple systems, global locations, numerous stakeholders, highly integrated processes.",
				"Medium to High - Multi-department processes, significant technology integration, moderate geographic spread.",
				"Medium - Several interconnected processes within one business unit.",
				"Low to Medium - Limited process complexity and organizational dependencies.",
				"Low - Routine administrative activities with standardized procedures.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000007"),
			Name:        "Fraud Exposure",
			Description: "Opportunity for fraud",
			ScoreGuidelines: createGuidelines(
				"High - High-value transactions, extensive management judgment, frequent manual overrides, prior fraud incidents, weak control environment, or elevated corruption risk.",
				"Medium to High - Significant cash/assets, complex transactions, extensive third-party involvement, or previous control weaknesses.",
				"Medium - Some exposure due to manual processes, moderate transaction volume, or limited segregation of duties.",
				"Low to Medium - Limited fraud opportunities with effective preventive and detective controls.",
				"Low - Strong controls, minimal cash handling, limited opportunity for fraud, no history of fraud.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000008"),
			Name:        "Information Security",
			Description: "Cyber security threat",
			ScoreGuidelines: createGuidelines(
				"High - Mission-critical systems, highly sensitive or regulated data, frequent cyber threats, complex IT environment, or previous cybersecurity incidents.",
				"Medium to High - Critical business systems, significant confidential data, internet-facing applications, or known security weaknesses.",
				"Medium - Moderate reliance on IT systems, some sensitive information, standard security controls in place.",
				"Low to Medium - Limited exposure with effective cybersecurity governance and monitoring.",
				"Low - Minimal use of sensitive information, strong security controls, limited external connectivity.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000009"),
			Name:        "Business Continuity",
			Description: "Operational disruption potential",
			ScoreGuidelines: createGuidelines(
				"High - Mission-critical process essential for organizational survival; disruption would cause severe financial, operational, legal, or reputational consequences.",
				"Medium to High - Critical process with strict recovery requirements; disruption significantly affects operations, customers, or compliance.",
				"Medium - Important process requiring documented continuity procedures; moderate financial or operational impact.",
				"Low to Medium - Limited operational impact, alternate procedures available, low customer impact.",
				"Low - Non-critical process, minimal business impact if interrupted, recovery within one business day without formal plans.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000010"),
			Name:        "Change Management",
			Description: "Organizational or process changes",
			ScoreGuidelines: createGuidelines(
				"High - Enterprise-wide transformation, multiple concurrent critical initiatives, major system implementation, or merger with high execution uncertainty and substantial operational impact.",
				"Medium to High - Significant organizational or technology changes affecting multiple functions; increased implementation risk.",
				"Medium - Several moderate initiatives with manageable implementation risk and adequate governance.",
				"Low to Medium - Limited changes affecting a small portion of the organization; changes are well managed.",
				"Low - Stable environment with minimal organizational or technology changes and mature change governance.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000011"),
			Name:        "Third-Party Dependency",
			Description: "Dependency on external providers",
			ScoreGuidelines: createGuidelines(
				"High - Mission-critical operations depend on one or more external providers with limited alternatives; failure of these providers could severely disrupt operations, regulatory compliance, customer service, or financial performance.",
				"Medium to High - Significant reliance on external providers for critical operations, technology, or supply chain activities; disruption would materially affect business performance.",
				"Medium - Moderate dependence on external providers supporting important, but not mission-critical, activities.",
				"Low to Medium - Limited use of third parties with effective oversight and readily available alternatives.",
				"Low - Minimal reliance on external parties; services are non-critical and easily replaceable.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000012"),
			Name:        "Reputational Impact",
			Description: "Potential damage to reputation",
			ScoreGuidelines: createGuidelines(
				"High - Severe and long-lasting reputational damage with major public attention, stakeholder distrust, litigation, or executive accountability.",
				"Medium to High - Significant damage to customer trust, investor confidence, or regulatory relationships. Likely national media coverage.",
				"Medium - Potential negative perception among customers, regulators, or business partners. Moderate media attention possible.",
				"Low to Medium - Limited impact affecting only a small group of stakeholders. Easily managed internally.",
				"Low - Minimal or no impact on organizational reputation. Issues are unlikely to become publicly known.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000013"),
			Name:        "Customer Impact",
			Description: "Risk of damage to customer engagement",
			ScoreGuidelines: createGuidelines(
				"High - Severe impact on a large customer base, widespread service failure, customer harm, substantial complaints, or long-term loss of customer trust.",
				"Medium to High - Significant customer disruption, financial impact, or major decline in customer satisfaction affecting many customers.",
				"Medium - Noticeable customer dissatisfaction or moderate service disruption affecting a defined customer group.",
				"Low to Medium - Limited customer effect involving a small number of customers. Minor inconvenience.",
				"Low - No direct customer interaction. Failures have little or no customer impact.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000014"),
			Name:        "Previous Audit Findings",
			Description: "Prior audit issues",
			ScoreGuidelines: createGuidelines(
				"High - Numerous High or Critical findings, unresolved recommendations, repeated deficiencies, or ineffective management response.",
				"Medium to High - Minor findings identified. Corrective actions completed on time. Strong management response.",
				"Medium - Several moderate findings or isolated overdue actions. Some improvement opportunities remain.",
				"Low to Medium - Minor findings identified. Corrective actions completed on time. Strong management response.",
				"Low - No significant findings in previous audits. Recommendations fully implemented with no repeat issues.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000015"),
			Name:        "Time Since Last Audit",
			Description: "Length of time since the last review",
			ScoreGuidelines: createGuidelines(
				"High – More than 5 years or never audited.",
				"Medium to High – 3-5 years.",
				"Medium – 2-3 years.",
				"Low to Medium – 1-2 years.",
				"Low - Less than 1 year.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000016"),
			Name:        "Management Concern",
			Description: "Management's perception of risk",
			ScoreGuidelines: createGuidelines(
				"High – Critical management concern requiring immediate independent assurance due to significant uncertainty, major incidents, or substantial business risk.",
				"Medium to High – Significant management concern regarding governance, compliance, operational performance, or strategic execution. Management requests priority audit coverage.",
				"Medium – Moderate concerns regarding process efficiency, control effectiveness, or emerging risks. Management requests periodic assurance.",
				"Low to Medium – Minor concerns expressed by management, with no significant operational or control issues.",
				"Low - No significant management concerns. Controls are considered effective and the area is operating as expected.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000017"),
			Name:        "HSE",
			Description: "Frequency of incidents (Health, Safety, and Environment)",
			ScoreGuidelines: createGuidelines(
				"High – Frequent or critical incidents causing major operational disruption, financial loss, regulatory action, safety events, cybersecurity breaches, fraud, or significant reputational damage.",
				"Medium to High – Multiple significant incidents affecting operations, customers, compliance, or financial performance. Repeat incidents indicate control weaknesses.",
				"Medium – Several moderate incidents or an increasing incident trend. Some recurring issues identified.",
				"Low to Medium – Few minor incidents with limited impact. Corrective actions were timely and effective.",
				"Low - No significant incidents during the assessment period. Strong incident prevention and response processes.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000018"),
			Name:        "Process Automation Level",
			Description: "Complexity of process automation",
			ScoreGuidelines: createGuidelines(
				"High – Predominantly manual processes with weak controls, or highly complex automation lacking effective governance, monitoring, cybersecurity, or change management.",
				"Medium to High – Significant manual processing or highly complex automation with governance weaknesses, increasing operational or technology risk.",
				"Medium – Mixed manual and automated processing. Some reliance on manual controls or moderate system complexity.",
				"Low to Medium – Mostly automated with effective controls and only minor manual activities.",
				"Low - Well-controlled automation with strong governance, effective IT controls, reliable monitoring, and minimal manual intervention.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000019"),
			Name:        "Competency of Employee",
			Description: "Competency fit of employees",
			ScoreGuidelines: createGuidelines(
				"High – Critical competency gaps, lack of qualified personnel, ineffective succession planning, or widespread skill deficiencies causing significant operational and control risks.",
				"Medium to High – Significant competency deficiencies, skill shortages, high turnover, or inadequate training affecting operational performance.",
				"Medium – Some competency gaps exist or moderate dependence on key personnel. Additional training is required.",
				"Low to Medium – Competency requirements are generally met with only minor development needs.",
				"Low - Highly competent workforce with appropriate qualifications, certifications, continuous training, and low turnover.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000020"),
			Name:        "Data Sensitivity",
			Description: "Level of data sensitivity",
			ScoreGuidelines: createGuidelines(
				"High – Handles highly confidential, mission-critical, or regulated information where unauthorized disclosure, alteration, or loss would have severe legal, financial, operational, or reputational consequences.",
				"Medium to High – Processes significant volumes of sensitive customer, financial, employee, or proprietary information subject to regulatory requirements.",
				"Medium – Handles confidential business information or moderate volumes of personal data requiring standard protection.",
				"Low to Medium – Processes limited internal-use information with relatively low confidentiality requirements.",
				"Low - Handles only public or non-confidential information with minimal business impact if disclosed.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000021"),
			Name:        "Project Criticality",
			Description: "The importance and potential impact of a project to the organization",
			ScoreGuidelines: createGuidelines(
				"High – Mission-critical enterprise project; failure would significantly affect strategy, operations, or financial performance.",
				"Medium to High – High-impact project involving multiple business units with substantial financial or operational implications.",
				"Medium – Moderate-impact project affecting one or more departments with manageable risks.",
				"Low to Medium – Limited-scope project with minimal organizational impact.",
				"Low - Routine or minor project with negligible impact on business objectives.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000022"),
			Name:        "Legal Exposure",
			Description: "Exposure to legal violations",
			ScoreGuidelines: createGuidelines(
				"High – Extensive legal exposure with major litigation, high-value contractual obligations, or activities where legal failure could materially impact the organization.",
				"Medium to High – Significant legal exposure due to complex contracts, international operations, or recurring legal matters.",
				"Medium – Moderate legal exposure involving important contracts or occasional legal disputes.",
				"Low to Medium – Routine legal obligations with limited contractual or litigation exposure",
				"Low - Minimal legal obligations and very limited exposure to litigation or contractual disputes.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000023"),
			Name:        "ESG Impact",
			Description: "Impact to the environment, social, and governance",
			ScoreGuidelines: createGuidelines(
				"High – Critical ESG impact where failures could materially affect sustainability performance, stakeholder trust, regulatory compliance, or long-term organizational value.",
				"Medium to High – Significant impact on ESG performance with substantial regulatory, operational, or reputational implications.",
				"Medium – Moderate influence on sustainability objectives or stakeholder expectations.",
				"Low to Medium – Limited ESG impact with relatively minor stakeholder implications.",
				"Low - Minimal influence on environmental, social, or governance outcomes.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          uuid.MustParse("10000000-0000-0000-0000-000000000024"),
			Name:        "Emerging Risk",
			Description: "The extent to which an auditable entity is exposed to new, evolving, or rapidly changing risks",
			ScoreGuidelines: createGuidelines(
				"High – Significant exposure to rapidly evolving risks with potentially severe strategic, operational, financial, or regulatory consequences. Governance and controls are still developing.",
				"Medium to High – High exposure to emerging risks requiring enhanced monitoring and management attention.",
				"Medium – Moderate exposure with identifiable risks and partially established controls.",
				"Low to Medium – Limited exposure to emerging risks with manageable uncertainty.",
				"Low - Minimal exposure to new or evolving risks; operating environment is relatively stable.",
			),
			CreatedAt: now, UpdatedAt: now,
		},
	}
}

func seedStandardAuditUniverse(db *gorm.DB) error {
	now := time.Now()
	
	library := map[string][]string{
		"Corporate Governance": {
			"Board of Directors", "Board Committees", "Audit Committee", "Corporate Secretary",
			"Strategic Planning", "Corporate Governance", "Ethics & Compliance", "Sustainability (ESG)",
			"Investor Relations",
		},
		"Enterprise Risk Management": {
			"Risk Management Framework", "Risk Register", "Business Continuity Management",
			"Crisis Management", "Insurance Management", "Fraud Risk Management", "Third Party Risk Management",
		},
		"Finance": {
			"Financial Reporting", "Treasury", "Tax", "Accounts Payable", "Accounts Receivables", "Budgeting",
		},
		"Human Resources": {
			"Recruitment", "Employee on Boarding", "Payroll", "Benefits Administration",
			"Learning & Development", "Performance Management", "Talent Management",
			"Succession Planning", "Employee Relations", "Employee Exit Process",
		},
		"Procurement": {
			"Vendor Selection", "Tender Process", "Contract Management", "Purches Requisition",
			"Purchase Order", "Goods Receipt", "Vendor Performance Evaluation",
		},
		"Supply Chain": {
			"Inventory Management", "Warehouse", "Logistics", "Fleet Management", "Demand Planning",
		},
		"Operations": {
			"Core Business Operations", "Production", "Manufacturing", "Service Delivery",
			"Maintenance", "Quality Control", "Asset Utilization", "Production Planning",
		},
		"Sales & Marketing": {
			"Sales Management", "Customer Relationship Management", "Pricing", "Promotion",
			"Marketing Campaign", "Customer Satisfaction", "Dealer Management", "Sales Incentives",
		},
		"Legal & Compliance": {
			"Contract Review", "Litigation", "Regulatory Compliance", "License Management",
			"Intellectual Property", "Anti-Bribery Compliance", "Personal Data Protection",
		},
		"Asset Management": {
			"Fixed Assets", "Property Management", "Equipment", "Vehicle Management", "Asset Disposal",
		},
		"Projects": {
			"Capital Projects", "Construction Projects", "Digital Transformation", "ERP Implementation",
			"PMO", "Project Governance",
		},
		"Environment, Health & Safety (HSE)": {
			"Occupational Safety", "Environmental Compliance", "Waste Management", "Emergency Response",
			"Incident Investigation",
		},
	}

	// 1. Seed Simple Categories
	for catName, subs := range library {
		catID := uuid.New()
		cat := models.StandardAuditUniverse{
			ID:        catID,
			Name:      catName,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&cat).Error; err != nil {
			return fmt.Errorf("failed to create category %s: %w", catName, err)
		}

		for _, subName := range subs {
			sub := models.StandardAuditUniverse{
				ID:        uuid.New(),
				Name:      subName,
				ParentID:  &catID,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := db.Create(&sub).Error; err != nil {
				return fmt.Errorf("failed to create sub %s under %s: %w", subName, catName, err)
			}
		}
	}

	// 2. Information Technology (Hierarchical)
	itID := uuid.New()
	it := models.StandardAuditUniverse{
		ID:        itID,
		Name:      "Information Technology",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&it).Error; err != nil {
		return err
	}

	itSubtree := map[string][]string{
		"IT Governance":  {"IT Strategy", "IT Organization", "IT Budget"},
		"Infrastructure": {"Network", "Servers", "Cloud", "Data Center"},
		"Cybersecurity":  {"Access Management", "Identity Management", "Vulnerability Management", "Incident Response", "Security Monitoring"},
		"Applications":   {"ERP", "HRIS", "CRM", "AuditSphere", "Finance System"},
		"IT Operations":  {"Change Management", "Backup", "Disaster Recovery", "Help Desk", "Patch Management"},
	}

	for subcatName, entities := range itSubtree {
		subcatID := uuid.New()
		subcat := models.StandardAuditUniverse{
			ID:        subcatID,
			Name:      subcatName,
			ParentID:  &itID,
			Category:  subcatName,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&subcat).Error; err != nil {
			return err
		}

		for _, entName := range entities {
			ent := models.StandardAuditUniverse{
				ID:        uuid.New(),
				Name:      entName,
				ParentID:  &subcatID,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := db.Create(&ent).Error; err != nil {
				return err
			}
		}
	}

	// 3. Subsidiaries/Business Unit
	subID := uuid.New()
	subsidiariesCat := models.StandardAuditUniverse{
		ID:        subID,
		Name:      "Subsidiaries/Business Unit",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&subsidiariesCat).Error; err != nil {
		return err
	}

	// PT A has children
	ptaID := uuid.New()
	pta := models.StandardAuditUniverse{
		ID:        ptaID,
		Name:      "PT A",
		ParentID:  &subID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&pta).Error; err != nil {
		return err
	}
	ptaChildren := []string{"Finance", "Operations", "Procurement", "HR", "IT", "Sales & Marketing"}
	for _, childName := range ptaChildren {
		c := models.StandardAuditUniverse{
			ID:        uuid.New(),
			Name:      childName,
			ParentID:  &ptaID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&c).Error; err != nil {
			return err
		}
	}

	// PT B, PT C
	ptb := models.StandardAuditUniverse{
		ID:        uuid.New(),
		Name:      "PT B",
		ParentID:  &subID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&ptb).Error; err != nil {
		return err
	}
	ptc := models.StandardAuditUniverse{
		ID:        uuid.New(),
		Name:      "PT C",
		ParentID:  &subID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&ptc).Error; err != nil {
		return err
	}

	// 4. Branches/Regional Offices
	brID := uuid.New()
	branchesCat := models.StandardAuditUniverse{
		ID:        brID,
		Name:      "Branches/Regional Offices",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&branchesCat).Error; err != nil {
		return err
	}
	branches := []string{"Jakarta", "Bandung", "Surabaya", "Medan", "Makassar"}
	for _, br := range branches {
		c := models.StandardAuditUniverse{
			ID:        uuid.New(),
			Name:      br,
			ParentID:  &brID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&c).Error; err != nil {
			return err
		}
	}

	// 5. Products/Services
	prodID := uuid.New()
	prodsCat := models.StandardAuditUniverse{
		ID:        prodID,
		Name:      "Products/Services",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&prodsCat).Error; err != nil {
		return err
	}
	prods := []string{"Product/Service A", "Product/Service B", "Product/Service C"}
	for _, p := range prods {
		c := models.StandardAuditUniverse{
			ID:        uuid.New(),
			Name:      p,
			ParentID:  &prodID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&c).Error; err != nil {
			return err
		}
	}

	// 6. Major Business Process
	procID := uuid.New()
	procsCat := models.StandardAuditUniverse{
		ID:        procID,
		Name:      "Major Business Process",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&procsCat).Error; err != nil {
		return err
	}
	procs := []string{"Order to Cash", "Procure to Pay", "Plan to Produce"}
	for _, pr := range procs {
		c := models.StandardAuditUniverse{
			ID:        uuid.New(),
			Name:      pr,
			ParentID:  &procID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&c).Error; err != nil {
			return err
		}
	}

	return nil
}

func seedCorporateRiskFactors(db *gorm.DB) error {
	now := time.Now()
	// Map Table 26 from PDF 1 page 14
	weights := map[string]float64{
		"Financial Materiality":          0.15,
		"Strategic Importance":           0.10,
		"Inherent Risk":                  0.07,
		"HSE":                            0.08,
		"Regulatory & Compliance Impact": 0.06,
		"Operational Complexity":         0.08,
		"Fraud Exposure":                 0.04,
		"Information Security":           0.06,
		"Business Continuity":            0.05,
		"Change Management":              0.05,
		"Third-Party Dependency":         0.06,
		"Reputational Impact":            0.05,
		"Customer Impact":                0.06,
		"Previous Audit Findings":        0.04,
		"ESG Impact":                     0.05,
	}

	for name, w := range weights {
		var sf models.StandardRiskFactor
		if err := db.Where("name = ?", name).First(&sf).Error; err != nil {
			return fmt.Errorf("standard risk factor %s not found for corporate weighting: %w", name, err)
		}
		cf := models.CorporateRiskFactor{
			ID:                   uuid.New(),
			StandardRiskFactorID: sf.ID,
			Weight:               w,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := db.Create(&cf).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedCorporateAuditUniverseAndScores(db *gorm.DB) error {
	now := time.Now()

	// 1. Select the 14 corporate auditable entities from standard catalog
	corporateEntities := []string{
		"Finance",
		"Human Resources",
		"Procurement",
		"Supply Chain",
		"Operations",
		"Sales & Marketing",
		"Information Technology",
		"Legal & Compliance",
		"Asset Management",
		"Projects",
		"Environment, Health & Safety (HSE)",
		"Subsidiaries/Business Unit",
		"Branches/Regional Offices",
		"Products/Services",
	}

	corporateMap := make(map[string]uuid.UUID)

	for _, name := range corporateEntities {
		var standard models.StandardAuditUniverse
		if err := db.Where("name = ? AND parent_id IS NULL", name).First(&standard).Error; err != nil {
			// Find nested branch or sub-entity if needed, otherwise skip or handle
			continue
		}

		corp := models.CorporateAuditUniverse{
			ID:                      uuid.New(),
			StandardAuditUniverseID: &standard.ID,
			Name:                    standard.Name,
			CreatedAt:               now,
			UpdatedAt:               now,
		}
		if err := db.Create(&corp).Error; err != nil {
			return err
		}
		corporateMap[name] = corp.ID

		// Also create corporate sub-entities for this parent standard entity
		var subStandards []models.StandardAuditUniverse
		db.Where("parent_id = ?", standard.ID).Find(&subStandards)
		for _, subSt := range subStandards {
			subID := subSt.ID
			corpSub := models.CorporateAuditUniverse{
				ID:                      uuid.New(),
				StandardAuditUniverseID: &subID,
				Name:                    subSt.Name,
				ParentID:                &corp.ID,
				CreatedAt:               now,
				UpdatedAt:               now,
			}
			db.Create(&corpSub)
		}
	}

	// Hand-craft a couple of specific ones (e.g. Branches/Regional Offices: Jakarta, Subsidiaries: PT A etc.)
	// PT A Subsidiary under Subsidiaries/Business Unit
	var subParent models.CorporateAuditUniverse
	if err := db.Where("name = ? AND parent_id IS NULL", "Subsidiaries/Business Unit").First(&subParent).Error; err == nil {
		pta := models.CorporateAuditUniverse{
			ID:        uuid.New(),
			Name:      "PT A",
			ParentID:  &subParent.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		db.Create(&pta)
		// PT A child Finance
		ptaFinance := models.CorporateAuditUniverse{
			ID:        uuid.New(),
			Name:      "Finance",
			ParentID:  &pta.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		db.Create(&ptaFinance)
	}

	// Jakarta branch under Branches/Regional Offices
	var brParent models.CorporateAuditUniverse
	if err := db.Where("name = ? AND parent_id IS NULL", "Branches/Regional Offices").First(&brParent).Error; err == nil {
		jkt := models.CorporateAuditUniverse{
			ID:        uuid.New(),
			Name:      "Jakarta",
			ParentID:  &brParent.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		db.Create(&jkt)
		corporateMap["Branches/Regional Offices: Jakarta"] = jkt.ID
	}

	// Fetch all corporate risk factors to perform scoring
	var corporateRiskFactors []models.CorporateRiskFactor
	if err := db.Preload("StandardRiskFactor").Find(&corporateRiskFactors).Error; err != nil {
		return err
	}

	// 2. Score 2026 Audit Universe Year
	// Pre-seed scores from PDF:
	// Let's create a mapping of Entity Name -> Risk Factor Name -> Score
	financeScores := map[string]int{
		"Financial Materiality":          5,
		"Strategic Importance":           5,
		"Inherent Risk":                  4,
		"HSE":                            2,
		"Regulatory & Compliance Impact": 5,
		"Operational Complexity":         3,
		"Fraud Exposure":                 4,
		"Information Security":           5,
		"Business Continuity":            5,
		"Change Management":              3,
		"Third-Party Dependency":         3,
		"Reputational Impact":            3,
		"Customer Impact":                4,
		"Previous Audit Findings":        5,
		"ESG Impact":                     3,
	}

	hrScores := map[string]int{
		"Financial Materiality":          4,
		"Strategic Importance":           5,
		"Inherent Risk":                  4,
		"HSE":                            3,
		"Regulatory & Compliance Impact": 4,
		"Operational Complexity":         4,
		"Fraud Exposure":                 4,
		"Information Security":           4,
		"Business Continuity":            4,
		"Change Management":              3,
		"Third-Party Dependency":         3,
		"Reputational Impact":            4,
		"Customer Impact":                4,
		"Previous Audit Findings":        4,
		"ESG Impact":                     4,
	}

	procurementScores := map[string]int{
		"Financial Materiality":          5,
		"Strategic Importance":           5,
		"Inherent Risk":                  4,
		"HSE":                            5,
		"Regulatory & Compliance Impact": 4,
		"Operational Complexity":         4,
		"Fraud Exposure":                 4,
		"Information Security":           4,
		"Business Continuity":            4,
		"Change Management":              3,
		"Third-Party Dependency":         4,
		"Reputational Impact":            3,
		"Customer Impact":                4,
		"Previous Audit Findings":        4,
		"ESG Impact":                     3,
	}

	// Default score map for other entities
	defaultScores := map[string]int{
		"Financial Materiality":          3,
		"Strategic Importance":           3,
		"Inherent Risk":                  3,
		"HSE":                            3,
		"Regulatory & Compliance Impact": 3,
		"Operational Complexity":         3,
		"Fraud Exposure":                 3,
		"Information Security":           3,
		"Business Continuity":            3,
		"Change Management":              3,
		"Third-Party Dependency":         3,
		"Reputational Impact":            3,
		"Customer Impact":                3,
		"Previous Audit Findings":        3,
		"ESG Impact":                     3,
	}

	entitySpecificScores := map[string]map[string]int{
		"Finance":         financeScores,
		"Human Resources": hrScores,
		"Procurement":     procurementScores,
	}

	for entityName, corpID := range corporateMap {
		scoreMap := defaultScores
		if specific, ok := entitySpecificScores[entityName]; ok {
			scoreMap = specific
		}

		// Calculate Weighted score
		totalWeighted := 0.0
		for _, rf := range corporateRiskFactors {
			scoreVal := scoreMap[rf.StandardRiskFactor.Name]
			if scoreVal == 0 {
				scoreVal = 3 // default
			}
			totalWeighted += rf.Weight * float64(scoreVal)
		}

		riskIndex := (totalWeighted / 5.0) * 100.0
		var riskLevel string
		var priority bool

		if riskIndex >= 80.0 {
			riskLevel = "High"
			priority = true
		} else if riskIndex >= 60.0 {
			riskLevel = "Medium to High"
			priority = true
		} else if riskIndex >= 40.0 {
			riskLevel = "Medium"
			priority = false
		} else if riskIndex >= 20.0 {
			riskLevel = "Low to Medium"
			priority = false
		} else {
			riskLevel = "Low"
			priority = false
		}

		// Create AuditUniverseYear
		auYear := models.AuditUniverseYear{
			ID:                       uuid.New(),
			CorporateAuditUniverseID: corpID,
			Year:                     2026,
			RiskIndex:                riskIndex,
			RiskLevel:                riskLevel,
			AuditPriority:            priority,
			CreatedAt:                now,
			UpdatedAt:                now,
		}
		if err := db.Create(&auYear).Error; err != nil {
			return err
		}

		// Create scores
		for _, rf := range corporateRiskFactors {
			scoreVal := scoreMap[rf.StandardRiskFactor.Name]
			if scoreVal == 0 {
				scoreVal = 3
			}
			weightedVal := rf.Weight * float64(scoreVal)

			sc := models.AuditUniverseRiskScore{
				ID:                    uuid.New(),
				AuditUniverseYearID:   auYear.ID,
				CorporateRiskFactorID: rf.ID,
				Score:                 scoreVal,
				WeightedScore:         weightedVal,
				CreatedAt:             now,
				UpdatedAt:             now,
			}
			db.Create(&sc)
		}
	}

	return nil
}
