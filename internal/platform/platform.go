// Package platform descreve a máquina onde o romanctl está rodando.
package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"romanctl/internal/i18n"
	"runtime"
	"strings"
)

// Gerenciadores de pacotes reconhecidos.
const (
	Apt    = "apt"
	Dnf    = "dnf"
	Pacman = "pacman"
	Zypper = "zypper"
	Apk    = "apk"
	Brew   = "brew"
)

// Info é o que os outros pacotes precisam saber sobre a máquina.
type Info struct {
	OS      string `json:"os"`      // linux ou darwin
	Arch    string `json:"arch"`    // amd64 ou arm64
	Proot   bool   `json:"proot"`   // Termux/proot-distro: sem Docker, namespaces nem cgroups
	WSL     bool   `json:"wsl"`     // Windows Subsystem for Linux
	Root    bool   `json:"root"`    // rodando como root
	Manager string `json:"manager"` // gerenciador de pacotes, ou vazio
	BinDir  string `json:"binDir"`  // onde os binários são publicados
	OptDir  string `json:"optDir"`  // onde ficam as ferramentas com vários arquivos

	Containers     bool   `json:"containers"`     // o kernel consegue rodar containers
	Virtualization bool   `json:"virtualization"` // há aceleração para VMs
	GUI            bool   `json:"gui"`            // há interface gráfica
	Libc           string `json:"libc,omitempty"` // glibc ou musl, só no Linux
}

// Current é a máquina atual.
var Current = Detect()

func Detect() Info {
	home, _ := os.UserHomeDir()
	version := procVersion()
	info := Info{
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
		Root:  os.Geteuid() == 0,
		WSL:   strings.Contains(strings.ToLower(version), "microsoft"),
		Proot: isProot(version, os.Getenv("ROMANCTL_PROOT")),
	}
	info.Manager = detectManager(info.OS, func(bin string) bool {
		_, err := exec.LookPath(bin)
		return err == nil
	})
	info.BinDir, info.OptDir = installDirs(info.Root, home)
	info.detectCapabilities(realProbes())
	if v := os.Getenv("ROMANCTL_BIN_DIR"); v != "" {
		info.BinDir = v
	}
	if v := os.Getenv("ROMANCTL_OPT_DIR"); v != "" {
		info.OptDir = v
	}
	return info
}

func procVersion() string {
	b, _ := os.ReadFile("/proc/version")
	return string(b)
}

func isProot(version, override string) bool {
	switch override {
	case "1":
		return true
	case "0":
		return false
	}
	return strings.Contains(version, "PRoot")
}

func detectManager(goos string, have func(string) bool) string {
	if goos == "darwin" {
		if have("brew") {
			return Brew
		}
		return ""
	}
	for _, m := range []struct{ bin, name string }{
		{"apt-get", Apt}, {"dnf", Dnf}, {"pacman", Pacman}, {"zypper", Zypper}, {"apk", Apk},
	} {
		if have(m.bin) {
			return m.name
		}
	}
	return ""
}

func installDirs(root bool, home string) (bin, opt string) {
	if root {
		return "/usr/local/bin", "/opt"
	}
	return filepath.Join(home, ".local", "bin"), filepath.Join(home, ".local", "share", "romanctl", "tools")
}

// probes isola o que vem do sistema, para os testes simularem qualquer máquina.
type probes struct {
	exists func(string) bool
	env    func(string) string
	glob   func(string) bool
	hvMac  func() bool
}

func realProbes() probes {
	return probes{
		exists: func(p string) bool { _, err := os.Stat(p); return err == nil },
		env:    os.Getenv,
		glob: func(pattern string) bool {
			m, _ := filepath.Glob(pattern)
			return len(m) > 0
		},
		hvMac: func() bool {
			out, err := exec.Command("sysctl", "-n", "kern.hv_support").Output()
			return err == nil && strings.TrimSpace(string(out)) == "1"
		},
	}
}

