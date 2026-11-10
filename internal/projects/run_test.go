package projects

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"romanctl/internal/manifest"
)

type syncBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func TestSupervisorPoeONomeEUsaPastaEAmbienteDeCadaProcesso(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Join(dir, "web app"), 0o755)
	var out syncBuffer
	code := Supervise(dir, []manifest.Process{
		{Name: "api", Run: "echo porta=$API_PORT; sleep 30", Env: map[string]string{"API_PORT": "3302"}},
		{Name: "web", Dir: "web app", Run: "sleep 0.5; pwd"},
	}, &out)
	text := out.String()
	if !strings.Contains(text, "[api] porta=3302") || !strings.Contains(text, "[web] "+filepath.Join(dir, "web app")) {
		t.Fatalf("saída: %q", text)
	}
	if code != 1 {
		t.Fatalf("um processo que termina sozinho derruba o projeto: código %d", code)
	}
}

func TestQuedaDeUmProcessoParaOsOutrosNaHora(t *testing.T) {
	var out syncBuffer
	start := time.Now()
	code := Supervise(t.TempDir(), []manifest.Process{
		{Name: "api", Run: "echo falhou; exit 7"},
		{Name: "web", Run: "sleep 30"},
	}, &out)
	if code != 7 {
		t.Fatalf("código %d, saída %q", code, out.String())
	}
	if time.Since(start) > 4*time.Second {
		t.Fatalf("demorou %s para parar o web", time.Since(start))
	}
	if !strings.Contains(out.String(), "[api] falhou") {
		t.Fatalf("saída: %q", out.String())
	}
}

func TestComandoDoManifestoPassaPeloShell(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, manifest.FileName), []byte("name: x\nrun: migrate && serve\n"), 0o644)
	cmd, err := RunCommand(dir, "x")
	if err != nil || len(cmd) != 3 || cmd[2] != "migrate && serve" {
		t.Fatalf("%v %v", cmd, err)
	}
}

func TestProjetoComProcessosRodaPeloSupervisor(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, manifest.FileName), []byte("name: x\nprocesses:\n  - name: a\n    run: true\n"), 0o644)
	cmd, err := RunCommand(dir, "x")
	if err != nil || len(cmd) != 2 || cmd[1] != SuperviseArg {
		t.Fatalf("%v %v", cmd, err)
	}
}
