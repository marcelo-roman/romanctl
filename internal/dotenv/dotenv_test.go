package dotenv

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadIgnoraComentariosEMantemOAmbiente(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(path, []byte("# comentário\nexport A=1\nB=\"dois espaços\"\nC='x'\n\nD=4\n"), 0o600)
	got := Load(path, func(k string) (string, bool) { return "", k == "D" })
	want := []string{"A=1", "B=dois espaços", "C=x"}
	if !slices.Equal(got, want) {
		t.Fatalf("obtido %v", got)
	}
	if Load(path+"x", os.LookupEnv) != nil {
		t.Fatal("arquivo inexistente deveria dar nil")
	}
}
