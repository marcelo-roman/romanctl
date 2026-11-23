package tools

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"romanctl/internal/i18n"
	"strings"
	"time"

	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

var httpClient = &http.Client{Transport: &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	ResponseHeaderTimeout: 60 * time.Second,
	TLSHandshakeTimeout:   20 * time.Second,
}}

// httpGet retries temporary server errors, common in GitHub downloads. headers come in
// name/value pairs.
func httpGet(url string, headers ...string) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < httpAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(httpBackoff * time.Duration(attempt))
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "romanctl")
		for i := 0; i+1 < len(headers); i += 2 {
			req.Header.Set(headers[i], headers[i+1])
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "x509") {
				// Retrying does not fix a certificate. Common in minimal container images.
				return nil, i18n.Errorf("%v; install the root certificates, such as the ca-certificates package", err)
			}
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		resp.Body.Close()
		lastErr = i18n.Errorf("%s: HTTP %d", url, resp.StatusCode)
		if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return nil, lastErr
		}
	}
	return nil, lastErr
}

const httpAttempts = 3

var httpBackoff = 2 * time.Second

func fetchText(url string) (string, error) {
	resp, err := httpGet(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(b)), err
}

func fetchJSON(url string, v any) error { return fetchJSONWith(url, v) }

func fetchJSONWith(url string, v any, headers ...string) error {
	resp, err := httpGet(url, headers...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

// download saves the URL into dir and verifies the SHA-256. Without a checksum published by the
// vendor, wantSum is empty and the user is warned.
func download(url, dir, wantSum string) (string, error) {
	file := filepath.Join(dir, path.Base(strings.SplitN(url, "?", 2)[0]))
	ui.Infof("Downloading %s", path.Base(file))
	resp, err := httpGet(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	out, err := os.Create(file)
	if err != nil {
		return "", err
	}
	defer out.Close()
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, hash), resp.Body); err != nil {
		return "", err
	}
	got := hex.EncodeToString(hash.Sum(nil))
	switch {
	case wantSum == "":
		ui.Warnf("The vendor does not publish a checksum for %s: only HTTPS was verified", path.Base(file))
	case !strings.EqualFold(got, wantSum):
		return "", i18n.Errorf("checksum of %s does not match (expected %s, got %s)", path.Base(file), wantSum, got)
	default:
		ui.Infof("SHA-256 checksum verified")
	}
	return file, nil
}

// sumFromFile finds filename's hash in a file in the "hash  name" format.
func sumFromFile(url, filename string) (string, error) {
	text, err := fetchText(url)
	if err != nil {
		return "", err
	}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && path.Base(strings.TrimPrefix(fields[len(fields)-1], "*")) == filename {
			return fields[0], nil
		}
	}
	// A file with only the hash, like kubectl.sha256.
	if fields := strings.Fields(text); len(lines) == 1 && len(fields) >= 1 {
		return fields[0], nil
	}
	return "", i18n.Errorf("%s does not appear in %s", filename, url)
}

// githubLatestTag follows the /releases/latest redirect, without consuming API rate limit.
func githubLatestTag(repo string) (string, error) {
	client := &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	url := "https://github.com/" + repo + "/releases/latest"
	// A renamed repository answers 301 to the new /releases/latest address.
	for hops := 0; hops < 5; hops++ {
		resp, err := client.Get(url)
		if err != nil {
			return "", err
		}
		resp.Body.Close()
		location := resp.Header.Get("Location")
		// The tag may contain a slash, like kustomize/v5.7.1.
		if _, tag, found := strings.Cut(location, "/releases/tag/"); found && tag != "" {
			return tag, nil
		}
		if !strings.HasSuffix(location, "/releases/latest") {
			break
		}
		url = location
	}
	return "", i18n.Errorf("could not find the latest version of %s", repo)
}

// githubAssetSum reads the SHA-256 GitHub computes for each file of a release.
func githubAssetSum(repo, tag, asset string) (string, error) {
	var release struct {
		Assets []struct {
			Name   string `json:"name"`
			Digest string `json:"digest"`
		} `json:"assets"`
	}
	if err := fetchJSON("https://api.github.com/repos/"+repo+"/releases/tags/"+tag, &release); err != nil {
		return "", err
	}
	for _, a := range release.Assets {
		if a.Name == asset {
			return strings.TrimPrefix(a.Digest, "sha256:"), nil
		}
	}
	return "", i18n.Errorf("%s does not appear in version %s of %s", asset, tag, repo)
}

func extractTar(file, dest string, strip int) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	if out, err := sys.Output("tar", "-xf", file, "-C", dest, i18n.Sprintf("--strip-components=%d", strip)); err != nil {
		return i18n.Errorf("tar: %v %s", err, out)
	}
	return nil
}

func extractZip(file, dest string) error {
	r, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	defer r.Close()
	root := filepath.Clean(dest) + string(os.PathSeparator)
	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(target, root) {
			return i18n.Errorf("unsafe path in zip: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := writeZipEntry(f, target); err != nil {
			return err
		}
	}
	return nil
}

func writeZipEntry(f *zip.File, target string) error {
	in, err := f.Open()
	if err != nil {
		return err
	}
	defer in.Close()
	if f.Mode()&os.ModeSymlink != 0 {
		link, err := io.ReadAll(in)
		if err != nil {
			return err
		}
		return os.Symlink(string(link), target)
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode().Perm()|0o600)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// swapDir replaces dest with src at once, so a failure never leaves the tool half-installed.
func swapDir(src, dest string) error {
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return os.Rename(src, dest)
}

// link publishes binaries to /usr/local/bin.
func link(paths ...string) error {
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	for _, p := range paths {
		target := filepath.Join(binDir, filepath.Base(p))
		os.Remove(target)
		if err := os.Symlink(p, target); err != nil {
			return err
		}
	}
	return nil
}

func installBinary(src, name string) error {
	if err := os.Chmod(src, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	target := filepath.Join(binDir, name)
	os.Remove(target)
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(target, b, 0o755)
}

// tempNear creates the working directory next to the destination, so the final Rename works.
func tempNear(dest string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	return os.MkdirTemp(filepath.Dir(dest), ".romanctl-")
}
