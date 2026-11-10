package scaffold

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"romanctl/internal/config"
)

func tarball(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		tw.Write([]byte(body))
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

func fakeCodeload(t *testing.T, archives map[string][]byte) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := archives[strings.TrimPrefix(r.URL.Path, "/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write(body)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("ROMANCTL_TEMPLATES", "")
	t.Setenv("ROMANCTL_TEMPLATES_URL", srv.URL)
	dataDir := config.DataDir
	config.DataDir = t.TempDir()
	t.Cleanup(func() { config.DataDir = dataDir })
}

func TestUpdateBaixaATagEDepoisAMain(t *testing.T) {
	fakeCodeload(t, map[string][]byte{
		"refs/tags/v9.9.8": tarball(t, map[string]string{"romanctl-9.9.8/README.md": "sem templates"}),
		"refs/heads/main": tarball(t, map[string]string{
			"romanctl-main/templates/projects/go/x/template.yaml": "id: x\n",
			"romanctl-main/internal/cli/root.go":                  "package cli",
		}),
	})
	version := Version
	Version = "9.9.9"
	t.Cleanup(func() { Version = version })

	r, err := Update("")
	if err != nil {
		t.Fatal(err)
	}
	Version = "9.9.8"
	if r2, err := Update(""); err != nil || !r2.Fallback {
		t.Fatalf("tag sem templates: %+v %v", r2, err)
	}
	if !r.Fallback || r.Ref != "refs/heads/main" || InstalledRef() != "refs/heads/main" {
		t.Fatalf("%+v %q", r, InstalledRef())
	}
	if _, err := os.Stat(filepath.Join(Root(), "projects/go/x/template.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(Root(), "internal")); err == nil {
		t.Fatal("extraiu o que não é template")
	}
}

func TestUpdateRecusaCaminhoParaFora(t *testing.T) {
	fakeCodeload(t, map[string][]byte{
		"refs/heads/main": tarball(t, map[string]string{"romanctl-main/templates/../../fora": "x"}),
	})
	if _, err := Update("refs/heads/main"); err == nil {
		t.Fatal("aceitou um arquivo sem templates")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(config.DataDir), "fora")); err == nil {
		t.Fatal("gravou fora da pasta")
	}
}
