//go:build unix

package provisioner

import (
	"os/exec"
	"syscall"
	"time"
)

// killGrace is how long the script's process group gets after SIGTERM
// before it is SIGKILLed.
const killGrace = 10 * time.Second

// configureProcessGroup starts the script in its own process group and makes
// cancellation signal the whole group. bash defers a trapped SIGTERM until its
// foreground child (e.g. `docker compose build`) exits, so signalling only the
// bash PID would leave the job running.
func configureProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		pgid := -cmd.Process.Pid
		err := syscall.Kill(pgid, syscall.SIGTERM)
		go func() {
			time.Sleep(killGrace)
			_ = syscall.Kill(pgid, syscall.SIGKILL)
		}()
		return err
	}
	cmd.WaitDelay = killGrace + 5*time.Second
}
