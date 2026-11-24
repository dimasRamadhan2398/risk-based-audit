package provisioner

import (
	"encoding/json"
	"strings"
)

// StepDef is one of the fixed provisioning steps (docs/provisioning-job-contract.md, "Steps").
type StepDef struct {
	Key   string
	Index int
	Title string
}

// Steps is the fixed, ordered step list the script reports against.
var Steps = []StepDef{
	{"generate", 1, "Generate tenant stack"},
	{"databases", 2, "Create databases"},
	{"build", 3, "Build images"},
	{"migrate", 4, "Migrate and seed"},
	{"storage", 5, "Create storage"},
	{"email", 6, "Provision email domain"},
	{"start", 7, "Start containers"},
	{"health", 8, "Wait for services to be healthy"},
	{"ingress", 9, "Configure nginx and TLS"},
}

// Step statuses. StepPending is the provisioner's initial state; the script
// only ever reports the other four.
const (
	StepPending = "pending"
	StepRunning = "running"
	StepDone    = "done"
	StepSkipped = "skipped"
	StepFailed  = "failed"
)

const (
	progressPrefix = "@@PROGRESS "
	resultPrefix   = "@@RESULT "
)

// Progress is the payload of an @@PROGRESS line.
type Progress struct {
	Key     string `json:"key"`
	Index   int    `json:"index"`
	Total   int    `json:"total"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ScriptResult is the payload of an @@RESULT line. It carries the initial
// admin password, so it must never be logged or persisted as-is.
type ScriptResult struct {
	Slug                 string `json:"slug"`
	FrontendURL          string `json:"frontend_url"`
	APIURL               string `json:"api_url"`
	AdminUsername        string `json:"admin_username"`
	AdminEmail           string `json:"admin_email"`
	AdminInitialPassword string `json:"admin_initial_password"`
}

// LineKind classifies one line of script output.
type LineKind int

const (
	LineLog LineKind = iota
	LineProgress
	LineResult
	// LineMalformed is a line that starts with a protocol prefix but does not
	// parse. Its text is not safe to log verbatim for @@RESULT.
	LineMalformed
)

// ParsedLine is the outcome of ParseLine.
type ParsedLine struct {
	Kind     LineKind
	Progress *Progress
	Result   *ScriptResult
	// Text is the log text (LineLog), or a description of the problem (LineMalformed).
	Text string
}

func stepByKey(key string) (StepDef, bool) {
	for _, s := range Steps {
		if s.Key == key {
			return s, true
		}
	}
	return StepDef{}, false
}

func stepByIndex(i int) (StepDef, bool) {
	if i >= 1 && i <= len(Steps) {
		return Steps[i-1], true
	}
	return StepDef{}, false
}

func validReportedStatus(s string) bool {
	switch s {
	case StepRunning, StepDone, StepSkipped, StepFailed:
		return true
	}
	return false
}

// ParseLine parses one stdout line of `onboard-tenant.sh --progress`. Lines
// without a protocol prefix are plain log lines.
func ParseLine(raw string) ParsedLine {
	line := strings.TrimRight(raw, "\r\n")
	switch {
	case strings.HasPrefix(line, progressPrefix):
		var p Progress
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, progressPrefix)), &p); err != nil {
			return ParsedLine{Kind: LineMalformed, Text: "malformed @@PROGRESS line: " + line}
		}
		if !validReportedStatus(p.Status) {
			return ParsedLine{Kind: LineMalformed, Text: "@@PROGRESS line with unknown status: " + line}
		}
		def, ok := stepByKey(p.Key)
		if !ok {
			// Fall back to the index when the key is missing or unknown.
			if def, ok = stepByIndex(p.Index); !ok {
				return ParsedLine{Kind: LineMalformed, Text: "@@PROGRESS line with unknown step: " + line}
			}
		}
		p.Key, p.Index, p.Total = def.Key, def.Index, len(Steps)
		return ParsedLine{Kind: LineProgress, Progress: &p}
	case strings.HasPrefix(line, resultPrefix):
		var r ScriptResult
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, resultPrefix)), &r); err != nil {
			// The line may hold the password: never echo it.
			return ParsedLine{Kind: LineMalformed, Text: "malformed @@RESULT line (content hidden)"}
		}
		return ParsedLine{Kind: LineResult, Result: &r}
	default:
		return ParsedLine{Kind: LineLog, Text: line}
	}
}
