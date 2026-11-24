package provisioner

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

// Job statuses.
const (
	JobQueued    = "queued"
	JobRunning   = "running"
	JobSucceeded = "succeeded"
	JobFailed    = "failed"
)

// Event types on the SSE stream.
const (
	EventProgress = "progress"
	EventLog      = "log"
	EventStatus   = "status"
	EventResult   = "result"
)

// Step is one entry of Job.Steps and the data of a "progress" event.
type Step struct {
	Key        string     `json:"key"`
	Index      int        `json:"index"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	Message    string     `json:"message"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

// Result is Job.Result. AdminInitialPassword is filled only in the one
// response that hands it to the creator; it is nil everywhere else,
// including on disk.
type Result struct {
	Slug                 string  `json:"slug,omitempty"`
	FrontendURL          string  `json:"frontend_url"`
	APIURL               string  `json:"api_url"`
	AdminUsername        string  `json:"admin_username"`
	AdminEmail           *string `json:"admin_email"` // null when the script got no --admin-email
	AdminInitialPassword *string `json:"admin_initial_password"`
	// TLS is false when the URLs are plain http: local mode, or production
	// with step 9 skipped (allow_no_tls).
	TLS bool `json:"tls"`
}

// Job is the provisioning job as stored and returned.
type Job struct {
	ID             string     `json:"id"`
	Status         string     `json:"status"`
	Slug           string     `json:"slug"`
	ClientName     string     `json:"client_name"`
	Mode           string     `json:"mode"`
	Domain         string     `json:"domain,omitempty"`
	FEPort         *int       `json:"fe_port"`
	APIPort        *int       `json:"api_port"`
	AdminName      string     `json:"admin_name,omitempty"`
	AdminEmail     string     `json:"admin_email"`
	WithDemoData   bool       `json:"with_demo_data"`
	ProvisionEmail bool       `json:"provision_email"`
	DryRun         bool       `json:"dry_run"`
	AllowNoTLS     bool       `json:"allow_no_tls"`
	HealthTimeout  *int       `json:"health_timeout_seconds"` // null: script default (120 s)
	CreatedBy      string     `json:"created_by"`
	RetryOf        string     `json:"retry_of,omitempty"` // failed job this one retries
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	CurrentStep    string     `json:"current_step"`
	Steps          []Step     `json:"steps"`
	Error          *string    `json:"error"`
	ExitCode       *int       `json:"exit_code,omitempty"`
	Result         *Result    `json:"result"`
}

// Event is one entry of a job's event log.
type Event struct {
	Seq  int64           `json:"seq"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
	TS   time.Time       `json:"ts"`
}

// LogData is the data of a "log" event.
type LogData struct {
	Line   string `json:"line"`
	Stream string `json:"stream"`
	TS     string `json:"ts"`
}

// StatusData is the data of a "status" event.
type StatusData struct {
	Status string  `json:"status"`
	Error  *string `json:"error"`
}

func (j *Job) terminal() bool { return j.Status == JobSucceeded || j.Status == JobFailed }

// clone returns a deep copy, so callers can hand it out without the lock.
func (j *Job) clone() Job {
	c := *j
	c.Steps = append([]Step(nil), j.Steps...)
	if j.Result != nil {
		r := *j.Result
		r.AdminInitialPassword = nil
		c.Result = &r
	}
	return c
}

func initialSteps() []Step {
	steps := make([]Step, len(Steps))
	for i, d := range Steps {
		steps[i] = Step{Key: d.Key, Index: d.Index, Title: d.Title, Status: StepPending}
	}
	return steps
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err) // crypto/rand never fails on supported platforms
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func ptr[T any](v T) *T { return &v }
