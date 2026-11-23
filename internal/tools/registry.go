// Package tools installs development tools on Linux and macOS.
package tools

import (
	"context"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/ui"
)

const (
	GroupCloud  = "cloud"
	GroupK8s    = "k8s"
	GroupIaC    = "iac"
	GroupImages = "containers"
	GroupSec    = "security"
	GroupDB     = "db"
	GroupObs    = "observability"
	GroupNet    = "network"
	GroupGit    = "git"
	GroupBuild  = "build"
	GroupShell  = "shell"
	GroupSDK    = "sdk"
	GroupNode   = "node-tools"
	GroupPython = "python-tools"
	GroupAI     = "ai"
)

// Groups defines the order and titles of the listing.
var Groups = []struct{ Key, Title string }{
	{GroupSDK, i18n.T("Programming languages and SDKs")}, {GroupNode, i18n.T("Node.js tools")}, {GroupPython, i18n.T("Python tools")},
	{GroupBuild, i18n.T("Build and automation")}, {GroupGit, i18n.T("Git and GitHub")}, {GroupCloud, i18n.T("Cloud")}, {GroupK8s, "Kubernetes"},
	{GroupIaC, i18n.T("Infrastructure as code")}, {GroupImages, i18n.T("Containers and images")}, {GroupSec, i18n.T("Security")},
	{GroupDB, i18n.T("Database clients")}, {GroupObs, i18n.T("Observability and load")}, {GroupNet, i18n.T("Network")},
	{GroupShell, i18n.T("Terminal and productivity")}, {GroupAI, i18n.T("AI")},
}

var groupAliases = map[string]string{
	"nuvem": GroupCloud, "kubernetes": GroupK8s, "sdks": GroupSDK, "linguagens": GroupSDK, "linguagem": GroupSDK,
	"languages": GroupSDK, "programacao": GroupSDK, "linguagens-de-programacao": GroupSDK, "programming-languages": GroupSDK,
	"infra": GroupIaC, "infraestrutura": GroupIaC, "docker-tools": GroupImages, "imagens": GroupImages, "seguranca": GroupSec,
	"banco": GroupDB, "bancos": GroupDB, "banco-de-dados": GroupDB, "bancos-de-dados": GroupDB, "database": GroupDB, "databases": GroupDB,
	"monitoramento": GroupObs, "observabilidade": GroupObs, "rede": GroupNet, "redes": GroupNet, "net": GroupNet, "terminal": GroupShell, "cli": GroupShell,
	"node": "", "ia": GroupAI,
}

// Tool is an installable tool. Detect replaces the version check when the binary
// is shared, like dotnet with several SDKs.
type Tool struct {
	Key         string
	Name        string
	Group       string
	Description string
	Aliases     []string
	Version     []string
	VersionLine string // the version line starts like this, when it is not the first line of output
	Detect      func() (string, bool)
	Install     func() error        // custom installer, used on Linux
	Portable    bool                // Install also works on macOS (npm, venv etc.)
	Pkgs        map[string][]string // packages per manager; "*" applies to all
	RemovePkgs  map[string][]string // what to remove, if different from Pkgs
	Brew        []string            // Homebrew formulas on macOS
	Cask        bool                // Brew entries are casks
	PostInstall func() error
	ProotReason string   // why it does not run under proot
	Arches      []string // empty means all
	Next        string   // what to do after installing
	Warning     string   // what else stops working on removal
	RemoveNpm   string   // global npm package to remove
	RemovePaths []string
	AfterRemove func() error
	Repo        string // GitHub project: README and page
	Npm         string // npm package, when it comes from there
	PyPI        string // PyPI package, when it comes from there
	Release     bool   // the installer downloads Repo's releases
	latest      func() (string, error)
	versions    func() ([]Version, error) // versions installAt knows how to install
	installAt   func(version string) error
	venvDest    string // dedicated Python environment, to install pkg==version
	venvBins    []string
}

func (t *Tool) aka(aliases ...string) *Tool { t.Aliases = aliases; return t }
func (t *Tool) ver(args ...string) *Tool    { t.Version = args; return t }
func (t *Tool) warn(w string) *Tool         { t.Warning = w; return t }

// on replaces a manager's packages; with no names, the tool does not exist there.
func (t *Tool) on(manager string, pkgs ...string) *Tool {
	if t.Pkgs == nil {
		t.Pkgs = map[string][]string{}
	}
	t.Pkgs[manager] = pkgs
	return t
}

