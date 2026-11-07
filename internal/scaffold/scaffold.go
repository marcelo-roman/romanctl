package scaffold

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/projects"
	"romanctl/internal/services"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"slices"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

type templateData struct {
	Name     string // nome em PascalCase, usado em namespaces e projetos
	Lower    string // nome do projeto como foi digitado
	DbName   string // nome válido para banco de dados
	Postgres bool
	Services map[string]bool
}

func newTemplateData(name string, withs []string) templateData {
	d := templateData{
		Name:     projects.DotnetName(name),
		Lower:    name,
		DbName:   strings.NewReplacer("-", "_", ".", "_").Replace(name),
		Services: map[string]bool{},
	}
	for _, s := range withs {
		d.Services[s] = true
	}
	d.Postgres = d.Services["postgres"]
	return d
}

type generatedFile struct {
	Path    string
	Content []byte
}

func conditionHolds(when string, withs map[string]bool) bool {
	if name, negated := strings.CutPrefix(when, "!"); negated {
		return !withs[name]
	}
	return withs[when]
}

func renderString(name, text string, data templateData) ([]byte, error) {
	out, err := renderAny(name, text, data)
	return []byte(out), err
}

func renderAny(name, text string, data any) (string, error) {
	tpl, err := template.New(name).Option("missingkey=error").Parse(text)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

func renderTemplate(t *Template, data templateData) ([]generatedFile, error) {
	var files []generatedFile
	err := fs.WalkDir(t.files, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || p == "template.yaml" {
			return nil
		}
		for _, c := range t.Conditional {
			if (p == c.Path || strings.HasPrefix(p, c.Path+"/")) && !conditionHolds(c.When, data.Services) {
				return nil
			}
		}
		b, err := fs.ReadFile(t.files, p)
		if err != nil {
			return err
		}
		content, err := renderString(p, string(b), data)
		if err != nil {
			return i18n.Errorf("template %s: %w", p, err)
		}
		files = append(files, generatedFile{strings.ReplaceAll(p, "__Name__", data.Name), content})
		return nil
	})
	return files, err
}

func validateScaffold(t *Template, name string, withs []string) error {
	if err := projects.ValidateName(name); err != nil {
		return ui.Usagef("%v", err)
	}
	if t.Stack == "dotnet" {
		if pascal := projects.DotnetName(name); pascal == "" || !unicode.IsLetter(rune(pascal[0])) {
			return ui.Usagef("em .NET o nome precisa começar com letra, porque vira namespace: %q", name)
		}
	}
	for _, s := range withs {
		if services.FindService(s) == nil {
			return ui.Usagef("serviço desconhecido %q", s)
		}
		if !slices.Contains(t.Services, services.FindService(s).Key) {
			return ui.Usagef("o template %s não oferece %q (oferece: %s)", t.Key(), s, strings.Join(t.Services, ", "))
		}
	}
	return nil
}

// ScaffoldResult é o que "new" informa, inclusive em --dry-run e --json.
type ScaffoldResult struct {
	OK       bool             `json:"ok"`
	DryRun   bool             `json:"dryRun"`
	Path     string           `json:"path"`
	Manifest manifest.Project `json:"manifest"`
	Files    []string         `json:"files"`
}

func buildManifest(t *Template, name string, withs []string, data templateData) (manifest.Project, error) {
	run, err := renderString("run", t.Run, data)
	if err != nil {
		return manifest.Project{}, err
	}
	return manifest.Project{Name: name, Stack: t.Stack, Template: t.ID, Arch: t.Arch,
		Services: withs, Run: string(run), Test: t.Test, URL: t.URL}, nil
}

// writeScaffold grava numa pasta temporária e só depois move, para nunca deixar
// um projeto pela metade.
func WriteScaffold(t *Template, name, dest string, withs []string) (ScaffoldResult, error) {
	data := newTemplateData(name, withs)
	files, err := renderTemplate(t, data)
	if err != nil {
		return ScaffoldResult{}, err
	}
	project, err := buildManifest(t, name, withs, data)
	if err != nil {
		return ScaffoldResult{}, err
	}
	manifestYAML, err := yaml.Marshal(project)
	if err != nil {
		return ScaffoldResult{}, err
	}
	files = append(files, generatedFile{manifest.FileName, manifestYAML})

	result := ScaffoldResult{OK: true, Path: dest, Manifest: project}
	for _, f := range files {
		result.Files = append(result.Files, f.Path)
	}
	sort.Strings(result.Files)
	if dest == "" {
		return result, nil
	}

	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return result, err
	}
	tmp, err := os.MkdirTemp(parent, ".gerando-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(tmp)
	for _, f := range files {
		target := filepath.Join(tmp, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return result, err
		}
		if err := os.WriteFile(target, f.Content, 0o644); err != nil {
			return result, err
		}
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		return result, err
	}
	return result, os.Rename(tmp, dest)
}

func NewFromTemplate(name, stack, archFlag, withFlag string, git, dryRun bool) error {
	arch, err := normalizeArch(splitList(archFlag))
	if err != nil {
		return err
	}
	t, err := findTemplate(strings.ToLower(stack), arch)
	if err != nil {
		return err
	}
	withs := SplitServices(withFlag)
	return ScaffoldProject(t, name, withs, git, dryRun)
}

func ScaffoldProject(t *Template, name string, withs []string, git, dryRun bool) error {
	if err := validateScaffold(t, name, withs); err != nil {
		return err
	}
	dest := filepath.Join(config.ProjDir, name)
	target := dest
	if dryRun {
		target = "" // writeScaffold só lista os arquivos, sem gravar nada
	}
	result, err := WriteScaffold(t, name, target, withs)
	if err != nil {
		return i18n.Errorf("falha ao gerar %s: %w", name, err)
	}
	result.Path, result.DryRun = dest, dryRun
	if !dryRun && git && sys.Have("git") {
		sys.Output("git", "-C", dest, "init", "-q")
	}
	if ui.JSONOutput {
		return ui.PrintJSON(result)
	}
	if dryRun {
		ui.Infof("Simulação: nada foi gravado. Seriam criados %d arquivos em %s:", len(result.Files), dest)
		for _, f := range result.Files {
			fmt.Println("  " + f)
		}
		return nil
	}
	ui.Okf("Projeto %s criado com o template %s (%d arquivos)", ui.SBoldTxt.Render(name), t.Key(), len(result.Files))
	i18n.Printf("   Pasta:   %s\n   Rodar:   romanctl run %s\n   Testar:  cd %s && %s\n   Abrir:   romanctl code %s\n",
		dest, name, dest, result.Manifest.Test, name)
	for _, s := range withs {
		if svc := services.FindService(s); svc != nil && !svc.Running() {
			ui.Warnf("%s não está rodando. Antes de rodar o projeto: romanctl start %s", svc.Name, svc.Key)
		}
	}
	return nil
}

// SplitServices lê --with, aceitando apelidos como pg e mongodb.
func SplitServices(withFlag string) []string {
	var out []string
	for _, s := range splitList(withFlag) {
		if svc := services.FindService(s); svc != nil {
			s = svc.Key
		}
		if !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
