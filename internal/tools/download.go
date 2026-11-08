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

// httpGet repete erros temporários do servidor, comuns em downloads do GitHub.
func httpGet(url string) (*http.Response, error) {
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
		resp, err := httpClient.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "x509") {
				// Certificado não melhora repetindo. Comum em imagens mínimas de container.
				return nil, i18n.Errorf("%v; instale os certificados raiz, como o pacote ca-certificates", err)
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

func fetchJSON(url string, v any) error {
	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

// download grava a URL em dir e confere o SHA-256. Sem checksum publicado pelo
// fornecedor, wantSum vem vazio e o usuário é avisado.
func download(url, dir, wantSum string) (string, error) {
	file := filepath.Join(dir, path.Base(strings.SplitN(url, "?", 2)[0]))
	ui.Infof("Baixando %s", path.Base(file))
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
		ui.Warnf("O fornecedor não publica checksum de %s: conferido só o HTTPS", path.Base(file))
	case !strings.EqualFold(got, wantSum):
		return "", i18n.Errorf("checksum de %s não confere (esperado %s, obtido %s)", path.Base(file), wantSum, got)
	default:
		ui.Infof("Checksum SHA-256 conferido")
	}
	return file, nil
}

// sumFromFile acha o hash de filename num arquivo no formato "hash  nome".
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
	// Arquivo só com o hash, como kubectl.sha256.
	if fields := strings.Fields(text); len(lines) == 1 && len(fields) >= 1 {
		return fields[0], nil
	}
	return "", i18n.Errorf("%s não aparece em %s", filename, url)
}

// githubLatestTag segue o redirecionamento de /releases/latest, sem gastar o limite da API.
func githubLatestTag(repo string) (string, error) {
	client := &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	url := "https://github.com/" + repo + "/releases/latest"
	// Repositório renomeado responde 301 para o novo endereço de /releases/latest.
	for hops := 0; hops < 5; hops++ {
		resp, err := client.Get(url)
		if err != nil {
			return "", err
		}
		resp.Body.Close()
		location := resp.Header.Get("Location")
		// A tag pode ter barra, como kustomize/v5.7.1.
		if _, tag, found := strings.Cut(location, "/releases/tag/"); found && tag != "" {
			return tag, nil
		}
		if !strings.HasSuffix(location, "/releases/latest") {
			break
		}
		url = location
	}
	return "", i18n.Errorf("não achei a última versão de %s", repo)
}

// githubAssetSum lê o SHA-256 que o GitHub calcula para cada arquivo de uma versão.
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
	return "", i18n.Errorf("%s não aparece na versão %s de %s", asset, tag, repo)
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
			return i18n.Errorf("caminho inseguro no zip: %s", f.Name)
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

// swapDir troca dest por src de uma vez, para uma falha não deixar a ferramenta pela metade.
func swapDir(src, dest string) error {
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return os.Rename(src, dest)
}

// link publica binários em /usr/local/bin.
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

// tempNear cria o diretório de trabalho ao lado do destino, para o Rename final funcionar.
func tempNear(dest string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	return os.MkdirTemp(filepath.Dir(dest), ".romanctl-")
}
