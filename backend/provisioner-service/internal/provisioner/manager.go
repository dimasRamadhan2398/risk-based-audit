package provisioner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrNotFound is returned for an unknown job id.
var ErrNotFound = errors.New("job not found")

// ErrRestarted is the error a job gets when the provisioner restarts mid-run.
const ErrRestarted = "provisioner restarted"

// Manager owns the jobs, the single-worker queue and the event logs.
type Manager struct {
	mu     sync.Mutex
	store  *Store
	runner Runner
	log    *slog.Logger
	now    func() time.Time
	jobs   map[string]*entry
	queue  []string
	wake   chan struct{}

	registryPath string
	portFree     PortFree
}

type entry struct {
	job Job
	// password is the initial admin password from @@RESULT. Memory only.
	password     string
	passwordRead bool

	events       []Event
	eventsLoaded bool
	nextSeq      int64
	// changed is closed (and replaced) whenever an event is appended or the
	// job changes, waking SSE streams.
	changed chan struct{}
}

// ManagerOptions configures NewManager.
type ManagerOptions struct {
	Store        *Store
	Runner       Runner
	Logger       *slog.Logger
	RegistryPath string   // backend/tenants/registry.tsv
	PortFree     PortFree // nil skips the OS port check
	Now          func() time.Time
}

// NewManager loads persisted jobs, fails the ones that were running and
// re-queues the ones that were queued.
func NewManager(o ManagerOptions) (*Manager, error) {
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
	if o.Now == nil {
		o.Now = func() time.Time { return time.Now().UTC() }
	}
	m := &Manager{
		store: o.Store, runner: o.Runner, log: o.Logger, now: o.Now,
		jobs: map[string]*entry{}, wake: make(chan struct{}, 1),
		registryPath: o.RegistryPath, portFree: o.PortFree,
	}
	jobs, errs := m.store.LoadJobs()
	for _, err := range errs {
		m.log.Warn("skipping unreadable job", "error", err)
	}
	sort.Slice(jobs, func(a, b int) bool { return jobs[a].CreatedAt.Before(jobs[b].CreatedAt) })
	for _, j := range jobs {
		e := &entry{job: *j, changed: make(chan struct{})}
		m.jobs[j.ID] = e
		if j.terminal() {
			continue
		}
		if err := m.loadEventsLocked(e); err != nil {
			m.log.Warn("cannot read job events", "job", j.ID, "error", err)
			e.eventsLoaded = true
		}
		switch j.Status {
		case JobRunning:
			m.finishLocked(e, JobFailed, ptr(ErrRestarted))
			m.log.Warn("job was running when the provisioner stopped; marked failed", "job", j.ID, "slug", j.Slug)
		case JobQueued:
			m.queue = append(m.queue, j.ID)
		}
	}
	return m, nil
}

func (m *Manager) loadEventsLocked(e *entry) error {
	if e.eventsLoaded {
		return nil
	}
	evs, err := m.store.LoadEvents(e.job.ID)
	if err != nil {
		return err
	}
	e.events, e.eventsLoaded = evs, true
	for _, ev := range evs {
		if ev.Seq > e.nextSeq {
			e.nextSeq = ev.Seq
		}
	}
	return nil
}

func (e *entry) notify() {
	close(e.changed)
	e.changed = make(chan struct{})
}

// emitLocked appends an event to memory and disk and wakes listeners.
func (m *Manager) emitLocked(e *entry, typ string, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		m.log.Error("marshal event", "error", err)
		return
	}
	e.nextSeq++
	ev := Event{Seq: e.nextSeq, Type: typ, Data: b, TS: m.now()}
	e.events = append(e.events, ev)
	if err := m.store.AppendEvent(e.job.ID, ev); err != nil {
		m.log.Error("persist event", "job", e.job.ID, "error", err)
	}
	e.notify()
}

func (m *Manager) saveLocked(e *entry) {
	if err := m.store.SaveJob(&e.job); err != nil {
		m.log.Error("persist job", "job", e.job.ID, "error", err)
	}
}

// reservationsLocked merges the registry with the active jobs.
func (m *Manager) reservationsLocked() (Reservations, error) {
	res, err := ReadRegistry(m.registryPath)
	if err != nil {
		return res, fmt.Errorf("read tenant registry: %w", err)
	}
	active := map[string]bool{}
	latest := map[string]*Job{} // latest finished real run per slug
	for _, e := range m.jobs {
		j := &e.job
		if !j.terminal() {
			active[j.Slug] = true
			res.Slugs[j.Slug] = true
			for _, p := range []*int{j.FEPort, j.APIPort} {
				if p != nil {
					res.Ports[*p] = true
				}
			}
			continue
		}
		if !j.DryRun && (latest[j.Slug] == nil || j.CreatedAt.After(latest[j.Slug].CreatedAt)) {
			latest[j.Slug] = j
		}
	}
	// Retry: a registered slug whose latest real job failed and that no
	// queued/running job holds.
	for slug, j := range latest {
		if j.Status == JobFailed && res.Slugs[slug] && !active[slug] {
			res.Retryable[slug] = j.ID
		}
	}
	return res, nil
}

