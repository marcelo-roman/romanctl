package services

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"romanctl/internal/platform"
)

// fakeEngine põe no PATH um "docker" que registra os comandos e simula o estado.
func fakeEngine(t *testing.T) (logPath string) {
	t.Helper()
	dir := t.TempDir()
	logPath = filepath.Join(dir, "calls")
	stateFile := filepath.Join(dir, "state")
	script := "#!/bin/sh\necho \"$*\" >> " + logPath + "\ncase \"$1\" in\n" +
		"inspect) [ -f " + stateFile + " ] || exit 1; cat " + stateFile + ";;\n" +
		"run|start) echo true > " + stateFile + ";;\n" +
		"stop) echo false > " + stateFile + ";;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(dir, "docker"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return logPath
}

func TestRunArgsPublicaSoNoLocalhost(t *testing.T) {
	got := strings.Join(runArgs(containerSpec{Key: "x", Image: "img:1", Ports: []string{"5672:5672", "15672:15672"},
		Env: []string{"A=1"}, Volume: "/data", Args: []string{"serve"}}), " ")
	want := "run -d --name romanctl-x -p 127.0.0.1:5672:5672 -p 127.0.0.1:15672:15672 -e A=1 -v romanctl-x:/data img:1 serve"
	if got != want {
		t.Fatalf("obtido %s", got)
	}
}

func TestProotNaoOferecePorContainer(t *testing.T) {
	saved := platform.Current
	defer func() { platform.Current = saved }()
	platform.Current = platform.Info{OS: "linux", Arch: "arm64", Proot: true}
	s := containerService(redisContainer)
	if s.Unsupported == "" || s.Installed() {
		t.Fatal("deveria ser indisponível no proot")
	}
}

func TestArquiteturaSemImagem(t *testing.T) {
	saved := platform.Current
	defer func() { platform.Current = saved }()
	platform.Current = platform.Info{OS: "linux", Arch: "arm64"}
	s := containerService(containerSpec{Key: "x", Image: "i", Ports: []string{"1:1"}, Arches: []string{"amd64"}})
	if !strings.Contains(s.Unsupported, "arm64") {
		t.Fatalf("mensagem: %q", s.Unsupported)
	}
}

func TestCicloDeVidaComMotorSimulado(t *testing.T) {
	logPath := fakeEngine(t)
	saved := platform.Current
	defer func() { platform.Current = saved }()
	platform.Current = platform.Info{OS: "linux", Arch: "amd64"}

	// Porta local aberta para a espera de prontidão terminar logo.
	ln := listenLocal(t)
	defer ln.Close()
	port := ln.Addr().String()[strings.LastIndex(ln.Addr().String(), ":")+1:]
	s := containerService(containerSpec{Key: "x", Name: "X", Image: "i", Ports: []string{port + ":1"}})

	if s.Running() {
		t.Fatal("não deveria estar rodando")
	}
	if err := s.start(); err != nil {
		t.Fatal(err)
	}
	if !s.Running() {
		t.Fatal("deveria estar rodando")
	}
	if err := s.stop(); err != nil {
		t.Fatal(err)
	}
	if err := s.start(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(logPath)
	calls := strings.Split(strings.TrimSpace(string(b)), "\n")
	var verbs []string
	for _, c := range calls {
		verb := strings.Fields(c)[0]
		if verb != "inspect" {
			verbs = append(verbs, verb)
		}
	}
	if !slices.Equal(verbs, []string{"run", "stop", "start"}) {
		t.Fatalf("a segunda subida deveria reaproveitar o container: %v", verbs)
	}
}
