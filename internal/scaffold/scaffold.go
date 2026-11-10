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

// templateData é o que os textos de template.yaml e generator.yaml enxergam, como next.
type templateData struct {
	Name     string
	Lower    string
	DbName   string
	Postgres bool
	Services map[string]bool
}

func newTemplateData(name string, withs []string) templateData {
	d := templateData{
		Name:     projects.DotnetName(name),
		Lower:    name,
		DbName:   strings.Join(words(name), "_"),
		Services: map[string]bool{},
	}
	for _, s := range withs {
		d.Services[s] = true
	}
	d.Postgres = d.Services["postgres"]
	return d
}

type generatedFile struct {
	Path       string
	Content    []byte
	Executable bool
}

func (f generatedFile) mode() os.FileMode {
	if f.Executable {
		return 0o755
	}
	return 0o644
}

type sourceFile struct {
	content    []byte
	executable bool
}

func conditionHolds(when string, withs map[string]bool) bool {
	if name, negated := strings.CutPrefix(when, "!"); negated {
		return !withs[name]
	}
	return withs[when]
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

// buildOutputs são pastas que os testes do template na própria pasta deixam para trás.
var buildOutputs = map[string]bool{".git": true, "node_modules": true, "bin": true, "obj": true, "target": true,
	".venv": true, "__pycache__": true, ".pytest_cache": true, ".mypy_cache": true, ".ruff_cache": true, "dist": true,
	".next": true, ".gradle": true, "build": true, "coverage": true, ".turbo": true}

var localFiles = map[string]bool{".env": true, "coverage.out": true, ".DS_Store": true}

func under(p, dir string) bool { return p == dir || strings.HasPrefix(p, dir+"/") }

// walkSource lê os arquivos de uma pasta do template, sem os artefatos de build.
func walkSource(fsys fs.FS, root string, visit func(rel string, f sourceFile) error) error {
	return fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			if p != root && buildOutputs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if localFiles[d.Name()] || !d.Type().IsRegular() {
			return nil
		}
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel := p
		if root != "." {
			rel = strings.TrimPrefix(p, root+"/")
		}
		return visit(rel, sourceFile{content: b, executable: info.Mode()&0o111 != 0})
	})
}

func excluded(conditional []conditionalPath, p string, withs map[string]bool) bool {
	for _, c := range conditional {
		if under(p, c.Path) && !conditionHolds(c.When, withs) {
			return true
		}
	}
	return false
}

// renderTemplate copia a raiz do template, aplica as variantes que valem e troca o nome de exemplo.
func renderTemplate(t *Template, name string, withs map[string]bool) ([]generatedFile, error) {
	files := map[string]sourceFile{}
	var order []string
	add := func(rel string, f sourceFile) error {
		if _, seen := files[rel]; !seen {
			order = append(order, rel)
		}
		files[rel] = f
		return nil
	}
	err := walkSource(t.files, ".", func(p string, f sourceFile) error {
		if p == "template.yaml" || under(p, variantsDir) || excluded(t.Conditional, p, withs) {
			return nil
		}
		return add(p, f)
	})
	if err != nil {
		return nil, err
	}
	for _, o := range t.Overlays {
		if !conditionHolds(o.When, withs) {
			continue
		}
		if err := walkSource(t.files, o.From, add); err != nil {
			return nil, err
		}
	}
	r := projectReplacer(name)
	out := make([]generatedFile, 0, len(order))
	for _, rel := range order {
		f := files[rel]
		out = append(out, generatedFile{r.Replace(rel), replaceContent(r, f.content), f.executable})
	}
	return out, nil
}

func validateScaffold(t *Template, name string, withs []string) error {
	if err := projects.ValidateName(name); err != nil {
		return ui.Usagef("%v", err)
	}
	if pascal := projects.DotnetName(name); pascal == "" || !unicode.IsLetter(rune(pascal[0])) {
		return ui.Usagef("o nome precisa começar com letra, porque vira namespace e pacote: %q", name)
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
	Missing  []string         `json:"missingTools,omitempty"`
	Setup    string           `json:"setup,omitempty"`
}

// Options são as escolhas de new para um template.
type Options struct {
	Git, DryRun, Offline bool
}

func buildManifest(t *Template, name string, withs []string) manifest.Project {
	r := projectReplacer(name)
	m := manifest.Project{Name: name, Stack: t.Stack, Template: t.ID, Arch: t.Arch, Services: withs,
		Run: r.Replace(t.Run), Test: r.Replace(t.Test), URL: t.URL, Requires: t.Requires, Publish: manifest.Publish{Kind: t.Publish.Kind}}
	for _, s := range withs {
		if vars, ok := t.Env[s]; ok {
			if m.Env == nil {
				m.Env = manifest.Env{}
			}
			m.Env[s] = vars
		}
	}
	for _, proc := range t.Processes {
		m.Processes = append(m.Processes, manifest.Process{Name: proc.Name, Dir: r.Replace(proc.Dir), Run: r.Replace(proc.Run), URL: proc.URL})
	}
	for key, value := range t.Publish.Vars {
		if m.Publish.Vars == nil {
			m.Publish.Vars = map[string]string{}
		}
		m.Publish.Vars[key] = r.Replace(value)
	}
	return m
}

// writeScaffold grava numa pasta temporária e só depois move, para nunca deixar
// um projeto pela metade.
func WriteScaffold(t *Template, name, dest string, withs []string) (ScaffoldResult, error) {
	files, err := renderTemplate(t, name, newTemplateData(name, withs).Services)
	if err != nil {
		return ScaffoldResult{}, err
	}
	project := buildManifest(t, name, withs)
	manifestYAML, err := yaml.Marshal(project)
	if err != nil {
		return ScaffoldResult{}, err
	}
	files = append(files, generatedFile{Path: manifest.FileName, Content: manifestYAML})

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
		if err := os.WriteFile(target, f.Content, f.mode()); err != nil {
			return result, err
		}
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		return result, err
	}
	return result, os.Rename(tmp, dest)
}

func NewFromTemplate(name, stack, archFlag, withFlag string, opts Options) error {
	arch, err := normalizeArch(splitList(archFlag))
	if err != nil {
		return err
	}
	t, err := findTemplate(strings.ToLower(stack), arch)
	if err != nil {
		return err
	}
	return ScaffoldProject(t, name, SplitServices(withFlag), opts)
}

func ScaffoldProject(t *Template, name string, withs []string, opts Options) error {
	if err := validateScaffold(t, name, withs); err != nil {
		return err
	}
	dest := filepath.Join(config.ProjDir, name)
	target := dest
	if opts.DryRun {
		target = "" // writeScaffold só lista os arquivos, sem gravar nada
	}
	result, err := WriteScaffold(t, name, target, withs)
	if err != nil {
		return i18n.Errorf("falha ao gerar %s: %w", name, err)
	}
	result.Path, result.DryRun = dest, opts.DryRun
	if opts.DryRun {
		result.Missing = missingTools(t.Requires)
		if ui.JSONOutput {
			return ui.PrintJSON(result)
		}
		ui.Infof("Simulação: nada foi gravado. Seriam criados %d arquivos em %s:", len(result.Files), dest)
		for _, f := range result.Files {
			fmt.Println("  " + f)
		}
		return nil
	}
	if opts.Git && sys.Have("git") {
		sys.Output("git", "-C", dest, "init", "-q")
	}
	result.Missing = ensureTools(t.Requires)
	result.Setup = runSetup(t, dest, opts.Offline, result.Missing)
	if ui.JSONOutput {
		return ui.PrintJSON(result)
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
