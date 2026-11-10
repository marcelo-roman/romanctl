package publish

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"romanctl/internal/scaffold"
)

func TestSensitive(t *testing.T) {
	got := sensitive([]string{".env", "api/.env.production", ".env.example", "certs/server.key", "app.pem", "id_rsa", "main.go", "keys.go", "docs/.env.sample"})
	want := []string{".env", "api/.env.production", "certs/server.key", "app.pem", "id_rsa"}
	if !slices.Equal(got, want) {
		t.Fatalf("obtido %v", got)
	}
}

func TestTodoTipoGeraArquivosValidos(t *testing.T) {
	data := fileData{"Name": "Loja", "Lower": "loja", "API": "src/Loja.Api", "Dll": "Loja.Api", "Main": ".", "Module": "loja", "Bin": "loja", "Dist": "loja"}
	for _, kind := range Kinds() {
		files, err := render(kind, data)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		for _, want := range []string{"Dockerfile", ".dockerignore", ".github/workflows/ci.yml"} {
			if _, ok := files[want]; !ok {
				t.Errorf("%s: falta %s", kind, want)
			}
		}
		var workflow map[string]any
		if err := yaml.Unmarshal(files[".github/workflows/ci.yml"], &workflow); err != nil || workflow["jobs"] == nil {
			t.Errorf("%s: workflow inválido: %v", kind, err)
		}
	}
	files, _ := render("dotnet-solution", data)
	if !strings.Contains(string(files["Dockerfile"]), `dotnet publish src/Loja.Api`) || !strings.Contains(string(files["Dockerfile"]), `"Loja.Api.dll"`) {
		t.Fatalf("Dockerfile: %s", files["Dockerfile"])
	}
}

// fakeGH simula o GitHub com um repositório local. O git continua sendo o de verdade.
func fakeGH(t *testing.T, signedIn bool) (remote string, calls string) {
	t.Helper()
	bin := t.TempDir()
	remote = filepath.Join(t.TempDir(), "remoto.git")
	calls = filepath.Join(bin, "calls")
	auth := "exit 0"
	if !signedIn {
		auth = "echo 'not logged in'; exit 1"
	}
	script := "#!/bin/sh\necho \"$*\" >> " + calls + "\ncase \"$1 $2\" in\n" +
		"'auth status') " + auth + ";;\n" +
		"'repo create') git init -q --bare " + remote + " && git remote add origin " + remote + " && git push -q -u origin HEAD 2>&1 && echo https://github.example/u/x;;\n" +
		"'repo view') echo https://github.example/u/x;;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	os.WriteFile(filepath.Join(home, "gitconfig"), []byte("[user]\n\tname = Teste\n\temail = t@e.st\n[init]\n\tdefaultBranch = main\n"), 0o644)
	return remote, calls
}

func goProject(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "loja")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module loja\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644)
	return dir
}

