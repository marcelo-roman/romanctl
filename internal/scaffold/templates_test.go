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
		t.Fatal("nenhum template em " + Root())
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

func TestTodoTemplateEstaCompleto(t *testing.T) {
	defer i18n.Set(i18n.Lang())
	i18n.Set(i18n.EN)
	for _, tpl := range repoTemplates(t) {
		t.Run(tpl.Key(), func(t *testing.T) {
			if tpl.Description == "" || i18n.T(tpl.Description) == tpl.Description {
				t.Errorf("descrição sem tradução em en.go: %q", tpl.Description)
			}
			if tpl.Run == "" || tpl.Test == "" || tpl.URL == "" {
				t.Errorf("run, test e url são obrigatórios")
			}
			if len(tpl.Requires) == 0 {
				t.Errorf("requires vazio: diga que ferramenta roda o projeto")
			}
			if tpl.Publish.Kind == "" {
				t.Errorf("publish.kind vazio")
			}
			if len(tpl.Arch) == 0 {
				t.Errorf("arch vazio")
			}
			for _, a := range tpl.Arch {
				if !slices.Contains(knownArch, a) {
					t.Errorf("arquitetura fora do vocabulário: %s", a)
				}
			}
			for _, key := range tpl.Requires {
				if tools.Find(key) == nil {
					t.Errorf("requires aponta para ferramenta que não existe: %s", key)
				}
			}
			for _, s := range tpl.Services {
				if services.FindService(s) == nil {
					t.Errorf("serviço desconhecido: %s", s)
				}
			}
			for s := range tpl.Env {
				if !slices.Contains(tpl.Services, s) {
					t.Errorf("env para serviço que o template não oferece: %s", s)
				}
			}
			for _, c := range tpl.Conditional {
				if !slices.Contains(tpl.Services, strings.TrimPrefix(c.When, "!")) {
					t.Errorf("conditional %s depende de serviço que o template não oferece: %s", c.Path, c.When)
				}
				if _, err := fs.Stat(tpl.files, c.Path); err != nil {
					t.Errorf("conditional aponta para caminho que não existe: %s", c.Path)
				}
			}
			for _, o := range tpl.Overlays {
				if !under(o.From, variantsDir) {
					t.Errorf("overlay fora de %s: %s", variantsDir, o.From)
				}
				if !slices.Contains(tpl.Services, strings.TrimPrefix(o.When, "!")) {
					t.Errorf("overlay %s depende de serviço que o template não oferece: %s", o.From, o.When)
				}
				if _, err := fs.Stat(tpl.files, o.From); err != nil {
					t.Errorf("overlay aponta para pasta que não existe: %s", o.From)
				}
			}
		})
	}
}

func TestTodoTemplateGeraComQualquerCombinacaoDeServicos(t *testing.T) {
	for _, tpl := range repoTemplates(t) {
		for _, withs := range subsets(tpl.Services) {
			name := tpl.Key() + "+" + strings.Join(withs, ",")
			t.Run(name, func(t *testing.T) {
				dest := filepath.Join(t.TempDir(), "minha-loja")
				r, err := WriteScaffold(tpl, "minha-loja", dest, withs)
				if err != nil {
					t.Fatal(err)
				}
				for _, f := range r.Files {
					if leftoverSample.MatchString(f) || strings.HasPrefix(f, variantsDir) {
						t.Errorf("nome de exemplo ou variante no caminho: %s", f)
					}
					b, _ := os.ReadFile(filepath.Join(dest, f))
					if m := leftoverSample.Find(b); m != nil {
						t.Errorf("%s ainda tem %q", f, m)
					}
					if strings.Contains(string(b), "{{.") && f != "romanctl.yaml" {
						t.Errorf("%s tem marcador de text/template, que os templates não usam mais", f)
					}
				}
				for _, s := range withs {
					conn, ok := services.FindService(s).Conn("minha_loja")
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

func TestTodaPastaDeTemplateCarrega(t *testing.T) {
	loaded := map[string]bool{}
	for _, tpl := range repoTemplates(t) {
		loaded[tpl.Key()] = true
	}
	dirs, _ := filepath.Glob(filepath.Join(Root(), "projects", "*", "*"))
	for _, dir := range dirs {
		key := filepath.Base(filepath.Dir(dir)) + "/" + filepath.Base(dir)
		if !loaded[key] {
			t.Errorf("%s não carregou: confira o template.yaml", key)
		}
	}
}

func TestGrafiasDoNome(t *testing.T) {
	r := projectReplacer("minha-loja.v2")
	in := "SampleApp sampleApp sample_app sample-app SAMPLE_APP Sample App sampleapp com/example/sampleapp"
	want := "MinhaLojaV2 minhaLojaV2 minha_loja_v2 minha-loja-v2 MINHA_LOJA_V2 Minha Loja V2 minhalojav2 com/example/minhalojav2"
	if got := r.Replace(in); got != want {
		t.Fatalf("obtido %q", got)
	}
	e := entityReplacer("loja", "ProductCategory", "Order")
	in = "SampleItems sampleItems sample_items sample-items sampleitems SampleItem sample-item SampleParents SampleParent SampleApp"
	want = "ProductCategories productCategories product_categories product-categories productcategories ProductCategory product-category Orders Order Loja"
	if got := e.Replace(in); got != want {
		t.Fatalf("obtido %q", got)
	}
}

func TestApelidosDeArquitetura(t *testing.T) {
	got, err := normalizeArch([]string{"ports-and-adapters", "vsa", "modulith", "fsd", "eda"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"event-driven", "feature-sliced", "hexagonal", "modular-monolith", "vertical-slice"}
	if !slices.Equal(got, want) {
		t.Fatalf("obtido %v", got)
	}
	for alias, target := range archAliases {
		if !slices.Contains(knownArch, target) {
			t.Errorf("apelido %s aponta para arquitetura desconhecida %s", alias, target)
		}
	}
}
