package up

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"romanctl/internal/manifest"
	"romanctl/internal/state"
)

func TestMergeEnvDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(path, []byte("# mine\nDATABASE_URL=mine\nOTHER=1"), 0o600)
	added, err := MergeEnv(path, []KV{{"DATABASE_URL", "romanctl"}, {"REDIS_URL", "redis://x"}, {"CS", "Host=a;Port=1"}})
	if err != nil || !slices.Equal(added, []string{"REDIS_URL", "CS"}) {
		t.Fatalf("%v %v", added, err)
	}
	b, _ := os.ReadFile(path)
	text := string(b)
	if !strings.Contains(text, "DATABASE_URL=mine\n") || !strings.Contains(text, "# mine") || !strings.Contains(text, "OTHER=1\nREDIS_URL=") {
		t.Fatalf("%s", text)
	}
	if !strings.Contains(text, `CS="Host=a;Port=1"`) {
		t.Fatalf("; needs quotes: %s", text)
	}
	added, _ = MergeEnv(path, []KV{{"REDIS_URL", "other"}})
	if len(added) != 0 {
		t.Fatalf("%v", added)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("%v", info.Mode())
	}
}

func TestProcessEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(path, []byte("A=1\nexport B=\"x;y\"\n#C=3\nD='z'\n"), 0o600)
	got := ProcessEnv(path, func(k string) (string, bool) { return "", k == "D" })
	if !slices.Equal(got, []string{"A=1", "B=x;y"}) {
		t.Fatalf("%v", got)
	}
	if ProcessEnv(path+"x", os.LookupEnv) != nil {
		t.Fatal("no file")
	}
}

func TestEnsureIgnored(t *testing.T) {
	dir := t.TempDir()
	changed, err := EnsureIgnored(dir)
	if err != nil || !changed {
		t.Fatalf("%v %v", changed, err)
	}
	if changed, _ := EnsureIgnored(dir); changed {
		t.Fatal("should not repeat")
	}
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("bin/"), 0o644)
	EnsureIgnored(dir)
	b, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if string(b) != "bin/\n.env\n" {
		t.Fatalf("%q", b)
	}
}

func TestSplitServices(t *testing.T) {
	others := map[string]state.UpEntry{"a": {Uses: []string{"postgres", "redis"}}, "b": {Uses: []string{"postgres"}}}
	stop, handOff := splitServices([]string{"postgres", "redis"}, others, "a")
	if !slices.Equal(stop, []string{"redis"}) || !slices.Equal(handOff, []string{"postgres"}) {
		t.Fatalf("%v %v", stop, handOff)
	}
	stop, handOff = splitServices(nil, others, "a")
	if len(stop)+len(handOff) != 0 {
		t.Fatal("what up did not start is left alone")
	}
}

func TestNamesAndPort(t *testing.T) {
	if DBName("My-Shop.v2") != "my_shop_v2" {
		t.Fatal(DBName("My-Shop.v2"))
	}
	if portOf("http://localhost:5080") != "5080" || portOf("nothing") != "" {
		t.Fatal("port")
	}
}

func TestTemplateVariablesUseServiceConnection(t *testing.T) {
	m := &manifest.Project{Services: []string{"postgres", "redis"}, Env: manifest.Env{
		"postgres": {"ConnectionStrings__Default": "Host={{.Host}};Port={{.Port}};Database={{.Database}};Username={{.User}};Password={{.Password}}"},
	}}
	vars, err := serviceEnv(m, "shop")
	if err != nil {
		t.Fatal(err)
	}
	want := []KV{
		{"DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5432/shop"},
		{"ConnectionStrings__Default", "Host=127.0.0.1;Port=5432;Database=shop;Username=postgres;Password=postgres"},
		{"REDIS_URL", "redis://127.0.0.1:6379"},
	}
	if !slices.Equal(vars, want) {
		t.Fatalf("got %v", vars)
	}
	m.Env["postgres"]["X"] = "{{.Missing}}"
	if _, err := serviceEnv(m, "shop"); err == nil {
		t.Fatal("accepted a field that does not exist")
	}
}
