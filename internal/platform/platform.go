// Package platform describes the machine romanctl is running on.
package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"romanctl/internal/i18n"
	"runtime"
	"strings"
)

// Recognized package managers.
const (
	Apt    = "apt"
	Dnf    = "dnf"
	Pacman = "pacman"
	Zypper = "zypper"
	Apk    = "apk"
	Brew   = "brew"
)

// Info is what other packages need to know about the machine.
type Info struct {
	OS      string `json:"os"`      // linux or darwin
	Arch    string `json:"arch"`    // amd64 or arm64
	Proot   bool   `json:"proot"`   // Termux/proot-distro: no Docker, namespaces or cgroups
	WSL     bool   `json:"wsl"`     // Windows Subsystem for Linux
	Root    bool   `json:"root"`    // running as root
	Manager string `json:"manager"` // package manager, or empty
	BinDir  string `json:"binDir"`  // where binaries are published
	OptDir  string `json:"optDir"`  // where multi-file tools live

	Containers     bool   `json:"containers"`     // the kernel can run containers
	Virtualization bool   `json:"virtualization"` // VM acceleration is available
	GUI            bool   `json:"gui"`            // a graphical interface is available
	Libc           string `json:"libc,omitempty"` // glibc or musl, Linux only
}

// Current is the current machine.
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

// probes isolates what comes from the system, so tests can simulate any machine.
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
		// On macOS, containers run inside a VM.
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
	// Allows forcing it when detection is wrong: ROMANCTL_CONTAINERS=1, for example.
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

// Requirements a tool may demand from the machine.
const (
	NeedContainers = "containers"
	NeedVM         = "vm"
	NeedGUI        = "gui"
	NeedGlibc      = "glibc"
)

// Missing explains why the machine does not meet the requirement; empty if it does.
func (i Info) Missing(need string) string {
	switch need {
	case NeedContainers:
		if !i.Containers {
			if i.Proot {
				return i18n.T("proot lacks the namespaces and cgroups containers need; for a remote Docker, use docker-cli")
			}
			return i18n.T("this machine cannot run containers")
		}
	case NeedVM:
		if !i.Virtualization {
			return i18n.T("this machine has no virtualization available")
		}
	case NeedGUI:
		if !i.GUI {
			return i18n.T("needs a graphical interface, and this session has none")
		}
	case NeedGlibc:
		if i.Libc == "musl" {
			return i18n.T("the official binary requires glibc, and this system uses musl")
		}
	}
	return ""
}

// Opt returns a path inside the tools directory.
func (i Info) Opt(name string) string { return filepath.Join(i.OptDir, name) }

// Sudo prefixes the command when it requires root. brew never runs as root.
func (i Info) Sudo(command ...string) []string {
	if i.Root || i.Manager == Brew {
		return command
	}
	if _, err := exec.LookPath("sudo"); err != nil {
		return command
	}
	return append([]string{"sudo"}, command...)
}

// UArch is the architecture in x86_64/aarch64 style.
func (i Info) UArch() string { return pick(i.Arch, "x86_64", "aarch64") }

// XArch is the architecture in x86_64/arm64 style, common in Go projects.
func (i Info) XArch() string { return pick(i.Arch, "x86_64", "arm64") }

// X64 is the architecture in x64/arm64 style, used by Node and others.
func (i Info) X64() string { return pick(i.Arch, "x64", "arm64") }

func pick(arch, amd, arm string) string {
	if arch == "amd64" {
		return amd
	}
	return arm
}

// Describe summarizes the platform in one line.
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
	parts = append(parts, yesNo(i.Containers, "containers", i18n.T("no containers")),
		yesNo(i.Virtualization, i18n.T("virtualization"), i18n.T("no virtualization")),
		yesNo(i.GUI, i18n.T("graphical interface"), i18n.T("no graphical interface")))
	return strings.Join(parts, " · ")
}
