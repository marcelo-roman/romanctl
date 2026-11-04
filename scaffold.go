package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"gopkg.in/yaml.v3"
)

//go:embed all:templates
var embeddedTemplates embed.FS

const manifestFile = "romanctl.yaml"

// Vocabulário reconhecido, mesmo quando ainda não há template para a combinação.
var knownArch = []string{"clean", "ddd", "cqrs", "event-sourcing", "bff", "trpc"}

var archAliases = map[string]string{
	"clean-architecture": "clean",
	"eventsourcing":      "event-sourcing",
	"es":                 "event-sourcing",
}

type conditionalPath struct {
	Path string `yaml:"path"`
	When string `yaml:"when"`
}

// Template é uma pasta com template.yaml e os arquivos a gerar.
type Template struct {
	ID          string            `yaml:"id" json:"id"`
	Stack       string            `yaml:"stack" json:"stack"`
	Arch        []string          `yaml:"arch" json:"arch"`
	Description string            `yaml:"description" json:"description"`
	Services    []string          `yaml:"services" json:"services"`
	Run         string            `yaml:"run" json:"run"`
	Test        string            `yaml:"test" json:"test"`
	URL         string            `yaml:"url" json:"url"`
	Conditional []conditionalPath `yaml:"conditional" json:"-"`
	Source      string            `yaml:"-" json:"source"`
	files       fs.FS
}

func (t *Template) Key() string { return t.Stack + "/" + t.ID }

// ProjectManifest é o romanctl.yaml gravado em cada projeto gerado.
type ProjectManifest struct {
	Name     string   `yaml:"name" json:"name"`
	Stack    string   `yaml:"stack" json:"stack"`
	Template string   `yaml:"template" json:"template"`
	Arch     []string `yaml:"arch" json:"arch"`
	Services []string `yaml:"services" json:"services"`
	Run      string   `yaml:"run" json:"run"`
	Test     string   `yaml:"test" json:"test"`
	URL      string   `yaml:"url" json:"url"`
}

func readManifest(projectPath string) (*ProjectManifest, bool) {
	b, err := os.ReadFile(filepath.Join(projectPath, manifestFile))
	if err != nil {
		return nil, false
	}
	var m ProjectManifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		warnf("%s inválido em %s: %v", manifestFile, projectPath, err)
		return nil, false
	}
	return &m, true
}

func userTemplatesDir() string { return filepath.Join(home, ".config", "romanctl", "templates") }

// loadTemplates lê os templates embutidos; os do usuário têm prioridade.
func loadTemplates() []*Template {
	found := map[string]*Template{}
	load := func(root fs.FS, source string) {
		manifests, _ := fs.Glob(root, "*/*/template.yaml")
		for _, manifest := range manifests {
			b, err := fs.ReadFile(root, manifest)
			if err != nil {
				continue
			}
			var t Template
			if err := yaml.Unmarshal(b, &t); err != nil {
				warnf("Template inválido %s: %v", manifest, err)
				continue
			}
			dir := path.Dir(manifest)
			sub, err := fs.Sub(root, dir)
			if err != nil {
				continue
			}
			t.Stack, t.ID = path.Dir(dir), path.Base(dir)
			t.files, t.Source = sub, source
			sort.Strings(t.Arch)
			found[t.Key()] = &t
		}
	}
	if sub, err := fs.Sub(embeddedTemplates, "templates"); err == nil {
		load(sub, "embutido")
	}
	load(os.DirFS(userTemplatesDir()), "usuario")

	list := make([]*Template, 0, len(found))
	for _, t := range found {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Key() < list[j].Key() })
	return list
}

