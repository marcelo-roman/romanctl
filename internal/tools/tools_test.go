package tools

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/ui"
)

func keys(list []*Tool) string {
	var out []string
	for _, t := range list {
		out = append(out, t.Key)
	}
	return strings.Join(out, ",")
}

func TestResolveAcceptsAliasesWithoutRepeats(t *testing.T) {
	got, err := Resolve([]string{"azure-cli", "az", "GOLANG"})
	if err != nil {
		t.Fatal(err)
	}
	if keys(got) != "az,go" {
		t.Fatalf("got %s", keys(got))
	}
}

func TestResolveExpandsGroupsAndAll(t *testing.T) {
	cloud, err := Resolve([]string{"cloud"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range cloud {
		if tool.Group != GroupCloud {
			t.Fatalf("%s is not in the cloud group", tool.Key)
		}
	}
	all, err := Resolve([]string{"sdks", "all"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range all {
		if why := tool.Unavailable(); why != "" {
			t.Errorf("all brought %s, which does not run here: %s", tool.Key, why)
		}
	}
	if len(all) == 0 {
		t.Fatal("all came back empty")
	}
}

func TestUnknownTool(t *testing.T) {
	if _, err := Resolve([]string{"does-not-exist"}); err == nil || !ui.IsUsage(err) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistryHasNoDuplicateKeys(t *testing.T) {
	seen := map[string]string{}
	for _, tool := range All {
		for _, name := range append([]string{tool.Key}, tool.Aliases...) {
			if other, dup := seen[name]; dup {
				t.Errorf("%q points to %s and %s", name, other, tool.Key)
			}
			seen[name] = tool.Key
		}
		if (tool.Install == nil && len(tool.Pkgs) == 0 && len(tool.Brew) == 0) || (tool.Detect == nil && len(tool.Version) == 0) {
			t.Errorf("%s is incomplete", tool.Key)
		}
	}
}

func TestFirstLineIgnoresWarningsAndSpaces(t *testing.T) {
	got := firstLine("\nWARNING: something\nVersion     v1.0  \nother\n")
	if got != "Version v1.0" {
		t.Fatalf("got %q", got)
	}
}

func TestVersionLine(t *testing.T) {
	gradle := "openjdk version \"21\"\n\n------\nGradle 9.2.0\n------"
	if got := versionLine(gradle, "Gradle "); got != "Gradle 9.2.0" {
		t.Fatalf("%q", got)
	}
	if got := versionLine("go: downloading go1.27.3 (linux/amd64)\ngo version go1.27.3 linux/amd64", "go version"); got != "go version go1.27.3 linux/amd64" {
		t.Fatalf("%q", got)
	}
	if got := firstLine("info: kotlinc-jvm 2.4.21 (JRE 21)"); got != "kotlinc-jvm 2.4.21 (JRE 21)" {
		t.Fatalf("%q", got)
	}
}

func TestFirstLineStripsColors(t *testing.T) {
	if got := firstLine("\x1b[1mApache Maven 3.8.7\x1b[m\nMaven home: /usr"); got != "Apache Maven 3.8.7" {
		t.Fatalf("%q", got)
	}
}

func TestToolNameBeatsGroup(t *testing.T) {
	got, err := Resolve([]string{"node"})
	if err != nil || keys(got) != "node" {
		t.Fatalf("got %s, %v", keys(got), err)
	}
}

func TestEveryGroupHasToolsAndEveryToolHasGroup(t *testing.T) {
	count := map[string]int{}
	for _, tool := range All {
		count[tool.Group]++
		if tool.Description == "" {
			t.Errorf("%s has no description", tool.Key)
		}
		if tool.RemoveNpm == "" && len(tool.Pkgs) == 0 && len(tool.RemovePkgs) == 0 && len(tool.RemovePaths) == 0 && tool.Brew == nil {
			t.Errorf("%s does not know how to remove itself", tool.Key)
		}
	}
	for _, g := range Groups {
		if count[g.Key] == 0 {
			t.Errorf("group %s is empty", g.Key)
		}
		delete(count, g.Key)
	}
	if len(count) > 0 {
		t.Errorf("groups missing from the listing: %v", count)
	}
}

func TestVitalPackagesAreProtected(t *testing.T) {
	for _, pkg := range []string{"apt", "python3", "python3.12", "libc6", "coreutils"} {
		if !aptProtected.MatchString(pkg) {
			t.Errorf("%s should be protected", pkg)
		}
	}
	for _, pkg := range []string{"jq", "php-cli", "python3-pip", "aptitude"} {
		if aptProtected.MatchString(pkg) {
			t.Errorf("%s should not be protected", pkg)
		}
	}
}

func TestAptBlockedRejectsExplicitlyInstalledPackages(t *testing.T) {
	got := aptBlocked([]string{"perl", "git", "libfoo", "apt"}, []string{"perl"}, []string{"perl", "git"})
	if strings.Join(got, ",") != "git,apt (vital)" {
		t.Fatalf("got %v", got)
	}
	if got := aptBlocked([]string{"php-cli", "php8.3-cli"}, []string{"php-cli"}, []string{"php-cli"}); len(got) != 0 {
		t.Fatalf("an automatic dependency should go: %v", got)
	}
}

func TestEveryToolHasMacOSRoute(t *testing.T) {
	for _, tool := range All {
		// Linux container and Kubernetes runtimes have no macOS version.
		linuxOnly := []string{"swift", "buildah", "containerd", "nerdctl", "crictl", "k3s", "k0s"}
		if !slices.Contains(linuxOnly, tool.Key) && !tool.Portable && len(tool.Brew) == 0 {
			t.Errorf("%s does not install on macOS", tool.Key)
		}
	}
}

func TestAssetWithoutHardcodedArchitecture(t *testing.T) {
	for _, tool := range All {
		for _, name := range tool.RemovePaths {
			if strings.Contains(name, "arm64") || strings.Contains(name, "aarch64") {
				t.Errorf("%s: hardcoded architecture in %s", tool.Key, name)
			}
		}
	}
	r := Release{Map: map[string]string{"amd64": "64bit", "arm64": "ARM64"}}
	got := assetVars(r, "v1.2", "1.2", platform.Info{Arch: "amd64"}).Replace("a-{ver}-{arch}-{uarch}-{xarch}-{x64}-{m}")
	if got != "a-1.2-amd64-x86_64-x86_64-x64-64bit" {
		t.Fatalf("got %s", got)
	}
	got = assetVars(r, "kustomize/v5", "5", platform.Info{Arch: "arm64"}).Replace("{shorttag}-{arch}-{uarch}-{xarch}-{x64}-{m}")
	if got != "v5-arm64-aarch64-arm64-arm64-ARM64" {
		t.Fatalf("got %s", got)
	}
}

func withPlatform(t *testing.T, p platform.Info) {
	t.Helper()
	saved := platform.Current
	platform.Current = p
	t.Cleanup(func() { platform.Current = saved })
}

func TestRequirementsPerMachine(t *testing.T) {
	desktop := platform.Info{OS: "linux", Arch: "amd64", Manager: platform.Dnf, Containers: true, Virtualization: true, GUI: true, Libc: "glibc"}
	server := desktop
	server.GUI, server.Virtualization = false, false
	proot := platform.Info{OS: "linux", Arch: "arm64", Manager: platform.Apt, Proot: true, Libc: "glibc"}
	alpine := platform.Info{OS: "linux", Arch: "amd64", Manager: platform.Apk, Containers: true, Libc: "musl"}
	mac := platform.Info{OS: "darwin", Arch: "arm64", Manager: platform.Brew, Containers: true, Virtualization: true, GUI: true}

	cases := []struct {
		machine platform.Info
		tool    string
		runs    bool
	}{
		{desktop, "docker", true}, {desktop, "rancher-desktop", true}, {desktop, "orbstack", false}, {desktop, "docker-desktop", false},
		{server, "docker", true}, {server, "rancher-desktop", false}, {server, "colima", false},
		{proot, "docker", false}, {proot, "podman", false}, {proot, "kind", false}, {proot, "docker-cli", true}, {proot, "kubectl", true}, {proot, "rancher", true},
		{alpine, "node", false}, {alpine, "aws", false}, {alpine, "kubectl", true}, {alpine, "docker", true},
		{mac, "orbstack", true}, {mac, "docker-desktop", true}, {mac, "containerd", false}, {mac, "node", true},
		{platform.Info{OS: "linux", Arch: "arm64", Manager: platform.Apt, Containers: true, GUI: true, Virtualization: true, Libc: "glibc"}, "rancher-desktop", false},
	}
	for _, c := range cases {
		withPlatform(t, c.machine)
		tool := Find(c.tool)
		if tool == nil {
			t.Fatalf("%s does not exist", c.tool)
		}
		if why := tool.Unavailable(); (why == "") != c.runs {
			t.Errorf("%s on %+v: reason %q, want runs=%v", c.tool, c.machine, why, c.runs)
		}
	}
	withPlatform(t, platform.Info{OS: "linux", Arch: "arm64", Manager: platform.Apt, Proot: true, Libc: "glibc"})
	if why := Find("orbstack").Unavailable(); !strings.Contains(why, "macOS") {
		t.Errorf("OrbStack on Linux should mention the OS: %q", why)
	}
	withPlatform(t, desktop)
	if got := strings.Join(Find("docker").pkgsFor(platform.Dnf), ","); got != "moby-engine,docker-compose" {
		t.Fatalf("wrong packages: %s", got)
	}
}

func TestGroupBringsOnlyWhatRuns(t *testing.T) {
	withPlatform(t, platform.Info{OS: "linux", Arch: "arm64", Manager: platform.Apt, Proot: true, Libc: "glibc"})
	got, err := Resolve([]string{"containers"})
	if err != nil {
		t.Fatal(err)
	}
	list := "," + keys(got) + ","
	if strings.Contains(list, ",docker,") || strings.Contains(list, ",podman,") || !strings.Contains(list, ",docker-cli,") {
		t.Fatalf("got %s", list)
	}
	// Requested by name, it passes and is refused at install time, with the reason.
	named, err := Resolve([]string{"docker"})
	if err != nil || keys(named) != "docker" {
		t.Fatalf("%s %v", keys(named), err)
	}
	// Without the binary on PATH: the machine running the test may have real Docker.
	t.Setenv("PATH", t.TempDir())
	if r := InstallOne(named[0], false); r.Action != "unavailable" || !strings.Contains(r.Error, "proot") {
		t.Fatalf("%+v", r)
	}
}

func TestRequirementsTablePointsToRealTools(t *testing.T) {
	for key := range requirements {
		if Find(key) == nil || Find(key).Key != key {
			t.Errorf("requirement for a nonexistent tool: %s", key)
		}
	}
}

func TestDescriptionsAndWarningsAreEnglish(t *testing.T) {
	defer i18n.Set(i18n.Lang())
	i18n.Set(i18n.EN)
	for _, tool := range All {
		for _, text := range []string{tool.Description, tool.Warning} {
			// Catalog texts are translated when the catalog is built, so in English none of
			// them may be a Portuguese sentence.
			if text != "" && portuguese.MatchString(text) {
				t.Errorf("%s: Portuguese text in the catalog: %q", tool.Key, text)
			}
		}
	}
}

var portuguese = regexp.MustCompile(`[áàâãéêíóôõúç]| (de|do|da|para|com|sem|em|e) `)

func fakeBin(t *testing.T, dir, name, script string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestSlowToolDoesNotHoldList(t *testing.T) {
	dir := t.TempDir()
	fakeBin(t, dir, "fast", "echo fast 1.0")
	fakeBin(t, dir, "slow", "sleep 30")
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	t.Setenv("HOME", t.TempDir())

	savedAll, savedBudget := All, statusBudget
	defer func() { All, statusBudget = savedAll, savedBudget }()
	All = []*Tool{
		{Key: "fast", Name: "Fast", Version: []string{"fast", "--version"}},
		{Key: "slow", Name: "Slow", Version: []string{"slow", "--version"}},
		{Key: "missing", Name: "Missing", Version: []string{"really-does-not-exist", "--version"}},
	}
	statusBudget = 300 * time.Millisecond

	began := time.Now()
	var calls int
	got := Statuses(func(done, total int) { calls++ })
	if elapsed := time.Since(began); elapsed > 3*time.Second {
		t.Fatalf("took %s", elapsed)
	}
	byKey := map[string]Status{}
	for _, s := range got {
		byKey[s.Key] = s
	}
	if s := byKey["fast"]; !s.Installed || s.Version != "fast 1.0" {
		t.Errorf("%+v", s)
	}
	if s := byKey["slow"]; !s.Installed || s.Version != "?" {
		t.Errorf("%+v", s)
	}
	if s := byKey["missing"]; s.Installed {
		t.Errorf("%+v", s)
	}
	if calls != 2 {
		t.Errorf("progress called %d times", calls)
	}
}

func TestCacheRerunsOnlyWhenBinaryChanges(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	fakeBin(t, dir, "tool", "echo x >> "+log+"; echo v1")
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	home := t.TempDir()
	cache := &versionCache{entries: map[string]cachedVersion{}}
	tool := &Tool{Key: "tool", Version: []string{"tool"}}
	_ = home

	for i := 0; i < 3; i++ {
		if v, ok := tool.cachedDetect(cache); !ok || v != "v1" {
			t.Fatalf("%q %v", v, ok)
		}
	}
	calls, _ := os.ReadFile(log)
	if strings.Count(string(calls), "x") != 1 {
		t.Fatalf("ran %d times", strings.Count(string(calls), "x"))
	}
	bin := filepath.Join(dir, "tool")
	future := time.Now().Add(time.Hour)
	fakeBin(t, dir, "tool", "echo v2")
	os.Chtimes(bin, future, future)
	if v, _ := tool.cachedDetect(cache); v != "v2" {
		t.Fatalf("after replacing the binary: %q", v)
	}
	os.Remove(bin)
	if _, ok := tool.cachedDetect(cache); ok {
		t.Fatal("a removed binary should not count as installed")
	}
}

func TestCategoryNames(t *testing.T) {
	for name, want := range map[string]string{
		"db": GroupDB, "database clients": GroupDB, "Database-Clients": GroupDB, "DATABASE": GroupDB, "banco de dados": GroupDB, "Banco-de-Dados": GroupDB, "BANCO": GroupDB, "Database clients": GroupDB,
		"linguagens": GroupSDK, "Linguagens de Programação": GroupSDK, "sdk": GroupSDK, "segurança": GroupSec, "k8s": GroupK8s,
	} {
		if got, ok := FindGroup(name); !ok || got != want {
			t.Errorf("%q: %q %v", name, got, ok)
		}
	}
	if _, ok := FindGroup("made-up"); ok {
		t.Error("made-up category")
	}
	if GroupTitle(GroupDB) == GroupDB {
		t.Error("no title")
	}
}

func TestBinaryThatIsNotOurs(t *testing.T) {
	ours := []string{"/home/a/.local/bin", "/home/a/.local/share/romanctl/tools"}
	cases := []struct {
		path, goos string
		managed    bool
		want       bool
	}{
		{"/usr/bin/jq", "darwin", false, true},
		{"/usr/bin/kubectl", "linux", false, true},            // the runner already had it
		{"/home/a/.local/bin/kubectl", "linux", false, false}, // ours is still there
		{"/usr/bin/jq", "linux", true, false},                 // apt removed it and it remained: failure
		{termuxPrefix + "/bin/jq", "linux", true, true},
		{"/home/a/.local/share/romanctl/tools/go/bin/go", "linux", false, false},
	}
	for _, c := range cases {
		if got := isExternal(c.path, c.goos, c.managed, ours); got != c.want {
			t.Errorf("%s on %s (managed=%v): %v", c.path, c.goos, c.managed, got)
		}
	}
}

func TestDownloadVerifiesChecksum(t *testing.T) {
	const body = "content"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(body))
	}))
	defer srv.Close()
	sum := sha256.Sum256([]byte(body))

	if _, err := download(srv.URL+"/a", t.TempDir(), hex.EncodeToString(sum[:])); err != nil {
		t.Fatalf("the right checksum failed: %v", err)
	}
	if _, err := download(srv.URL+"/a", t.TempDir(), strings.Repeat("0", 64)); err == nil {
		t.Fatal("accepted a wrong checksum")
	}
	if _, err := download(srv.URL+"/missing", t.TempDir(), ""); err == nil {
		t.Fatal("accepted HTTP 404")
	}
}

func TestSumFromFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("aaa  other.tar.gz\nbbb *gh_linux_arm64.tar.gz\n"))
	}))
	defer srv.Close()
	got, err := sumFromFile(srv.URL, "gh_linux_arm64.tar.gz")
	if err != nil || got != "bbb" {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := sumFromFile(srv.URL, "not-there"); err == nil {
		t.Fatal("should fail when the file is not in the list")
	}
}

func TestExtractZipRejectsPathEscapingDestination(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "evil.zip")
	f, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	entry, _ := w.Create("../outside.txt")
	entry.Write([]byte("x"))
	w.Close()
	f.Close()

	dest := filepath.Join(dir, "destination")
	if err := extractZip(file, dest); err == nil {
		t.Fatal("accepted an unsafe path")
	}
	if _, err := os.Stat(filepath.Join(dir, "outside.txt")); err == nil {
		t.Fatal("wrote outside the destination")
	}
}
