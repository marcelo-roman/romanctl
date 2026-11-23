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

// Outputs that translate their first argument.
var translating = map[string]bool{
	"ui.Okf": true, "ui.Failf": true, "ui.Infof": true, "ui.Warnf": true, "ui.Usagef": true,
	"Okf": true, "Failf": true, "Infof": true, "Warnf": true, "Usagef": true,
	"i18n.T": true, "i18n.Sprintf": true, "i18n.Errorf": true, "i18n.Printf": true,
}

var fmtVerb = regexp.MustCompile(`%[-+# 0]*[0-9.*]*[a-zA-Z%]`)
var word = regexp.MustCompile(`[a-zA-Z\x{c0}-\x{ff}]{3,}`)

// Technical text, the same in both languages: formats, commands and proper names.
var technicalPrefix = regexp.MustCompile(`^(git [a-z]+:|gh repo|tar:|gcc:|gopls:|apt-mark:|go env:|go get:|go mod |npm init:|dotnet new %s|template %s|%s: HTTP|%s install|%s remove|brew install|brew uninstall|npm install -g|npm uninstall -g|(use|usage): romanctl|romanctl|    romanctl run| \[--with|  --stack|\.NET|arq-|--strip-components|%s/pg_ctl)`)

// Same text in both languages.
var sameInBoth = map[string]bool{
	"Realm:   %s\n": true, "API:     %s\n": true, "  (romanctl claude install)": true, "%s (realm %s)\n": true,
	"client_id:      %s\nclient_secret:  %s\ntoken endpoint: %s\n":                     true,
	"client_id:      %s\nclient_secret:  %s\ntoken endpoint: %s\napi:            %s\n": true,
	"Templates": true, "%s %s, via %s": true, "Log": true, "logs": true,
}

func isTechnical(v string) bool {
	return sameInBoth[v] || technicalPrefix.MatchString(v) || !word.MatchString(fmtVerb.ReplaceAllString(v, ""))
}

var accent = regexp.MustCompile(`[\x{c0}-\x{d6}\x{d8}-\x{f6}\x{f8}-\x{ff}]`)

// Generated code and scripts: their content is not a romanctl message.
var generatedFiles = []string{"projects/create.go", "tools/custom.go"}

func TestEveryMessageHasTranslation(t *testing.T) {
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
						if _, ok := pt[v]; !ok && !isTechnical(v) {
							t.Errorf("%s:%d: missing translation: %q", rel, fset.Position(lit.Pos()).Line, v)
						}
					}
				}
			case *ast.BasicLit:
				if x.Kind != token.STRING || generated || check[x] {
					return true
				}
				if v, _ := strconv.Unquote(x.Value); accent.MatchString(v) && len([]rune(v)) > 1 {
					t.Errorf("%s:%d: accented text outside the catalog: %q", rel, fset.Position(x.Pos()).Line, v)
				}
			}
			return true
		})
		return nil
	})
}