func (t *Tool) brew(formulas ...string) *Tool { t.Brew = formulas; return t }
func (t *Tool) cask(casks ...string) *Tool    { t.Brew, t.Cask = casks, true; return t }
func (t *Tool) noBrew() *Tool                 { t.Brew = nil; return t }
func (t *Tool) next(text string) *Tool        { t.Next = text; return t }
func (t *Tool) only(arches ...string) *Tool   { t.Arches = arches; return t }

// Unavailable explains why the tool cannot be installed on this machine.
func (t *Tool) Unavailable() string {
	// The OS comes first: "only exists for macOS" explains more than a missing capability.
	if why := t.unsupportedSystem(); why != "" {
		return why
	}
	p := platform.Current
	for _, need := range requirements[t.Key] {
		// glibc only matters where Linux installs the official binary.
		if need == platform.NeedGlibc && p.OS == "darwin" {
			continue
		}
		if why := p.Missing(need); why != "" {
			return why
		}
	}
	return ""
}

func (t *Tool) unsupportedSystem() string {
	p := platform.Current
	if len(t.Arches) > 0 && !slices.Contains(t.Arches, p.Arch) {
		return i18n.Sprintf("there is no version for %s", p.Arch)
	}
	if p.OS == "darwin" {
		if t.Portable {
			return ""
		}
		if len(t.Brew) == 0 {
			return i18n.T("only exists for Linux")
		}
		if p.Manager != platform.Brew {
			return i18n.T("needs Homebrew: https://brew.sh")
		}
		return ""
	}
	if t.Install != nil {
		return ""
	}
	if len(t.Pkgs) == 0 {
		return i18n.T("only exists for macOS")
	}
	if len(t.pkgsFor(p.Manager)) == 0 {
		if p.Manager == "" {
			return i18n.T("no recognized package manager on this machine")
		}
		return i18n.Sprintf("there is no package for %s", p.Manager)
	}
	return ""
}

func (t *Tool) pkgsFor(manager string) []string {
	if pkgs, ok := t.Pkgs[manager]; ok {
		return pkgs
	}
	return t.Pkgs["*"]
}

func (t *Tool) removePkgsFor(manager string) []string {
	if t.RemovePkgs != nil {
		if pkgs, ok := t.RemovePkgs[manager]; ok {
			return pkgs
		}
		return t.RemovePkgs["*"]
	}
	return t.pkgsFor(manager)
}

// runInstall picks the right installation route for this machine.
func (t *Tool) runInstall(version string) error {
	p := platform.Current
	var err error
	switch {
	case version != "":
		err = t.installVersion(version)
	case p.OS == "darwin" && !t.Portable:
		err = brewInstall(t.Brew, t.Cask)
	case t.Install != nil:
		err = t.Install()
	default:
		err = pkgInstall(t.pkgsFor(p.Manager))
	}
	if err == nil && t.PostInstall != nil {
		err = t.PostInstall()
	}
	return err
}

// All is the full catalog, defined in the catalog_*.go files.
var All = slices.Concat(catalogSDK(), catalogDev(), catalogInfra(), catalogUtils())

// Find looks up a tool by name or alias.
func Find(name string) *Tool {
	name = strings.ToLower(name)
	for _, t := range All {
		if t.Key == name || slices.Contains(t.Aliases, name) {
			return t
		}
	}
	return nil
}

func isGroup(name string) (string, bool) {
	// Accepts "Database clients", "database-clients" and "DATABASE", plus the Portuguese titles.
	name = strings.Join(strings.Fields(strings.ReplaceAll(ui.Fold(name), "-", " ")), "-")
	if alias, ok := groupAliases[name]; ok && alias != "" {
		return alias, true
	}
	for _, g := range Groups {
		title := strings.Join(strings.Fields(ui.Fold(g.Title)), "-")
		if g.Key == name || title == name {
			return g.Key, true
		}
	}
	return "", false
}

// FindGroup translates a category name into its key.
func FindGroup(name string) (string, bool) { return isGroup(name) }

// GroupTitle returns a category's name.
func GroupTitle(key string) string {
	for _, g := range Groups {
		if g.Key == key {
			return g.Title
		}
	}
	return key
}

// GroupKeys lists the tools of a group or of "all"; nil if name is not a group.
// A tool name beats a group: "node" is the tool.
func GroupKeys(name string) []string {
	var keys []string
	if strings.EqualFold(name, "all") || strings.EqualFold(name, "tudo") {
		for _, t := range All {
			keys = append(keys, t.Key)
		}
		return keys
	}
	if Find(name) != nil {
		return nil
	}
	group, ok := isGroup(name)
	if !ok {
		return nil
	}
	for _, t := range All {
		if t.Group == group {
			keys = append(keys, t.Key)
		}
	}
	return keys
}

