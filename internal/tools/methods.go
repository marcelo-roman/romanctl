package tools

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"romanctl/internal/i18n"
	"slices"
	"strings"

	"romanctl/internal/platform"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

const termuxPrefix = "/data/data/com.termux/files/usr"

var (
	binDir  = platform.Current.BinDir
	venvDir = platform.Current.Opt("venvs")
	arch    = platform.Current.Arch
)

func opt(name string) string { return platform.Current.Opt(name) }

// ---------------------------------------------------------------- apt

var (
	aptUpdated bool
	// Outro apt rodando (em outra sessão, por exemplo) não é erro: espera a trava.
	aptLock = []string{"-o", "DPkg::Lock::Timeout=600"}
	aptEnv  = []string{"DEBIAN_FRONTEND=noninteractive"}
	// Remover estes pacotes quebraria o Ubuntu ou o próprio apt.
	aptProtected = regexp.MustCompile(`^(apt|dpkg|bash|coreutils|libc6|libc-bin|login|passwd|sed|grep|tar|gzip|findutils|perl-base|base-files|base-passwd|python3|python3-minimal|python3\.\d+|python3\.\d+-minimal|ubuntu-.*|sudo)$`)
	aptRemv      = regexp.MustCompile(`(?m)^Remv (\S+)`)
)

func privileged(env []string, command ...string) error {
	c := platform.Current.Sudo(command...)
	return sys.Quiet("", env, c[0], c[1:]...)
}

func pkgInstall(pkgs []string) error {
	if len(pkgs) == 0 {
		return i18n.Errorf("nenhum pacote definido para esta distribuição")
	}
	ui.Infof("%s install %s", platform.Current.Manager, strings.Join(pkgs, " "))
	switch platform.Current.Manager {
	case platform.Apt:
		return aptInstall(pkgs...)
	case platform.Dnf:
		return privileged(nil, slices.Concat([]string{"dnf", "install", "-y"}, pkgs)...)
	case platform.Pacman:
		return privileged(nil, slices.Concat([]string{"pacman", "-Sy", "--noconfirm", "--needed"}, pkgs)...)
	case platform.Zypper:
		return privileged(nil, slices.Concat([]string{"zypper", "--non-interactive", "install"}, pkgs)...)
	case platform.Apk:
		return privileged(nil, slices.Concat([]string{"apk", "add"}, pkgs)...)
	case platform.Brew:
		return brewInstall(pkgs, false)
	}
	return i18n.Errorf("nenhum gerenciador de pacotes reconhecido nesta máquina")
}

func pkgRemove(pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}
	ui.Infof("%s remove %s", platform.Current.Manager, strings.Join(pkgs, " "))
	switch platform.Current.Manager {
	case platform.Apt:
		return aptRemove(pkgs)
	case platform.Dnf:
		return privileged(nil, slices.Concat([]string{"dnf", "remove", "-y"}, pkgs)...)
	case platform.Pacman:
		return privileged(nil, slices.Concat([]string{"pacman", "-R", "--noconfirm"}, pkgs)...)
	case platform.Zypper:
		return privileged(nil, slices.Concat([]string{"zypper", "--non-interactive", "remove"}, pkgs)...)
	case platform.Apk:
		return privileged(nil, slices.Concat([]string{"apk", "del"}, pkgs)...)
	case platform.Brew:
		return brewRemove(pkgs, false)
	}
	return nil
}

func brewInstall(names []string, cask bool) error {
	if len(names) == 0 {
		return i18n.Errorf("não há fórmula do Homebrew para esta ferramenta")
	}
	args := []string{"install"}
	if cask {
		args = append(args, "--cask")
	}
	ui.Infof("brew install %s", strings.Join(names, " "))
	return sys.Quiet("", []string{"HOMEBREW_NO_AUTO_UPDATE=1", "NONINTERACTIVE=1"}, "brew", append(args, names...)...)
}

func brewRemove(names []string, cask bool) error {
	if len(names) == 0 {
		return nil
	}
	args := []string{"uninstall", "--ignore-dependencies"}
	if cask {
		args = append(args, "--cask")
	}
	ui.Infof("brew uninstall %s", strings.Join(names, " "))
	return sys.Quiet("", nil, "brew", append(args, names...)...)
}

