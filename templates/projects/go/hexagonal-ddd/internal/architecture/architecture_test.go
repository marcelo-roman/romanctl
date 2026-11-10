package architecture

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const module = "sample-app"

var allowed = map[string][]string{
	"domain":      {"internal/domain"},
	"application": {"internal/domain", "internal/application"},
	"adapters":    {"internal/domain", "internal/application", "internal/adapters"},
}

func TestDependenciesPointInward(t *testing.T) {
	root := ".."
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		layer := strings.Split(filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator))), "/")[0]
		rules, ok := allowed[layer]
		if !ok {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			inner, internal := strings.CutPrefix(p, module+"/")
			if !internal {
				if layer == "domain" && strings.Contains(strings.Split(p, "/")[0], ".") {
					t.Errorf("%s: the domain must not depend on %s", path, p)
				}
				continue
			}
			if !permitted(inner, rules) {
				t.Errorf("%s: %s must not depend on %s", path, layer, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func permitted(pkg string, rules []string) bool {
	for _, r := range rules {
		if pkg == r || strings.HasPrefix(pkg, r+"/") {
			return true
		}
	}
	return false
}
