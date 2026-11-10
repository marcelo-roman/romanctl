package services

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"romanctl/internal/platform"
)

// fakeEngine põe no PATH um "docker" que registra os comandos e simula containers e imagens.
func fakeEngine(t *testing.T) (logPath string) {
	t.Helper()
	forgetEngine()
	t.Cleanup(forgetEngine)
	dir := t.TempDir()
	logPath = filepath.Join(dir, "calls")
	state := filepath.Join(dir, "containers")
	images := filepath.Join(dir, "images")
	script := "#!/bin/sh\necho \"$*\" >> " + logPath + "\ncase \"$1\" in\n" +
		"ps) [ -f " + state + " ] && cat " + state + "; exit 0;;\n" +
		"image) [ -f " + images + " ] && cat " + images + "; exit 0;;\n" +
		"pull) echo \"$2\" >> " + images + ";;\n" +
		"create) printf 'romanctl-x\\tcreated\\n' > " + state + ";;\n" +
		"run) printf 'romanctl-x\\trunning\\n' > " + state + ";;\n" +
		"start) printf 'romanctl-x\\trunning\\n' > " + state + ";;\n" +
		"stop) printf 'romanctl-x\\texited\\n' > " + state + ";;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(dir, "docker"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return logPath
}

func TestTerDockerNaoEhTerOServico(t *testing.T) {
	fakeEngine(t)
	saved := platform.Current
	defer func() { platform.Current = saved }()
	platform.Current = platform.Info{OS: "linux", Arch: "amd64"}
	s := containerService(containerSpec{Key: "x", Name: "X", Image: "img:1", Ports: []string{"1:1"}})
	if s.Installed() {
		t.Fatal("sem imagem nem container, o serviço não está instalado")
	}
	if err := s.install(); err != nil {
		t.Fatal(err)
	}
	if !s.Installed() {
		t.Fatal("instalar cria o container, e com ele o serviço está instalado")
	}
	if s.Running() {
		t.Fatal("instalar não liga o serviço")
	}
}

func TestImagemDeOutroProjetoNaoInstala(t *testing.T) {
	logPath := fakeEngine(t)
	saved := platform.Current
	defer func() { platform.Current = saved }()
	platform.Current = platform.Info{OS: "linux", Arch: "amd64"}
	os.WriteFile(filepath.Join(filepath.Dir(logPath), "images"), []byte("img:1\n"), 0o644)
	s := containerService(containerSpec{Key: "x", Name: "X", Image: "img:1", Ports: []string{"1:1"}})
	if s.Installed() {
		t.Fatal("a imagem baixada por outro projeto não faz o serviço instalado")
	}
	if err := s.install(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(logPath); strings.Contains(string(b), "pull") {
		t.Fatal("a imagem já estava aqui: não precisava baixar de novo")
	}
}

func TestCreateUsaOsMesmosArgumentosDoRun(t *testing.T) {
	c := containerSpec{Key: "x", Image: "img:1", Ports: []string{"1:1"}, Volume: "/d"}
	if got := strings.Join(createArgs(c), " "); got != "create --name romanctl-x -p 127.0.0.1:1:1 -v romanctl-x:/d img:1" {
		t.Fatal(got)
	}
}

func TestImageRefCompletaATag(t *testing.T) {
	for in, want := range map[string]string{"redis": "redis:latest", "redis:8": "redis:8",
		"localhost:5000/app": "localhost:5000/app:latest", "mcr.microsoft.com/mssql/server:2022-latest": "mcr.microsoft.com/mssql/server:2022-latest"} {
		if got := imageRef(in); got != want {
			t.Errorf("%s: %s", in, got)
		}
	}
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
		if verb != "ps" && verb != "image" {
			verbs = append(verbs, verb)
		}
	}
	if !slices.Equal(verbs, []string{"run", "stop", "start"}) {
		t.Fatalf("a segunda subida deveria reaproveitar o container: %v", verbs)
	}
}
