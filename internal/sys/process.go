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

// processIs confere se o PID está vivo e é o programa esperado. O PID sozinho não
// basta: depois de uma queda, o Android pode reusá-lo em outro processo.
func ProcessIs(pid int, name string) bool {
	if pid <= 0 || syscall.Kill(pid, 0) != nil {
		return false
	}
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil {
		// Sem /proc, como no macOS, pergunta ao ps.
		out, psErr := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
		return psErr == nil && strings.Contains(string(out), name)
	}
	return strings.Contains(string(b), name)
}

// pgrep: o padrão deve usar o truque "[x]" para não casar com o próprio comando.
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

// startDetached inicia um processo em nova sessão, com saída num arquivo de log,
// para que ele continue rodando depois que o romanctl terminar.
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

// StartDetachedEnv é como StartDetached, com variáveis extras, e devolve o PID. O processo lidera
// o próprio grupo, para que parar o grupo pegue também os filhos, como o app do dotnet run.
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
	// Colhe o processo quando ele terminar: sem isso, ele ficaria zumbi e pareceria vivo.
	go cmd.Wait()
	return pid, nil
}

// GroupAlive diz se ainda há algum processo no grupo.
func GroupAlive(pgid int) bool {
	return pgid > 0 && syscall.Kill(-pgid, 0) == nil
}

// StopGroup pede para o grupo terminar e força depois do prazo.
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
