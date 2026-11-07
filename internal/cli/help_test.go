package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"romanctl/internal/i18n"
)

// Todo caso do dispatch precisa ter ajuda: comando novo sem --help falha aqui.
func TestTodoComandoTemAjuda(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "root.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "dispatch" {
			return true
		}
		ast.Inspect(fn, func(m ast.Node) bool {
			c, ok := m.(*ast.CaseClause)
			if !ok {
				return true
			}
			for _, e := range c.List {
				if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					word, _ := strconv.Unquote(lit.Value)
					found++
					if findCommand([]string{word}) == nil {
						t.Errorf("comando %q sem ajuda", word)
					}
				}
			}
			return true
		})
		return false
	})
	if found < 20 {
		t.Fatalf("só %d comandos encontrados", found)
	}
}

func TestSubcomandosTemAjuda(t *testing.T) {
	for _, words := range [][]string{
		{"tools", "list"}, {"tools", "ls"}, {"tools", "install"}, {"tools", "uninstall"}, {"tools", "rm"}, {"tools", "categories"},
		{"system", "status"}, {"system", "battery"}, {"system", "watch"},
		{"offline", "status"}, {"offline", "prep"}, {"offline", "test"},
		{"restore", "status"}, {"restore", "save"}, {"restore", "clear"}, {"config", "lang"},
	} {
		c := findCommand(words)
		if c == nil || !strings.Contains(c.path, " ") {
			t.Errorf("%v caiu em %+v", words, c)
		}
	}
	if c := findCommand([]string{"start", "postgres"}); c == nil || c.path != "start" {
		t.Errorf("o nome do serviço não deveria mudar a ajuda: %+v", c)
	}
	if findCommand([]string{"inexistente"}) != nil {
		t.Error("comando inexistente")
	}
}

var accent = regexp.MustCompile(`[áàâãéêíóôõúç]`)

func TestAjudaCompletaNosDoisIdiomas(t *testing.T) {
	check := func(c, what string, x text) {
		if x.pt == "" || x.en == "" {
			t.Errorf("%s: %s sem tradução: %+v", c, what, x)
		}
		if accent.MatchString(x.en) {
			t.Errorf("%s: %s com português no inglês: %q", c, what, x.en)
		}
	}
	seen := map[string]bool{}
	for _, c := range commands {
		if seen[c.path] {
			t.Errorf("%s repetido", c.path)
		}
		seen[c.path] = true
		check(c.path, "resumo", c.summary)
		if c.details != (text{}) {
			check(c.path, "detalhes", c.details)
		}
		if !strings.HasPrefix(c.usage, "romanctl "+strings.Fields(c.path)[0]) {
			t.Errorf("%s: uso errado: %s", c.path, c.usage)
		}
		for _, o := range c.options {
			check(c.path, o.flag, o.desc)
		}
		for _, e := range c.examples {
			if !strings.HasPrefix(e, "romanctl "+strings.Fields(c.path)[0]) {
				t.Errorf("%s: exemplo errado: %s", c.path, e)
			}
		}
	}
	for _, g := range usageGroups {
		for _, p := range g.paths {
			if findCommand([]string{p}) == nil {
				t.Errorf("ajuda geral aponta para %s", p)
			}
		}
	}
	covered := map[string]bool{}
	for _, g := range usageGroups {
		for _, p := range g.paths {
			covered[p] = true
		}
	}
	for _, c := range commands {
		if !strings.Contains(c.path, " ") && !covered[c.path] {
			t.Errorf("%s fora da ajuda geral", c.path)
		}
	}
}

func TestUsoEmInglesSemPortugues(t *testing.T) {
	defer i18n.Set(i18n.Lang())
	i18n.Set(i18n.EN)
	pt := regexp.MustCompile(`[áàâãéêíóôõúç]|\b(nome|nomes|tipo|projeto|pasta|grupos|agregado|texto|comando)\b`)
	for _, c := range commands {
		if got := localized(c.usage); pt.MatchString(got) {
			t.Errorf("%s: %s", c.path, got)
		}
		for _, o := range c.options {
			if got := localized(o.flag); pt.MatchString(got) {
				t.Errorf("%s: %s", c.path, got)
			}
		}
	}
}

func TestTextoSegueOIdioma(t *testing.T) {
	defer i18n.Set(i18n.Lang())
	x := text{"sim", "yes"}
	i18n.Set(i18n.EN)
	if x.String() != "yes" {
		t.Fatal(x.String())
	}
	i18n.Set(i18n.PT)
	if x.String() != "sim" {
		t.Fatal(x.String())
	}
}
