package scaffold

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"

	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/projects"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

type requirement struct {
	Path string `yaml:"path"`
	Hint string `yaml:"hint"`
}

type patch struct {
	When   string `yaml:"when"`
	File   string `yaml:"file"`
	Before string `yaml:"before"`
	SkipIf string `yaml:"skipIf"`
	Text   string `yaml:"text"`
}

// Generator creates files in an existing project. It lives in
// generators/<target>/<kind>/generator.yaml, with its files next to it, in the files directory.
type Generator struct {
	Kind        string            `yaml:"-" json:"kind"`
	Description string            `yaml:"description" json:"description"`
	Aliases     []string          `yaml:"aliases" json:"aliases,omitempty"`
	NeedsIn     bool              `yaml:"needsIn" json:"needsIn"`
	Includes    []string          `yaml:"includes" json:"includes,omitempty"`
	Requires    []requirement     `yaml:"requires" json:"-"`
	Conditional []conditionalPath `yaml:"conditional" json:"-"`
	Patches     []patch           `yaml:"patches" json:"-"`
	After       []string          `yaml:"after" json:"-"`
	Next        string            `yaml:"next" json:"-"`
	files       fs.FS
}

func userGeneratorsDir() string {
	return filepath.Join(config.Home, ".config", "romanctl", "generators")
}

// LoadGenerators reads a target's generators; the user's take precedence.
func LoadGenerators(target string) []*Generator {
	ensureInstalled()
	found := map[string]*Generator{}
	language, name, _ := strings.Cut(target, "/")
	load := func(root fs.FS, dir string) {
		manifests, _ := fs.Glob(root, dir+"/*/generator.yaml")
		for _, file := range manifests {
			b, err := fs.ReadFile(root, file)
			if err != nil {
				continue
			}
			var g Generator
			if err := yaml.Unmarshal(b, &g); err != nil {
				ui.Warnf("Invalid generator %s: %v", file, err)
				continue
			}
			dir := path.Dir(file)
			g.Kind = path.Base(dir)
			if sub, err := fs.Sub(root, dir+"/files"); err == nil {
				g.files = sub
			}
			found[g.Kind] = &g
		}
	}
	load(os.DirFS(Root()), path.Join(language, generatorsSubdir, name))
	load(os.DirFS(userGeneratorsDir()), target)
	list := make([]*Generator, 0, len(found))
	for _, g := range found {
		list = append(list, g)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Kind < list[j].Kind })
	return list
}

// AllTargets lists the targets that have generators, for the catalog.
func AllTargets() []string {
	ensureInstalled()
	dirs, _ := filepath.Glob(filepath.Join(Root(), "*", generatorsSubdir, "*"))
	var out []string
	for _, dir := range dirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			out = append(out, filepath.Base(filepath.Dir(filepath.Dir(dir)))+"/"+filepath.Base(dir))
		}
	}
	return out
}

// TargetOf works out which generators apply to a project.
func TargetOf(projectPath string) string {
	if m, ok := manifest.Read(projectPath); ok {
		return m.Stack + "/" + m.Template
	}
	switch projects.DetectType(projectPath) {
	case "next":
		return "node/next"
	case "react/vite":
		return "react/vite"
	case "express":
		return "node/express"
	case "go":
		return "go/simple"
	case "dotnet":
		return "dotnet/simple"
	}
	return ""
}

// FindGenerator looks up by kind or by alias.
func FindGenerator(list []*Generator, name string) *Generator {
	name = strings.ToLower(name)
	for _, g := range list {
		if g.Kind == name || slices.Contains(g.Aliases, name) {
			return g
		}
	}
	return nil
}

// ---------------------------------------------------------------- names

