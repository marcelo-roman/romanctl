// Package selfupdate updates the binary itself from GitHub releases.
package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"romanctl/internal/i18n"
)

// Repo is where releases come from.
const Repo = "marcelo-roman/romanctl"

// baseURL can be overridden by ROMANCTL_RELEASES_URL, to test against a local server.
func baseURL() string {
	if v := os.Getenv("ROMANCTL_RELEASES_URL"); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return "https://github.com/" + Repo + "/releases"
}

var client = &http.Client{Timeout: 2 * time.Minute}

func get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "romanctl")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// Latest finds the latest version. It uses the release's VERSION file and, if missing,
// the /releases/latest redirect; neither consumes GitHub API quota.
func Latest() (string, error) {
	if b, err := get(baseURL() + "/latest/download/VERSION"); err == nil {
		if v := strings.TrimSpace(string(b)); v != "" {
			return Normalize(v), nil
		}
	}
	noFollow := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := noFollow.Get(baseURL() + "/latest")
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	if _, tag, ok := strings.Cut(resp.Header.Get("Location"), "/releases/tag/"); ok && tag != "" {
		return Normalize(tag), nil
	}
	return "", i18n.Errorf("could not find the latest version of romanctl")
}

// Normalize strips the "v": v0.2.0 and 0.2.0 are the same version.
func Normalize(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }

// Compare returns -1, 0 or 1. Non-numeric parts, like -rc1, count as zero.
func Compare(a, b string) int {
	pa, pb := strings.Split(Normalize(a), "."), strings.Split(Normalize(b), ".")
	for i := 0; i < max(len(pa), len(pb)); i++ {
		x, y := 0, 0
		if i < len(pa) {
			x, _ = strconv.Atoi(strings.SplitN(pa[i], "-", 2)[0])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(strings.SplitN(pb[i], "-", 2)[0])
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

// Asset is the name of this platform's binary in the release.
func Asset() string { return "romanctl-" + runtime.GOOS + "-" + runtime.GOARCH }

func checksumFor(sums, asset string) string {
	for _, line := range strings.Split(sums, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == asset {
			return fields[0]
		}
	}
	return ""
}

// Install downloads the version, verifies it and replaces the current executable. If the new
// version does not run, the previous binary is put back.
func Install(version, target string) error {
	base := baseURL() + "/download/v" + Normalize(version)
	asset := Asset()
	sums, err := get(base + "/SHA256SUMS")
	if err != nil {
		return i18n.Errorf("could not download the checksum list for version %s: %v", version, err)
	}
	want := checksumFor(string(sums), asset)
	if want == "" {
		return i18n.Errorf("version %s has no binary for %s/%s", version, runtime.GOOS, runtime.GOARCH)
	}
	binary, err := get(base + "/" + asset)
	if err != nil {
		return i18n.Errorf("could not download %s: %v", asset, err)
	}
	sum := sha256.Sum256(binary)
	if got := hex.EncodeToString(sum[:]); got != want {
		return i18n.Errorf("checksum does not match (expected %s, got %s); nothing was replaced", want, got)
	}

	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".romanctl-new-")
	if err != nil {
		return i18n.Errorf("no permission to write to %s; run with sudo or reinstall with install.sh", dir)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(binary); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return err
	}
	// The new version runs before taking its place: a broken binary never replaces the good one.
	out, err := exec.Command(tmp.Name(), "version").CombinedOutput()
	if err != nil || !strings.Contains(string(out), Normalize(version)) {
		return i18n.Errorf("the downloaded version did not run on this machine (%v: %s); nothing was replaced", err, strings.TrimSpace(string(out)))
	}
	backup := target + ".previous"
	os.Remove(backup)
	if err := os.Rename(target, backup); err != nil {
		return i18n.Errorf("no permission to replace %s; run with sudo or reinstall with install.sh", target)
	}
	if err := os.Rename(tmp.Name(), target); err != nil {
		os.Rename(backup, target)
		return err
	}
	os.Remove(backup)
	return nil
}

// Executable is the real path of the running binary, with links resolved.
func Executable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}
