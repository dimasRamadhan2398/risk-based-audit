package provisioner

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// CreateRequest is the body of POST /jobs.
type CreateRequest struct {
	ClientName     string `json:"client_name"`
	Slug           string `json:"slug"`
	Mode           string `json:"mode"`
	Domain         string `json:"domain"`
	FEPort         *int   `json:"fe_port"`
	APIPort        *int   `json:"api_port"`
	AdminName      string `json:"admin_name"`
	AdminEmail     string `json:"admin_email"`
	WithDemoData   bool   `json:"with_demo_data"`
	ProvisionEmail *bool  `json:"provision_email"` // omitted means true
	DryRun         bool   `json:"dry_run"`
	AllowNoTLS     bool   `json:"allow_no_tls"`           // production only; true is rejected in local mode
	HealthTimeout  *int   `json:"health_timeout_seconds"` // 30-900; omitted means the script default (120)
	CreatedBy      string `json:"created_by"`             // set by auth-service from the JWT
}

// SlugPattern is the contract's slug rule (also enforced by generate-tenant-stack.sh).
// Bounds for health_timeout_seconds. The script itself accepts 1-1800.
const (
	MinHealthTimeout = 30
	MaxHealthTimeout = 900
)

var SlugPattern = regexp.MustCompile(`^[a-z][a-z0-9]{1,30}$`)

var hostLabel = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// ValidationError lists the invalid fields. Conflict is true when every
// problem is "already taken" (slug or port), which maps to 409 instead of 400.
type ValidationError struct {
	Fields   map[string]string
	Conflict bool
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for k, v := range e.Fields {
		parts = append(parts, k+": "+v)
	}
	return "invalid request: " + strings.Join(parts, "; ")
}

// Reservations is what the semantic checks compare against: the registry and
// the jobs that are still queued or running.
type Reservations struct {
	Slugs map[string]bool
	Ports map[int]bool
	// SlugPorts are the registry ports of each registered slug.
	SlugPorts map[string][]int
	// Retryable maps a registered slug whose latest real (non dry-run)
	// provisioner job failed to that job's id. Such a slug may be provisioned
	// again: the failed run left it in the registry, and the script is
	// idempotent. Its own registry ports count as free for the retry.
	Retryable map[string]string
}

// PortFree reports whether the OS has nothing listening on the port.
type PortFree func(port int) bool

// OSPortFree tries to bind the port on all interfaces.
func OSPortFree(port int) bool {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}

// ReadRegistry reads backend/tenants/registry.tsv (slug, client_name, domain,
// api_domain, fe_port, kong_port, kong_admin_port, redis_db, created_at). A
// missing file is an empty registry.
func ReadRegistry(path string) (Reservations, error) {
	res := Reservations{Slugs: map[string]bool{}, Ports: map[int]bool{}, SlugPorts: map[string][]int{}, Retryable: map[string]string{}}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return res, nil
	}
	if err != nil {
		return res, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		if first {
			first = false
			continue // header
		}
		cols := strings.Split(sc.Text(), "\t")
		if len(cols) == 0 || strings.TrimSpace(cols[0]) == "" {
			continue
		}
		res.Slugs[cols[0]] = true
		for _, idx := range []int{4, 5, 6} { // fe_port, kong_port, kong_admin_port
			if idx < len(cols) {
				if p, err := strconv.Atoi(strings.TrimSpace(cols[idx])); err == nil {
					res.Ports[p] = true
					res.SlugPorts[cols[0]] = append(res.SlugPorts[cols[0]], p)
				}
			}
		}
	}
	return res, sc.Err()
}

// textOK rejects control characters (newlines would break the TSV registry
// and log lines) and a leading '-' (onboard-tenant.sh would read it as a flag).
// Everything else, including quotes, $() and backticks, is plain data: it is
// passed as one argv entry and never reaches a shell parser.
func textOK(s string) string {
	if !utf8.ValidString(s) {
		return "must be valid UTF-8"
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return "must not contain control characters or line breaks"
		}
	}
	if strings.HasPrefix(s, "-") {
		return "must not start with '-'"
	}
	return ""
}

func validEmail(s string) bool {
	if s == "" || len(s) > 254 || strings.ContainsAny(s, " \t<>\"'`$\\") || strings.HasPrefix(s, "-") {
		return false
	}
	a, err := mail.ParseAddress(s)
	if err != nil || a.Address != s {
		return false
	}
	at := strings.LastIndex(s, "@")
	return at > 0 && strings.Contains(s[at+1:], ".")
}

func validHostname(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	labels := strings.Split(s, ".")
	if len(labels) < 2 {
		return false
	}
	for _, l := range labels {
		if !hostLabel.MatchString(l) {
			return false
		}
	}
	return true
}

// Normalize trims the free-text fields and applies defaults.
func (r *CreateRequest) Normalize() {
	r.ClientName = strings.TrimSpace(r.ClientName)
	r.Slug = strings.TrimSpace(r.Slug)
	r.Mode = strings.TrimSpace(r.Mode)
	r.Domain = strings.ToLower(strings.TrimSpace(r.Domain))
	r.AdminName = strings.TrimSpace(r.AdminName)
	r.AdminEmail = strings.TrimSpace(r.AdminEmail)
	r.CreatedBy = strings.TrimSpace(r.CreatedBy)
	if r.ProvisionEmail == nil {
		r.ProvisionEmail = ptr(true)
	}
}

