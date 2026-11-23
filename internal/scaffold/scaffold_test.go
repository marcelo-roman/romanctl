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

func TestInMemoryVariant(t *testing.T) {
	r, err := WriteScaffold(dotnetTemplate(t), "my-shop", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !hasPath(r.Files, "Persistence/InMemory/") || hasPath(r.Files, "Persistence/EfCore/") {
		t.Fatalf("wrong persistence: %v", r.Files)
	}
	if !hasPath(r.Files, "src/MyShop.Domain/MyShop.Domain.csproj") {
		t.Fatalf("name not applied: %v", r.Files)
	}
	if hasPath(r.Files, "SampleApp") || hasPath(r.Files, "template.yaml") {
		t.Fatalf("a template file was left over: %v", r.Files)
	}
}

func TestPostgresVariant(t *testing.T) {
	r, err := WriteScaffold(dotnetTemplate(t), "my-shop", "", []string{"postgres"})
	if err != nil {
		t.Fatal(err)
	}
	if hasPath(r.Files, "Persistence/InMemory/") || !hasPath(r.Files, "Persistence/EfCore/") {
		t.Fatalf("wrong persistence: %v", r.Files)
	}
	if r.Manifest.Run != "dotnet run --project src/MyShop.Api" {
		t.Fatalf("wrong run: %s", r.Manifest.Run)
	}
}

func TestWithoutDestinationNothingIsWritten(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := WriteScaffold(dotnetTemplate(t), "x", "", nil); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("wrote files: %v", entries)
	}
}

func TestWritesProjectWithManifest(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "shop")
	if _, err := WriteScaffold(dotnetTemplate(t), "shop", dest, []string{"postgres"}); err != nil {
		t.Fatal(err)
	}
	m, ok := manifest.Read(dest)
	if !ok || m.Template != "clean-ddd-cqrs" || !slices.Equal(m.Services, []string{"postgres"}) {
		t.Fatalf("manifesto errado: %+v", m)
	}
	settings, err := os.ReadFile(filepath.Join(dest, "src/Shop.Api/appsettings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(settings), "Database=shop;") {
		t.Fatalf("wrong connection string: %s", settings)
	}
}

func TestRejectsInvalidCombinations(t *testing.T) {
	tpl := dotnetTemplate(t)
	cases := []error{
		validateScaffold(tpl, "1app", nil),
		validateScaffold(tpl, "Invalid Name", nil),
		validateScaffold(tpl, "ok", []string{"does-not-exist"}),
		validateScaffold(tpl, "ok", []string{"redis"}),
	}
	for i, err := range cases {
		if err == nil || !ui.IsUsage(err) {
			t.Errorf("case %d: unexpected error: %v", i, err)
		}
	}
	if _, err := normalizeArch([]string{"inventada"}); !ui.IsUsage(err) {
		t.Errorf("accepted a made-up architecture: %v", err)
	}
	if _, err := findTemplate("dotnet", []string{"trpc"}); !ui.IsUsage(err) {
		t.Errorf("accepted a combination without a template: %v", err)
	}
}

func TestExecutableFileStaysExecutable(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "template.yaml"), []byte("id: x\n"), 0o644)
	os.WriteFile(filepath.Join(root, "mvnw"), []byte("#!/bin/sh\n"), 0o755)
	os.WriteFile(filepath.Join(root, "README.md"), []byte("sample-app\n"), 0o644)
	dest := filepath.Join(t.TempDir(), "shop")
	if _, err := WriteScaffold(&Template{ID: "x", Stack: "s", files: os.DirFS(root)}, "shop", dest, nil); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]os.FileMode{"mvnw": 0o755, "README.md": 0o644} {
		info, err := os.Stat(filepath.Join(dest, name))
		if err != nil || info.Mode().Perm() != want {
			t.Fatalf("%s: %v %v", name, info.Mode(), err)
		}
	}
}
