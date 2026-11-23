package scaffold

import (
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/ui"
	"slices"
	"sort"
	"strings"
)

// Recognized vocabulary, even when there is no template for the combination yet.
var knownArch = []string{"bff", "clean", "cqrs", "ddd", "event-driven", "event-sourcing", "feature-sliced", "hexagonal",
	"layered", "microservices", "modular-monolith", "trpc", "vertical-slice"}

var archAliases = map[string]string{
	"clean-architecture": "clean",
	"eventsourcing":      "event-sourcing",
	"es":                 "event-sourcing",
	"ports-and-adapters": "hexagonal",
	"ports-adapters":     "hexagonal",
	"n-tier":             "layered",
	"mvc":                "layered",
	"vsa":                "vertical-slice",
	"vertical-slices":    "vertical-slice",
	"modular":            "modular-monolith",
	"modulith":           "modular-monolith",
	"eda":                "event-driven",
	"fsd":                "feature-sliced",
}

type conditionalPath struct {
	Path string `yaml:"path"`
	When string `yaml:"when"`
}

// overlay is a directory, usually under _variants, whose files replace the root's when When holds.
type overlay struct {
	From string `yaml:"from"`
	When string `yaml:"when"`
}

const variantsDir = "_variants"

// setupStep is a setup command. In YAML, a string or {run, when}: when restricts the step to a
// variant, like npm ci with postgres and npm install without it, to reconcile the lockfile.
type setupStep struct {
	Run  string `yaml:"run" json:"run"`
	When string `yaml:"when" json:"when,omitempty"`
}

func (s *setupStep) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		s.Run = node.Value
		return nil
	}
	type plain setupStep
	return node.Decode((*plain)(s))
}

// setupFor returns the setup commands that apply to the chosen services.
func (t *Template) setupFor(withs map[string]bool) []string {
	var out []string
	for _, step := range t.Setup {
		if step.When == "" || conditionHolds(step.When, withs) {
			out = append(out, step.Run)
		}
	}
	return out
}

// Template is a directory with template.yaml and the files to generate.
type Template struct {
	ID          string             `yaml:"id" json:"id"`
	Stack       string             `yaml:"stack" json:"stack"`
	Arch        []string           `yaml:"arch" json:"arch"`
	Description string             `yaml:"description" json:"description"`
	Services    []string           `yaml:"services" json:"services"`
	Run         string             `yaml:"run" json:"run"`
	Test        string             `yaml:"test" json:"test"`
	URL         string             `yaml:"url" json:"url"`
	Requires    []string           `yaml:"requires" json:"requires,omitempty"`
	Setup       []setupStep        `yaml:"setup" json:"setup,omitempty"`
	Env         manifest.Env       `yaml:"env" json:"env,omitempty"`
	Publish     manifest.Publish   `yaml:"publish" json:"publish,omitempty"`
	Processes   []manifest.Process `yaml:"processes" json:"processes,omitempty"`
	Conditional []conditionalPath  `yaml:"conditional" json:"-"`
	Overlays    []overlay          `yaml:"overlays" json:"-"`
	Source      string             `yaml:"-" json:"source"`
	files       fs.FS
}

func (t *Template) Key() string { return t.Stack + "/" + t.ID }

func userTemplatesDir() string { return filepath.Join(config.Home, ".config", "romanctl", "templates") }

// LoadTemplates reads romanctl's templates; the user's take precedence.
func LoadTemplates() []*Template {
	ensureInstalled()
	found := map[string]*Template{}
	load := func(root fs.FS, pattern, source string) {
		manifests, _ := fs.Glob(root, pattern)
		for _, manifest := range manifests {
			b, err := fs.ReadFile(root, manifest)
			if err != nil {
				continue
			}
			var t Template
			if err := yaml.Unmarshal(b, &t); err != nil {
				ui.Warnf("Invalid template %s: %v", manifest, err)
				continue
			}
			dir := path.Dir(manifest)
			sub, err := fs.Sub(root, dir)
			if err != nil {
				continue
			}
			t.Stack, t.ID = strings.SplitN(dir, "/", 2)[0], path.Base(dir)
			t.files, t.Source = sub, source
			sort.Strings(t.Arch)
			found[t.Key()] = &t
		}
	}
	load(os.DirFS(Root()), projectsPattern, "romanctl")
	load(os.DirFS(userTemplatesDir()), "*/*/template.yaml", "user")

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
			return nil, ui.Usagef("unknown architecture %q (known: %s)", a, strings.Join(knownArch, ", "))
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
		return i18n.T("  none")
	}
	return strings.Join(lines, "\n")
}

func findTemplate(stack string, arch []string) (*Template, error) {
	all := LoadTemplates()
	if stack == "" {
		return nil, ui.Usagef("specify --stack. Available combinations:\n%s", describeTemplates(all))
	}
	if len(arch) == 0 {
		return nil, ui.Usagef("specify --arch. Available combinations:\n%s", describeTemplates(all))
	}
	for _, t := range all {
		if t.Stack == stack && slices.Equal(t.Arch, arch) {
			return t, nil
		}
	}
	return nil, ui.Usagef("there is no template yet for --stack %s --arch %s. Available combinations:\n%s",
		stack, strings.Join(arch, ","), describeTemplates(all))
}

// MatchTemplates filters by what was given; empty fields do not restrict.
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

// FindByKey finds a project's template by the stack and id stored in romanctl.yaml.
func FindByKey(stack, id string) *Template {
	for _, t := range LoadTemplates() {
		if t.Stack == stack && t.ID == id {
			return t
		}
	}
	return nil
}

// AvailableCombinations describes the templates, for error messages.
func AvailableCombinations() string { return describeTemplates(LoadTemplates()) }