func splitList(s string) []string {
	var out []string
	for _, item := range strings.Split(s, ",") {
		if item = strings.ToLower(strings.TrimSpace(item)); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func normalizeArch(in []string) ([]string, error) {
	var out []string
	for _, a := range in {
		if alias, ok := archAliases[a]; ok {
			a = alias
		}
		if !slices.Contains(knownArch, a) {
			return nil, usagef("arquitetura desconhecida %q (conhecidas: %s)", a, strings.Join(knownArch, ", "))
		}
		if !slices.Contains(out, a) {
			out = append(out, a)
		}
	}
	sort.Strings(out)
	return out, nil
}

func describeTemplates(templates []*Template) string {
	var lines []string
	for _, t := range templates {
		lines = append(lines, fmt.Sprintf("  --stack %s --arch %s", t.Stack, strings.Join(t.Arch, ",")))
	}
	if len(lines) == 0 {
		return "  nenhum"
	}
	return strings.Join(lines, "\n")
}

func findTemplate(stack string, arch []string) (*Template, error) {
	all := loadTemplates()
	if stack == "" {
		return nil, usagef("informe --stack. Combinações disponíveis:\n%s", describeTemplates(all))
	}
	if len(arch) == 0 {
		return nil, usagef("informe --arch. Combinações disponíveis:\n%s", describeTemplates(all))
	}
	for _, t := range all {
		if t.Stack == stack && slices.Equal(t.Arch, arch) {
			return t, nil
		}
	}
	return nil, usagef("ainda não existe template para --stack %s --arch %s. Combinações disponíveis:\n%s",
		stack, strings.Join(arch, ","), describeTemplates(all))
}

type templateData struct {
	Name     string // nome em PascalCase, usado em namespaces e projetos
	Lower    string // nome do projeto como foi digitado
	DbName   string // nome válido para banco de dados
	Postgres bool
	Services map[string]bool
}

func newTemplateData(name string, services []string) templateData {
	d := templateData{
		Name:     dotnetName(name),
		Lower:    name,
		DbName:   strings.NewReplacer("-", "_", ".", "_").Replace(name),
		Services: map[string]bool{},
	}
	for _, s := range services {
		d.Services[s] = true
	}
	d.Postgres = d.Services["postgres"]
	return d
}

type generatedFile struct {
	Path    string
	Content []byte
}

func conditionHolds(when string, services map[string]bool) bool {
	if name, negated := strings.CutPrefix(when, "!"); negated {
		return !services[name]
	}
	return services[when]
}

func renderString(name, text string, data templateData) ([]byte, error) {
	tpl, err := template.New(name).Option("missingkey=error").Parse(text)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
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
			return fmt.Errorf("template %s: %w", p, err)
		}
		files = append(files, generatedFile{strings.ReplaceAll(p, "__Name__", data.Name), content})
		return nil
	})
	return files, err
}

func validateScaffold(t *Template, name string, services []string) error {
	if err := validateName(name); err != nil {
		return usagef("%v", err)
	}
	if t.Stack == "dotnet" {
		if pascal := dotnetName(name); pascal == "" || !unicode.IsLetter(rune(pascal[0])) {
			return usagef("em .NET o nome precisa começar com letra, porque vira namespace: %q", name)
		}
	}
	for _, s := range services {
		if findService(s) == nil {
			return usagef("serviço desconhecido %q", s)
		}
		if !slices.Contains(t.Services, findService(s).Key) {
			return usagef("o template %s não oferece %q (oferece: %s)", t.Key(), s, strings.Join(t.Services, ", "))
		}
	}
	return nil
}

// ScaffoldResult é o que "new" informa, inclusive em --dry-run e --json.
type ScaffoldResult struct {
	OK       bool            `json:"ok"`
	DryRun   bool            `json:"dryRun"`
	Path     string          `json:"path"`
	Manifest ProjectManifest `json:"manifest"`
	Files    []string        `json:"files"`
}

func buildManifest(t *Template, name string, services []string, data templateData) (ProjectManifest, error) {
	run, err := renderString("run", t.Run, data)
	if err != nil {
		return ProjectManifest{}, err
	}
	return ProjectManifest{Name: name, Stack: t.Stack, Template: t.ID, Arch: t.Arch,
		Services: services, Run: string(run), Test: t.Test, URL: t.URL}, nil
}

