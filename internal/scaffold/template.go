package scaffold

import (
	"embed"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/ui"
	"slices"
	"sort"
	"strings"
)

//go:embed all:templates
var embeddedTemplates embed.FS

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

func userTemplatesDir() string { return filepath.Join(config.Home, ".config", "romanctl", "templates") }

// loadTemplates lê os templates embutidos; os do usuário têm prioridade.
func LoadTemplates() []*Template {
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
				ui.Warnf("Template inválido %s: %v", manifest, err)
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
			return nil, ui.Usagef("arquitetura desconhecida %q (conhecidas: %s)", a, strings.Join(knownArch, ", "))
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
		lines = append(lines, i18n.Sprintf("  --stack %s --arch %s", t.Stack, strings.Join(t.Arch, ",")))
	}
	if len(lines) == 0 {
		return i18n.T("  nenhum")
	}
	return strings.Join(lines, "\n")
}

func findTemplate(stack string, arch []string) (*Template, error) {
	all := LoadTemplates()
	if stack == "" {
		return nil, ui.Usagef("informe --stack. Combinações disponíveis:\n%s", describeTemplates(all))
	}
	if len(arch) == 0 {
		return nil, ui.Usagef("informe --arch. Combinações disponíveis:\n%s", describeTemplates(all))
	}
	for _, t := range all {
		if t.Stack == stack && slices.Equal(t.Arch, arch) {
			return t, nil
		}
	}
	return nil, ui.Usagef("ainda não existe template para --stack %s --arch %s. Combinações disponíveis:\n%s",
		stack, strings.Join(arch, ","), describeTemplates(all))
}

// MatchTemplates filtra pelo que já foi informado; os campos vazios não restringem.
func MatchTemplates(stack, archFlag string) ([]*Template, error) {
	arch, err := normalizeArch(splitList(archFlag))
	if err != nil {
		return nil, err
	}
	var out []*Template
	for _, t := range LoadTemplates() {
		if stack != "" && !strings.EqualFold(t.Stack, stack) {
			continue
		}
		if !slices.ContainsFunc(arch, func(a string) bool { return !slices.Contains(t.Arch, a) }) {
			out = append(out, t)
		}
	}
	return out, nil
}

// AvailableCombinations descreve os templates, para mensagens de erro.
func AvailableCombinations() string { return describeTemplates(LoadTemplates()) }