var (
	identifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
	wordEdge   = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

// Pascal accepts product-item, product_item and ProductItem.
func Pascal(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	var out strings.Builder
	for _, p := range parts {
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		out.WriteString(string(r))
	}
	return out.String()
}

func camel(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func separated(s, sep string) string {
	return strings.ToLower(wordEdge.ReplaceAllString(s, "${1}"+sep+"${2}"))
}

// Plural follows English rules, because the generated code is in English.
func Plural(s string) string {
	lower := strings.ToLower(s)
	switch {
	case len(s) > 1 && strings.HasSuffix(lower, "y") && !strings.ContainsAny(lower[len(lower)-2:len(lower)-1], "aeiou"):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(lower, "s"), strings.HasSuffix(lower, "x"), strings.HasSuffix(lower, "z"),
		strings.HasSuffix(lower, "ch"), strings.HasSuffix(lower, "sh"):
		return s + "es"
	}
	return s + "s"
}

// addData is what generator.yaml texts see, like next.
type addData struct {
	templateData
	Entity, EntityPlural, EntitySnake, EntityPluralSnake, EntityKebab, EntityPluralKebab string
	In                                                                                   string
	names                                                                                *strings.Replacer
	jsExt                                                                                bool
}

func projectIdentity(projectPath string) (name string, services []string) {
	base := filepath.Base(projectPath)
	if m, ok := manifest.Read(projectPath); ok {
		return m.Name, m.Services
	}
	if files, _ := filepath.Glob(filepath.Join(projectPath, "*.csproj")); len(files) > 0 {
		return strings.TrimSuffix(filepath.Base(files[0]), ".csproj"), nil
	}
	return base, nil
}

func newAddData(projectPath, entity, plural, in string) addData {
	name, services := projectIdentity(projectPath)
	base := newTemplateData(name, services)
	if plural == "" {
		plural = Plural(entity)
	}
	d := addData{templateData: base, Entity: entity, EntityPlural: plural, In: in,
		EntitySnake: separated(entity, "_"), EntityPluralSnake: separated(plural, "_"),
		EntityKebab: separated(entity, "-"), EntityPluralKebab: separated(plural, "-"),
		names: entityReplacer(name, entity, plural, in)}
	if _, err := os.Stat(filepath.Join(projectPath, "tsconfig.json")); err != nil {
		d.jsExt = true
	}
	return d
}

// expand replaces the sample names. In a React project without TypeScript, .tsx and .ts become .jsx and .js.
func (d addData) expand(s string) string { return d.names.Replace(s) }

func (d addData) expandPath(p string) string {
	p = d.expand(p)
	if !d.jsExt {
		return p
	}
	if base, ok := strings.CutSuffix(p, ".tsx"); ok {
		return base + ".jsx"
	}
	if base, ok := strings.CutSuffix(p, ".ts"); ok {
		return base + ".js"
	}
	return p
}

func renderData(name, text string, data addData) (string, error) {
	return renderAny(name, text, data)
}

// ---------------------------------------------------------------- execution

// AddOptions are the add command's choices.
type AddOptions struct {
	Kind, Name, In, Plural string
	Force, DryRun          bool
}

// AddResult is what add reports, including with --dry-run and --json.
type AddResult struct {
	OK      bool     `json:"ok"`
	DryRun  bool     `json:"dryRun"`
	Project string   `json:"project"`
	Target  string   `json:"target"`
	Kind    string   `json:"kind"`
	Name    string   `json:"name"`
	Files   []string `json:"files"`
	Patched []string `json:"patched"`
	Manual  []string `json:"manualSteps,omitempty"`
	Next    []string `json:"next,omitempty"`
}

type plannedPatch struct {
	file, before, skipIf string
	lines                []string
}

func collect(all []*Generator, g *Generator, seen map[string]bool, out *[]*Generator) error {
	if seen[g.Kind] {
		return nil
	}
	seen[g.Kind] = true
	for _, inc := range g.Includes {
		child := FindGenerator(all, inc)
		if child == nil {
			return i18n.Errorf("generator %s includes %s, which does not exist", g.Kind, inc)
		}
		if err := collect(all, child, seen, out); err != nil {
			return err
		}
	}
	*out = append(*out, g)
	return nil
}

// Add generates the files and registers whatever is needed in existing files.
func Add(projectPath string, opts AddOptions) (AddResult, error) {
	target := TargetOf(projectPath)
	result := AddResult{Project: projectPath, Target: target, DryRun: opts.DryRun}
	if target == "" {
		return result, ui.Usagef("there are no generators for this type of project: %s", projectPath)
	}
	all := LoadGenerators(target)
	g := FindGenerator(all, opts.Kind)
	if g == nil {
		return result, ui.Usagef("unknown generator %q for %s (available: %s)", opts.Kind, target, kinds(all))
	}
	entity := Pascal(opts.Name)
	if !identifier.MatchString(entity) {
		return result, ui.Usagef("invalid name %q: use letters and numbers, starting with a letter (e.g. Product)", opts.Name)
	}
	in := Pascal(opts.In)
	var chain []*Generator
	if err := collect(all, g, map[string]bool{}, &chain); err != nil {
		return result, err
	}
	needsIn := slices.ContainsFunc(chain, func(x *Generator) bool { return x.NeedsIn })
	if needsIn && in == "" {
		return result, ui.Usagef("%s needs --in with the aggregate name (e.g. --in Product)", g.Kind)
	}
	result.Kind, result.Name = g.Kind, entity

	data := newAddData(projectPath, entity, Pascal(opts.Plural), in)
	var files []generatedFile
	var patches []plannedPatch
	var requires []requirement
	var after []string
	for _, step := range chain {
		for _, command := range step.After {
			if !slices.Contains(after, command) {
				after = append(after, command)
			}
		}
		rendered, err := renderGenerator(step, data)
		if err != nil {
			return result, err
		}
		files = append(files, rendered...)
		requires = append(requires, step.Requires...)
		for _, p := range step.Patches {
			if planned, ok := planPatch(p, data); ok {
				patches = append(patches, planned)
			}
		}
		if step.Next != "" {
			text, err := renderData("next", data.expand(i18n.T(step.Next)), data)
			if err != nil {
				return result, err
			}
			if text != "" {
				result.Next = append(result.Next, text)
			}
		}
	}
	for _, f := range files {
		result.Files = append(result.Files, f.Path)
	}
	sort.Strings(result.Files)

	// Everything is checked before anything is written.
	for _, r := range requires {
		p := data.expandPath(r.Path)
		if slices.Contains(result.Files, p) {
			continue
		}
		if _, err := os.Stat(filepath.Join(projectPath, filepath.FromSlash(p))); err != nil {
			hint := data.expand(r.Hint)
			return result, ui.Usagef("%s is missing; first run: %s", p, hint)
		}
	}
	if !opts.Force {
		var conflicts []string
		for _, p := range result.Files {
			if _, err := os.Stat(filepath.Join(projectPath, filepath.FromSlash(p))); err == nil {
				conflicts = append(conflicts, p)
			}
		}
		if len(conflicts) > 0 {
			return result, ui.Usagef("these files already exist, use --force to overwrite: %s", strings.Join(conflicts, ", "))
		}
	}
	for _, p := range patches {
		changed, manual, err := applyPatch(projectPath, p, opts.DryRun)
		if err != nil {
			return result, err
		}
		if changed {
			result.Patched = appendUnique(result.Patched, p.file)
		}
		result.Manual = append(result.Manual, manual...)
	}
	if !opts.DryRun {
		for _, f := range files {
			dest := filepath.Join(projectPath, filepath.FromSlash(f.Path))
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return result, err
			}
			if err := os.WriteFile(dest, f.Content, f.mode()); err != nil {
				return result, err
			}
		}
		for _, command := range after {
			if err := sys.Quiet(projectPath, nil, "sh", "-c", command); err != nil {
				ui.Warnf("%s failed after generating: %v", command, err)
			}
		}
	}
	result.OK = true
	return result, nil
}

// containsBlock says whether the patch lines are already in the file, together and in the same order.
// A lone matching line, like a ")", does not count: the block is only skipped once fully inserted.
func containsBlock(lines, block []string) bool {
	if len(block) == 0 {
		return true
	}
	for i := 0; i+len(block) <= len(lines); i++ {
		match := true
		for j, want := range block {
			if strings.TrimSpace(lines[i+j]) != strings.TrimSpace(want) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func appendUnique(list []string, v string) []string {
	if slices.Contains(list, v) {
		return list
	}
	return append(list, v)
}

func kinds(list []*Generator) string {
	var out []string
	for _, g := range list {
		out = append(out, g.Kind)
	}
	return strings.Join(out, ", ")
}

func renderGenerator(g *Generator, data addData) ([]generatedFile, error) {
	if g.files == nil {
		return nil, nil
	}
	var files []generatedFile
	err := walkSource(g.files, ".", func(p string, f sourceFile) error {
		if excluded(g.Conditional, p, data.Services) {
			return nil
		}
		files = append(files, generatedFile{data.expandPath(p), replaceContent(data.names, f.content), f.executable})
		return nil
	})
	return files, err
}

func planPatch(p patch, data addData) (plannedPatch, bool) {
	if p.When != "" && !conditionHolds(p.When, data.Services) {
		return plannedPatch{}, false
	}
	planned := plannedPatch{file: data.expand(p.File), before: p.Before, skipIf: data.expand(p.SkipIf)}
	planned.lines = dedent(strings.Split(data.expand(p.Text), "\n"))
	return planned, true
}

// dedent strips the block's common indentation and the blank lines at its ends: each line keeps its
// relative indent, and the patch adds the marker's indent to it.
func dedent(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	common := ""
	first := true
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if first {
			common, first = indent, false
			continue
		}
		for !strings.HasPrefix(indent, common) {
			common = common[:len(common)-1]
		}
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, strings.TrimRight(strings.TrimPrefix(line, common), " \t"))
	}
	return out
}

func applyPatch(projectPath string, p plannedPatch, dryRun bool) (changed bool, manual []string, err error) {
	dest := filepath.Join(projectPath, filepath.FromSlash(p.file))
	b, readErr := os.ReadFile(dest)
	manualSteps := func() []string {
		var out []string
		for _, line := range p.lines {
			out = append(out, p.file+": "+line)
		}
		return out
	}
	if readErr != nil {
		return false, manualSteps(), nil
	}
	content := string(b)
	if p.skipIf != "" && strings.Contains(content, p.skipIf) {
		return false, nil, nil
	}
	lines := strings.Split(content, "\n")
	if containsBlock(lines, p.lines) {
		return false, nil, nil
	}
	missing := p.lines
	idx := slices.IndexFunc(lines, func(l string) bool { return strings.Contains(l, p.before) })
	if idx < 0 {
		return false, manualSteps(), nil
	}
	if dryRun {
		return true, nil, nil
	}
	indent := lines[idx][:len(lines[idx])-len(strings.TrimLeft(lines[idx], " \t"))]
	var inserted []string
	for _, line := range missing {
		if line == "" {
			inserted = append(inserted, "")
			continue
		}
		inserted = append(inserted, indent+line)
	}
	lines = slices.Insert(lines, idx, inserted...)
	return true, nil, os.WriteFile(dest, []byte(strings.Join(lines, "\n")), 0o644)
}
