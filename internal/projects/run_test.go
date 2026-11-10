package projects

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"romanctl/internal/manifest"
	"romanctl/internal/sys"
)

func TestSupervisorSobeTodosComONomeNaFrente(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "web app"), 0o755)
	script := supervisor([]manifest.Process{
		{Name: "api", Run: "echo pronta"},
		{Name: "web", Dir: "web app", Run: "pwd"},
	})
	cmd := exec.Command("sh", "-c", script)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	text := string(out)
	if !strings.Contains(text, "[api] pronta") || !strings.Contains(text, "[web] "+filepath.Join(dir, "web app")) {
		t.Fatalf("saída: %q", text)
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

func TestPararOGrupoParaTodosOsProcessos(t *testing.T) {
	dir := t.TempDir()
	script := supervisor([]manifest.Process{{Name: "a", Run: "sleep 30"}, {Name: "b", Run: "sleep 30"}})
	pid, err := sys.StartDetachedEnv(filepath.Join(dir, "log"), dir, nil, "sh", "-c", script)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	if !sys.GroupAlive(pid) {
		t.Fatal("o grupo deveria estar vivo")
	}
	sys.StopGroup(pid, 3*time.Second)
	if !sys.WaitUntil(2*time.Second, func() bool { return !sys.GroupAlive(pid) }) {
		t.Fatal("sobrou processo do grupo")
	}
}
