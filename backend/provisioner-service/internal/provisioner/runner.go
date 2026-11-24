package provisioner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// maxLineBytes caps one output line; longer lines are split.
const maxLineBytes = 256 * 1024

// Runner runs onboard-tenant.sh for a job.
type Runner interface {
	// Run starts the script with args and calls onLine for every output line
	// (stream is "stdout" or "stderr"). onLine may be called from two
	// goroutines. It returns the exit code; startErr is set only when the
	// process could not be started.
	Run(ctx context.Context, args []string, onLine func(stream, line string)) (exitCode int, startErr error)
}

// ExecRunner runs Command (program + fixed leading args, e.g. ["bash",
// "<repo>/scripts/onboard-tenant.sh"]) followed by the job's args. It uses
// exec directly with an argv array: no shell ever parses the arguments.
type ExecRunner struct {
	Command []string
	Dir     string
	Env     []string // appended to the filtered environment (see ScriptEnv)
}

func (r *ExecRunner) Run(ctx context.Context, args []string, onLine func(stream, line string)) (int, error) {
	if len(r.Command) == 0 {
		return -1, errors.New("no command configured")
	}
	argv := append(append([]string(nil), r.Command[1:]...), args...)
	cmd := exec.CommandContext(ctx, r.Command[0], argv...)
	cmd.Dir = r.Dir
	cmd.Env = append(ScriptEnv(os.Environ()), r.Env...)
	cmd.Stdin = nil
	configureProcessGroup(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return -1, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return -1, err
	}
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); scanLines(stdout, "stdout", onLine) }()
	go func() { defer wg.Done(); scanLines(stderr, "stderr", onLine) }()
	wg.Wait()
	err = cmd.Wait()
	if err == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if code := ee.ExitCode(); code >= 0 {
			return code, nil
		}
		return 128, nil // killed by a signal
	}
	return 1, nil
}

func scanLines(rd io.Reader, stream string, onLine func(stream, line string)) {
	br := bufio.NewReaderSize(rd, 64*1024)
	var buf []byte
	for {
		chunk, isPrefix, err := br.ReadLine()
		buf = append(buf, chunk...)
		if err != nil {
			if len(buf) > 0 {
				onLine(stream, string(buf))
			}
			return
		}
		if isPrefix && len(buf) < maxLineBytes {
			continue
		}
		// A line longer than maxLineBytes is emitted in pieces.
		onLine(stream, string(buf))
		buf = buf[:0]
	}
}

// PortsFunc returns the next free frontend and API ports.
type PortsFunc func(ctx context.Context) (NextPorts, error)

// NextPorts is the body of GET /ports/next.
type NextPorts struct {
	FEPort  int `json:"fe_port"`
	APIPort int `json:"api_port"`
}

// ExecNextPorts runs `generate-tenant-stack.sh --next-ports` (Command holds
// the program and the script path) and parses the JSON it prints. The last
// stdout line that looks like a JSON object is used, so stray log lines
// before it do not matter.
func ExecNextPorts(command []string, dir string) PortsFunc {
	return func(ctx context.Context) (NextPorts, error) {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		argv := append(append([]string(nil), command[1:]...), "--next-ports")
		cmd := exec.CommandContext(ctx, command[0], argv...)
		cmd.Dir = dir
		var out, errOut bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errOut
		if err := cmd.Run(); err != nil {
			return NextPorts{}, fmt.Errorf("generate-tenant-stack.sh --next-ports: %v: %s", err, lastLine(errOut.String()))
		}
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			l := strings.TrimSpace(lines[i])
			if !strings.HasPrefix(l, "{") {
				continue
			}
			var np NextPorts
			if err := json.Unmarshal([]byte(l), &np); err != nil {
				return NextPorts{}, fmt.Errorf("parse --next-ports output: %w", err)
			}
			if np.FEPort <= 0 || np.APIPort <= 0 {
				return NextPorts{}, errors.New("--next-ports output is missing fe_port or api_port")
			}
			return np, nil
		}
		return NextPorts{}, errors.New("--next-ports printed no JSON object")
	}
}

func lastLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndex(s, "\n"); i >= 0 {
		s = s[i+1:]
	}
	if len(s) > 500 {
		s = s[:500]
	}
	return s
}

// scriptEnvNames are the variables passed through to onboard-tenant.sh (and
// so to docker compose, certbot and curl). Everything else in the
// provisioner's environment, PROVISIONER_TOKEN above all, is dropped.
//
// Besides the basics, this is what the scripts read from the environment:
// generate-tenant-stack.sh (BASE_DOMAIN, TENANT_STORAGE_ROOT, IMAGE_TAG),
// onboard-tenant.sh (DRY_RUN_DELAY, BUILDKIT_PROGRESS) and
// provision-tenant-email.sh (AUDITSPHERE_EMAIL_ENV, EMAIL_SENDING_DOMAIN,
// EMAIL_FROM_LOCALPART, RESEND_REGION, RESEND_CURL_*). The Resend master key
// is read from /etc/auditsphere/email.env, not from the environment.
var scriptEnvNames = map[string]bool{
	"PATH": true, "HOME": true, "USER": true, "LOGNAME": true, "SHELL": true,
	"LANG": true, "LANGUAGE": true, "TZ": true, "TERM": true, "TMPDIR": true,
	"XDG_RUNTIME_DIR": true, // rootless docker socket
	"SSL_CERT_FILE":   true, "SSL_CERT_DIR": true,
	"HTTP_PROXY": true, "HTTPS_PROXY": true, "NO_PROXY": true,
	"BASE_DOMAIN": true, "TENANT_STORAGE_ROOT": true, "IMAGE_TAG": true,
	"DRY_RUN_DELAY":         true,
	"AUDITSPHERE_EMAIL_ENV": true, "EMAIL_SENDING_DOMAIN": true, "EMAIL_FROM_LOCALPART": true,
	"RESEND_REGION": true, "RESEND_CURL_CONNECT_TIMEOUT": true, "RESEND_CURL_MAX_TIME": true,
}

// scriptEnvPrefixes pass through whole families of variables.
var scriptEnvPrefixes = []string{"LC_", "DOCKER_", "COMPOSE_", "BUILDKIT_"}

// windowsEnvNames are needed for Git Bash and its tools to start on Windows
// (local dry runs only).
var windowsEnvNames = map[string]bool{
	"SYSTEMROOT": true, "SYSTEMDRIVE": true, "WINDIR": true, "COMSPEC": true, "PATHEXT": true,
	"TEMP": true, "TMP": true, "USERPROFILE": true, "HOMEDRIVE": true, "HOMEPATH": true,
	"APPDATA": true, "LOCALAPPDATA": true, "PROGRAMDATA": true, "PROGRAMFILES": true,
	"MSYSTEM": true, "NUMBER_OF_PROCESSORS": true, "PROCESSOR_ARCHITECTURE": true,
}

// ScriptEnv returns the allow-listed subset of environ (KEY=value entries).
func ScriptEnv(environ []string) []string {
	return filterEnv(environ, runtime.GOOS == "windows")
}

func filterEnv(environ []string, windows bool) []string {
	out := make([]string, 0, len(environ))
	for _, kv := range environ {
		name, _, ok := strings.Cut(kv, "=")
		if !ok || name == "" {
			continue
		}
		key := name
		if windows {
			key = strings.ToUpper(name) // Windows names are case-insensitive
		} else if upper := strings.ToUpper(name); upper != name && strings.HasSuffix(upper, "_PROXY") && scriptEnvNames[upper] {
			key = upper // curl also honours http_proxy / https_proxy / no_proxy
		}
		if envAllowed(key, windows) {
			out = append(out, kv)
		}
	}
	return out
}

func envAllowed(key string, windows bool) bool {
	if scriptEnvNames[key] || (windows && windowsEnvNames[key]) {
		return true
	}
	for _, p := range scriptEnvPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}
