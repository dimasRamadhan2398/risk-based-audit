package provisioner

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Store persists jobs and their event logs as files:
//
//	<data>/jobs/<id>/job.json      the Job (never contains the admin password)
//	<data>/jobs/<id>/events.jsonl  one Event per line, append-only
type Store struct {
	dir string
}

var idPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// NewStore creates the data directory (mode 0700) if needed.
func NewStore(dataDir string) (*Store, error) {
	dir := filepath.Join(dataDir, "jobs")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

func (s *Store) jobDir(id string) string { return filepath.Join(s.dir, id) }

// SaveJob writes job.json atomically (temp file + rename).
func (s *Store) SaveJob(j *Job) error {
	if !idPattern.MatchString(j.ID) {
		return fmt.Errorf("invalid job id %q", j.ID)
	}
	c := j.clone() // drops any password
	b, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return err
	}
	dir := s.jobDir(j.ID)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "job-*.tmp")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(dir, "job.json"))
}

// AppendEvent appends one event line.
func (s *Store) AppendEvent(id string, ev Event) error {
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(s.jobDir(id), "events.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	_, werr := f.Write(append(b, '\n'))
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	return cerr
}

// LoadJobs reads every job.json. Unreadable entries are returned in errs and skipped.
func (s *Store) LoadJobs() (jobs []*Job, errs []error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, []error{err}
	}
	for _, e := range entries {
		if !e.IsDir() || !idPattern.MatchString(e.Name()) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.dir, e.Name(), "job.json"))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		var j Job
		if err := json.Unmarshal(b, &j); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", e.Name(), err))
			continue
		}
		jobs = append(jobs, &j)
	}
	return jobs, errs
}

// LoadEvents reads a job's event log. A torn last line (crash mid-write) is ignored.
func (s *Store) LoadEvents(id string) ([]Event, error) {
	f, err := os.Open(filepath.Join(s.jobDir(id), "events.jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), maxLineBytes*2)
	for sc.Scan() {
		var ev Event
		if json.Unmarshal(sc.Bytes(), &ev) == nil {
			out = append(out, ev)
		}
	}
	return out, sc.Err()
}