// Create validates the request and queues a job.
func (m *Manager) Create(req CreateRequest) (Job, error) {
	req.Normalize()
	if verr := ValidateFormat(&req); verr != nil {
		return Job{}, verr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	taken, err := m.reservationsLocked()
	if err != nil {
		return Job{}, err
	}
	if verr := ValidateAvailability(&req, taken, m.portFree); verr != nil {
		return Job{}, verr
	}
	j := Job{
		ID: newID(), Status: JobQueued, Slug: req.Slug, ClientName: req.ClientName, Mode: req.Mode,
		Domain: req.Domain, FEPort: req.FEPort, APIPort: req.APIPort,
		AdminName: req.AdminName, AdminEmail: req.AdminEmail,
		WithDemoData: req.WithDemoData, ProvisionEmail: *req.ProvisionEmail, DryRun: req.DryRun,
		AllowNoTLS: req.AllowNoTLS, HealthTimeout: req.HealthTimeout,
		CreatedBy: req.CreatedBy, CreatedAt: m.now(), Steps: initialSteps(),
	}
	j.RetryOf = taken.Retryable[req.Slug]
	if j.Mode == "local" {
		j.Domain = ""
	}
	e := &entry{job: j, eventsLoaded: true, changed: make(chan struct{})}
	if err := m.store.SaveJob(&e.job); err != nil {
		return Job{}, fmt.Errorf("persist job: %w", err)
	}
	m.jobs[j.ID] = e
	m.queue = append(m.queue, j.ID)
	select {
	case m.wake <- struct{}{}:
	default:
	}
	m.log.Info("job queued", "job", j.ID, "slug", j.Slug, "retry_of", j.RetryOf, "mode", j.Mode, "dry_run", j.DryRun, "created_by", j.CreatedBy)
	return e.job.clone(), nil
}

// List returns up to limit jobs, newest first, without secrets.
func (m *Manager) List(limit int) []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Job, 0, len(m.jobs))
	for _, e := range m.jobs {
		out = append(out, e.job.clone())
	}
	sort.Slice(out, func(a, b int) bool {
		if !out[a].CreatedAt.Equal(out[b].CreatedAt) {
			return out[a].CreatedAt.After(out[b].CreatedAt)
		}
		return out[a].ID > out[b].ID
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// Get returns one job. When requester is the job's creator and the initial
// admin password has not been handed out yet, it is included this one time
// and then forgotten.
func (m *Manager) Get(id, requester string) (Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.jobs[id]
	if !ok {
		return Job{}, ErrNotFound
	}
	j := e.job.clone()
	if j.Result != nil && e.password != "" && !e.passwordRead && requester != "" && requester == e.job.CreatedBy {
		j.Result.AdminInitialPassword = ptr(e.password)
		e.passwordRead = true
		e.password = ""
		m.log.Info("initial admin password handed to the job creator", "job", id)
	}
	return j, nil
}

// EventsAfter returns the events with seq > since, whether the job has
// finished, and a channel that is closed on the next change.
func (m *Manager) EventsAfter(id string, since int64) ([]Event, bool, <-chan struct{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.jobs[id]
	if !ok {
		return nil, false, nil, ErrNotFound
	}
	if err := m.loadEventsLocked(e); err != nil {
		return nil, false, nil, err
	}
	i := sort.Search(len(e.events), func(i int) bool { return e.events[i].Seq > since })
	out := append([]Event(nil), e.events[i:]...)
	return out, e.job.terminal(), e.changed, nil
}

// Run is the single worker. It returns when ctx is cancelled.
func (m *Manager) Run(ctx context.Context) {
	for {
		m.mu.Lock()
		var id string
		if len(m.queue) > 0 {
			id, m.queue = m.queue[0], m.queue[1:]
		}
		m.mu.Unlock()
		if id == "" {
			select {
			case <-ctx.Done():
				return
			case <-m.wake:
				continue
			}
		}
		m.runJob(ctx, id)
		if ctx.Err() != nil {
			return
		}
	}
}

// secretPattern masks obvious secrets in log lines (KEY=value, "password": "x").
var secretPattern = regexp.MustCompile(`(?i)((?:password|passwd|secret|token|api[_-]?key)[A-Za-z0-9_]*["']?\s*[:=]\s*["']?)([^\s"',]+)`)

func (m *Manager) redact(e *entry, line string) string {
	if e.password != "" {
		line = strings.ReplaceAll(line, e.password, "[redacted]")
	}
	return secretPattern.ReplaceAllString(line, "${1}[redacted]")
}

func (m *Manager) runJob(ctx context.Context, id string) {
	m.mu.Lock()
	e, ok := m.jobs[id]
	if !ok || e.job.Status != JobQueued {
		m.mu.Unlock()
		return
	}
	e.job.Status = JobRunning
	e.job.StartedAt = ptr(m.now())
	m.saveLocked(e)
	m.emitLocked(e, EventStatus, StatusData{Status: JobRunning})
	args := ScriptArgs(&e.job)
	m.mu.Unlock()

	m.log.Info("job started", "job", id, "slug", e.job.Slug, "dry_run", e.job.DryRun)
	exitCode, startErr := m.runner.Run(ctx, args, func(stream, line string) { m.handleLine(e, stream, line) })

	m.mu.Lock()
	defer m.mu.Unlock()
	e.job.ExitCode = ptr(exitCode)
	switch {
	case startErr != nil:
		m.finishLocked(e, JobFailed, ptr("could not start onboard-tenant.sh: "+startErr.Error()))
	case ctx.Err() != nil:
		m.finishLocked(e, JobFailed, ptr("provisioner stopped while the job was running"))
	case exitCode != 0:
		msg := fmt.Sprintf("onboard-tenant.sh exited with code %d", exitCode)
		for _, s := range e.job.Steps {
			if s.Status == StepFailed && s.Message != "" {
				msg = s.Title + ": " + s.Message
				break
			}
		}
		m.finishLocked(e, JobFailed, ptr(msg))
	default:
		if e.job.Result == nil {
			m.emitLocked(e, EventLog, LogData{Line: "warning: onboard-tenant.sh finished without an @@RESULT line", Stream: "provisioner", TS: m.now().Format(time.RFC3339)})
		}
		m.finishLocked(e, JobSucceeded, nil)
	}
	m.log.Info("job finished", "job", id, "status", e.job.Status, "exit_code", exitCode)
}

// finishLocked moves a job to a terminal status, closes any step still
// running, emits the result (if any) and the final status event, and saves.
func (m *Manager) finishLocked(e *entry, status string, errMsg *string) {
	now := m.now()
	for i := range e.job.Steps {
		s := &e.job.Steps[i]
		if s.Status == StepRunning {
			if status == JobFailed {
				s.Status = StepFailed
				if errMsg != nil {
					s.Message = *errMsg
				}
			} else {
				s.Status = StepDone
			}
			s.FinishedAt = ptr(now)
			m.emitLocked(e, EventProgress, *s)
		}
	}
	if e.job.Result != nil {
		m.emitLocked(e, EventResult, e.job.clone().Result) // password excluded
	}
	e.job.Status = status
	e.job.Error = errMsg
	e.job.FinishedAt = ptr(now)
	m.saveLocked(e)
	m.emitLocked(e, EventStatus, StatusData{Status: status, Error: errMsg})
}

func (m *Manager) handleLine(e *entry, stream, line string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	if stream == "stdout" {
		p := ParseLine(line)
		switch p.Kind {
		case LineProgress:
			m.applyProgressLocked(e, p.Progress, now)
			return
		case LineResult:
			r := p.Result
			e.password = r.AdminInitialPassword
			e.passwordRead = false
			e.job.Result = &Result{Slug: r.Slug, FrontendURL: r.FrontendURL, APIURL: r.APIURL, AdminUsername: r.AdminUsername}
			if r.AdminEmail != "" {
				e.job.Result.AdminEmail = ptr(r.AdminEmail)
			}
			m.applyTLSLocked(e)
			m.saveLocked(e)
			return
		case LineMalformed:
			line = p.Text
		}
	}
	m.emitLocked(e, EventLog, LogData{Line: m.redact(e, line), Stream: stream, TS: now.Format(time.RFC3339Nano)})
}

func (m *Manager) applyProgressLocked(e *entry, p *Progress, now time.Time) {
	s := &e.job.Steps[p.Index-1]
	s.Status = p.Status
	s.Message = m.redact(e, p.Message)
	if s.StartedAt == nil {
		s.StartedAt = ptr(now)
	}
	if p.Status == StepRunning {
		s.FinishedAt = nil
		e.job.CurrentStep = s.Key
	} else {
		s.FinishedAt = ptr(now)
		if p.Status == StepFailed {
			e.job.CurrentStep = s.Key
		}
	}
	m.saveLocked(e)
	m.emitLocked(e, EventProgress, *s)
}

// applyTLSLocked sets Result.TLS. When step 9 (ingress) was skipped, nginx
// serves the tenant on plain HTTP only (local mode, or production under
// --allow-no-tls), so any https:// URL is rewritten to http://. The script
// already reports http in that case; this only guards the contract.
func (m *Manager) applyTLSLocked(e *entry) {
	r := e.job.Result
	skipped := false
	for _, s := range e.job.Steps {
		if s.Key == "ingress" && s.Status == StepSkipped {
			skipped = true
		}
	}
	if skipped {
		r.FrontendURL = toHTTP(r.FrontendURL)
		r.APIURL = toHTTP(r.APIURL)
	}
	r.TLS = strings.HasPrefix(r.FrontendURL, "https://") && strings.HasPrefix(r.APIURL, "https://")
}

func toHTTP(u string) string {
	if strings.HasPrefix(u, "https://") {
		return "http://" + strings.TrimPrefix(u, "https://")
	}
	return u
}
