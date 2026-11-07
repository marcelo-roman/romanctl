package platform

import "testing"

func TestIsProot(t *testing.T) {
	if !isProot("Linux version 6.17.0-PRoot-Distro", "") {
		t.Error("não detectou proot")
	}
	if isProot("Linux version 6.8.0-generic", "") {
		t.Error("detectou proot em Linux comum")
	}
	if isProot("6.17.0-PRoot-Distro", "0") || !isProot("generic", "1") {
		t.Error("variável de ambiente deveria ter prioridade")
	}
}

func TestDetectManager(t *testing.T) {
	only := func(bins ...string) func(string) bool {
		return func(b string) bool {
			for _, x := range bins {
				if x == b {
					return true
				}
			}
			return false
		}
	}
	cases := []struct {
		goos string
		bins []string
		want string
	}{
		{"linux", []string{"apt-get"}, Apt}, {"linux", []string{"dnf"}, Dnf}, {"linux", []string{"pacman"}, Pacman},
		{"linux", []string{"zypper"}, Zypper}, {"linux", []string{"apk"}, Apk}, {"linux", nil, ""},
		{"darwin", []string{"brew", "apt-get"}, Brew}, {"darwin", nil, ""},
	}
	for _, c := range cases {
		if got := detectManager(c.goos, only(c.bins...)); got != c.want {
			t.Errorf("%s %v: obtido %q, esperado %q", c.goos, c.bins, got, c.want)
		}
	}
}

func TestInstallDirs(t *testing.T) {
	if bin, opt := installDirs(true, "/root"); bin != "/usr/local/bin" || opt != "/opt" {
		t.Errorf("root: %s %s", bin, opt)
	}
	if bin, opt := installDirs(false, "/home/ana"); bin != "/home/ana/.local/bin" || opt != "/home/ana/.local/share/romanctl/tools" {
		t.Errorf("usuário: %s %s", bin, opt)
	}
}

func TestArquiteturas(t *testing.T) {
	amd, arm := Info{Arch: "amd64"}, Info{Arch: "arm64"}
	if amd.UArch() != "x86_64" || arm.UArch() != "aarch64" || amd.XArch() != "x86_64" || arm.XArch() != "arm64" ||
		amd.X64() != "x64" || arm.X64() != "arm64" {
		t.Error("nomes de arquitetura errados")
	}
}

func fakeProbes(paths []string, env map[string]string, musl, hv bool) probes {
	return probes{
		exists: func(p string) bool {
			for _, x := range paths {
				if x == p {
					return true
				}
			}
			return false
		},
		env:   func(k string) string { return env[k] },
		glob:  func(string) bool { return musl },
		hvMac: func() bool { return hv },
	}
}

func TestCapacidades(t *testing.T) {
	kernel := []string{"/proc/self/ns/pid", "/sys/fs/cgroup"}
	cases := []struct {
		name                string
		info                Info
		probes              probes
		containers, vm, gui bool
		libc                string
	}{
		{"desktop Linux", Info{OS: "linux"}, fakeProbes(append(kernel, "/dev/kvm"), map[string]string{"WAYLAND_DISPLAY": "wayland-0"}, false, false), true, true, true, "glibc"},
		{"servidor", Info{OS: "linux"}, fakeProbes(kernel, nil, false, false), true, false, false, "glibc"},
		{"proot", Info{OS: "linux", Proot: true}, fakeProbes(append(kernel, "/dev/kvm"), map[string]string{"DISPLAY": ":0"}, false, false), false, false, false, "glibc"},
		{"Android", Info{OS: "linux"}, fakeProbes(append(kernel, "/system/bin/app_process"), nil, false, false), false, false, false, "glibc"},
		{"Alpine", Info{OS: "linux"}, fakeProbes(kernel, nil, true, false), true, false, false, "musl"},
		{"Mac", Info{OS: "darwin"}, fakeProbes(nil, nil, false, true), true, true, true, ""},
		{"Mac em VM sem aninhamento", Info{OS: "darwin"}, fakeProbes(nil, nil, false, false), false, false, true, ""},
		{"forçado", Info{OS: "linux"}, fakeProbes(nil, map[string]string{"ROMANCTL_CONTAINERS": "1", "ROMANCTL_GUI": "1", "ROMANCTL_VM": "1"}, false, false), true, true, true, "glibc"},
	}
	for _, c := range cases {
		i := c.info
		i.detectCapabilities(c.probes)
		if i.Containers != c.containers || i.Virtualization != c.vm || i.GUI != c.gui || i.Libc != c.libc {
			t.Errorf("%s: %+v", c.name, i)
		}
	}
}

func TestMissing(t *testing.T) {
	ok := Info{Containers: true, Virtualization: true, GUI: true, Libc: "glibc"}
	for _, need := range []string{NeedContainers, NeedVM, NeedGUI, NeedGlibc} {
		if why := ok.Missing(need); why != "" {
			t.Errorf("%s: %s", need, why)
		}
		if (Info{Libc: "musl"}).Missing(need) == "" {
			t.Errorf("%s deveria faltar", need)
		}
	}
}
