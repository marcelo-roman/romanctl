package tools

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"romanctl/internal/i18n"
	"slices"
	"sort"
	"strings"

	"romanctl/internal/platform"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

func kubectlLatest() (string, error) { return fetchText("https://dl.k8s.io/release/stable.txt") }

func installKubectl() error { return installKubectlAt("") }

func installKubectlAt(want string) error {
	version, err := orLatest(want, kubectlLatest)
	if err != nil {
		return err
	}
	url := "https://dl.k8s.io/release/" + version + "/bin/linux/" + arch + "/kubectl"
	sum, err := fetchText(url + ".sha256")
	if err != nil {
		return err
	}
	return installFromURL(url, sum, []string{"kubectl"})
}

func installHelm() error { return installHelmAt("") }

func installHelmAt(want string) error {
	tag, err := resolveTag("helm/helm", want)
	if err != nil {
		return err
	}
	name := "helm-" + tag + "-linux-" + arch + ".tar.gz"
	url := "https://get.helm.sh/" + name
	sum, err := sumFromFile(url+".sha256sum", name)
	if err != nil {
		return err
	}
	return installFromURL(url, sum, []string{"helm"})
}

func installAWS() error {
	tmp, err := os.MkdirTemp("", "romanctl-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	file, err := download("https://awscli.amazonaws.com/awscli-exe-linux-"+platform.Current.UArch()+".zip", tmp, "")
	if err != nil {
		return err
	}
	if err := extractZip(file, tmp); err != nil {
		return err
	}
	ui.Infof("Rodando o instalador oficial da AWS...")
	return sys.Quiet(tmp, nil, filepath.Join(tmp, "aws", "install"), "-i", opt("aws-cli"), "-b", binDir, "--update")
}

func installGcloud() error {
	url := "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-linux-" +
		map[string]string{"amd64": "x86_64", "arm64": "arm"}[arch] + ".tar.gz"
	return installToOpt(url, "", opt("google-cloud-sdk"), 1, "bin/gcloud", "bin/gsutil", "bin/bq")
}

func installDockerCLI() error {
	base := "https://download.docker.com/linux/static/stable/" + platform.Current.UArch() + "/"
	page, err := fetchText(base)
	if err != nil {
		return err
	}
	latest, best := "", [3]int{}
	for _, m := range dockerAsset.FindAllStringSubmatch(page, -1) {
		var v [3]int
		fmt.Sscanf(m[1], "%d.%d.%d", &v[0], &v[1], &v[2])
		if v[0] > best[0] || (v[0] == best[0] && (v[1] > best[1] || (v[1] == best[1] && v[2] >= best[2]))) {
			best, latest = v, m[0]
		}
	}
	if latest == "" {
		return i18n.Errorf("não achei o Docker em %s", base)
	}
	return installFromURL(base+latest, "", []string{"docker"})
}

func installLima() error { return installLimaAt("") }

func installLimaAt(want string) error {
	const repo = "lima-vm/lima"
	tag, err := resolveTag(repo, want)
	if err != nil {
		return err
	}
	base := "https://github.com/" + repo + "/releases/download/" + tag + "/"
	asset := "lima-" + strings.TrimPrefix(tag, "v") + "-Linux-" + platform.Current.UArch() + ".tar.gz"
	sum, err := sumFromFile(base+"SHA256SUMS", asset)
	if err != nil {
		return err
	}
	return installToOpt(base+asset, sum, opt("lima"), 0, "bin/limactl", "bin/lima")
}

func installOllama() error { return installOllamaAt("") }

func installOllamaAt(want string) error {
	if !sys.Have("zstd") {
		if err := pkgInstall([]string{"zstd"}); err != nil {
			return err
		}
	}
	const repo = "ollama/ollama"
	tag, err := resolveTag(repo, want)
	if err != nil {
		return err
	}
	base := "https://github.com/" + repo + "/releases/download/" + tag + "/"
	asset := "ollama-linux-" + arch + ".tar.zst"
	sum, err := sumFromFile(base+"sha256sum.txt", asset)
	if err != nil {
		return err
	}
	// O binário procura as bibliotecas ao lado dele, então fica em uma pasta própria.
	return installToOpt(base+asset, sum, opt("ollama"), 0, "bin/ollama")
}

func nodeLTS() (string, error) {
	var releases []struct {
		Version string `json:"version"`
		LTS     any    `json:"lts"`
	}
	if err := fetchJSON("https://nodejs.org/dist/index.json", &releases); err != nil {
		return "", err
	}
	for _, r := range releases {
		if r.LTS != false {
			return r.Version, nil
		}
	}
	return "", i18n.Errorf("não achei uma versão LTS do Node")
}

func installNode() error { return installNodeAt("") }

func installNodeAt(want string) error {
	version, err := orLatest(want, nodeLTS)
	if err != nil {
		return err
	}
	base := "https://nodejs.org/dist/" + version + "/"
	name := "node-" + version + "-linux-" + platform.Current.X64() + ".tar.xz"
	sum, err := sumFromFile(base+"SHASUMS256.txt", name)
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "romanctl-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	file, err := download(base+name, tmp, sum)
	if err != nil {
		return err
	}
	// Extrai por cima, sem apagar: preserva os pacotes globais, como o mongosh.
	dest := opt("node")
	if err := extractTar(file, dest, 1); err != nil {
		return err
	}
	return link(dest+"/bin/node", dest+"/bin/npm", dest+"/bin/npx", dest+"/bin/corepack")
}

type goRelease struct {
	Version string `json:"version"`
	Files   []struct {
		Filename string `json:"filename"`
		OS       string `json:"os"`
		Arch     string `json:"arch"`
		Kind     string `json:"kind"`
		Sha256   string `json:"sha256"`
	} `json:"files"`
}

func goReleases() ([]goRelease, error) {
	var releases []goRelease
	if err := fetchJSON("https://go.dev/dl/?mode=json", &releases); err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, i18n.Errorf("lista de versões do Go vazia")
	}
	return releases, nil
}

