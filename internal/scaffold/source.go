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

// Repo é de onde vêm os templates: a pasta templates do repositório do romanctl.
const Repo = "marcelo-roman/romanctl"

// Version é a versão do binário. Os templates baixados são os da mesma tag.
var Version = "dev"

const refFile = ".ref"

// Root é a pasta com projects e generators: ROMANCTL_TEMPLATES, para desenvolver, ou a cópia baixada.
func Root() string {
	if v := os.Getenv("ROMANCTL_TEMPLATES"); v != "" {
		return v
	}
	return filepath.Join(config.DataDir, "templates")
}

func projectsDir() string   { return filepath.Join(Root(), "projects") }
func generatorsDir() string { return filepath.Join(Root(), "generators") }

// Installed diz se há templates nesta máquina.
func Installed() bool { return sys.Exists(projectsDir()) }

// InstalledRef diz de que versão vieram os templates baixados.
func InstalledRef() string {
	b, err := os.ReadFile(filepath.Join(Root(), refFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func versionRef() string {
	if Version == "" || Version == "dev" {
		return "refs/heads/main"
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

// ensureInstalled baixa os templates na primeira vez que alguém precisa deles.
func ensureInstalled() {
	ensureOnce.Do(func() {
		if Installed() || os.Getenv("ROMANCTL_TEMPLATES") != "" {
			return
		}
		if !sys.IsOnline() {
			ui.Warnf("Os templates ainda não foram baixados. Com internet, rode: romanctl templates update")
			return
		}
		var err error
		ui.Spin(i18n.T("Baixando os templates..."), func(func(string)) { _, err = Update("") })
		if err != nil {
			ui.Warnf("Não consegui baixar os templates: %v", err)
		}
	})
}

// UpdateResult é o que templates update informa.
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

// Update baixa a pasta templates de uma tag ou branch. Vazio, a da versão do binário; se a tag
// ainda não existe, a da main.
func Update(ref string) (UpdateResult, error) {
	if os.Getenv("ROMANCTL_TEMPLATES") != "" {
		return UpdateResult{}, i18n.Errorf("ROMANCTL_TEMPLATES aponta para %s; os templates de lá não são baixados", Root())
	}
	result := UpdateResult{Path: Root()}
	wanted := ref
	if wanted == "" {
		wanted = versionRef()
	}
	body, status, err := download(archiveURL(wanted))
	if status == http.StatusNotFound && ref == "" && wanted != "refs/heads/main" {
		wanted, result.Fallback = "refs/heads/main", true
		body, _, err = download(archiveURL(wanted))
	}
	if err != nil {
		return result, err
	}
	defer body.Close()
	result.Ref = wanted

	dest := Root()
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return result, err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dest), ".templates-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(tmp)
	if err := extractTemplates(body, tmp); err != nil {
		return result, err
	}
	if !sys.Exists(filepath.Join(tmp, "projects")) {
		return result, i18n.Errorf("%s não tem a pasta templates/projects", wanted)
	}
	if err := os.WriteFile(filepath.Join(tmp, refFile), []byte(wanted+"\n"), 0o644); err != nil {
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

// extractTemplates tira do tar.gz do GitHub só o que está em <raiz>/templates/.
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
			return i18n.Errorf("caminho inválido no arquivo: %s", h.Name)
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
