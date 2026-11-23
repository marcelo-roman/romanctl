package scaffold

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"romanctl/internal/i18n"
	"romanctl/internal/ui"
)

func TestNames(t *testing.T) {
	for in, want := range map[string]string{"product-item": "ProductItem", "product_item": "ProductItem", "ProductItem": "ProductItem", "order": "Order"} {
		if got := Pascal(in); got != want {
			t.Errorf("%s: %s", in, got)
		}
	}
	for in, want := range map[string]string{"Product": "Products", "Category": "Categories", "Day": "Days", "Address": "Addresses", "Box": "Boxes", "Batch": "Batches"} {
		if got := Plural(in); got != want {
			t.Errorf("%s: %s", in, got)
		}
	}
	if separated("ProductItem", "-") != "product-item" || separated("ProductItems", "_") != "product_items" {
		t.Error("wrong separation")
	}
}

func dotnetProject(t *testing.T, services []string) string {
	t.Helper()
	dest := filepath.Join(t.TempDir(), "shop")
	if _, err := WriteScaffold(dotnetTemplate(t), "shop", dest, services); err != nil {
		t.Fatal(err)
	}
	return dest
}

func pt(text string) string {
	defer i18n.Set(i18n.Lang())
	i18n.Set(i18n.PT)
	return i18n.T(text)
}

func TestEveryGeneratorRenders(t *testing.T) {
	for _, target := range AllTargets() {
		list := LoadGenerators(target)
		if len(list) == 0 {
			t.Errorf("%s has no generators", target)
		}
		for _, g := range list {
			if g.Description == "" {
				t.Errorf("%s/%s has no description", target, g.Kind)
			}
			for _, text := range []string{g.Description, g.Next} {
				if text != "" && pt(text) == text {
					t.Errorf("%s/%s missing from pt.go: %q", target, g.Kind, text)
				}
			}
			for _, services := range [][]string{nil, {"postgres"}} {
				d := newAddData(filepath.Join(t.TempDir(), "shop"), "ProductItem", "", "Order")
				for _, s := range services {
					d.Services[s] = true
				}
				d.Postgres = len(services) > 0
				files, err := renderGenerator(g, d)
				if err != nil {
					t.Errorf("%s/%s: %v", target, g.Kind, err)
				}
				for _, f := range files {
					if leftoverSample.MatchString(f.Path) || leftoverSample.Match(f.Content) || strings.Contains(string(f.Content), "{{.") {
						t.Errorf("%s/%s: sample name or marker not replaced in %s", target, g.Kind, f.Path)
					}
				}
				for _, p := range g.Patches {
					planned, ok := planPatch(p, d)
					if ok && (leftoverSample.MatchString(planned.file) || leftoverSample.MatchString(strings.Join(planned.lines, "\n"))) {
						t.Errorf("%s/%s: sample name not replaced in the patch of %s", target, g.Kind, planned.file)
					}
				}
			}
		}
	}
}