func goAllReleases() ([]goRelease, error) {
	var releases []goRelease
	if err := fetchJSON("https://go.dev/dl/?mode=json&include=all", &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func goLatest() (string, error) {
	releases, err := goReleases()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(releases[0].Version, "go"), nil
}

func installGo() error { return installGoAt("") }

func installGoAt(want string) error {
	releases, err := goReleases()
	if want != "" {
		releases, err = goAllReleases()
	}
	if err != nil {
		return err
	}
	release := releases[0]
	if want != "" {
		i := slices.IndexFunc(releases, func(r goRelease) bool { return r.Version == "go"+strings.TrimPrefix(want, "go") })
		if i < 0 {
			return i18n.Errorf("o Go %s não está na lista de versões", want)
		}
		release = releases[i]
	}
	for _, f := range release.Files {
		if f.OS == "linux" && f.Arch == arch && f.Kind == "archive" {
			return installToOpt("https://go.dev/dl/"+f.Filename, f.Sha256, goDir(), 1, "bin/go", "bin/gofmt")
		}
	}
	return i18n.Errorf("não achei o Go para linux/%s", arch)
}

func goDir() string {
	if platform.Current.Root {
		return "/usr/local/go"
	}
	return opt("go")
}

func zigIndex() (map[string]map[string]any, string, error) {
	var index map[string]map[string]any
	if err := fetchJSON("https://ziglang.org/download/index.json", &index); err != nil {
		return nil, "", err
	}
	var versions []string
	for v := range index {
		if v != "master" {
			versions = append(versions, v)
		}
	}
	sort.Slice(versions, func(i, j int) bool { return semverLess(versions[i], versions[j]) })
	if len(versions) == 0 {
		return nil, "", i18n.Errorf("lista de versões do Zig vazia")
	}
	return index, versions[len(versions)-1], nil
}

func zigLatest() (string, error) {
	_, version, err := zigIndex()
	return version, err
}

func installZig() error { return installZigAt("") }

func installZigAt(want string) error {
	index, version, err := zigIndex()
	if want != "" {
		version = want
	}
	if err != nil {
		return err
	}
	target, ok := index[version][platform.Current.UArch()+"-linux"].(map[string]any)
	if !ok {
		return i18n.Errorf("Zig %s não tem versão para %s-linux", version, platform.Current.UArch())
	}
	url, _ := target["tarball"].(string)
	sum, _ := target["shasum"].(string)
	return installToOpt(url, sum, opt("zig"), 1, "zig")
}

func semverLess(a, b string) bool {
	var x, y [3]int
	fmt.Sscanf(a, "%d.%d.%d", &x[0], &x[1], &x[2])
	fmt.Sscanf(b, "%d.%d.%d", &y[0], &y[1], &y[2])
	for i := range x {
		if x[i] != y[i] {
			return x[i] < y[i]
		}
	}
	return false
}

func installDart() error {
	asset := "dartsdk-linux-" + platform.Current.X64() + "-release.zip"
	url := "https://storage.googleapis.com/dart-archive/channels/stable/release/latest/sdk/" + asset
	sum, err := sumFromFile(url+".sha256sum", asset)
	if err != nil {
		return err
	}
	return installToOpt(url, sum, opt("dart-sdk"), 1, "bin/dart")
}

func installKotlin() error { return installKotlinAt("") }

// installGradle usa a distribuição oficial: o pacote gradle do Ubuntu parou na 4.4.
func installGradle() error {
	var current struct {
		Version     string `json:"version"`
		DownloadURL string `json:"downloadUrl"`
		ChecksumURL string `json:"checksumUrl"`
	}
	if err := fetchJSON("https://services.gradle.org/versions/current", &current); err != nil {
		return err
	}
	if current.DownloadURL == "" {
		return i18n.Errorf("não achei a versão atual do Gradle")
	}
	sum, err := sumFromFile(current.ChecksumURL, path.Base(current.DownloadURL))
	if err != nil {
		return err
	}
	return installToOpt(current.DownloadURL, sum, opt("gradle"), 1, "bin/gradle")
}

func installKotlinAt(want string) error {
	const repo = "JetBrains/kotlin"
	tag, err := resolveTag(repo, want)
	if err != nil {
		return err
	}
	asset := "kotlin-compiler-" + strings.TrimPrefix(tag, "v") + ".zip"
	url := "https://github.com/" + repo + "/releases/download/" + tag + "/" + asset
	sum, err := sumFromFile(url+".sha256", asset)
	if err != nil {
		return err
	}
	return installToOpt(url, sum, opt("kotlin"), 1, "bin/kotlin", "bin/kotlinc")
}

func installSwift() error {
	var releases []struct {
		Name string `json:"name"`
		Tag  string `json:"tag"`
	}
	if err := fetchJSON("https://www.swift.org/api/v1/install/releases.json", &releases); err != nil {
		return err
	}
	if len(releases) == 0 {
		return i18n.Errorf("lista de versões do Swift vazia")
	}
	tag := releases[len(releases)-1].Tag
	if platform.Current.Manager == platform.Apt {
		// Dependências da toolchain, conforme a documentação do Swift.
		if err := aptInstall("libncurses6", "libedit2", "libxml2", "libz3-4", "libcurl4t64", "libsqlite3-0", "libpython3.12t64", "binutils", "tzdata"); err != nil {
			return err
		}
	}
	dir, suffix := "ubuntu2404", "ubuntu24.04"
	if arch == "arm64" {
		dir, suffix = "ubuntu2404-aarch64", "ubuntu24.04-aarch64"
	}
	url := "https://download.swift.org/" + strings.ToLower(tag) + "/" + dir + "/" + tag + "/" + tag + "-" + suffix + ".tar.gz"
	return installToOpt(url, "", opt("swift"), 1, "usr/bin/swift", "usr/bin/swiftc")
}

func installNeovim() error { return installNeovimAt("") }

func installNeovimAt(want string) error {
	const repo = "neovim/neovim"
	tag, err := resolveTag(repo, want)
	if err != nil {
		return err
	}
	asset := "nvim-linux-" + platform.Current.XArch() + ".tar.gz"
	sum, err := githubAssetSum(repo, tag, asset)
	if err != nil {
		return err
	}
	return installToOpt("https://github.com/"+repo+"/releases/download/"+tag+"/"+asset, sum, opt("nvim"), 1, "bin/nvim")
}

func installDotnetEF() error {
	if !sys.Have("dotnet") {
		return i18n.Errorf("precisa do .NET: romanctl tools install dotnet")
	}
	if err := sys.Quiet("", dotnetToolEnv, "dotnet", "tool", "update", "--global", "dotnet-ef"); err != nil {
		return err
	}
	if !platform.Current.Proot {
		return link(dotnetEFPath())
	}
	// No proot, o executável da ferramenta não passa pelo wrapper e precisa do limite de heap.
	script := "#!/bin/sh\n: \"${DOTNET_GCHeapHardLimit:=0x40000000}\"\nexport DOTNET_GCHeapHardLimit\nexec " + dotnetEFPath() + " \"$@\"\n"
	return os.WriteFile(binDir+"/dotnet-ef", []byte(script), 0o755)
}

var dotnetToolEnv = []string{"DOTNET_NOLOGO=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1"}

func dotnetEFPath() string { return filepath.Join(os.Getenv("HOME"), ".dotnet", "tools", "dotnet-ef") }

func removeDotnetEF() error {
	if !sys.Have("dotnet") {
		return nil
	}
	if _, err := os.Stat(dotnetEFPath()); err != nil {
		return nil
	}
	return sys.Quiet("", dotnetToolEnv, "dotnet", "tool", "uninstall", "--global", "dotnet-ef")
}

func installRust() error {
	base := "https://static.rust-lang.org/rustup/dist/" + platform.Current.UArch() + "-unknown-linux-gnu/rustup-init"
	text, err := fetchText(base + ".sha256")
	if err != nil {
		return err
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return i18n.Errorf("checksum do rustup vazio")
	}
	tmp, err := os.MkdirTemp("", "romanctl-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	file, err := download(base, tmp, fields[0])
	if err != nil {
		return err
	}
	if err := os.Chmod(file, 0o755); err != nil {
		return err
	}
	ui.Infof("Instalando o Rust com o rustup (demora)...")
	if err := sys.Quiet(tmp, nil, file, "-y", "--no-modify-path", "--profile", "minimal"); err != nil {
		return err
	}
	cargoBin := filepath.Join(os.Getenv("HOME"), ".cargo", "bin")
	return link(filepath.Join(cargoBin, "cargo"), filepath.Join(cargoBin, "rustc"), filepath.Join(cargoBin, "rustup"))
}

func rustPaths() []string {
	h := os.Getenv("HOME")
	return []string{"cargo", "rustc", "rustup", filepath.Join(h, ".cargo"), filepath.Join(h, ".rustup")}
}

func installCodeServer() error { return installCodeServerAt("") }

func installCodeServerAt(want string) error {
	const repo = "coder/code-server"
	tag, err := resolveTag(repo, want)
	if err != nil {
		return err
	}
	name := "code-server-" + strings.TrimPrefix(tag, "v") + "-linux-" + arch + ".tar.gz"
	sum, err := githubAssetSum(repo, tag, name)
	if err != nil {
		return err
	}
	url := "https://github.com/" + repo + "/releases/download/" + tag + "/" + name
	return installToOpt(url, sum, opt("code-server"), 1, "bin/code-server")
}

// No proot, o GC do .NET falha ao reservar memória (0x8007000E) sem limite de heap,
// e o dotnet test não reconhece o binário do dotnet.
const (
	dotnetRunSettingsPath = "/usr/local/etc/dotnet-proot.runsettings"
	dotnetRunSettings     = `<?xml version="1.0" encoding="utf-8"?>
<!-- No proot, o dotnet test não reconhece a arquitetura do binário do dotnet. -->
<RunSettings>
  <RunConfiguration>
    <DotNetHostPath>/usr/lib/dotnet/dotnet</DotNetHostPath>
  </RunConfiguration>
</RunSettings>
`
	dotnetWrapper = `#!/bin/sh
# Wrapper: no proot/Android o GC do .NET falha ao reservar memória (0x8007000E).
# Limitar o heap resolve. Ajuste/remova DOTNET_GCHeapHardLimit se precisar.
: "${DOTNET_GCHeapHardLimit:=0x40000000}"
export DOTNET_GCHeapHardLimit
# O dotnet test não reconhece o binário no proot: esta configuração indica o caminho.
# O MSBuild lê variáveis de ambiente como propriedades; um projeto pode sobrescrevê-la.
: "${RunSettingsFilePath:=` + dotnetRunSettingsPath + `}"
export RunSettingsFilePath
exec /usr/lib/dotnet/dotnet "$@"
`
)

// writeDotnetWrapper grava os contornos que só o proot precisa.
func writeDotnetWrapper() error {
	if !platform.Current.Proot {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dotnetRunSettingsPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dotnetRunSettingsPath, []byte(dotnetRunSettings), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(binDir+"/dotnet", []byte(dotnetWrapper), 0o755); err != nil {
		return err
	}
	ui.Infof("Wrapper do .NET para o proot gravado em %s/dotnet", binDir)
	return nil
}

// O wrapper só sai junto com o último SDK.
func removeDotnetWrapperIfUnused() error {
	if !platform.Current.Proot {
		return nil
	}
	sdks, _ := filepath.Glob("/usr/lib/dotnet/sdk/*")
	if len(sdks) > 0 {
		return nil
	}
	os.Remove(binDir + "/dotnet")
	os.Remove(dotnetRunSettingsPath)
	return nil
}

func dotnetHasSDK(major string) func() (string, bool) {
	return func() (string, bool) {
		out, err := versionOutput("dotnet", "--list-sdks")
		if err != nil {
			return "", false
		}
		var found []string
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, major+".") {
				found = append(found, strings.Fields(line)[0])
			}
		}
		if len(found) == 0 {
			return "", false
		}
		sort.Strings(found)
		return found[len(found)-1], true
	}
}

// O Python sozinho não basta: o venv fica num pacote separado no Ubuntu.
func detectPython() (string, bool) {
	if _, err := versionOutput("python3", "-c", "import ensurepip, venv"); err != nil {
		return "", false
	}
	version, err := versionOutput("python3", "--version")
	return strings.TrimSpace(version), err == nil
}
