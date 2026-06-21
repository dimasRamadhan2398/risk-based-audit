package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"risk-service/models"
	"risk-service/pkg/database"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start risk-service HTTP server",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

// Static mappings between Branch names (frontend) and UUIDs (backend DB)
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

type RiskResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Impact      int    `json:"impact"`
	Likelihood  int    `json:"likelihood"`
	Severity    int    `json:"severity"`
	Category    string `json:"category"`
	Branch      string `json:"branch"`
	Description string `json:"description"`
}

type RiskRequest struct {
	Name        string `json:"name"`
	Impact      int    `json:"impact"`
	Likelihood  int    `json:"likelihood"`
	Severity    int    `json:"severity"`
	Category    string `json:"category"`
	Branch      string `json:"branch"`
	Description string `json:"description"`
}

func runServe(cmd *cobra.Command, args []string) error {
	if err := initConfig(); err != nil {
		return err
	}
	if err := initLogger(); err != nil {
		return err
	}

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return err
	}

	// Seed initial risks if the DB is empty
	if err := seedInitialRisks(db); err != nil {
		log.Printf("Warning: Seeding failed: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"service": "risk-service",
		})
	})

	// GET & POST risks
	mux.HandleFunc("/api/v1/risks", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case http.MethodGet:
			var registers []models.RiskRegister
			if err := db.Preload("Profile").Find(&registers).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query risks")
				return
			}

			data := make([]RiskResponse, 0)
			for _, reg := range registers {
				branch := uuidToBranch[reg.Profile.DepartmentID.String()]
				if branch == "" {
					branch = "Head Office"
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
				})
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": "Risks fetched successfully",
				"data":    data,
			})

		case http.MethodPost:
			var req RiskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
				return
			}

			pID := uuid.New()
			profile := models.RiskProfile{
				ID:           pID,
				DepartmentID: uuid.MustParse(branchToUUID[req.Branch]),
				OwnerID:      uuid.Nil,
				Category:     req.Category,
				Description:  req.Description,
			}
			if err := db.Create(&profile).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to create risk profile")
				return
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
			if err := db.Create(&register).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to create risk register")
				return
			}

			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"success": true,
				"message": "Risk created successfully",
				"data": RiskResponse{
					ID:          regID.String(),
					Name:        req.Name,
					Impact:      req.Impact,
					Likelihood:  req.Likelihood,
					Severity:    req.Severity,
					Category:    req.Category,
					Branch:      req.Branch,
					Description: req.Description,
				},
			})

		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	})

	// PUT & DELETE risks
	mux.HandleFunc("/api/v1/risks/", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/risks/")
		regID, err := uuid.Parse(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid risk ID format")
			return
		}

		switch r.Method {
		case http.MethodPut:
			var req RiskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
				return
			}

			var register models.RiskRegister
			if err := db.Preload("Profile").First(&register, "id = ?", regID).Error; err != nil {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Risk register not found")
				return
			}

			// Update register fields
			register.RiskEvent = req.Name
			register.InherentLikelihood = req.Likelihood
			register.InherentImpact = req.Impact
			register.InherentScore = req.Severity
			register.ResidualScore = req.Severity
			if err := db.Save(&register).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to update risk register")
				return
			}

			// Update associated profile fields
			register.Profile.Category = req.Category
			register.Profile.Description = req.Description
			register.Profile.DepartmentID = uuid.MustParse(branchToUUID[req.Branch])
			if err := db.Save(&register.Profile).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to update risk profile")
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": "Risk updated successfully",
				"data": RiskResponse{
					ID:          regID.String(),
					Name:        req.Name,
					Impact:      req.Impact,
					Likelihood:  req.Likelihood,
					Severity:    req.Severity,
					Category:    req.Category,
					Branch:      req.Branch,
					Description: req.Description,
				},
			})

		case http.MethodDelete:
			var register models.RiskRegister
			if err := db.First(&register, "id = ?", regID).Error; err != nil {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Risk register not found")
				return
			}

			// Delete register (GORM will soft-delete if DeletedAt is set, otherwise hard delete)
			if err := db.Delete(&register).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to delete risk")
				return
			}

			// Delete associated profile
			db.Delete(&models.RiskProfile{}, "id = ?", register.ProfileID)

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": "Risk deleted successfully",
			})

		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	})

	// GET & POST mitigations
	mux.HandleFunc("/api/v1/mitigations", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case http.MethodGet:
			riskIdQuery := r.URL.Query().Get("riskId")
			var mitigations []models.RiskMitigation
			query := db
			if riskIdQuery != "" {
				parsedRiskId, err := uuid.Parse(riskIdQuery)
				if err == nil {
					query = query.Where("risk_id = ?", parsedRiskId)
				}
			}
			if err := query.Find(&mitigations).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query mitigations")
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": "Mitigations fetched successfully",
				"data":    mitigations,
			})

		case http.MethodPost:
			var req models.RiskMitigation
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
				return
			}
			if req.ID == uuid.Nil {
				req.ID = uuid.New()
			}
			if err := db.Create(&req).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to create mitigation")
				return
			}

			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"success": true,
				"message": "Mitigation created successfully",
				"data":    req,
			})

		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	})

	// PUT & DELETE mitigations
	mux.HandleFunc("/api/v1/mitigations/", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/mitigations/")
		mitID, err := uuid.Parse(idStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid mitigation ID format")
			return
		}

		switch r.Method {
		case http.MethodPut:
			var req models.RiskMitigation
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
				return
			}

			var mitigation models.RiskMitigation
			if err := db.First(&mitigation, "id = ?", mitID).Error; err != nil {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Mitigation not found")
				return
			}

			// Update fields
			mitigation.RiskEvent = req.RiskEvent
			mitigation.MitigationPlan = req.MitigationPlan
			mitigation.Supervisor = req.Supervisor
			mitigation.PIC = req.PIC
			mitigation.UnitInCharge = req.UnitInCharge
			mitigation.StartDate = req.StartDate
			mitigation.EndDate = req.EndDate
			mitigation.Notes = req.Notes
			if err := db.Save(&mitigation).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to update mitigation")
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": "Mitigation updated successfully",
				"data":    mitigation,
			})

		case http.MethodDelete:
			var mitigation models.RiskMitigation
			if err := db.First(&mitigation, "id = ?", mitID).Error; err != nil {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Mitigation not found")
				return
			}

			if err := db.Delete(&mitigation).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to delete mitigation")
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": "Mitigation deleted successfully",
			})

		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	addr := ":" + port
	log.Printf("Starting risk-service HTTP server on %s", addr)

	return http.ListenAndServe(addr, mux)
}