func TestResourceGeneratesAndRegisters(t *testing.T) {
	for _, variant := range []struct {
		services []string
		want     string
		other    string
	}{{nil, "InMemory/InMemoryProductRepository.cs", "EfCore"}, {[]string{"postgres"}, "EfCore/ProductRepository.cs", "InMemory/InMemoryProduct"}} {
		dir := dotnetProject(t, variant.services)
		r, err := Add(dir, AddOptions{Kind: "crud", Name: "product"})
		if err != nil {
			t.Fatal(err)
		}
		if r.Kind != "resource" || r.Name != "Product" || !hasPath(r.Files, variant.want) || hasPath(r.Files, variant.other+"/Product") {
			t.Fatalf("%+v", r)
		}
		di, _ := os.ReadFile(filepath.Join(dir, "src/Shop.Infrastructure/DependencyInjection.cs"))
		if !strings.Contains(string(di), "IProductRepository") || !strings.Contains(string(di), "        services.AddScoped<Shop.Domain.Products.IProductRepository") {
			t.Fatalf("registration without indentation:\n%s", di)
		}
		program, _ := os.ReadFile(filepath.Join(dir, "src/Shop.Api/Program.cs"))
		if !strings.Contains(string(program), "app.MapProductEndpoints();\n// romanctl:endpoints") {
			t.Fatalf("Program.cs:\n%s", program)
		}
		if len(r.Manual) != 0 {
			t.Fatalf("manual: %v", r.Manual)
		}
		// Repeating is rejected and leaves the registrations unchanged.
		if _, err := Add(dir, AddOptions{Kind: "resource", Name: "Product"}); !ui.IsUsage(err) || !strings.Contains(err.Error(), "--force") {
			t.Fatalf("error: %v", err)
		}
		if _, err := Add(dir, AddOptions{Kind: "resource", Name: "Product", Force: true}); err != nil {
			t.Fatal(err)
		}
		again, _ := os.ReadFile(filepath.Join(dir, "src/Shop.Api/Program.cs"))
		if strings.Count(string(again), "MapProductEndpoints") != 1 {
			t.Fatalf("registered twice:\n%s", again)
		}
	}
}

func TestMissingDependencies(t *testing.T) {
	dir := dotnetProject(t, nil)
	for _, c := range []AddOptions{{Kind: "controller", Name: "Product"}, {Kind: "command", Name: "ArchiveProduct", In: "Product"}, {Kind: "repository", Name: "Product"}} {
		_, err := Add(dir, c)
		if !ui.IsUsage(err) || !strings.Contains(err.Error(), "romanctl add") {
			t.Errorf("%s: %v", c.Kind, err)
		}
	}
	if entries, _ := os.ReadDir(filepath.Join(dir, "src/Shop.Api")); slices.ContainsFunc(entries, func(e os.DirEntry) bool { return strings.Contains(e.Name(), "Product") }) {
		t.Fatal("wrote files even with a missing dependency")
	}
}