func (i *Info) detectCapabilities(p probes) {
	if i.OS == "darwin" {
		// No macOS, containers rodam dentro de uma VM.
		i.Virtualization = p.hvMac()
		i.Containers = i.Virtualization
		i.GUI = true
	} else {
		android := p.exists("/system/bin/app_process")
		i.Containers = !i.Proot && !android && p.exists("/proc/self/ns/pid") && p.exists("/sys/fs/cgroup")
		i.Virtualization = !i.Proot && !android && p.exists("/dev/kvm")
		i.GUI = !i.Proot && !android && (p.env("DISPLAY") != "" || p.env("WAYLAND_DISPLAY") != "")
		i.Libc = "glibc"
		if p.glob("/lib/ld-musl-*") {
			i.Libc = "musl"
		}
	}
	// Permite forçar quando a detecção errar: ROMANCTL_CONTAINERS=1, por exemplo.
	force := func(name string, target *bool) {
		switch p.env(name) {
		case "1":
			*target = true
		case "0":
			*target = false
		}
	}
	force("ROMANCTL_CONTAINERS", &i.Containers)
	force("ROMANCTL_VM", &i.Virtualization)
	force("ROMANCTL_GUI", &i.GUI)
}

// Requisitos que uma ferramenta pode exigir da máquina.
const (
	NeedContainers = "containers"
	NeedVM         = "vm"
	NeedGUI        = "gui"
	NeedGlibc      = "glibc"
)

// Missing explica por que a máquina não atende ao requisito; vazio se atende.
func (i Info) Missing(need string) string {
	switch need {
	case NeedContainers:
		if !i.Containers {
			if i.Proot {
				return i18n.T("o proot não tem os namespaces e cgroups de que containers precisam; para um Docker remoto, use docker-cli")
			}
			return i18n.T("esta máquina não consegue rodar containers")
		}
	case NeedVM:
		if !i.Virtualization {
			return i18n.T("esta máquina não tem virtualização disponível")
		}
	case NeedGUI:
		if !i.GUI {
			return i18n.T("precisa de interface gráfica, e esta sessão não tem")
		}
	case NeedGlibc:
		if i.Libc == "musl" {
			return i18n.T("o binário oficial exige glibc, e este sistema usa musl")
		}
	}
	return ""
}

// Opt devolve um caminho dentro da pasta de ferramentas.
func (i Info) Opt(name string) string { return filepath.Join(i.OptDir, name) }

// Sudo prefixa o comando quando ele exige root. O brew nunca roda como root.
func (i Info) Sudo(command ...string) []string {
	if i.Root || i.Manager == Brew {
		return command
	}
	if _, err := exec.LookPath("sudo"); err != nil {
		return command
	}
	return append([]string{"sudo"}, command...)
}

// UArch é a arquitetura no estilo x86_64/aarch64.
func (i Info) UArch() string { return pick(i.Arch, "x86_64", "aarch64") }

// XArch é a arquitetura no estilo x86_64/arm64, comum em projetos Go.
func (i Info) XArch() string { return pick(i.Arch, "x86_64", "arm64") }

// X64 é a arquitetura no estilo x64/arm64, usado por Node e outros.
func (i Info) X64() string { return pick(i.Arch, "x64", "arm64") }

func pick(arch, amd, arm string) string {
	if arch == "amd64" {
		return amd
	}
	return arm
}

// Describe resume a plataforma em uma linha.
func (i Info) Describe() string {
	parts := []string{i.OS + "/" + i.Arch}
	if i.Proot {
		parts = append(parts, "proot")
	}
	if i.WSL {
		parts = append(parts, "WSL")
	}
	if i.Manager != "" {
		parts = append(parts, i.Manager)
	}
	if i.Libc == "musl" {
		parts = append(parts, "musl")
	}
	yesNo := func(ok bool, yes, no string) string {
		if ok {
			return yes
		}
		return no
	}
	parts = append(parts, yesNo(i.Containers, "containers", i18n.T("sem containers")),
		yesNo(i.Virtualization, i18n.T("virtualização"), i18n.T("sem virtualização")),
		yesNo(i.GUI, i18n.T("interface gráfica"), i18n.T("sem interface gráfica")))
	return strings.Join(parts, " · ")
}
