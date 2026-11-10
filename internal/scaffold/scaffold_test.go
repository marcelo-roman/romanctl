package scaffold

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"romanctl/internal/manifest"
	"romanctl/internal/ui"
)

func dotnetTemplate(t *testing.T) *Template {
	t.Helper()
	arch, err := normalizeArch([]string{"cqrs", "clean-architecture", "ddd", "ddd"})
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := findTemplate("dotnet", arch)
	if err != nil {
		t.Fatal(err)
	}
	return tpl
}

func hasPath(files []string, part string) bool {
	return slices.ContainsFunc(files, func(f string) bool { return strings.Contains(f, part) })
}

func TestVariacaoEmMemoria(t *testing.T) {
	r, err := WriteScaffold(dotnetTemplate(t), "minha-loja", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !hasPath(r.Files, "Persistence/InMemory/") || hasPath(r.Files, "Persistence/EfCore/") {
		t.Fatalf("persistência errada: %v", r.Files)
	}
	if !hasPath(r.Files, "src/MinhaLoja.Domain/MinhaLoja.Domain.csproj") {
		t.Fatalf("nome não aplicado: %v", r.Files)
	}
	if hasPath(r.Files, "SampleApp") || hasPath(r.Files, "template.yaml") {
		t.Fatalf("sobrou arquivo do template: %v", r.Files)
	}
}

func TestVariacaoComPostgres(t *testing.T) {
	r, err := WriteScaffold(dotnetTemplate(t), "minha-loja", "", []string{"postgres"})
	if err != nil {
		t.Fatal(err)
	}
	if hasPath(r.Files, "Persistence/InMemory/") || !hasPath(r.Files, "Persistence/EfCore/") {
		t.Fatalf("persistência errada: %v", r.Files)
	}
	if r.Manifest.Run != "dotnet run --project src/MinhaLoja.Api" {
		t.Fatalf("run errado: %s", r.Manifest.Run)
	}
}

func TestSemDestinoNadaEGravado(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := WriteScaffold(dotnetTemplate(t), "x", "", nil); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("gravou arquivos: %v", entries)
	}
}

func TestGravaProjetoComManifesto(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "loja")
	if _, err := WriteScaffold(dotnetTemplate(t), "loja", dest, []string{"postgres"}); err != nil {
		t.Fatal(err)
	}
	m, ok := manifest.Read(dest)
	if !ok || m.Template != "clean-ddd-cqrs" || !slices.Equal(m.Services, []string{"postgres"}) {
		t.Fatalf("manifesto errado: %+v", m)
	}
	settings, err := os.ReadFile(filepath.Join(dest, "src/Loja.Api/appsettings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(settings), "Database=loja;") {
		t.Fatalf("connection string errada: %s", settings)
	}
}

func TestRecusaCombinacoesInvalidas(t *testing.T) {
	tpl := dotnetTemplate(t)
	cases := []error{
		validateScaffold(tpl, "1app", nil),
		validateScaffold(tpl, "Nome Invalido", nil),
		validateScaffold(tpl, "ok", []string{"nao-existe"}),
		validateScaffold(tpl, "ok", []string{"redis"}),
	}
	for i, err := range cases {
		if err == nil || !ui.IsUsage(err) {
			t.Errorf("caso %d: erro inesperado: %v", i, err)
		}
	}
	if _, err := normalizeArch([]string{"inventada"}); !ui.IsUsage(err) {
		t.Errorf("aceitou arquitetura inventada: %v", err)
	}
	if _, err := findTemplate("dotnet", []string{"trpc"}); !ui.IsUsage(err) {
		t.Errorf("aceitou combinação sem template: %v", err)
	}
}

func TestArquivoExecutavelContinuaExecutavel(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "template.yaml"), []byte("id: x\n"), 0o644)
	os.WriteFile(filepath.Join(root, "mvnw"), []byte("#!/bin/sh\n"), 0o755)
	os.WriteFile(filepath.Join(root, "README.md"), []byte("sample-app\n"), 0o644)
	dest := filepath.Join(t.TempDir(), "loja")
	if _, err := WriteScaffold(&Template{ID: "x", Stack: "s", files: os.DirFS(root)}, "loja", dest, nil); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]os.FileMode{"mvnw": 0o755, "README.md": 0o644} {
		info, err := os.Stat(filepath.Join(dest, name))
		if err != nil || info.Mode().Perm() != want {
			t.Fatalf("%s: %v %v", name, info.Mode(), err)
		}
	}
}
