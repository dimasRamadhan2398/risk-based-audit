package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
			if err := db.Preload("Profile").Preload("Assessments").Find(&registers).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query risks")
				return
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

			// Create assessments if provided, or default for 2026
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
					if err := db.Create(&ast).Error; err != nil {
						writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to create assessment")
						return
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
				if err := db.Create(&ast).Error; err != nil {
					writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to create default assessment")
					return
				}
			}

			var createdAssessments []models.RiskAssessment
			db.Find(&createdAssessments, "risk_register_id = ?", regID)

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
					Assessments: assessmentsRes,
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

			// Update assessments
			if len(req.Assessments) > 0 {
				for _, astReq := range req.Assessments {
					var existing models.RiskAssessment
					err := db.First(&existing, "risk_register_id = ? AND year = ?", regID, astReq.Year).Error
					if err != nil {
						newAst := models.RiskAssessment{
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
						db.Create(&newAst)
					} else {
						existing.ImpactQ1 = astReq.ImpactQ1
						existing.ImpactQ2 = astReq.ImpactQ2
						existing.ImpactQ3 = astReq.ImpactQ3
						existing.ImpactQ4 = astReq.ImpactQ4
						existing.LikelihoodQ1 = astReq.LikelihoodQ1
						existing.LikelihoodQ2 = astReq.LikelihoodQ2
						existing.LikelihoodQ3 = astReq.LikelihoodQ3
						existing.LikelihoodQ4 = astReq.LikelihoodQ4
						db.Save(&existing)
					}
				}
			}

			var updatedAssessments []models.RiskAssessment
			db.Find(&updatedAssessments, "risk_register_id = ?", regID)

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
					Assessments: assessmentsRes,
				},
			})

		case http.MethodDelete:
			var register models.RiskRegister
			if err := db.First(&register, "id = ?", regID).Error; err != nil {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Risk register not found")
				return
			}

			if err := db.Delete(&register).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to delete risk")
				return
			}

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

			for i := range mitigations {
				if mitigations[i].MonitoringData != "" {
					json.Unmarshal([]byte(mitigations[i].MonitoringData), &mitigations[i].Monitoring)
				}
				if len(mitigations[i].Monitoring) == 0 {
					mitigations[i].Monitoring = models.GenerateMonitoringChecks(mitigations[i].StartDate, mitigations[i].EndDate)
				}
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

			if len(req.Monitoring) == 0 {
				req.Monitoring = models.GenerateMonitoringChecks(req.StartDate, req.EndDate)
			}
			if data, err := json.Marshal(req.Monitoring); err == nil {
				req.MonitoringData = string(data)
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
			
			dateChanged := !mitigation.StartDate.Equal(req.StartDate) || !mitigation.EndDate.Equal(req.EndDate)
			mitigation.StartDate = req.StartDate
			mitigation.EndDate = req.EndDate
			mitigation.Notes = req.Notes

			// Update monitoring
			if len(req.Monitoring) > 0 {
				mitigation.Monitoring = req.Monitoring
			} else if dateChanged {
				mitigation.Monitoring = models.GenerateMonitoringChecks(req.StartDate, req.EndDate)
			} else {
				if mitigation.MonitoringData != "" {
					json.Unmarshal([]byte(mitigation.MonitoringData), &mitigation.Monitoring)
				}
			}

			if data, err := json.Marshal(mitigation.Monitoring); err == nil {
				mitigation.MonitoringData = string(data)
			}

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
		Name         string `json:"name"`
		Impact       int    `json:"impact"`
		Likelihood   int    `json:"likelihood"`
		Severity     int    `json:"severity"`
		Category     string `json:"category"`
		Branch       string `json:"branch"`
		Description  string `json:"description"`
		ImpactQ1     int
		ImpactQ2     int
		ImpactQ3     int
		ImpactQ4     int
		LikelihoodQ1 int
		LikelihoodQ2 int
		LikelihoodQ3 int
		LikelihoodQ4 int
	}

	rawSeeds := []SeedRisk{
		{
			Name: "Target pendapatan dan laba tidak tercapai", Impact: 5, Likelihood: 4, Severity: 98, Category: "Financial", Branch: "Head Office", Description: "Terget pendapatan tidak tercapai karena kinerja tim marketing yang kurang maksimal dan strategi marketing yang tidak efektif.",
			ImpactQ1: 5, ImpactQ2: 4, ImpactQ3: 3, ImpactQ4: 2,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 2,
		},
		{
			Name: "Target efisiensi biaya operasional dan umum tidak tercapai", Impact: 5, Likelihood: 4, Severity: 95, Category: "Financial", Branch: "Head Office", Description: "Target efisiensi biaya operasional dan umum tidak tercapai karena kinerja tim keuangan yang kurang maksimal dan strategi keuangan yang tidak efektif.",
			ImpactQ1: 5, ImpactQ2: 4, ImpactQ3: 3, ImpactQ4: 2,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 3, LikelihoodQ4: 2,
		},
		{
			Name: "Ancaman terhadap Cyber Security dan perlindungan data pribadi", Impact: 5, Likelihood: 4, Severity: 88, Category: "Technology", Branch: "Head Office", Description: "Ancaman terhadap cyber security dan kebocoran data pelanggan/karyawan.",
			ImpactQ1: 5, ImpactQ2: 4, ImpactQ3: 3, ImpactQ4: 2,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 2,
		},
		{
			Name: "Terjadinya fraud", Impact: 4, Likelihood: 4, Severity: 92, Category: "Financial", Branch: "Head Office", Description: "Penyalahgunaan wewenang atau kecurangan keuangan di lingkungan internal.",
			ImpactQ1: 4, ImpactQ2: 3, ImpactQ3: 2, ImpactQ4: 1,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 1,
		},
		{
			Name: "Implementasi teknologi dan digitalisasi tidak berhasil", Impact: 4, Likelihood: 3, Severity: 72, Category: "Technology", Branch: "Bali Branch", Description: "Kegagalan implementasi sistem baru yang menghambat operasional.",
			ImpactQ1: 4, ImpactQ2: 3, ImpactQ3: 2, ImpactQ4: 2,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 2,
		},
		{
			Name: "Pengembangan kompetensi karyawan tidak terlaksana sesuai rencana", Impact: 4, Likelihood: 3, Severity: 58, Category: "Human Resources", Branch: "Head Office", Description: "Kesenjangan keahlian karyawan akibat program training tidak berjalan.",
			ImpactQ1: 4, ImpactQ2: 3, ImpactQ3: 2, ImpactQ4: 2,
			LikelihoodQ1: 3, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 2,
		},
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

	var risk1ID uuid.UUID
	var risk2ID uuid.UUID

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

		if s.Name == "Target pendapatan dan laba tidak tercapai" {
			risk1ID = register.ID
		} else if s.Name == "Target efisiensi biaya operasional dan umum tidak tercapai" {
			risk2ID = register.ID
		}

		ast := models.RiskAssessment{
			ID:             uuid.New(),
			RiskRegisterID: register.ID,
			Year:           2026,
			ImpactQ1:       s.ImpactQ1,
			ImpactQ2:       s.ImpactQ2,
			ImpactQ3:       s.ImpactQ3,
			ImpactQ4:       s.ImpactQ4,
			LikelihoodQ1:   s.LikelihoodQ1,
			LikelihoodQ2:   s.LikelihoodQ2,
			LikelihoodQ3:   s.LikelihoodQ3,
			LikelihoodQ4:   s.LikelihoodQ4,
		}
		if ast.ImpactQ1 == 0 {
			ast.ImpactQ1 = s.Impact
			ast.ImpactQ2 = s.Impact
			ast.ImpactQ3 = s.Impact
			ast.ImpactQ4 = s.Impact
			ast.LikelihoodQ1 = s.Likelihood
			ast.LikelihoodQ2 = s.Likelihood
			ast.LikelihoodQ3 = s.Likelihood
			ast.LikelihoodQ4 = s.Likelihood
		}
		if err := db.Create(&ast).Error; err != nil {
			return err
		}
	}

	// Seed mitigations for Risk 1
	if risk1ID != uuid.Nil {
		mit1 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk1ID,
			RiskEvent:      "Target pendapatan dan laba usaha tidak tercapai",
			MitigationPlan: "Melakukan kampanye promosi penjualan secara konsisten pada seluruh saluran promosi",
			Supervisor:     "Budi Hartanto",
			PIC:            "Carolina Wijaya",
			UnitInCharge:   "Sales & Marketing",
			StartDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit1.Monitoring = models.GenerateMonitoringChecks(mit1.StartDate, mit1.EndDate)
		if data, err := json.Marshal(mit1.Monitoring); err == nil {
			mit1.MonitoringData = string(data)
		}
		db.Create(&mit1)

		mit2 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk1ID,
			RiskEvent:      "Target pendapatan dan laba usaha tidak tercapai",
			MitigationPlan: "Meningkatkan jumlah customer dan nilai pembelian customer",
			Supervisor:     "Budi Hartanto",
			PIC:            "Carolina Wijaya",
			UnitInCharge:   "Sales & Marketing",
			StartDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit2.Monitoring = models.GenerateMonitoringChecks(mit2.StartDate, mit2.EndDate)
		if data, err := json.Marshal(mit2.Monitoring); err == nil {
			mit2.MonitoringData = string(data)
		}
		db.Create(&mit2)

		mit3 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk1ID,
			RiskEvent:      "Target pendapatan dan laba usaha tidak tercapai",
			MitigationPlan: "Memaksimalkan jumlah customer visit",
			Supervisor:     "Budi Hartanto",
			PIC:            "Carolina Wijaya",
			UnitInCharge:   "Sales & Marketing",
			StartDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit3.Monitoring = models.GenerateMonitoringChecks(mit3.StartDate, mit3.EndDate)
		if data, err := json.Marshal(mit3.Monitoring); err == nil {
			mit3.MonitoringData = string(data)
		}
		db.Create(&mit3)

		mit4 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk1ID,
			RiskEvent:      "Target pendapatan dan laba usaha tidak tercapai",
			MitigationPlan: "Peluncuran produk/jasa baru ke pasar",
			Supervisor:     "Anton Hermawan",
			PIC:            "Budi Hartanto & Carolina Wijaya",
			UnitInCharge:   "Product Development",
			StartDate:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit4.Monitoring = models.GenerateMonitoringChecks(mit4.StartDate, mit4.EndDate)
		if data, err := json.Marshal(mit4.Monitoring); err == nil {
			mit4.MonitoringData = string(data)
		}
		db.Create(&mit4)
	}

	// Seed mitigations for Risk 2
	if risk2ID != uuid.Nil {
		mit5 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk2ID,
			RiskEvent:      "Target efisiensi biaya operasional dan umum tidak tercapai",
			MitigationPlan: "Mengontrol biaya operasional dan umum agar efisien namun tetap efektif",
			Supervisor:     "Indarto",
			PIC:            "Wahyu Hidayat & Arief Kuncoro",
			UnitInCharge:   "Operasional",
			StartDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit5.Monitoring = models.GenerateMonitoringChecks(mit5.StartDate, mit5.EndDate)
		if data, err := json.Marshal(mit5.Monitoring); err == nil {
			mit5.MonitoringData = string(data)
		}
		db.Create(&mit5)
	}

	log.Println("Seeding finished successfully.")
	return nil
}