func enableCors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, statusCode int, code, message string) {
	writeJSON(w, statusCode, map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

func seedInitialRisks(db *gorm.DB) error {
	var count int64
	db.Model(&models.RiskRegister{}).Count(&count)
	if count > 0 {
		return nil
	}

	log.Println("Seeding initial risk data...")

	type SeedRisk struct {
		Name        string `json:"name"`
		Impact      int    `json:"impact"`
		Likelihood  int    `json:"likelihood"`
		Severity    int    `json:"severity"`
		Category    string `json:"category"`
		Branch      string `json:"branch"`
		Description string `json:"description"`
	}

	rawSeeds := []SeedRisk{
		{Name: "Target pendapatan dan laba tidak tercapai", Impact: 5, Likelihood: 4, Severity: 98, Category: "Financial", Branch: "Head Office", Description: "Terget pendapatan tidak tercapai karena kinerja tim marketing yang kurang maksimal dan strategi marketing yang tidak efektif."},
		{Name: "Target efisiensi biaya operasional dan umum tidak tercapai", Impact: 5, Likelihood: 4, Severity: 95, Category: "Financial", Branch: "Head Office", Description: "Target efisiensi biaya operasional dan umum tidak tercapai karena kinerja tim keuangan yang kurang maksimal dan strategi keuangan yang tidak efektif."},
		{Name: "Regulatory Non-Compliance", Impact: 4, Likelihood: 4, Severity: 88, Category: "Compliance", Branch: "Surabaya Branch", Description: "Failure to adhere to government regulations, industry standards, or legal requirements."},
		{Name: "Abuse of Power / Authority", Impact: 5, Likelihood: 5, Severity: 92, Category: "Governance", Branch: "Head Office", Description: "Misuse of managerial or executive authority for personal gain or organizational harm."},
		{Name: "Supply Chain Disruption", Impact: 4, Likelihood: 3, Severity: 72, Category: "Operations", Branch: "Bandung Branch", Description: "Critical interruptions in the supply chain due to vendor failures, logistics, or global events."},
		{Name: "Data Privacy Violation", Impact: 5, Likelihood: 3, Severity: 85, Category: "Compliance", Branch: "Bali Branch", Description: "Breaches of customer or employee data privacy, violating GDPR/local data protection laws."},
		{Name: "Market Volatility Exposure", Impact: 3, Likelihood: 4, Severity: 65, Category: "Financial", Branch: "Jakarta Branch", Description: "Financial losses due to unpredictable market fluctuations, currency risks, or commodity prices."},
		{Name: "Talent Attrition / Brain Drain", Impact: 3, Likelihood: 3, Severity: 50, Category: "Human Resources", Branch: "Head Office", Description: "Loss of key employees and institutional knowledge affecting operational continuity."},
		{Name: "Reputational Damage", Impact: 4, Likelihood: 2, Severity: 75, Category: "Strategic", Branch: "Surabaya Branch", Description: "Significant brand damage due to public scandals, social media crises, or product failures."},
		{Name: "Environmental Compliance Failure", Impact: 3, Likelihood: 2, Severity: 55, Category: "Compliance", Branch: "Bandung Branch", Description: "Violations of environmental regulations leading to fines, shutdowns, or cleanup obligations."},
		{Name: "Operational System Failure", Impact: 4, Likelihood: 3, Severity: 70, Category: "Technology", Branch: "Bali Branch", Description: "Critical failure in core business systems causing operational downtime and revenue loss."},
		{Name: "Insider Trading", Impact: 5, Likelihood: 2, Severity: 90, Category: "Financial", Branch: "Head Office", Description: "Illegal trading of securities based on material, non-public information by employees."},
		{Name: "Workplace Safety Incident", Impact: 3, Likelihood: 2, Severity: 58, Category: "Human Resources", Branch: "Surabaya Branch", Description: "Accidents or hazardous conditions leading to employee injury or regulatory action."},
		{Name: "Third-Party Vendor Risk", Impact: 2, Likelihood: 3, Severity: 40, Category: "Operations", Branch: "Jakarta Branch", Description: "Risks arising from outsourced vendors failing to meet service, security, or compliance standards."},
		{Name: "Intellectual Property Theft", Impact: 4, Likelihood: 2, Severity: 78, Category: "Strategic", Branch: "Bandung Branch", Description: "Unauthorized copying, use, or distribution of company trade secrets and proprietary technology."},
		{Name: "Natural Disaster Impact", Impact: 5, Likelihood: 1, Severity: 60, Category: "Operations", Branch: "Bali Branch", Description: "Disruption from earthquakes, floods, hurricanes, or other catastrophic natural events."},
		{Name: "Interest Rate Fluctuation", Impact: 2, Likelihood: 4, Severity: 35, Category: "Financial", Branch: "Head Office", Description: "Exposure to changing interest rates affecting debt servicing and investment returns."},
		{Name: "Political / Geopolitical Risk", Impact: 3, Likelihood: 3, Severity: 52, Category: "Strategic", Branch: "Jakarta Branch", Description: "Business disruption from political instability, sanctions, trade wars, or regime changes."},
		{Name: "Product Liability", Impact: 4, Likelihood: 1, Severity: 68, Category: "Operations", Branch: "Surabaya Branch", Description: "Legal liability from defective products causing harm to consumers or businesses."},
		{Name: "Pandemic / Health Crisis", Impact: 5, Likelihood: 2, Severity: 82, Category: "Operations", Branch: "Head Office", Description: "Widespread health emergencies causing workforce disruption and operational shutdowns."},
	}

	for _, s := range rawSeeds {
		pID := uuid.New()
		profile := models.RiskProfile{
			ID:           pID,
			DepartmentID: uuid.MustParse(branchToUUID[s.Branch]),
			OwnerID:      uuid.Nil,
			Category:     s.Category,
			Description:  s.Description,
		}
		if err := db.Create(&profile).Error; err != nil {
			return err
		}

		register := models.RiskRegister{
			ID:                   uuid.New(),
			ProfileID:            pID,
			RiskSource:           models.RiskSourceDirect,
			RiskEvent:            s.Name,
			InherentLikelihood:   s.Likelihood,
			InherentImpact:       s.Impact,
			InherentScore:        s.Severity,
			ControlEffectiveness: 0,
			ResidualScore:        s.Severity,
			FinalRiskLevel:       models.RiskFinalLevelHigh,
			Status:               models.RiskRegisterStatusApproved,
		}
		if err := db.Create(&register).Error; err != nil {
			return err
		}
	}

	log.Println("Seeding finished successfully.")
	return nil
}
