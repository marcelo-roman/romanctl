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

func TestMergeEnvNaoSobrescreve(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(path, []byte("# meu\nDATABASE_URL=minha\nOUTRA=1"), 0o600)
	added, err := MergeEnv(path, []KV{{"DATABASE_URL", "romanctl"}, {"REDIS_URL", "redis://x"}, {"CS", "Host=a;Port=1"}})
	if err != nil || !slices.Equal(added, []string{"REDIS_URL", "CS"}) {
		t.Fatalf("%v %v", added, err)
	}
	b, _ := os.ReadFile(path)
	text := string(b)
	if !strings.Contains(text, "DATABASE_URL=minha\n") || !strings.Contains(text, "# meu") || !strings.Contains(text, "OUTRA=1\nREDIS_URL=") {
		t.Fatalf("%s", text)
	}
	if !strings.Contains(text, `CS="Host=a;Port=1"`) {
		t.Fatalf("; precisa de aspas: %s", text)
	}
	added, _ = MergeEnv(path, []KV{{"REDIS_URL", "outro"}})
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
		t.Fatal("sem arquivo")
	}
}

func TestEnsureIgnored(t *testing.T) {
	dir := t.TempDir()
	changed, err := EnsureIgnored(dir)
	if err != nil || !changed {
		t.Fatalf("%v %v", changed, err)
	}
	if changed, _ := EnsureIgnored(dir); changed {
		t.Fatal("não deveria repetir")
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
		t.Fatal("o que o up não ligou não é tocado")
	}
}

func TestNomesEPorta(t *testing.T) {
	if DBName("Minha-Loja.v2") != "minha_loja_v2" {
		t.Fatal(DBName("Minha-Loja.v2"))
	}
	if portOf("http://localhost:5080") != "5080" || portOf("nada") != "" {
		t.Fatal("porta")
	}
}

func TestVariaveisDoTemplateUsamAConexaoDoServico(t *testing.T) {
	m := &manifest.Project{Services: []string{"postgres", "redis"}, Env: manifest.Env{
		"postgres": {"ConnectionStrings__Default": "Host={{.Host}};Port={{.Port}};Database={{.Database}};Username={{.User}};Password={{.Password}}"},
	}}
	vars, err := serviceEnv(m, "loja")
	if err != nil {
		t.Fatal(err)
	}
	want := []KV{
		{"DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5432/loja"},
		{"ConnectionStrings__Default", "Host=127.0.0.1;Port=5432;Database=loja;Username=postgres;Password=postgres"},
		{"REDIS_URL", "redis://127.0.0.1:6379"},
	}
	if !slices.Equal(vars, want) {
		t.Fatalf("obtido %v", vars)
	}
	m.Env["postgres"]["X"] = "{{.Inexistente}}"
	if _, err := serviceEnv(m, "loja"); err == nil {
		t.Fatal("aceitou campo que não existe")
	}
}
