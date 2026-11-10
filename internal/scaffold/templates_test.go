package scaffold

import (
	"io/fs"
	"reflect"
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

var leftoverMarker = regexp.MustCompile(`__[A-Z][A-Za-z]*__`)

func embeddedOnly(t *testing.T) []*Template {
	t.Helper()
	home := config.Home
	config.Home = t.TempDir()
	t.Cleanup(func() { config.Home = home })
	list := LoadTemplates()
	if len(list) == 0 {
		t.Fatal("nenhum template embutido")
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
	for _, tpl := range embeddedOnly(t) {
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
				name := strings.TrimPrefix(c.When, "!")
				if !slices.Contains(tpl.Services, name) {
					t.Errorf("conditional %s depende de serviço que o template não oferece: %s", c.Path, name)
				}
				if _, err := fs.Stat(tpl.files, c.Path); err != nil {
					t.Errorf("conditional aponta para caminho que não existe: %s", c.Path)
				}
			}
		})
	}
}

func TestTodoTemplateGeraComQualquerCombinacaoDeServicos(t *testing.T) {
	for _, tpl := range embeddedOnly(t) {
		for _, withs := range subsets(tpl.Services) {
			name := tpl.Key() + "+" + strings.Join(withs, ",")
			t.Run(name, func(t *testing.T) {
				r, err := WriteScaffold(tpl, "minha-loja", "", withs)
				if err != nil {
					t.Fatal(err)
				}
				for _, f := range r.Files {
					if m := leftoverMarker.FindString(f); m != "" {
						t.Errorf("marcador %s sobrou em %s", m, f)
					}
				}
				if strings.Contains(r.Manifest.Run, "{{") || strings.Contains(r.Manifest.Test, "{{") {
					t.Errorf("run ou test sem renderizar: %q %q", r.Manifest.Run, r.Manifest.Test)
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

func TestDadosDoNome(t *testing.T) {
	d := newTemplateData("minha-loja.v2", nil)
	want := templateData{Name: "MinhaLojaV2", Lower: "minha-loja.v2", Snake: "minha_loja_v2", Kebab: "minha-loja-v2",
		Camel: "minhaLojaV2", DbName: "minha_loja_v2", Module: "minha-loja.v2", Package: "com.example.minhalojav2",
		PackagePath: "com/example/minhalojav2"}
	d.Services = nil
	if !reflect.DeepEqual(d, want) {
		t.Fatalf("obtido %+v", d)
	}
	if got := d.pathReplacer().Replace("src/main/java/__PackagePath__/__Name__App.java"); got != "src/main/java/com/example/minhalojav2/MinhaLojaV2App.java" {
		t.Fatalf("caminho: %s", got)
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
