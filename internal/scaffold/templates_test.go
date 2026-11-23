package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"text/template"

	"gopkg.in/yaml.v3"

	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/services"
	"romanctl/internal/tools"
)

var leftoverSample = regexp.MustCompile(`(?i)sample[-_ ]?(app|items?|parents?)`)

func repoTemplates(t *testing.T) []*Template {
	t.Helper()
	home := config.Home
	config.Home = t.TempDir()
	t.Cleanup(func() { config.Home = home })
	list := LoadTemplates()
	if len(list) == 0 {
		t.Fatal("no templates in " + Root())
	}
	return list
}

func subsets(items []string) [][]string {
	out := [][]string{nil}
	for _, item := range items {
		for _, prev := range out {
			out = append(out, append(slices.Clone(prev), item))
		}
	}
	return out
}

func TestEveryTemplateIsComplete(t *testing.T) {
	defer i18n.Set(i18n.Lang())
	i18n.Set(i18n.PT)
	for _, tpl := range repoTemplates(t) {
		t.Run(tpl.Key(), func(t *testing.T) {
			if tpl.Description == "" || i18n.T(tpl.Description) == tpl.Description {
				t.Errorf("description missing from pt.go: %q", tpl.Description)
			}
			if (tpl.Run == "" && len(tpl.Processes) == 0) || tpl.Test == "" || tpl.URL == "" {
				t.Errorf("run (or processes), test and url are required")
			}
			if len(tpl.Requires) == 0 {
				t.Errorf("empty requires: say which tool runs the project")
			}
			if tpl.Publish.Kind == "" {
				t.Errorf("empty publish.kind")
			}
			if len(tpl.Arch) == 0 {
				t.Errorf("empty arch")
			}
			for _, a := range tpl.Arch {
				if !slices.Contains(knownArch, a) {
					t.Errorf("architecture outside the vocabulary: %s", a)
				}
			}
			for _, key := range tpl.Requires {
				if tools.Find(key) == nil {
					t.Errorf("requires points to a tool that does not exist: %s", key)
				}
			}
			for _, s := range tpl.Services {
				if services.FindService(s) == nil {
					t.Errorf("unknown service: %s", s)
				}
			}
			for s := range tpl.Env {
				if !slices.Contains(tpl.Services, s) {
					t.Errorf("env for a service the template does not offer: %s", s)
				}
			}
			for _, c := range tpl.Conditional {
				if !slices.Contains(tpl.Services, strings.TrimPrefix(c.When, "!")) {
					t.Errorf("conditional %s depends on a service the template does not offer: %s", c.Path, c.When)
				}
				if _, err := fs.Stat(tpl.files, c.Path); err != nil {
					t.Errorf("conditional points to a path that does not exist: %s", c.Path)
				}
			}
			for _, step := range tpl.Setup {
				if step.Run == "" {
					t.Errorf("setup step without run")
				}
				if step.When != "" && !slices.Contains(tpl.Services, strings.TrimPrefix(step.When, "!")) {
					t.Errorf("setup %q depends on a service the template does not offer: %s", step.Run, step.When)
				}
			}
			for _, p := range tpl.Processes {
				if p.Name == "" || p.Run == "" {
					t.Errorf("process without name or run: %+v", p)
				}
				if p.Dir != "" {
					if _, err := fs.Stat(tpl.files, p.Dir); err != nil {
						t.Errorf("process %s points to a directory that does not exist: %s", p.Name, p.Dir)
					}
				}
			}
			for _, o := range tpl.Overlays {
				if !under(o.From, variantsDir) {
					t.Errorf("overlay outside %s: %s", variantsDir, o.From)
				}
				if !slices.Contains(tpl.Services, strings.TrimPrefix(o.When, "!")) {
					t.Errorf("overlay %s depends on a service the template does not offer: %s", o.From, o.When)
				}
				if _, err := fs.Stat(tpl.files, o.From); err != nil {
					t.Errorf("overlay points to a directory that does not exist: %s", o.From)
				}
			}
		})
	}
}

