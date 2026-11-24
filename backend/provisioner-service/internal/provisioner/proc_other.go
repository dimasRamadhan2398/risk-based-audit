//go:build !unix

package provisioner

import (
	"os/exec"
	"time"
)

// configureProcessGroup: no process groups here (Windows is only used for
// local dry runs); cancellation kills the direct child.
func configureProcessGroup(cmd *exec.Cmd) {
	cmd.WaitDelay = 10 * time.Second
}