// Resolve translates names, groups and "all" into tools, without repeats. Groups bring only
// what runs on this machine; a name requested directly passes, and installation explains the refusal.
func Resolve(names []string) ([]*Tool, error) {
	var out []*Tool
	add := func(t *Tool) {
		if !slices.Contains(out, t) {
			out = append(out, t)
		}
	}
	for _, name := range names {
		if strings.EqualFold(name, "all") || strings.EqualFold(name, "tudo") {
			for _, t := range All {
				if t.Unavailable() == "" {
					add(t)
				}
			}
			continue
		}
		// A tool name beats a group: "node" is the tool.
		if t := Find(name); t != nil {
			add(t)
			continue
		}
		if group, ok := isGroup(name); ok {
			for _, t := range All {
				if t.Group == group && t.Unavailable() == "" {
					add(t)
				}
			}
			continue
		}
		return nil, ui.Usagef("unknown tool or group %q (see: romanctl tools list)", name)
	}
	return out, nil
}

// Detected reports the installed version. The tool only counts as installed if it runs.
func (t *Tool) Detected() (string, bool) {
	if t.Detect != nil {
		return t.Detect()
	}
	if _, err := exec.LookPath(t.Version[0]); err != nil {
		return "", false
	}
	out, err := versionOutput(t.Version...)
	if err != nil {
		return "", false
	}
	return versionLine(out, t.VersionLine), true
}

// versionOutput runs a version command with a timeout. WaitDelay covers programs that leave a
// child process holding the output open: without it, the wait would exceed the timeout.
func versionOutput(command ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.WaitDelay = 2 * time.Second
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func versionLine(out, prefix string) string {
	if prefix == "" {
		return firstLine(out)
	}
	for _, line := range strings.Split(ansi.Strip(out), "\n") {
		if line = strings.TrimSpace(line); strings.HasPrefix(line, prefix) {
			return line
		}
	}
	return firstLine(out)
}

func firstLine(s string) string {
	for _, line := range strings.Split(ansi.Strip(s), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "WARNING") {
			return strings.Join(strings.Fields(strings.TrimPrefix(line, "info: ")), " ")
		}
	}
	return ""
}