func aptInstall(pkgs ...string) error {
	args := slices.Concat([]string{"apt-get"}, aptLock, []string{"install", "-y", "--no-install-recommends"}, pkgs)
	err := privileged(aptEnv, args...)
	if err != nil && !aptUpdated {
		ui.Infof("Atualizando a lista de pacotes e tentando de novo...")
		aptUpdated = true
		if uerr := privileged(aptEnv, slices.Concat([]string{"apt-get"}, aptLock, []string{"update"})...); uerr != nil {
			return uerr
		}
		err = privileged(aptEnv, args...)
	}
	return err
}

func aptIsInstalled(pkg string) bool {
	out, err := sys.Output("dpkg-query", "-W", "-f=${Status}", pkg)
	return err == nil && strings.Contains(out, "install ok installed")
}

// aptRemove simula antes para não levar nada vital junto. O autoremove só atinge
// pacotes que o apt instalou como dependência, nunca os pedidos explicitamente.
func aptRemove(pkgs []string) error {
	var present []string
	for _, p := range pkgs {
		if aptIsInstalled(p) {
			present = append(present, p)
		}
	}
	if len(present) == 0 {
		return nil
	}
	sim := slices.Concat([]string{"-s", "remove", "--autoremove"}, present)
	out, err := sys.OutputEnv(aptEnv, "apt-get", sim...)
	if err != nil {
		return i18n.Errorf("simulação do apt: %v\n%s", err, sys.Tail(out, 8))
	}
	manualOut, err := sys.Output("apt-mark", "showmanual")
	if err != nil {
		return i18n.Errorf("apt-mark: %v", err)
	}
	var removed []string
	for _, m := range aptRemv.FindAllStringSubmatch(out, -1) {
		removed = append(removed, m[1])
	}
	if blocked := aptBlocked(removed, present, strings.Fields(manualOut)); len(blocked) > 0 {
		return i18n.Errorf("remoção recusada: levaria junto %s", strings.Join(blocked, ", "))
	}
	return privileged(aptEnv, slices.Concat([]string{"apt-get"}, aptLock, []string{"remove", "-y", "--autoremove"}, present)...)
}

// aptBlocked lista o que não pode sair junto: pacotes vitais e qualquer pacote instalado
// explicitamente que não foi pedido. Remover perl, por exemplo, levaria o git.
func aptBlocked(removed, requested, manual []string) []string {
	var blocked []string
	for _, pkg := range removed {
		if aptProtected.MatchString(pkg) {
			blocked = append(blocked, pkg+" (vital)")
		} else if !slices.Contains(requested, pkg) && slices.Contains(manual, strings.TrimSuffix(pkg, ":arm64")) {
			blocked = append(blocked, pkg)
		}
	}
	return blocked
}

// ---------------------------------------------------------------- remoção

func removalPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(binDir, p)
}

func hasLeftovers(t *Tool) bool {
	for _, p := range t.RemovePaths {
		if _, err := os.Lstat(removalPath(p)); err == nil {
			return true
		}
	}
	return false
}

func (t *Tool) remove() error {
	if t.RemoveNpm != "" && sys.Have("npm") {
		ui.Infof("npm uninstall -g %s", t.RemoveNpm)
		if err := sys.Quiet("", nil, "npm", "uninstall", "-g", t.RemoveNpm); err != nil {
			return err
		}
	}
	if platform.Current.OS == "darwin" {
		if !t.Portable {
			if err := brewRemove(t.Brew, t.Cask); err != nil {
				return err
			}
		}
	} else if err := pkgRemove(t.removePkgsFor(platform.Current.Manager)); err != nil {
		return err
	}
	for _, p := range t.RemovePaths {
		if err := os.RemoveAll(removalPath(p)); err != nil {
			return err
		}
	}
	if t.AfterRemove != nil {
		return t.AfterRemove()
	}
	return nil
}

// ---------------------------------------------------------------- construtores do catálogo

// pkgTool usa o gerenciador de pacotes da máquina; .on() troca nomes por distribuição.
func pkgTool(key, name, group, desc, bin string, pkgs ...string) *Tool {
	return &Tool{Key: key, Name: name, Group: group, Description: desc, Version: []string{bin, "--version"},
		Pkgs: map[string][]string{"*": pkgs}, Brew: pkgs}
}