func TestCommandNeedsAggregate(t *testing.T) {
	dir := dotnetProject(t, nil)
	if _, err := Add(dir, AddOptions{Kind: "command", Name: "Archive"}); !ui.IsUsage(err) || !strings.Contains(err.Error(), "--in") {
		t.Fatalf("error: %v", err)
	}
	if _, err := Add(dir, AddOptions{Kind: "resource", Name: "Product"}); err != nil {
		t.Fatal(err)
	}
	r, err := Add(dir, AddOptions{Kind: "command", Name: "ArchiveProduct", In: "Product"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(r.Files, "src/Shop.Application/Products/Commands/ArchiveProduct.cs") {
		t.Fatalf("%v", r.Files)
	}
	if _, err := Add(dir, AddOptions{Kind: "value-object", Name: "Sku", In: "Product"}); err != nil {
		t.Fatal(err)
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	dir := dotnetProject(t, nil)
	before, _ := os.ReadFile(filepath.Join(dir, "src/Shop.Api/Program.cs"))
	r, err := Add(dir, AddOptions{Kind: "resource", Name: "Product", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !r.DryRun || len(r.Files) == 0 || !slices.Contains(r.Patched, "src/Shop.Api/Program.cs") {
		t.Fatalf("%+v", r)
	}
	after, _ := os.ReadFile(filepath.Join(dir, "src/Shop.Api/Program.cs"))
	if string(before) != string(after) {
		t.Fatal("changed Program.cs")
	}
	if _, err := os.Stat(filepath.Join(dir, "src/Shop.Domain/Products")); err == nil {
		t.Fatal("created files")
	}
}

func TestWithoutMarkerAsksForManualAction(t *testing.T) {
	dir := dotnetProject(t, nil)
	program := filepath.Join(dir, "src/Shop.Api/Program.cs")
	b, _ := os.ReadFile(program)
	os.WriteFile(program, []byte(strings.ReplaceAll(string(b), "// romanctl:endpoints\n", "")), 0o644)
	r, err := Add(dir, AddOptions{Kind: "resource", Name: "Product"})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Manual) != 1 || !strings.Contains(r.Manual[0], "app.MapProductEndpoints();") {
		t.Fatalf("%+v", r.Manual)
	}
}

func TestInvalidNameAndWrongKind(t *testing.T) {
	dir := dotnetProject(t, nil)
	if _, err := Add(dir, AddOptions{Kind: "resource", Name: "1abc"}); !ui.IsUsage(err) {
		t.Fatalf("%v", err)
	}
	if _, err := Add(dir, AddOptions{Kind: "page", Name: "X"}); !ui.IsUsage(err) || !strings.Contains(err.Error(), "resource") {
		t.Fatalf("%v", err)
	}
}

func TestSimpleProjects(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"express":"1"}}`), 0o644)
	os.WriteFile(filepath.Join(dir, "index.js"), []byte("const app = express();\n\nconst PORT = 3000;\n"), 0o644)
	r, err := Add(dir, AddOptions{Kind: "controller", Name: "Product"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Target != "node/express" || !slices.Contains(r.Files, "routes/products.js") {
		t.Fatalf("%+v", r)
	}
	index, _ := os.ReadFile(filepath.Join(dir, "index.js"))
	if !strings.Contains(string(index), "app.use('/products', require('./routes/products'));\nconst PORT") {
		t.Fatalf("%s", index)
	}

	react := t.TempDir()
	os.WriteFile(filepath.Join(react, "package.json"), []byte(`{"devDependencies":{"vite":"1"}}`), 0o644)
	r, err = Add(react, AddOptions{Kind: "component", Name: "user-card"})
	if err != nil || !slices.Contains(r.Files, "src/components/UserCard.jsx") {
		t.Fatalf("%+v %v", r, err)
	}
	os.WriteFile(filepath.Join(react, "tsconfig.json"), []byte("{}"), 0o644)
	r, err = Add(react, AddOptions{Kind: "hook", Name: "cart"})
	if err != nil || !slices.Contains(r.Files, "src/hooks/useCart.ts") {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestPatchSkipsOnlyWholeBlock(t *testing.T) {
	file := []string{"routes = [", "    orders,", ")", "    # romanctl:routes", "]"}
	if containsBlock(file, []string{"products,", ")"}) {
		t.Fatal("a lone ) should not count as the block already inserted")
	}
	if !containsBlock(file, []string{"orders,", ")"}) {
		t.Fatal("the block exists, in order")
	}
}

func TestPatchKeepsRelativeIndent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.yaml")
	os.WriteFile(file, []byte("services:\n  app:\n    image: x\n  # romanctl:services\n"), 0o644)
	planned := plannedPatch{file: "config.yaml", before: "# romanctl:services",
		lines: dedent(strings.Split("\n    worker:\n      image: y\n      env:\n        - A=1\n", "\n"))}
	if _, _, err := applyPatch(dir, planned, false); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(file)
	want := "services:\n  app:\n    image: x\n  worker:\n    image: y\n    env:\n      - A=1\n  # romanctl:services\n"
	if string(b) != want {
		t.Fatalf("got:\n%s", b)
	}
	if changed, _, _ := applyPatch(dir, planned, false); changed {
		t.Fatal("the already inserted block was inserted again")
	}
}

func TestGivenPluralBeatsTheRule(t *testing.T) {
	d := newAddData(filepath.Join(t.TempDir(), "shop"), "Person", "People", "")
	if got := d.expand("SampleItems sample_items sample-items SampleItem"); got != "People people people Person" {
		t.Fatalf("got %q", got)
	}
	if got := newAddData(filepath.Join(t.TempDir(), "shop"), "Person", "", "").expand("SampleItems"); got != "Persons" {
		t.Fatalf("without --plural, the rule applies: %q", got)
	}
}
