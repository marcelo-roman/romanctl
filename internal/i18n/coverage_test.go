package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Saídas que traduzem o primeiro argumento.
var translating = map[string]bool{
	"ui.Okf": true, "ui.Failf": true, "ui.Infof": true, "ui.Warnf": true, "ui.Usagef": true,
	"Okf": true, "Failf": true, "Infof": true, "Warnf": true, "Usagef": true,
	"i18n.T": true, "i18n.Sprintf": true, "i18n.Errorf": true, "i18n.Printf": true,
}

var fmtVerb = regexp.MustCompile(`%[-+# 0]*[0-9.*]*[a-zA-Z%]`)
var word = regexp.MustCompile(`[a-zA-ZÀ-ÿ]{3,}`)

// Texto técnico, igual nos dois idiomas: formatos, comandos e nomes próprios.
var technicalPrefix = regexp.MustCompile(`^(git [a-z]+:|gh repo|Publica o|Atualiza o|tar:|gcc:|gopls:|apt-mark:|go env:|go get:|go mod |npm init:|dotnet new %s|template %s|%s: HTTP|%s install|%s remove|brew install|brew uninstall|npm install -g|npm uninstall -g|use: romanctl (offline|restore|system|config)|romanctl|    romanctl run| \[--with|  --stack|\.NET|arq-|--strip-components|%s/pg_ctl)`)

func isTechnical(v string) bool {
	return technicalPrefix.MatchString(v) || !word.MatchString(fmtVerb.ReplaceAllString(v, ""))
}

var accent = regexp.MustCompile(`[áàâãéêíóôõúçÁÉÍÓÚÃÕÇ]`)

// Código gerado e scripts: o conteúdo não é mensagem do romanctl.
var generatedFiles = []string{"projects/create.go", "tools/custom.go"}

func TestToda_MensagemTemTraducao(t *testing.T) {
	root := ".."
	filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") ||
			strings.Contains(p, "/templates/") || strings.Contains(p, "i18n/") || strings.HasSuffix(p, "cli/help.go") {
			return nil
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		rel := strings.TrimPrefix(p, "../")
		generated := false
		for _, g := range generatedFiles {
			generated = generated || rel == g
		}
		check := map[*ast.BasicLit]bool{}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.ImportSpec:
				return false
			case *ast.CallExpr:
				var name string
				switch fn := x.Fun.(type) {
				case *ast.SelectorExpr:
					if id, ok := fn.X.(*ast.Ident); ok {
						name = id.Name + "." + fn.Sel.Name
					}
				case *ast.Ident:
					name = fn.Name
				}
				if translating[name] && len(x.Args) > 0 {
					if lit, ok := x.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						v, _ := strconv.Unquote(lit.Value)
						check[lit] = true
						if _, ok := en[v]; !ok && !isTechnical(v) {
							t.Errorf("%s:%d: sem tradução: %q", rel, fset.Position(lit.Pos()).Line, v)
						}
					}
				}
			case *ast.BasicLit:
				if x.Kind != token.STRING || generated || check[x] {
					return true
				}
				// Qualquer outro texto com acento precisa estar no catálogo.
				// Uma letra sozinha é dado, como a tabela de acentos do filtro.
				if v, _ := strconv.Unquote(x.Value); accent.MatchString(v) && len([]rune(v)) > 1 {
					if _, ok := en[v]; !ok {
						t.Errorf("%s:%d: texto com acento fora do catálogo: %q", rel, fset.Position(x.Pos()).Line, v)
					}
				}
			}
			return true
		})
		return nil
	})
}