func npmTool(key, name, group, desc, pkg string, bins ...string) *Tool {
	return &Tool{Key: key, Name: name, Group: group, Description: desc, Version: []string{bins[0], "--version"},
		Portable: true, Npm: pkg, RemoveNpm: pkg, RemovePaths: bins, Install: func() error { return npmInstall(pkg, bins) }}
}

// npmInstall aceita "pacote" ou "pacote@versão".
func npmInstall(spec string, bins []string) error {
	if !sys.Have("npm") {
		return i18n.Errorf("precisa do Node: romanctl tools install node")
	}
	ui.Infof("npm install -g %s", spec)
	if err := sys.Quiet("", nil, "npm", "install", "-g", spec); err != nil {
		return err
	}
	prefix, err := sys.Output("npm", "prefix", "-g")
	if err != nil {
		return err
	}
	var paths []string
	for _, b := range bins {
		paths = append(paths, filepath.Join(prefix, "bin", b))
	}
	return link(paths...)
}

// venvTool isola cada ferramenta Python no próprio ambiente, sem mexer no Python do sistema.
func venvTool(key, name, group, desc, pkg string, bins ...string) *Tool {
	dest := filepath.Join(venvDir, key)
	return &Tool{Key: key, Name: name, Group: group, Description: desc, Version: []string{bins[0], "--version"}, Portable: true, PyPI: pkg,
		venvDest: dest, venvBins: bins, RemovePaths: append([]string{dest}, bins...), Install: func() error { return installVenv(dest, pkg, bins) }}
}

func installVenv(dest, pkg string, bins []string) error {
	// No Debian e no Ubuntu, o venv fica num pacote separado.
	if platform.Current.Manager == platform.Apt {
		if err := aptInstall("python3-venv"); err != nil {
			return err
		}
	}
	os.RemoveAll(dest)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	ui.Infof("Criando o ambiente Python e instalando %s...", pkg)
	if err := sys.Quiet("", nil, "python3", "-m", "venv", dest); err != nil {
		return err
	}
	if err := sys.Quiet("", nil, dest+"/bin/pip", "install", "--no-input", pkg); err != nil {
		os.RemoveAll(dest)
		return err
	}
	var paths []string
	for _, b := range bins {
		paths = append(paths, filepath.Join(dest, "bin", b))
	}
	return link(paths...)
}

// Release descreve um binário para Linux publicado no GitHub. Em Asset e Sums, {tag} e
// {ver} viram a versão, com e sem o "v", e {arch}, {uarch}, {xarch} e {x64} a arquitetura.
// {m} vem de Map, para os projetos que usam um nome próprio. Bins aceita "origem=destino".
type Release struct {
	Repo  string
	Asset string
	Sums  string // lista de checksums; vazio se o projeto não publica
	Bins  []string
	Map   map[string]string // arquitetura → valor de {m}
}

func releaseTool(key, name, group, desc string, r Release) *Tool {
	var paths []string
	for _, b := range r.Bins {
		_, dst := binNames(b)
		paths = append(paths, dst)
	}
	_, first := binNames(r.Bins[0])
	t := &Tool{Key: key, Name: name, Group: group, Description: desc, Version: []string{first, "--version"},
		Brew: []string{key}, RemovePaths: paths, Release: true, Install: func() error { return installRelease(r) },
		installAt: func(v string) error { return installReleaseAt(r, v) }}
	t.versions = func() ([]Version, error) { return githubVersions(r.Repo) }
	return t.from(r.Repo)
}

func binNames(spec string) (src, dst string) {
	if a, b, ok := strings.Cut(spec, "="); ok {
		return a, b
	}
	return spec, spec
}

func installRelease(r Release) error { return installReleaseAt(r, "") }