// ValidateFormat checks the request on its own (no registry or OS lookups).
func ValidateFormat(r *CreateRequest) *ValidationError {
	f := map[string]string{}

	if !SlugPattern.MatchString(r.Slug) {
		f["slug"] = "must match ^[a-z][a-z0-9]{1,30}$ (2-31 lowercase letters and digits, starting with a letter)"
	}
	if n := utf8.RuneCountInString(r.ClientName); n < 2 || n > 120 {
		f["client_name"] = "must be 2-120 characters"
	} else if msg := textOK(r.ClientName); msg != "" {
		f["client_name"] = msg
	}
	switch r.Mode {
	case "local", "production":
	default:
		f["mode"] = "must be 'local' or 'production'"
	}
	if r.Domain != "" && r.Mode != "local" {
		if !validHostname(r.Domain) {
			f["domain"] = "must be a lowercase hostname such as bai.example.com"
		} else if !strings.HasPrefix(r.Domain, r.Slug+".") {
			f["domain"] = "must be <slug>.<base domain>, e.g. " + r.Slug + ".auditsphere.app"
		}
	}
	if (r.FEPort == nil) != (r.APIPort == nil) {
		f["fe_port"] = "fe_port and api_port must be given together or both omitted"
	}
	for name, p := range map[string]*int{"fe_port": r.FEPort, "api_port": r.APIPort} {
		if p != nil && (*p < 1024 || *p > 65535) {
			f[name] = "must be between 1024 and 65535"
		}
	}
	if r.FEPort != nil && r.APIPort != nil && *r.FEPort == *r.APIPort {
		f["api_port"] = "must differ from fe_port"
	}
	if utf8.RuneCountInString(r.AdminName) > 120 {
		f["admin_name"] = "must be at most 120 characters"
	} else if r.AdminName != "" {
		if msg := textOK(r.AdminName); msg != "" {
			f["admin_name"] = msg
		}
	}
	if r.AllowNoTLS && r.Mode == "local" {
		f["allow_no_tls"] = "only applies to production mode (local mode never configures TLS)"
	}
	if r.HealthTimeout != nil && (*r.HealthTimeout < MinHealthTimeout || *r.HealthTimeout > MaxHealthTimeout) {
		f["health_timeout_seconds"] = fmt.Sprintf("must be between %d and %d seconds", MinHealthTimeout, MaxHealthTimeout)
	}
	if !validEmail(r.AdminEmail) {
		f["admin_email"] = "must be a valid email address"
	}
	if r.CreatedBy == "" {
		f["created_by"] = "is required (set by auth-service)"
	}
	if len(f) > 0 {
		return &ValidationError{Fields: f}
	}
	return nil
}

// ValidateAvailability checks the slug and ports against the registry, the
// active jobs and (for ports) the OS.
func ValidateAvailability(r *CreateRequest, taken Reservations, portFree PortFree) *ValidationError {
	f := map[string]string{}
	_, retry := taken.Retryable[r.Slug]
	if taken.Slugs[r.Slug] && !retry {
		f["slug"] = fmt.Sprintf("tenant '%s' already exists or is being provisioned", r.Slug)
	}
	own := map[int]bool{}
	if retry {
		for _, p := range taken.SlugPorts[r.Slug] {
			own[p] = true
		}
	}
	for name, p := range map[string]*int{"fe_port": r.FEPort, "api_port": r.APIPort} {
		if p == nil || own[*p] {
			continue // a retry may reuse the ports the failed run registered
		}
		if taken.Ports[*p] {
			f[name] = fmt.Sprintf("port %d is already allocated to a tenant or a queued job", *p)
		} else if portFree != nil && !portFree(*p) {
			f[name] = fmt.Sprintf("port %d is in use on the host", *p)
		}
	}
	if len(f) > 0 {
		return &ValidationError{Fields: f, Conflict: true}
	}
	return nil
}

// ScriptArgs builds the argv for scripts/onboard-tenant.sh (after the script
// path). Each value is its own element; nothing is joined or quoted.
//
// Options come first, then "--", then the positionals, so the client name
// can never be read as an option (contract, "Script CLI details").
func ScriptArgs(j *Job) []string {
	args := []string{"--progress"}
	if j.DryRun {
		args = append(args, "--dry-run")
	}
	if j.Mode == "local" {
		args = append(args, "--local")
	} else if j.Domain != "" {
		args = append(args, "--domain", j.Domain)
	}
	if j.FEPort != nil && j.APIPort != nil {
		args = append(args, "--fe-port", strconv.Itoa(*j.FEPort), "--api-port", strconv.Itoa(*j.APIPort))
	}
	if j.AdminName != "" {
		args = append(args, "--admin-name", j.AdminName)
	}
	args = append(args, "--admin-email", j.AdminEmail)
	if j.WithDemoData {
		args = append(args, "--with-demo-data")
	} else {
		args = append(args, "--empty-data")
	}
	if !j.ProvisionEmail {
		args = append(args, "--skip-email")
	}
	if j.HealthTimeout != nil {
		args = append(args, "--health-timeout", strconv.Itoa(*j.HealthTimeout))
	}
	if j.AllowNoTLS && j.Mode != "local" {
		args = append(args, "--allow-no-tls")
	}
	return append(args, "--", j.Slug, j.ClientName)
}
