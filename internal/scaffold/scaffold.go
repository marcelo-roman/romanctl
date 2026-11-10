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
	Name        string // PascalCase: namespaces, classes e projetos .NET
	Lower       string // o nome como foi digitado
	Snake       string // snake_case: pacotes Python e bancos
	Kebab       string // kebab-case: pacotes npm, imagens e serviços
	Camel       string // camelCase
	DbName      string // nome válido para banco de dados
	Module      string // módulo Go
	Package     string // pacote Java e Kotlin
	PackagePath string // o pacote em pastas: com/example/loja
	Postgres    bool
	Services    map[string]bool
}

const packagePrefix = "com.example."

func newTemplateData(name string, withs []string) templateData {
	pascal := projects.DotnetName(name)
	snake := strings.NewReplacer("-", "_", ".", "_").Replace(name)
	pkg := packagePrefix + strings.ReplaceAll(snake, "_", "")
	d := templateData{
		Name:        pascal,
		Lower:       name,
		Snake:       snake,
		Kebab:       strings.NewReplacer("_", "-", ".", "-").Replace(name),
		DbName:      snake,
		Module:      name,
		Package:     pkg,
		PackagePath: strings.ReplaceAll(pkg, ".", "/"),
		Services:    map[string]bool{},
	}
	if pascal != "" {
		d.Camel = strings.ToLower(pascal[:1]) + pascal[1:]
	}
	for _, s := range withs {
		d.Services[s] = true
	}
	d.Postgres = d.Services["postgres"]
	return d
}

// pathReplacer troca os marcadores dos caminhos, como __Name__ e __PackagePath__.
func (d templateData) pathReplacer() *strings.Replacer {
	return strings.NewReplacer("__Name__", d.Name, "__Lower__", d.Lower, "__Snake__", d.Snake, "__Kebab__", d.Kebab,
		"__Camel__", d.Camel, "__PackagePath__", d.PackagePath)
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
	paths := data.pathReplacer()
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
		files = append(files, generatedFile{paths.Replace(p), content})
		return nil
	})
	return files, err
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

func buildManifest(t *Template, name string, withs []string, data templateData) (manifest.Project, error) {
	run, err := renderString("run", t.Run, data)
	if err != nil {
		return manifest.Project{}, err
	}
	test, err := renderString("test", t.Test, data)
	if err != nil {
		return manifest.Project{}, err
	}
	m := manifest.Project{Name: name, Stack: t.Stack, Template: t.ID, Arch: t.Arch, Services: withs,
		Run: string(run), Test: string(test), URL: t.URL, Requires: t.Requires, Publish: manifest.Publish{Kind: t.Publish.Kind}}
	for _, s := range withs {
		if vars, ok := t.Env[s]; ok {
			if m.Env == nil {
				m.Env = manifest.Env{}
			}
			m.Env[s] = vars
		}
	}
	for key, value := range t.Publish.Vars {
		rendered, err := renderString("publish."+key, value, data)
		if err != nil {
			return manifest.Project{}, err
		}
		if m.Publish.Vars == nil {
			m.Publish.Vars = map[string]string{}
		}
		m.Publish.Vars[key] = string(rendered)
	}
	return m, nil
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