// writeScaffold grava numa pasta temporária e só depois move, para nunca deixar
// um projeto pela metade.
func writeScaffold(t *Template, name, dest string, services []string) (ScaffoldResult, error) {
	data := newTemplateData(name, services)
	files, err := renderTemplate(t, data)
	if err != nil {
		return ScaffoldResult{}, err
	}
	manifest, err := buildManifest(t, name, services, data)
	if err != nil {
		return ScaffoldResult{}, err
	}
	manifestYAML, err := yaml.Marshal(manifest)
	if err != nil {
		return ScaffoldResult{}, err
	}
	files = append(files, generatedFile{manifestFile, manifestYAML})

	result := ScaffoldResult{OK: true, Path: dest, Manifest: manifest}
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

func newFromTemplate(name, stack, archFlag, withFlag string, git, dryRun bool) error {
	arch, err := normalizeArch(splitList(archFlag))
	if err != nil {
		return err
	}
	t, err := findTemplate(strings.ToLower(stack), arch)
	if err != nil {
		return err
	}
	var services []string
	for _, s := range splitList(withFlag) {
		if svc := findService(s); svc != nil {
			s = svc.Key
		}
		if !slices.Contains(services, s) {
			services = append(services, s)
		}
	}
	sort.Strings(services)
	return scaffoldProject(t, name, services, git, dryRun)
}

func scaffoldProject(t *Template, name string, services []string, git, dryRun bool) error {
	if err := validateScaffold(t, name, services); err != nil {
		return err
	}
	dest := filepath.Join(projDir, name)
	target := dest
	if dryRun {
		target = "" // writeScaffold só lista os arquivos, sem gravar nada
	}
	result, err := writeScaffold(t, name, target, services)
	if err != nil {
		return fmt.Errorf("falha ao gerar %s: %w", name, err)
	}
	result.Path, result.DryRun = dest, dryRun
	if !dryRun && git && have("git") {
		output("git", "-C", dest, "init", "-q")
	}
	if jsonOutput {
		return printJSON(result)
	}
	if dryRun {
		infof("Simulação: nada foi gravado. Seriam criados %d arquivos em %s:", len(result.Files), dest)
		for _, f := range result.Files {
			fmt.Println("  " + f)
		}
		return nil
	}
	okf("Projeto %s criado com o template %s (%d arquivos)", sBoldTxt.Render(name), t.Key(), len(result.Files))
	fmt.Printf("   Pasta:   %s\n   Rodar:   romanctl run %s\n   Testar:  cd %s && %s\n   Abrir:   romanctl code %s\n",
		dest, name, dest, result.Manifest.Test, name)
	for _, s := range services {
		if svc := findService(s); svc != nil && !svc.Running() {
			warnf("%s não está rodando. Antes de rodar o projeto: romanctl start %s", svc.Name, svc.Key)
		}
	}
	return nil
}

// ---------------------------------------------------------------- catalog

type catalogOutput struct {
	Stacks    []string         `json:"stacks"`
	Arch      []string         `json:"arch"`
	Services  []string         `json:"services"`
	Templates []*Template      `json:"templates"`
	Simple    []map[string]any `json:"simpleTypes"`
}

func catalog() error {
	templates := loadTemplates()
	out := catalogOutput{Arch: knownArch, Templates: templates}
	for _, t := range templates {
		if !slices.Contains(out.Stacks, t.Stack) {
			out.Stacks = append(out.Stacks, t.Stack)
		}
	}
	for _, s := range services {
		out.Services = append(out.Services, s.Key)
	}
	for _, t := range projectTypes {
		out.Simple = append(out.Simple, map[string]any{"type": t.Key, "description": t.Label})
	}
	if jsonOutput {
		return printJSON(out)
	}
	fmt.Println(sTitle.Render("Templates de arquitetura"))
	for _, t := range templates {
		fmt.Printf("  %s  %s\n", sBoldTxt.Render(t.Key()), sGray.Render("("+t.Source+")"))
		fmt.Printf("    %s\n", t.Description)
		fmt.Printf("    romanctl new <nome> --stack %s --arch %s", t.Stack, strings.Join(t.Arch, ","))
		if len(t.Services) > 0 {
			fmt.Printf(" [--with %s]", strings.Join(t.Services, ","))
		}
		fmt.Println()
	}
	fmt.Println("\n" + sTitle.Render("Projetos simples"))
	fmt.Println("  romanctl new <tipo> <nome>   tipos: " + typeKeys())
	fmt.Printf("\nTemplates próprios: %s/<stack>/<id>/template.yaml\n", userTemplatesDir())
	return nil
}

// ---------------------------------------------------------------- saída

var jsonOutput bool

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// usageError indica uso incorreto: sai com código 2, separado de falha de execução.
type usageError struct{ msg string }

func (e usageError) Error() string { return e.msg }

func usagef(f string, a ...any) error { return usageError{fmt.Sprintf(f, a...)} }

func isUsage(err error) bool {
	var u usageError
	return errors.As(err, &u)
}