func installReleaseAt(r Release, want string) error {
	tag, err := resolveTag(r.Repo, want)
	if err != nil {
		return err
	}
	ver := tag[strings.LastIndex(tag, "/")+1:]
	ver = strings.TrimPrefix(ver, "v")
	expand := assetVars(r, tag, ver, platform.Current).Replace
	asset := expand(r.Asset)
	base := "https://github.com/" + r.Repo + "/releases/download/" + tag + "/"
	sum := ""
	if r.Sums != "" {
		if sum, err = sumFromFile(base+expand(r.Sums), asset); err != nil {
			return err
		}
	}
	return installFromURL(base+asset, sum, r.Bins)
}

func assetVars(r Release, tag, ver string, p platform.Info) *strings.Replacer {
	return strings.NewReplacer("{tag}", tag, "{ver}", ver, "{shorttag}", tag[strings.LastIndex(tag, "/")+1:],
		"{arch}", p.Arch, "{uarch}", p.UArch(), "{xarch}", p.XArch(), "{x64}", p.X64(), "{m}", r.Map[p.Arch])
}

// installFromURL aceita binário solto, .tar.* ou .zip e procura os binários dentro.
func installFromURL(url, sum string, bins []string) error {
	tmp, err := os.MkdirTemp("", "romanctl-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	file, err := download(url, tmp, sum)
	if err != nil {
		return err
	}
	unpacked := filepath.Join(tmp, "x")
	switch {
	case strings.HasSuffix(file, ".zip"):
		err = extractZip(file, unpacked)
	case strings.Contains(filepath.Base(file), ".tar.") || strings.HasSuffix(file, ".tgz"):
		err = extractTar(file, unpacked, 0)
	default:
		_, dst := binNames(bins[0])
		return installBinary(file, dst)
	}
	if err != nil {
		return err
	}
	for _, spec := range bins {
		src, dst := binNames(spec)
		found, err := findFile(unpacked, src)
		if err != nil {
			return err
		}
		if err := installBinary(found, dst); err != nil {
			return err
		}
	}
	return nil
}

func findFile(root, name string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == name && found == "" {
			found = p
		}
		return nil
	})
	if err == nil && found == "" {
		err = i18n.Errorf("%s não está no pacote baixado", name)
	}
	return found, err
}

// installToOpt extrai um pacote inteiro numa pasta, para ferramentas com vários arquivos.
func installToOpt(url, sum, dest string, strip int, bins ...string) error {
	tmp, err := tempNear(dest)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	file, err := download(url, tmp, sum)
	if err != nil {
		return err
	}
	unpacked := filepath.Join(tmp, "x")
	if strings.HasSuffix(file, ".zip") {
		if err := extractZip(file, unpacked); err != nil {
			return err
		}
		if entries, _ := os.ReadDir(unpacked); strip == 1 && len(entries) == 1 {
			unpacked = filepath.Join(unpacked, entries[0].Name())
		}
	} else if err := extractTar(file, unpacked, strip); err != nil {
		return err
	}
	if err := swapDir(unpacked, dest); err != nil {
		return err
	}
	var paths []string
	for _, b := range bins {
		paths = append(paths, filepath.Join(dest, b))
	}
	return link(paths...)
}

// hashicorpTool cobre os produtos da HashiCorp, que seguem o mesmo padrão de download.
func hashicorpTool(key, name, group, desc string) *Tool {
	latest := func() (string, error) { return hashicorpLatest(key) }
	install := func(want string) error {
		version, err := orLatest(want, latest)
		if err != nil {
			return err
		}
		base := "https://releases.hashicorp.com/" + key + "/" + version + "/"
		asset := key + "_" + version + "_linux_" + arch + ".zip"
		sum, err := sumFromFile(base+key+"_"+version+"_SHA256SUMS", asset)
		if err != nil {
			return err
		}
		return installFromURL(base+asset, sum, []string{key})
	}
	return &Tool{Key: key, Name: name, Group: group, Description: desc, Version: []string{key, "version"}, Repo: "hashicorp/" + key, latest: latest,
		Brew: []string{"hashicorp/tap/" + key}, RemovePaths: []string{key}, Install: func() error { return install("") }, installAt: install,
		versions: func() ([]Version, error) { return hashicorpVersions(key) }}
}

func hashicorpLatest(key string) (string, error) {
	var check struct {
		Version string `json:"current_version"`
	}
	err := fetchJSON("https://checkpoint-api.hashicorp.com/v1/check/"+key, &check)
	return check.Version, err
}
