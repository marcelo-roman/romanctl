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

type source struct {
	path    string
	pkg     string
	test    bool
	imports []string
}

func sources(t *testing.T) []source {
	t.Helper()
	root := ".."
	var out []source
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, filepath.Dir(path))
		s := source{path: path, pkg: "internal/" + filepath.ToSlash(rel), test: strings.HasSuffix(path, "_test.go")}
		for _, imp := range file.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			s.imports = append(s.imports, p)
		}
		out = append(out, s)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("no source found")
	}
	return out
}

func service(pkg string) string {
	rest, ok := strings.CutPrefix(pkg, "internal/services/")
	if !ok {
		return ""
	}
	return strings.Split(rest, "/")[0]
}

func local(imp string) (string, bool) { return strings.CutPrefix(imp, module+"/") }

func standard(imp string) bool {
	return !strings.Contains(strings.Split(imp, "/")[0], ".") && !strings.HasPrefix(imp, module+"/")
}

func within(pkg, dir string) bool { return pkg == dir || strings.HasPrefix(pkg, dir+"/") }

func TestServicesDoNotImportEachOther(t *testing.T) {
	for _, s := range sources(t) {
		own := service(s.pkg)
		if own == "" {
			continue
		}
		for _, imp := range s.imports {
			inner, ok := local(imp)
			if other := service(inner); ok && other != "" && other != own {
				t.Errorf("%s: %s must not depend on %s", s.path, own, imp)
			}
		}
	}
}

func TestPlatformDoesNotDependOnServices(t *testing.T) {
	for _, s := range sources(t) {
		if !within(s.pkg, "internal/platform") {
			continue
		}
		for _, imp := range s.imports {
			if inner, ok := local(imp); ok && within(inner, "internal/services") {
				t.Errorf("%s: platform must not depend on %s", s.path, imp)
			}
		}
	}
}

func TestDomainImportsNoAdapters(t *testing.T) {
	for _, s := range sources(t) {
		own := service(s.pkg)
		domain := "internal/services/" + own + "/domain"
		if own == "" || s.test || !within(s.pkg, domain) {
			continue
		}
		for _, imp := range s.imports {
			inner, ok := local(imp)
			if standard(imp) || (ok && within(inner, domain)) {
				continue
			}
			t.Errorf("%s: the domain must not depend on %s", s.path, imp)
		}
	}
}

func TestApplicationDependsOnlyOnDomainAndMessagingContracts(t *testing.T) {
	for _, s := range sources(t) {
		own := service(s.pkg)
		base := "internal/services/" + own
		if own == "" || s.test || !within(s.pkg, base+"/application") {
			continue
		}
		for _, imp := range s.imports {
			inner, ok := local(imp)
			if standard(imp) || (ok && (within(inner, base+"/domain") || within(inner, base+"/application") || inner == "internal/platform/messaging")) {
				continue
			}
			t.Errorf("%s: the application layer must not depend on %s", s.path, imp)
		}
	}
}