func remoteFiles(t *testing.T, remote string) []string {
	t.Helper()
	out, err := exec.Command("git", "-C", remote, "ls-tree", "-r", "--name-only", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	return strings.Fields(string(out))
}

func TestPublicaEmRepositorioPrivadoPorPadrao(t *testing.T) {
	remote, calls := fakeGH(t, true)
	dir := goProject(t)
	result, err := Run(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || !result.Committed || result.URL != "https://github.example/u/x" {
		t.Fatalf("%+v", result)
	}
	got := remoteFiles(t, remote)
	for _, want := range []string{"main.go", "Dockerfile", ".dockerignore", ".github/workflows/ci.yml"} {
		if !slices.Contains(got, want) {
			t.Errorf("faltou %s no remoto: %v", want, got)
		}
	}
	log, _ := os.ReadFile(calls)
	if !strings.Contains(string(log), "--private") || strings.Contains(string(log), "--public") {
		t.Fatalf("visibilidade errada: %s", log)
	}
}

func TestPublicoSoQuandoPedido(t *testing.T) {
	_, calls := fakeGH(t, true)
	if _, err := Run(goProject(t), Options{Public: true, Description: "Uma loja"}); err != nil {
		t.Fatal(err)
	}
	log, _ := os.ReadFile(calls)
	if !strings.Contains(string(log), "--public") || !strings.Contains(string(log), "--description Uma loja") {
		t.Fatalf("%s", log)
	}
}

func TestSegundaPublicacaoSoEnvia(t *testing.T) {
	remote, calls := fakeGH(t, true)
	dir := goProject(t)
	if _, err := Run(dir, Options{}); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "novo.go"), []byte("package main\n"), 0o644)
	os.WriteFile(calls, nil, 0o644)
	result, err := Run(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Exists || !result.Committed {
		t.Fatalf("%+v", result)
	}
	log, _ := os.ReadFile(calls)
	if strings.Contains(string(log), "repo create") {
		t.Fatalf("não deveria criar outro repositório: %s", log)
	}
	if !slices.Contains(remoteFiles(t, remote), "novo.go") {
		t.Fatal("novo.go não chegou ao remoto")
	}
}

func TestSegredoCancelaAntesDoCommit(t *testing.T) {
	_, calls := fakeGH(t, true)
	dir := goProject(t)
	os.WriteFile(filepath.Join(dir, ".env"), []byte("SENHA=123\n"), 0o644)
	_, err := Run(dir, Options{})
	if err == nil || !strings.Contains(err.Error(), ".env") {
		t.Fatalf("erro: %v", err)
	}
	log, _ := os.ReadFile(calls)
	if strings.Contains(string(log), "repo create") {
		t.Fatal("não deveria ter criado o repositório")
	}
	if out, _ := exec.Command("git", "-C", dir, "log", "--oneline").CombinedOutput(); strings.Contains(string(out), "Publica") {
		t.Fatalf("não deveria ter feito commit: %s", out)
	}
	// Com o arquivo no .gitignore, a publicação segue.
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".env\n"), 0o644)
	if _, err := Run(dir, Options{}); err != nil {
		t.Fatal(err)
	}
}

func TestNaoSobrescreveOQueExiste(t *testing.T) {
	remote, _ := fakeGH(t, true)
	dir := goProject(t)
	mine := "FROM meu-proprio\n"
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(mine), 0o644)
	result, err := Run(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(result.Generated, "Dockerfile") {
		t.Fatalf("%+v", result)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if string(b) != mine {
		t.Fatalf("sobrescreveu: %s", b)
	}
	_ = remote
}

func TestSemCI(t *testing.T) {
	remote, _ := fakeGH(t, true)
	if _, err := Run(goProject(t), Options{NoCI: true}); err != nil {
		t.Fatal(err)
	}
	if got := remoteFiles(t, remote); slices.Contains(got, "Dockerfile") {
		t.Fatalf("%v", got)
	}
}

func TestSimulacaoNaoMudaNada(t *testing.T) {
	_, calls := fakeGH(t, true)
	dir := goProject(t)
	result, err := Plan(dir, Options{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Generated) != 3 || !result.OK {
		t.Fatalf("%+v", result)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		t.Fatal("criou .git")
	}
	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
		t.Fatal("criou Dockerfile")
	}
	log, _ := os.ReadFile(calls)
	if strings.Contains(string(log), "repo create") {
		t.Fatal("criou repositório")
	}
}

func TestSemLoginExplica(t *testing.T) {
	fakeGH(t, false)
	if _, err := Plan(goProject(t), Options{}); err == nil || !strings.Contains(err.Error(), "gh auth login") {
		t.Fatalf("erro: %v", err)
	}
}

func TestTodoTemplateTemPublicacao(t *testing.T) {
	for _, tpl := range scaffold.LoadTemplates() {
		dest := filepath.Join(t.TempDir(), "minha-loja")
		if _, err := scaffold.WriteScaffold(tpl, "minha-loja", dest, nil); err != nil {
			t.Fatalf("%s: %v", tpl.Key(), err)
		}
		kind, data := KindOf(dest)
		if !slices.Contains(Kinds(), kind) {
			t.Errorf("%s: publish.kind %q não existe (existem: %v)", tpl.Key(), kind, Kinds())
			continue
		}
		if _, err := render(kind, data); err != nil {
			t.Errorf("%s: %v", tpl.Key(), err)
		}
	}
}
