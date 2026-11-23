package sys

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// processIs checks that the PID is alive and is the expected program. The PID alone is not
// enough: after a crash, Android may reuse it for another process.
func ProcessIs(pid int, name string) bool {
	if pid <= 0 || syscall.Kill(pid, 0) != nil {
		return false
	}
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil {
		// Without /proc, as on macOS, ask ps.
		out, psErr := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
		return psErr == nil && strings.Contains(string(out), name)
	}
	return strings.Contains(string(b), name)
}

// pgrep: the pattern should use the "[x]" trick so it does not match the command itself.
func Pgrep(pattern string) bool { return exec.Command("pgrep", "-f", pattern).Run() == nil }

func WaitUntil(timeout time.Duration, f func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if f() {
			return true
		}
		time.Sleep(300 * time.Millisecond)
	}
	return f()
}

// startDetached starts a process in a new session, with output to a log file,
// so it keeps running after romanctl exits.
func StartDetached(logFile, dir, name string, args ...string) error {
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return err
	}
	lf, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer lf.Close()
	devnull, _ := os.Open(os.DevNull)
	defer devnull.Close()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = devnull, lf, lf
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// StartDetachedEnv is like StartDetached, with extra variables, and returns the PID. The process
// leads its own group, so stopping the group also catches its children, like dotnet run's app.
func StartDetachedEnv(logFile, dir string, env []string, name string, args ...string) (int, error) {
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return 0, err
	}
	lf, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, err
	}
	defer lf.Close()
	devnull, _ := os.Open(os.DevNull)
	defer devnull.Close()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = devnull, lf, lf
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	pid := cmd.Process.Pid
	// Reaps the process when it exits: otherwise it would become a zombie and look alive.
	go cmd.Wait()
	return pid, nil
}

// GroupAlive says whether any process is still in the group.
func GroupAlive(pgid int) bool {
	return pgid > 0 && syscall.Kill(-pgid, 0) == nil
}

// StopGroup asks the group to exit and forces it after the timeout.
func StopGroup(pgid int, grace time.Duration) {
	if !GroupAlive(pgid) {
		return
	}
	syscall.Kill(-pgid, syscall.SIGTERM)
	if WaitUntil(grace, func() bool { return !GroupAlive(pgid) }) {
		return
	}
	syscall.Kill(-pgid, syscall.SIGKILL)
	WaitUntil(2*time.Second, func() bool { return !GroupAlive(pgid) })
}