// Status is a tool's state, on screen or in --json.
type Status struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Group       string `json:"group"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
	Version     string `json:"version"`
	Latest      string `json:"latest,omitempty"` // the latest known, stored less than 6 hours ago
	Unavailable string `json:"unavailable,omitempty"`
}

// Statuses checks the tools in parallel. The version comes from the cache while the binary is
// unchanged, because running 170 programs on every listing is what froze the screen.
func Statuses(progress func(done, total int)) []Status {
	cache := loadVersionCache()
	type finished struct {
		index     int
		version   string
		installed bool
	}
	results := make(chan finished, len(All))
	slots := make(chan struct{}, 10)
	for i, t := range All {
		go func() {
			slots <- struct{}{}
			defer func() { <-slots }()
			version, installed := t.cachedDetect(cache)
			results <- finished{i, version, installed}
		}()
	}
	out := make([]Status, len(All))
	seen := make([]bool, len(All))
	deadline := time.After(statusBudget)
	done := 0
collect:
	for done < len(All) {
		select {
		case r := <-results:
			t := All[r.index]
			latest, _ := CachedLatest(t.Key)
			out[r.index] = Status{Key: t.Key, Name: t.Name, Group: t.Group, Description: i18n.T(t.Description), Installed: r.installed,
				Version: r.version, Latest: latest.Latest, Unavailable: t.Unavailable()}
			seen[r.index] = true
			done++
			if progress != nil {
				progress(done, len(All))
			}
		case <-deadline:
			break collect
		}
	}
	// A tool that does not respond must not hold up the whole list.
	for i, t := range All {
		if !seen[i] {
			installed := false
			if len(t.Version) > 0 {
				_, err := exec.LookPath(t.Version[0])
				installed = err == nil
			}
			out[i] = Status{Key: t.Key, Name: t.Name, Group: t.Group, Description: i18n.T(t.Description), Installed: installed, Version: "?", Unavailable: t.Unavailable()}
		}
	}
	cache.save()
	return out
}

// statusBudget is the longest a listing waits.
var statusBudget = 12 * time.Second

// Result is the outcome of installing a tool.
type Result struct {
	Key     string `json:"key"`
	Action  string `json:"action"` // installed, skipped or failed
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

// InstallOne installs and confirms the tool really runs afterwards.
func InstallOne(t *Tool, force bool) Result { return InstallVersion(t, "", force) }

// InstallVersion installs a chosen version; empty means the latest. Asking for a version other than
// the installed one is a reinstall, with no need for force.
func InstallVersion(t *Tool, want string, force bool) Result {
	installed, ok := t.Detected()
	if ok && !force && (want == "" || UpToDate(installed, want)) {
		ui.Okf("%s is already installed: %s", t.Name, installed)
		return Result{Key: t.Key, Action: "skipped", Version: installed}
	}
	if why := t.Unavailable(); why != "" {
		ui.Failf("%s: %s", t.Name, why)
		return Result{Key: t.Key, Action: "unavailable", Error: why}
	}
	chosen := ""
	if want != "" {
		chosen = t.resolveVersion(want)
		ui.Infof("%s %s, chosen version", t.Name, strings.TrimPrefix(want, "v"))
	} else if u := t.Upstream(false); u.Latest != "" {
		ui.Infof("%s %s, via %s", t.Name, u.Latest, u.Source)
	}
	var err0 error
	ui.Spin(i18n.Sprintf("Installing %s...", t.Name), func(func(string)) { err0 = t.runInstall(chosen) })
	if err := err0; err != nil {
		ui.Failf("%s: %v", t.Name, err)
		return Result{Key: t.Key, Action: "failed", Error: err.Error()}
	}
	version, ok := t.Detected()
	if !ok {
		err := i18n.Sprintf("%s was installed but did not run", t.Name)
		ui.Failf("%s", err)
		// Does not leave behind something that does not work, sometimes gigabytes of it.
		if rmErr := t.remove(); rmErr != nil {
			ui.Warnf("%v", rmErr)
		}
		return Result{Key: t.Key, Action: "failed", Error: err}
	}
	ui.Okf("%s installed: %s", t.Name, version)
	if want != "" && !UpToDate(version, want) {
		ui.Warnf("asked for %s, but the binary reports %s: another %s may come first in PATH", want, version, t.Name)
	}
	if t.Next != "" {
		ui.Infof("Next step: %s", t.Next)
	}
	return Result{Key: t.Key, Action: "installed", Version: version}
}

// isExternal says whether the leftover binary is not ours: Termux's, the ones shipped with macOS, like
// jq in /usr/bin, and the ones the machine already had, like kubectl on GitHub runners. A tool
// installed by the package manager that is still in /usr/bin is a real failure.
func isExternal(path, goos string, managedByPackages bool, ours []string) bool {
	if strings.HasPrefix(path, termuxPrefix) {
		return true
	}
	if goos == "darwin" && (strings.HasPrefix(path, "/usr/bin/") || strings.HasPrefix(path, "/bin/")) {
		return true
	}
	if managedByPackages {
		return false
	}
	for _, dir := range ours {
		if dir != "" && strings.HasPrefix(path, dir+"/") {
			return false
		}
	}
	return true
}

// UninstallOne removes what the tool's declaration says and checks the result.
func UninstallOne(t *Tool) Result {
	if _, ok := t.Detected(); !ok && !hasLeftovers(t) {
		ui.Infof("%s is not installed", t.Name)
		return Result{Key: t.Key, Action: "skipped"}
	}
	var rmErr0 error
	ui.Spin(i18n.Sprintf("Removing %s...", t.Name), func(func(string)) { rmErr0 = t.remove() })
	if err := rmErr0; err != nil {
		ui.Failf("%s: %v", t.Name, err)
		return Result{Key: t.Key, Action: "failed", Error: err.Error()}
	}
	if _, still := t.Detected(); still {
		where := t.Key
		if len(t.Version) > 0 {
			if p, err := exec.LookPath(t.Version[0]); err == nil {
				where = p
			}
		}
		p := platform.Current
		managed := p.OS != "darwin" && len(t.removePkgsFor(p.Manager)) > 0
		if isExternal(where, p.OS, managed, []string{p.BinDir, p.OptDir}) {
			ui.Warnf("%s is still available from the system, outside romanctl: %s", t.Name, where)
			return Result{Key: t.Key, Action: "external", Version: where}
		}
		err := i18n.Sprintf("%s is still installed in %s", t.Name, where)
		ui.Failf("%s", err)
		return Result{Key: t.Key, Action: "failed", Error: err}
	}
	ui.Okf("%s removed", t.Name)
	return Result{Key: t.Key, Action: "removed"}
}