func TestEveryTemplateGeneratesWithAnyServiceCombination(t *testing.T) {
	for _, tpl := range repoTemplates(t) {
		for _, withs := range subsets(tpl.Services) {
			name := tpl.Key() + "+" + strings.Join(withs, ",")
			t.Run(name, func(t *testing.T) {
				dest := filepath.Join(t.TempDir(), "my-shop")
				r, err := WriteScaffold(tpl, "my-shop", dest, withs)
				if err != nil {
					t.Fatal(err)
				}
				for _, f := range r.Files {
					if leftoverSample.MatchString(f) || strings.HasPrefix(f, variantsDir) {
						t.Errorf("sample name or variant in the path: %s", f)
					}
					b, _ := os.ReadFile(filepath.Join(dest, f))
					if m := leftoverSample.Find(b); m != nil {
						t.Errorf("%s still has %q", f, m)
					}
					if strings.Contains(string(b), "{{.") && f != "romanctl.yaml" {
						t.Errorf("%s has a text/template marker, which templates no longer use", f)
					}
				}
				for _, s := range withs {
					conn, ok := services.FindService(s).Conn("my_shop")
					if !ok {
						continue
					}
					for key, value := range r.Manifest.Env[s] {
						tp, err := template.New(key).Option("missingkey=error").Parse(value)
						if err == nil {
							err = tp.Execute(&strings.Builder{}, conn)
						}
						if err != nil {
							t.Errorf("env %s.%s: %v", s, key, err)
						}
					}
				}
			})
		}
	}
}

func TestEveryTemplateDirectoryLoads(t *testing.T) {
	loaded := map[string]bool{}
	for _, tpl := range repoTemplates(t) {
		loaded[tpl.Key()] = true
	}
	dirs, _ := filepath.Glob(filepath.Join(Root(), "*", "projects", "*"))
	for _, dir := range dirs {
		key := filepath.Base(filepath.Dir(filepath.Dir(dir))) + "/" + filepath.Base(dir)
		if !loaded[key] {
			t.Errorf("%s did not load: check its template.yaml", key)
		}
	}
}

func TestNameSpellings(t *testing.T) {
	r := projectReplacer("my-shop.v2")
	in := "SampleApp sampleApp sample_app sample-app SAMPLE_APP Sample App sample app sampleapp com/example/sampleapp"
	want := "MyShopV2 myShopV2 my_shop_v2 my-shop-v2 MY_SHOP_V2 My Shop V2 my shop v2 myshopv2 com/example/myshopv2"
	if got := r.Replace(in); got != want {
		t.Fatalf("got %q", got)
	}
	e := entityReplacer("shop", "ProductCategory", "ProductCategories", "Order")
	in = "SampleItems sampleItems sample_items sample-items sampleitems sample items SampleItem sample-item sample item SampleParents SampleParent SampleApp"
	want = "ProductCategories productCategories product_categories product-categories productcategories product categories ProductCategory product-category product category Orders Order Shop"
	if got := e.Replace(in); got != want {
		t.Fatalf("got %q", got)
	}
}

func TestArchitectureAliases(t *testing.T) {
	got, err := normalizeArch([]string{"ports-and-adapters", "vsa", "modulith", "fsd", "eda"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"event-driven", "feature-sliced", "hexagonal", "modular-monolith", "vertical-slice"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v", got)
	}
	for alias, target := range archAliases {
		if !slices.Contains(knownArch, target) {
			t.Errorf("alias %s points to unknown architecture %s", alias, target)
		}
	}
}

func TestSetupAcceptsStringOrStepWithWhen(t *testing.T) {
	var tpl Template
	if err := yaml.Unmarshal([]byte("setup:\n  - go mod tidy\n  - run: npm ci\n    when: postgres\n  - run: npm install\n    when: \"!postgres\"\n"), &tpl); err != nil {
		t.Fatal(err)
	}
	if got := tpl.setupFor(map[string]bool{}); !slices.Equal(got, []string{"go mod tidy", "npm install"}) {
		t.Fatalf("without postgres: %v", got)
	}
	if got := tpl.setupFor(map[string]bool{"postgres": true}); !slices.Equal(got, []string{"go mod tidy", "npm ci"}) {
		t.Fatalf("com postgres: %v", got)
	}
}
