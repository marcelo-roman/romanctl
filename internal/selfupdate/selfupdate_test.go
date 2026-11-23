package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.1.0", "0.2.0", -1}, {"v0.2.0", "0.2.0", 0}, {"0.10.0", "0.9.9", 1}, {"1.0", "1.0.0", 0},
		{"1.2.3", "1.2.10", -1}, {"0.2.0-rc1", "0.2.0", 0}, {"2.0.0", "v1.9.9", 1},
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("%s vs %s: %d", c.a, c.b, got)
		}
	}
}

func TestChecksumFor(t *testing.T) {
	sums := "aaa  romanctl-linux-amd64\nbbb *romanctl-linux-arm64\n"
	if checksumFor(sums, "romanctl-linux-arm64") != "bbb" || checksumFor(sums, "romanctl-darwin-arm64") != "" {
		t.Fatal("wrong checksum")
	}
}

// release sets up a server with the same URL layout as GitHub releases.
func release(t *testing.T, version, script string, corrupt bool) string {
	t.Helper()
	binary := []byte("#!/bin/sh\n" + script + "\n")
	sum := sha256.Sum256(binary)
	sums := hex.EncodeToString(sum[:]) + "  " + Asset() + "\n"
	if corrupt {
		binary = append(binary, '#')
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest/download/VERSION":
			w.Write([]byte("v" + version + "\n"))
		case "/download/v" + version + "/SHA256SUMS":
			w.Write([]byte(sums))
		case "/download/v" + version + "/" + Asset():
			w.Write(binary)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("ROMANCTL_RELEASES_URL", srv.URL)
	return srv.URL
}

func current(t *testing.T) string {
	t.Helper()
	target := filepath.Join(t.TempDir(), "romanctl")
	os.WriteFile(target, []byte("#!/bin/sh\necho romanctl 0.1.0\n"), 0o755)
	return target
}

func TestUpdates(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	release(t, "0.2.0", "echo romanctl 0.2.0", false)
	if v, err := Latest(); err != nil || v != "0.2.0" {
		t.Fatalf("%q %v", v, err)
	}
	target := current(t)
	if err := Install("0.2.0", target); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(target)
	if !strings.Contains(string(b), "0.2.0") {
		t.Fatalf("not replaced: %s", b)
	}
	if _, err := os.Stat(target + ".previous"); err == nil {
		t.Fatal("left the backup behind")
	}
}

func TestWrongChecksumDoesNotReplace(t *testing.T) {
	release(t, "0.2.0", "echo romanctl 0.2.0", true)
	target := current(t)
	if err := Install("0.2.0", target); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("%v", err)
	}
	if b, _ := os.ReadFile(target); !strings.Contains(string(b), "0.1.0") {
		t.Fatal("replaced anyway")
	}
}

func TestBinaryThatDoesNotRunDoesNotReplace(t *testing.T) {
	release(t, "0.2.0", "exit 1", false)
	target := current(t)
	if err := Install("0.2.0", target); err == nil || !strings.Contains(err.Error(), "did not run") {
		t.Fatalf("%v", err)
	}
	if b, _ := os.ReadFile(target); !strings.Contains(string(b), "0.1.0") {
		t.Fatal("replaced with a broken binary")
	}
}

func TestMissingVersion(t *testing.T) {
	release(t, "0.2.0", "echo romanctl 0.2.0", false)
	if err := Install("9.9.9", current(t)); err == nil {
		t.Fatal("should fail")
	}
}

func TestDirectoryWithoutPermission(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	release(t, "0.2.0", "echo romanctl 0.2.0", false)
	target := current(t)
	os.Chmod(filepath.Dir(target), 0o555)
	defer os.Chmod(filepath.Dir(target), 0o755)
	if err := Install("0.2.0", target); err == nil || !strings.Contains(err.Error(), "sudo") {
		t.Fatalf("%v", err)
	}
}

func TestLatestFromRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest" {
			http.Redirect(w, r, "/releases/tag/v0.3.1", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("ROMANCTL_RELEASES_URL", srv.URL)
	if v, err := Latest(); err != nil || v != "0.3.1" {
		t.Fatalf("%q %v", v, err)
	}
}
