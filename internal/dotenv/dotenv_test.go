package dotenv

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadIgnoresCommentsAndKeepsEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(path, []byte("# comment\nexport A=1\nB=\"two spaces\"\nC='x'\n\nD=4\n"), 0o600)
	got := Load(path, func(k string) (string, bool) { return "", k == "D" })
	want := []string{"A=1", "B=two spaces", "C=x"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v", got)
	}
	if Load(path+"x", os.LookupEnv) != nil {
		t.Fatal("a missing file should return nil")
	}
}
