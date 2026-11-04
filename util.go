package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	home, _  = os.UserHomeDir()
	dataDir  = filepath.Join(home, ".local", "share", "romanctl")
	logDir   = filepath.Join(dataDir, "logs")
	projDir  = envOr("ROMANCTL_PROJETOS", filepath.Join(home, "projetos"))
	sGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	sRed     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	sYellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	sGray    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	sBlue    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	sTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	sBoldTxt = lipgloss.NewStyle().Bold(true)
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func okf(f string, a ...any)   { fmt.Println(sGreen.Render("✔ ") + fmt.Sprintf(f, a...)) }
func failf(f string, a ...any) { fmt.Println(sRed.Render("✘ ") + fmt.Sprintf(f, a...)) }
func infof(f string, a ...any) { fmt.Println(sBlue.Render("• ") + fmt.Sprintf(f, a...)) }
func warnf(f string, a ...any) { fmt.Println(sYellow.Render("! ") + fmt.Sprintf(f, a...)) }

// run executa um comando com stdin/stdout/stderr ligados ao terminal.
func run(dir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), env...)
	return cmd.Run()
}

// output executa um comando e devolve a saída (stdout+stderr) sem espaços nas pontas.
func output(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// outputEnv é como output, com variáveis de ambiente extras.
func outputEnv(env []string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// outputIn é como output, mas executa dentro do diretório dir.
func outputIn(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func have(bin string) bool { _, err := exec.LookPath(bin); return err == nil }

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

// pgrep: o padrão deve usar o truque "[x]" para não casar com o próprio comando.
func pgrep(pattern string) bool { return exec.Command("pgrep", "-f", pattern).Run() == nil }

func waitUntil(timeout time.Duration, f func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if f() {
			return true
		}
		time.Sleep(300 * time.Millisecond)
	}
	return f()
}

func httpUp(url string) bool {
	c := http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := c.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// startDetached inicia um processo em nova sessão, com saída num arquivo de log,
// para que ele continue rodando depois que o romanctl terminar.
func startDetached(logFile, dir, name string, args ...string) error {
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

func dirSize(p string) int64 {
	var total int64
	filepath.Walk(p, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func human(n int64) string {
	const u = 1024
	if n < u {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(u), 0
	for m := n / u; m >= u; m /= u {
		div *= u
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
