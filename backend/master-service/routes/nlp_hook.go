package routes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func resolvePythonAIURL() (string, string) {
	pyURL := os.Getenv("PYTHON_AI_URL")
	if pyURL == "" {
		if _, err := os.Stat("/.dockerenv"); err == nil {
			pyURL = "http://python-ai:8000"
		} else {
			pyURL = "http://localhost:8000"
		}
	}
	apiKey := os.Getenv("PYTHON_AI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("AI_API_KEY")
	}
	if apiKey == "" {
		apiKey = "dev-ai-api-key"
	}
	return strings.TrimRight(pyURL, "/"), apiKey
}

// dispatchMasterDocumentToNLP asynchronously sends an imported QA / Consulting document
// or note to the Python AI IndoBERT service (/predict/text-analysis).
func dispatchMasterDocumentToNLP(docID, title, notes, source, sourceType, fileName, riskCategory string, fileBytes []byte) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[Master NLP Hook] Recovered: %v\n", r)
			}
		}()

		var b64 string
		if len(fileBytes) > 0 && len(fileBytes) <= 4*1024*1024 {
			b64 = base64.StdEncoding.EncodeToString(fileBytes)
		}

		ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(fileName), "."))
		if ext == "" {
			if sourceType == "Internal Note" {
				ext = "NOTE"
			} else if sourceType == "External Database" {
				ext = "DB_RECORD"
			} else {
				ext = "PDF"
			}
		}

		payload := map[string]interface{}{
			"docId":             docID,
			"title":             title,
			"text":              notes,
			"source":            source,
			"sourceType":        sourceType,
			"fileName":          fileName,
			"fileType":          ext,
			"fileContentBase64": b64,
			"riskCategory":      riskCategory,
			"persist":           true,
		}

		pyURL, apiKey := resolvePythonAIURL()
		body, err := json.Marshal(payload)
		if err != nil {
			return
		}
		client := &http.Client{Timeout: 12 * time.Second}
		req, err := http.NewRequest("POST", pyURL+"/predict/text-analysis", bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
}

// triggerExternalDBNLPSync triggers multi-source synchronization on Python AI
// when an external database connection is synced in Data Sources.
func triggerExternalDBNLPSync(connName string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[Master NLP Sync] Recovered: %v\n", r)
			}
		}()
		pyURL, apiKey := resolvePythonAIURL()
		client := &http.Client{Timeout: 15 * time.Second}
		req, err := http.NewRequest("POST", pyURL+"/predict/text-analysis/sync-sources", bytes.NewReader([]byte("{}")))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
}

