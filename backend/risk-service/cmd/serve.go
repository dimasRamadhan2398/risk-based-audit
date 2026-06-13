package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start risk-service HTTP server",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"service": "risk-service",
		})
	})

	mux.HandleFunc("/api/v1/risks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": "Risks fetched successfully",
				"data":    []interface{}{},
			})
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"error": map[string]interface{}{
					"code":    "METHOD_NOT_ALLOWED",
					"message": "Method not allowed",
				},
			})
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

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write json response: %v", err)
	}
}
