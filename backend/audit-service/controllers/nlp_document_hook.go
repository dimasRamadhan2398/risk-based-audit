package controllers

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// resolveFeatureSourceFromDir maps an upload storage directory to its AuditSphere feature name.
func resolveFeatureSourceFromDir(dir string) string {
	d := strings.ToLower(dir)
	switch {
	case strings.Contains(d, "working-paper"):
		return "Working Paper"
	case strings.Contains(d, "audit-result"):
		return "Audit Result Report"
	case strings.Contains(d, "executive-summar"):
		return "Executive Summary"
	case strings.Contains(d, "annual-plan"):
		return "Annual Audit Plan"
	case strings.Contains(d, "strategic-plan"):
		return "Strategic Audit Plan"
	case strings.Contains(d, "plan-document"):
		return "Audit Activity Plan"
	case strings.Contains(d, "assignment-letter"):
		return "Assignment Letter"
	case strings.Contains(d, "consulting"):
		return "Consulting Service"
	case strings.Contains(d, "performance-report"):
		return "Performance Report"
	case strings.Contains(d, "charter"):
		return "Audit Charter"
	case strings.Contains(d, "fieldwork"):
		return "Fieldwork Document"
	default:
		return "Uploaded Audit Document"
	}
}

// ExtractDocumentTextGo performs fast native Go text extraction from DOCX, XLSX, PDF, CSV, and TXT files.
func ExtractDocumentTextGo(fileName string, data []byte) string {
	if len(data) == 0 {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".docx":
		if txt := extractDocxXMLText(data); txt != "" {
			return txt
		}
	case ".xlsx":
		if txt := extractXlsxXMLText(data); txt != "" {
			return txt
		}
	case ".pdf":
		if txt := extractPDFStreamText(data); txt != "" {
			return txt
		}
	case ".csv", ".txt", ".md", ".json":
		return strings.TrimSpace(string(data))
	}
	return ""
}

func extractDocxXMLText(data []byte) string {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ""
	}
	var sb strings.Builder
	for _, f := range zr.File {
		if f.Name == "word/document.xml" || strings.HasPrefix(f.Name, "word/header") || strings.HasPrefix(f.Name, "word/footer") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			dec := xml.NewDecoder(rc)
			for {
				tok, err := dec.Token()
				if err != nil {
					break
				}
				switch se := tok.(type) {
				case xml.StartElement:
					if se.Name.Local == "t" {
						var content string
						if err := dec.DecodeElement(&content, &se); err == nil && strings.TrimSpace(content) != "" {
							sb.WriteString(content)
							sb.WriteString(" ")
						}
					} else if se.Name.Local == "p" || se.Name.Local == "tr" {
						sb.WriteString("\n")
					}
				}
			}
			_ = rc.Close()
		}
	}
	return strings.TrimSpace(sb.String())
}

func extractXlsxXMLText(data []byte) string {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ""
	}
	var parts []string
	for _, f := range zr.File {
		if f.Name == "xl/sharedStrings.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			dec := xml.NewDecoder(rc)
			for {
				tok, err := dec.Token()
				if err != nil {
					break
				}
				if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "t" {
					var val string
					if err := dec.DecodeElement(&val, &se); err == nil && strings.TrimSpace(val) != "" {
						parts = append(parts, strings.TrimSpace(val))
					}
				}
			}
			_ = rc.Close()
		}
	}
	return strings.Join(parts, " | ")
}

var pdfTextParenRegex = regexp.MustCompile(`\(([^()\\]{3,240})\)`)

func extractPDFStreamText(data []byte) string {
	var chunks []string
	raw := string(data)
	streamIdx := 0
	for {
		sPos := strings.Index(raw[streamIdx:], "stream")
		if sPos == -1 {
			break
		}
		sPos += streamIdx + 6
		if sPos < len(raw) && raw[sPos] == '\r' {
			sPos++
		}
		if sPos < len(raw) && raw[sPos] == '\n' {
			sPos++
		}
		ePos := strings.Index(raw[sPos:], "endstream")
		if ePos == -1 {
			break
		}
		streamBytes := data[sPos : sPos+ePos]
		streamIdx = sPos + ePos + 9

		zr, err := zlib.NewReader(bytes.NewReader(streamBytes))
		var target string
		if err == nil {
			decompressed, _ := io.ReadAll(io.LimitReader(zr, 2*1024*1024))
			_ = zr.Close()
			target = string(decompressed)
		} else {
			target = string(streamBytes)
		}

		matches := pdfTextParenRegex.FindAllStringSubmatch(target, 60)
		for _, m := range matches {
			if len(m) > 1 {
				clean := strings.TrimSpace(m[1])
				if len(clean) >= 3 && !strings.HasPrefix(clean, "PDF-") && !strings.HasPrefix(clean, "Identity-H") {
					chunks = append(chunks, clean)
				}
			}
		}
	}
	return strings.Join(chunks, " ")
}

// TriggerNLPDocumentAnalysis asynchronously extracts text/OCR and sends the imported document
// to the Python AI IndoBERT service (/predict/text-analysis).
func TriggerNLPDocumentAnalysis(title, description, source, fileName string, data []byte) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[NLP Hook] Recovered in TriggerNLPDocumentAnalysis: %v\n", r)
			}
		}()

		extractedText := ExtractDocumentTextGo(fileName, data)
		if description != "" {
			if extractedText != "" {
				extractedText = description + ". " + extractedText
			} else {
				extractedText = description
			}
		}
		if title == "" && fileName != "" {
			title = strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
		}

		// Encode up to 4MB for Python OCR/deep extraction
		var b64Content string
		if len(data) > 0 && len(data) <= 4*1024*1024 {
			b64Content = base64.StdEncoding.EncodeToString(data)
		}

		ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(fileName), "."))
		if ext == "" {
			ext = "PDF"
		}

		payload := map[string]interface{}{
			"text":              extractedText,
			"title":             title,
			"source":            source,
			"sourceType":        "Imported Document",
			"fileName":          fileName,
			"fileType":          ext,
			"fileContentBase64": b64Content,
			"persist":           true,
		}

		sendToPythonAI(payload)
	}()
}

// TriggerNLPNoteAnalysis asynchronously sends internal feature notes/descriptions to IndoBERT NLP.
func TriggerNLPNoteAnalysis(docID, title, noteText, sourceLabel, riskCategory string) {
	if len(strings.TrimSpace(noteText)) < 12 {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[NLP Hook] Recovered in TriggerNLPNoteAnalysis: %v\n", r)
			}
		}()

		payload := map[string]interface{}{
			"docId":        docID,
			"title":        title,
			"text":         noteText,
			"source":       sourceLabel,
			"sourceType":   "Internal Note",
			"fileName":     sourceLabel,
			"fileType":     "NOTE",
			"riskCategory": riskCategory,
			"persist":      true,
		}
		sendToPythonAI(payload)
	}()
}

func sendToPythonAI(payload map[string]interface{}) {
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

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 12 * time.Second}
	req, err := http.NewRequest("POST", strings.TrimRight(pyURL, "/")+"/predict/text-analysis", bytes.NewReader(bodyBytes))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[NLP Hook] Python AI unreachable (%v)\n", err)
		return
	}
	_ = resp.Body.Close()
}

