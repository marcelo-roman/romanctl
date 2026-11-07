package sys

import (
	"os"
	"os/exec"
	"path/filepath"
	"romanctl/internal/i18n"
	"strings"
)

// quiet esconde a saída do comando e só a mostra em caso de erro.
func Quiet(dir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return i18n.Errorf("%v\n%s", err, Tail(string(out), 12))
	}
	return nil
}

func Tail(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// run executa um comando com stdin/stdout/stderr ligados ao terminal.
func Run(dir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), env...)
	return cmd.Run()
}

// output executa um comando e devolve a saída (stdout+stderr) sem espaços nas pontas.
func Output(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// outputEnv é como output, com variáveis de ambiente extras.
func OutputEnv(env []string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// outputIn é como output, mas executa dentro do diretório dir.
func OutputIn(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func Have(bin string) bool { _, err := exec.LookPath(bin); return err == nil }

func Exists(p string) bool { _, err := os.Stat(p); return err == nil }

func DirSize(p string) int64 {
	var total int64
	filepath.Walk(p, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func Human(n int64) string {
	const u = 1024
	if n < u {
		return i18n.Sprintf("%d B", n)
	}
	div, exp := int64(u), 0
	for m := n / u; m >= u; m /= u {
		div *= u
		exp++
	}
	return i18n.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
