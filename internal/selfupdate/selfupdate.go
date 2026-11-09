// Package selfupdate atualiza o próprio binário a partir das releases do GitHub.
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

// Repo é de onde vêm as releases.
const Repo = "marcelo-roman/romanctl"

// baseURL pode ser trocada por ROMANCTL_RELEASES_URL, para testar contra um servidor local.
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

// Latest descobre a última versão. Usa o arquivo VERSION da release e, se ele não existir,
// o redirecionamento de /releases/latest; nenhum dos dois gasta a cota da API do GitHub.
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
	return "", i18n.Errorf("não consegui descobrir a última versão do romanctl")
}

// Normalize tira o "v": v0.2.0 e 0.2.0 são a mesma versão.
func Normalize(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }

// Compare devolve -1, 0 ou 1. Partes que não são número, como -rc1, contam como zero.
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

// Asset é o nome do binário desta plataforma na release.
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

// Install baixa a versão, confere e troca o executável atual. Se a versão nova não rodar, o
// binário anterior volta para o lugar.
func Install(version, target string) error {
	base := baseURL() + "/download/v" + Normalize(version)
	asset := Asset()
	sums, err := get(base + "/SHA256SUMS")
	if err != nil {
		return i18n.Errorf("não consegui baixar a lista de checksums da versão %s: %v", version, err)
	}
	want := checksumFor(string(sums), asset)
	if want == "" {
		return i18n.Errorf("a versão %s não tem binário para %s/%s", version, runtime.GOOS, runtime.GOARCH)
	}
	binary, err := get(base + "/" + asset)
	if err != nil {
		return i18n.Errorf("não consegui baixar %s: %v", asset, err)
	}
	sum := sha256.Sum256(binary)
	if got := hex.EncodeToString(sum[:]); got != want {
		return i18n.Errorf("checksum não confere (esperado %s, obtido %s); nada foi trocado", want, got)
	}

	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".romanctl-novo-")
	if err != nil {
		return i18n.Errorf("sem permissão para gravar em %s; rode com sudo ou reinstale pelo install.sh", dir)
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
	// A versão nova roda antes de entrar no lugar: um binário quebrado nunca substitui o bom.
	out, err := exec.Command(tmp.Name(), "version").CombinedOutput()
	if err != nil || !strings.Contains(string(out), Normalize(version)) {
		return i18n.Errorf("a versão baixada não rodou nesta máquina (%v: %s); nada foi trocado", err, strings.TrimSpace(string(out)))
	}
	backup := target + ".anterior"
	os.Remove(backup)
	if err := os.Rename(target, backup); err != nil {
		return i18n.Errorf("sem permissão para trocar %s; rode com sudo ou reinstale pelo install.sh", target)
	}
	if err := os.Rename(tmp.Name(), target); err != nil {
		os.Rename(backup, target)
		return err
	}
	os.Remove(backup)
	return nil
}

// Executable é o caminho real do binário em uso, sem links.
func Executable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}
