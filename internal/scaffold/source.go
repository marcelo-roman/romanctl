package scaffold

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

// Repo is where templates come from: the templates directory of romanctl's repository.
const Repo = "marcelo-roman/romanctl"

// Version is the binary's version. Downloaded templates come from the same tag.
var Version = "dev"

const refFile = ".ref"

// Root is the directory with projects and generators: ROMANCTL_TEMPLATES, for development, or the downloaded copy.
func Root() string {
	if v := os.Getenv("ROMANCTL_TEMPLATES"); v != "" {
		return v
	}
	return filepath.Join(config.DataDir, "templates")
}

// The directory has one subdirectory per language, each with projects/<id> and generators/<target>/<kind>.
const (
	projectsPattern  = "*/projects/*/template.yaml"
	generatorsSubdir = "generators"
	languageProjects = "projects"
)

// Installed says whether there are templates on this machine.
func Installed() bool { return hasProjects(Root()) }

func hasProjects(root string) bool {
	matches, _ := filepath.Glob(filepath.Join(root, "*", languageProjects))
	return len(matches) > 0
}

// InstalledRef says which version the downloaded templates came from.
func InstalledRef() string {
	b, err := os.ReadFile(filepath.Join(Root(), refFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func versionRef() string {
	if Version == "" || Version == "dev" {
		return mainRef
	}
	return "refs/tags/v" + strings.TrimPrefix(Version, "v")
}

func archiveURL(ref string) string {
	base := "https://codeload.github.com/" + Repo + "/tar.gz"
	if v := os.Getenv("ROMANCTL_TEMPLATES_URL"); v != "" {
		base = strings.TrimSuffix(v, "/")
	}
	return base + "/" + ref
}

var ensureOnce sync.Once

// ensureInstalled downloads the templates the first time something needs them.
func ensureInstalled() {
	ensureOnce.Do(func() {
		if Installed() || os.Getenv("ROMANCTL_TEMPLATES") != "" {
			return
		}
		if !sys.IsOnline() {
			ui.Warnf("The templates have not been downloaded yet. With internet, run: romanctl templates update")
			return
		}
		var err error
		ui.Spin(i18n.T("Downloading the templates..."), func(func(string)) { _, err = Update("") })
		if err != nil {
			ui.Warnf("Could not download the templates: %v", err)
		}
	})
}

// UpdateResult is what templates update reports.
type UpdateResult struct {
	OK       bool   `json:"ok"`
	Ref      string `json:"ref"`
	Path     string `json:"path"`
	Fallback bool   `json:"fallback,omitempty"`
}

var httpClient = &http.Client{Timeout: 2 * time.Minute}

func download(url string) (io.ReadCloser, int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "romanctl")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, resp.StatusCode, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	return resp.Body, resp.StatusCode, nil
}

// Update downloads the templates directory of a tag or branch. Empty means the binary's version; if
// that tag does not exist or has no templates yet, main.
func Update(ref string) (UpdateResult, error) {
	if os.Getenv("ROMANCTL_TEMPLATES") != "" {
		return UpdateResult{}, i18n.Errorf("ROMANCTL_TEMPLATES points to %s; templates there are not downloaded", Root())
	}
	result := UpdateResult{Path: Root()}
	dest := Root()
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return result, err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dest), ".templates-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(tmp)

	result.Ref = ref
	if ref == "" {
		result.Ref = versionRef()
	}
	err = fetch(result.Ref, tmp)
	if err != nil && ref == "" && result.Ref != mainRef {
		os.RemoveAll(tmp)
		os.MkdirAll(tmp, 0o755)
		result.Ref, result.Fallback = mainRef, true
		err = fetch(mainRef, tmp)
	}
	if err != nil {
		return result, err
	}
	if err := os.WriteFile(filepath.Join(tmp, refFile), []byte(result.Ref+"\n"), 0o644); err != nil {
		return result, err
	}
	old := dest + ".old"
	os.RemoveAll(old)
	if sys.Exists(dest) {
		if err := os.Rename(dest, old); err != nil {
			return result, err
		}
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Rename(old, dest)
		return result, err
	}
	os.RemoveAll(old)
	result.OK = true
	return result, nil
}

const mainRef = "refs/heads/main"

func fetch(ref, dest string) error {
	body, _, err := download(archiveURL(ref))
	if err != nil {
		return err
	}
	defer body.Close()
	if err := extractTemplates(body, dest); err != nil {
		return err
	}
	if !hasProjects(dest) {
		return i18n.Errorf("%s has no templates in templates/<language>/projects", ref)
	}
	return nil
}

// extractTemplates extracts from GitHub's tar.gz only what is under <root>/templates/.
func extractTemplates(r io.Reader, dest string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		parts := strings.SplitN(path.Clean(h.Name), "/", 3)
		if len(parts) < 3 || parts[1] != "templates" {
			continue
		}
		rel := parts[2]
		if strings.HasPrefix(rel, "../") || path.IsAbs(rel) {
			return i18n.Errorf("invalid path in the archive: %s", h.Name)
		}
		target := filepath.Join(dest, filepath.FromSlash(rel))
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(h.Mode)&0o755|0o644)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		}
	}
}
