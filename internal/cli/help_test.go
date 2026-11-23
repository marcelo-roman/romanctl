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

// Every dispatch case needs help: a new command without --help fails here.
func TestEveryCommandHasHelp(t *testing.T) {
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
						t.Errorf("command %q has no help", word)
					}
				}
			}
			return true
		})
		return false
	})
	if found < 20 {
		t.Fatalf("only %d commands found", found)
	}
}

func TestSubcommandsHaveHelp(t *testing.T) {
	for _, words := range [][]string{
		{"tools", "list"}, {"tools", "ls"}, {"tools", "install"}, {"tools", "uninstall"}, {"tools", "rm"}, {"tools", "categories"},
		{"system", "status"}, {"system", "battery"}, {"system", "watch"},
		{"offline", "status"}, {"offline", "prep"}, {"offline", "test"},
		{"restore", "status"}, {"restore", "save"}, {"restore", "clear"}, {"config", "lang"},
	} {
		c := findCommand(words)
		if c == nil || !strings.Contains(c.path, " ") {
			t.Errorf("%v resolved to %+v", words, c)
		}
	}
	if c := findCommand([]string{"start", "postgres"}); c == nil || c.path != "start" {
		t.Errorf("the service name should not change the help: %+v", c)
	}
	if findCommand([]string{"nonexistent"}) != nil {
		t.Error("nonexistent command")
	}
}

var accent = regexp.MustCompile(`[\x{c0}-\x{d6}\x{d8}-\x{f6}\x{f8}-\x{ff}]`)

func TestHelpCompleteInBothLanguages(t *testing.T) {
	check := func(c, what string, x text) {
		if x.pt == "" || x.en == "" {
			t.Errorf("%s: %s missing a translation: %+v", c, what, x)
		}
		if accent.MatchString(x.en) {
			t.Errorf("%s: %s has Portuguese in the English text: %q", c, what, x.en)
		}
	}
	seen := map[string]bool{}
	for _, c := range commands {
		if seen[c.path] {
			t.Errorf("%s repeated", c.path)
		}
		seen[c.path] = true
		check(c.path, "summary", c.summary)
		if c.details != (text{}) {
			check(c.path, "details", c.details)
		}
		if !strings.HasPrefix(c.usage, "romanctl "+strings.Fields(c.path)[0]) {
			t.Errorf("%s: wrong usage: %s", c.path, c.usage)
		}
		for _, o := range c.options {
			check(c.path, o.flag, o.desc)
		}
		for _, e := range c.examples {
			if !strings.HasPrefix(e, "romanctl "+strings.Fields(c.path)[0]) {
				t.Errorf("%s: wrong example: %s", c.path, e)
			}
		}
	}
	for _, g := range usageGroups {
		for _, p := range g.paths {
			if findCommand([]string{p}) == nil {
				t.Errorf("general help points to %s", p)
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
			t.Errorf("%s missing from the general help", c.path)
		}
	}
}

func TestUsageInEnglishHasNoPortuguese(t *testing.T) {
	defer i18n.Set(i18n.Lang())
	i18n.Set(i18n.EN)
	pt := regexp.MustCompile(`[\x{c0}-\x{d6}\x{d8}-\x{f6}\x{f8}-\x{ff}]|\b(nome|nomes|tipo|projeto|pasta|grupos|agregado|texto|comando)\b`)
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

func TestUsageInPortugueseTranslatesOnlyPlaceholders(t *testing.T) {
	defer i18n.Set(i18n.Lang())
	i18n.Set(i18n.PT)
	cases := map[string]string{
		"romanctl start [service|all]":                  "romanctl start [servi\u00e7o|all]",
		"romanctl add [type] [name] [--in <aggregate>]": "romanctl add [tipo] [nome] [--in <agregado>]",
		"romanctl iam users <realm> [--search X]":       "romanctl iam users <realm> [--search X]",
		"romanctl down [project] [--keep-services]":     "romanctl down [projeto] [--keep-services]",
	}
	for in, want := range cases {
		if got := localized(in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}

func TestTextFollowsLanguage(t *testing.T) {
	defer i18n.Set(i18n.Lang())
	x := text{"yes", "sim"}
	i18n.Set(i18n.EN)
	if x.String() != "yes" {
		t.Fatal(x.String())
	}
	i18n.Set(i18n.PT)
	if x.String() != "sim" {
		t.Fatal(x.String())
	}
}

func TestCommandListHasAllInSections(t *testing.T) {
	opts := commandOptions()
	if len(opts) != len(commands) {
		var missing []string
		for _, c := range commands {
			found := false
			for _, o := range opts {
				found = found || o.Value == c.path
			}
			if !found {
				missing = append(missing, c.path)
			}
		}
		t.Fatalf("%d of %d commands; missing %v", len(opts), len(commands), missing)
	}
	for _, o := range opts {
		if o.Section == "" {
			t.Errorf("%s has no section", o.Value)
		}
	}
}
