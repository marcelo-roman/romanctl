package tui

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"romanctl/internal/sysinfo"
	"romanctl/internal/tools"
)

type fakeBackend struct {
	commands  [][]string
	upstreams []string
}

func (f *fakeBackend) Header() string { return "linux/arm64" }
func (f *fakeBackend) Services() []Row {
	return []Row{{Key: "postgres", Status: "stopped", Columns: []string{"postgres", "PostgreSQL", "stopped", "pg://"}},
		{Key: "redis", Status: "running", Columns: []string{"redis", "Redis", "running", "redis://"}},
		{Key: "kafka", Status: "not-installed", Columns: []string{"kafka", "Kafka", "not installed", ""}},
		{Key: "docker", Status: "running", Group: engineGroup, Columns: []string{"docker", "Docker", "running", "unix:///var/run/docker.sock"}}}
}
func (f *fakeBackend) Projects() []Row {
	return []Row{{Key: "shop", Status: "stopped", Columns: []string{"shop", "dotnet", "stopped", "http://x", "01/01"}, Detail: [][2]string{{"Folder", "/p/shop"}}},
		{Key: "site", Status: "running", Columns: []string{"site", "next", "running", "http://y", "01/01"}}}
}
func (f *fakeBackend) Tools() []Row {
	return []Row{{Key: "kubectl", Status: "installed", Group: "k8s", Columns: []string{"kubectl", "k8s", "installed", "v1.0.0", "", "CLI"}},
		{Key: "helm", Status: "available", Group: "k8s", Columns: []string{"helm", "k8s", "available", "", "", "Packages"}},
		{Key: "psql", Status: "available", Group: "db", Columns: []string{"psql", "db", "available", "", "", "Client"}}}
}
func (f *fakeBackend) System() sysinfo.Snapshot {
	return sysinfo.Snapshot{Memory: sysinfo.Memory{Total: 1 << 33, Available: 1 << 32}, Disk: sysinfo.Disk{Size: 1 << 34, Free: 1 << 33},
		Battery: &sysinfo.Battery{Percent: 80, Power: 3.5, HoursLeft: 2.5}, Top: []sysinfo.Proc{{PID: 7, Name: "claude", CPU: 12, Memory: 1 << 28}}}
}
func (f *fakeBackend) Upstream(key string) tools.Upstream {
	f.upstreams = append(f.upstreams, key)
	return tools.Upstream{Latest: "2.0.0", Source: "GitHub releases", URL: "https://github.com/x/" + key, Readme: "# " + key + "\n\nUsage guide for " + key + "."}
}
func (f *fakeBackend) Latest(key string) tools.Upstream {
	return tools.Upstream{Latest: "2.0.0", Source: "GitHub releases"}
}
func (f *fakeBackend) Versions(key string) ([]tools.Version, error) {
	if key == "psql" {
		return nil, errors.New("apt only installs the current version from the repository")
	}
	return []tools.Version{{ID: "v2.0.0", Label: "2.0.0"}, {ID: "v1.9.0", Label: "1.9.0"}, {ID: "v1.0.0", Label: "1.0.0"}}, nil
}
func (f *fakeBackend) Logs(view, key string) string { return "line 1\nline 2\n" + key }
func (f *fakeBackend) Command(args ...string) *exec.Cmd {
	f.commands = append(f.commands, args)
	return exec.Command("true")
}

func setup(t *testing.T) (*Model, *fakeBackend) {
	t.Helper()
	f := &fakeBackend{}
	m := New(f)
	m.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	load(m, viewProjects)
	return m, f
}

// load runs the loading synchronously.
func load(m *Model, view string) {
	if msg := m.load(view)(); msg != nil {
		m.Update(msg)
	}
}

func key(m *Model, keys ...string) tea.Cmd {
	var last tea.Cmd
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "right":
			msg = tea.KeyMsg{Type: tea.KeyRight}
		case "left":
			msg = tea.KeyMsg{Type: tea.KeyLeft}
		case "backspace":
			msg = tea.KeyMsg{Type: tea.KeyBackspace}
		case "ctrl+d":
			msg = tea.KeyMsg{Type: tea.KeyCtrlD}
		case "ctrl+c":
			msg = tea.KeyMsg{Type: tea.KeyCtrlC}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		_, last = m.Update(msg)
	}
	return last
}

func screen(m *Model) string { return ansi.Strip(m.View()) }

func TestSwitchScreen(t *testing.T) {
	m, _ := setup(t)
	if !strings.Contains(screen(m), "shop") {
		t.Fatal(screen(m))
	}
	key(m, "tab")
	if m.view != viewServices {
		t.Fatal(m.view)
	}
	load(m, viewServices)
	if !strings.Contains(screen(m), "postgres") {
		t.Fatal(screen(m))
	}
	key(m, "3")
	if m.view != viewTools {
		t.Fatal(m.view)
	}
	key(m, "1")
	if m.view != viewProjects {
		t.Fatal(m.view)
	}
}

func TestCommands(t *testing.T) {
	m, _ := setup(t)
	for text, want := range map[string]string{"svc": viewServices, "tools": viewTools, "sys": viewSystem, "prj": viewProjects, "ser": viewServices} {
		key(m, ":", text, "enter")
		if m.view != want {
			t.Errorf(":%s went to %s", text, m.view)
		}
	}
	key(m, ":", "does-not-exist", "enter")
	if !strings.Contains(screen(m), "does-not-exist") {
		t.Fatal(screen(m))
	}
	key(m, ":", "s", "e", "tab")
	if m.command != "services" {
		t.Fatalf("complete: %q", m.command)
	}
	key(m, "esc")
	cmd := key(m, ":", "q", "enter")
	if !m.quit || cmd == nil {
		t.Fatal("did not quit")
	}
}

func TestCategoryCommand(t *testing.T) {
	m, _ := setup(t)
	key(m, ":", "db", "enter")
	if m.view != viewTools {
		t.Fatal(m.view)
	}
	load(m, viewTools)
	rows := m.visible()
	if len(rows) != 1 || rows[0].Key != "psql" {
		t.Fatalf("%+v", rows)
	}
}

func TestFilter(t *testing.T) {
	m, _ := setup(t)
	key(m, "/", "si")
	if rows := m.visible(); len(rows) != 1 || rows[0].Key != "site" {
		t.Fatalf("%+v", rows)
	}
	key(m, "enter")
	if m.mode != modeTable || m.filter != "si" {
		t.Fatalf("%v %q", m.mode, m.filter)
	}
	key(m, "esc")
	if m.filter != "" || len(m.visible()) != 2 {
		t.Fatal(m.filter)
	}
	key(m, "/", "zzz")
	if !strings.Contains(screen(m), "Nothing found") {
		t.Fatal(screen(m))
	}
}

func TestKeysArrivingTogether(t *testing.T) {
	m, _ := setup(t)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/si")})
	if m.mode != modeFilter || m.filter != "si" {
		t.Fatalf("mode %v, filter %q", m.mode, m.filter)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":svc")})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != viewServices {
		t.Fatal(m.view)
	}
}

func TestFilterLettersDoNotTriggerActions(t *testing.T) {
	m, f := setup(t)
	key(m, "/", "u", "r", "x", "q")
	if len(f.commands) != 0 || m.quit {
		t.Fatalf("%v %v", f.commands, m.quit)
	}
	key(m, "esc", ":", "u", "q")
	if m.quit || len(f.commands) != 0 {
		t.Fatal("the command does not trigger keys either")
	}
}

func TestActionRunsInBackgroundWithStateOnRow(t *testing.T) {
	m, f := setup(t)
	cmd := key(m, "u")
	if m.running["projects/shop"] == "" {
		t.Fatal("no progress")
	}
	if !strings.Contains(screen(m), "bring up") {
		t.Fatal(screen(m))
	}
	if cmd2 := key(m, "u"); cmd2 != nil {
		t.Fatal("repeating while running should not fire again")
	}
	msg := cmd()
	if got := strings.Join(f.commands[0], " "); got != "up shop -d" {
		t.Fatal(got)
	}
	m.Update(msg)
	if m.running["projects/shop"] != "" || !strings.Contains(screen(m), "done") {
		t.Fatal(screen(m))
	}
}

func TestFailureShowsInFooter(t *testing.T) {
	m, _ := setup(t)
	m.Update(doneMsg{id: "projects/shop", label: "up", err: exec.ErrNotFound, output: "x\n✘ port 5080 is in use\n"})
	if !strings.Contains(screen(m), "port 5080") {
		t.Fatal(screen(m))
	}
}

func TestActionsDependOnState(t *testing.T) {
	m, _ := setup(t)
	load(m, viewServices)
	key(m, "2")
	load(m, viewServices)
	view := screen(m)
	if !hasHint(view, "s", "start") || hasHint(view, "x", "stop") {
		t.Fatalf("stopped:\n%s", view)
	}
	key(m, "down")
	view = screen(m)
	if !hasHint(view, "x", "stop") || hasHint(view, "s", "start") {
		t.Fatalf("running:\n%s", view)
	}
	cmd := key(m, "s")
	if cmd != nil {
		t.Fatal("starting what is already running should be ignored")
	}
}

func TestRemoveAsksForConfirmation(t *testing.T) {
	m, f := setup(t)
	if cmd := key(m, "ctrl+d"); cmd != nil {
		t.Fatal("must not remove right away")
	}
	if m.mode != modeConfirm || !strings.Contains(screen(m), "shop") {
		t.Fatal(screen(m))
	}
	key(m, "n")
	if m.mode != modeTable || len(f.commands) != 0 {
		t.Fatal("n should cancel")
	}
	key(m, "ctrl+d")
	key(m, "enter")
	if len(f.commands) != 0 {
		t.Fatal("Enter does not confirm: the default is no")
	}
	key(m, "ctrl+d")
	cmd := key(m, "y")
	if cmd == nil {
		t.Fatal("y should confirm")
	}
	cmd()
	if got := strings.Join(f.commands[0], " "); got != "rm shop --yes" {
		t.Fatal(got)
	}
}

func TestEnterShowsActions(t *testing.T) {
	m, f := setup(t)
	key(m, "enter")
	if m.mode != modeActions || !strings.Contains(screen(m), "publish") {
		t.Fatal(screen(m))
	}
	key(m, "esc")
	if m.mode != modeTable {
		t.Fatal(m.mode)
	}
	key(m, "enter")
	cmd := key(m, "enter")
	if m.mode != modeTable || cmd == nil {
		t.Fatal("Enter should run the first action")
	}
	cmd()
	if got := strings.Join(f.commands[0], " "); got != "up shop -d" {
		t.Fatal(got)
	}
}

func TestCategories(t *testing.T) {
	m, _ := setup(t)
	key(m, "3")
	load(m, viewTools)
	if len(m.visible()) != 3 {
		t.Fatal(m.visible())
	}
	key(m, "right")
	if rows := m.visible(); len(rows) != 2 || rows[0].Group != "k8s" {
		t.Fatalf("%+v", rows)
	}
	key(m, "/", "helm", "enter")
	if rows := m.visible(); len(rows) != 1 {
		t.Fatalf("%+v", rows)
	}
	key(m, "esc", "left")
	if m.category != len(m.categories()) {
		t.Fatal(m.category)
	}
}

func TestLogs(t *testing.T) {
	m, _ := setup(t)
	cmd := key(m, "l")
	if m.mode != modeLogs {
		t.Fatal(m.mode)
	}
	m.Update(cmd())
	if !strings.Contains(screen(m), "line 2") {
		t.Fatal(screen(m))
	}
	key(m, "esc")
	if m.mode != modeTable {
		t.Fatal(m.mode)
	}
}

func TestDetailsAndHelp(t *testing.T) {
	m, _ := setup(t)
	key(m, "d")
	if !strings.Contains(screen(m), "/p/shop") {
		t.Fatal(screen(m))
	}
	key(m, "esc", "?")
	if !strings.Contains(screen(m), "NAVIGATE") {
		t.Fatal(screen(m))
	}
	key(m, "esc")
	if m.mode != modeTable {
		t.Fatal(m.mode)
	}
}

func TestSystem(t *testing.T) {
	m, _ := setup(t)
	key(m, "4")
	load(m, viewSystem)
	view := screen(m)
	for _, want := range []string{"Memory", "Battery", "80%", "claude", "#"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
}

func TestLongListFitsSmallScreen(t *testing.T) {
	m, _ := setup(t)
	var many []Row
	for i := 0; i < 200; i++ {
		many = append(many, Row{Key: "item", Status: "stopped", Columns: []string{"item", "t", "stopped", "a very long url here to overflow the width of the small phone screen", "x"}})
	}
	m.rows[viewProjects] = many
	for _, size := range [][2]int{{80, 24}, {40, 12}, {200, 60}} {
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		for i := 0; i < 120; i++ {
			key(m, "down")
		}
		lines := strings.Split(screen(m), "\n")
		if len(lines) != size[1] {
			t.Errorf("%dx%d: %d lines", size[0], size[1], len(lines))
		}
		for _, l := range lines {
			if w := len([]rune(l)); w > size[0] {
				t.Errorf("%dx%d: line of %d: %q", size[0], size[1], w, l)
			}
		}
	}
}

func TestNoIcons(t *testing.T) {
	m, _ := setup(t)
	for _, v := range []string{"1", "2", "3", "4", "?"} {
		key(m, v)
		load(m, m.view)
		for _, r := range screen(m) {
			if r >= 0x2190 && r != 0x2190 && r != 0x2192 && !isBoxDrawing(r) {
				t.Errorf("screen %s has the symbol %q", v, r)
			}
		}
		key(m, "esc")
	}
}

func TestCtrlCQuitsFromAnywhere(t *testing.T) {
	for _, keys := range [][]string{{"/"}, {":"}, {"?"}, {"d"}, {"l"}, {"enter"}, {"ctrl+d"}} {
		m, _ := setup(t)
		key(m, keys...)
		if cmd := key(m, "ctrl+c"); cmd == nil || !m.quit {
			t.Errorf("%v: did not quit", keys)
		}
	}
}

// The frame is not an icon: box drawing exists in every terminal.
func isBoxDrawing(r rune) bool { return r >= 0x2500 && r <= 0x257F }

func hasHint(view, key, label string) bool {
	return regexp.MustCompile(`<` + regexp.QuoteMeta(key) + `>\s+` + regexp.QuoteMeta(label) + `\b`).MatchString(view)
}

func toolsView(t *testing.T) (*Model, *fakeBackend) {
	m, f := setup(t)
	key(m, "3")
	load(m, viewTools)
	return m, f
}

func TestAvailableVersionInList(t *testing.T) {
	m, f := toolsView(t)
	cmd := m.fetchUpstream("kubectl")
	m.Update(cmd())
	if len(f.upstreams) != 1 || !regexp.MustCompile(`kubectl\s+k8s\s+installed\s+v1\.0\.0\s+2\.0\.0`).MatchString(screen(m)) {
		t.Fatal(screen(m))
	}
	if m.fetchUpstream("kubectl") != nil {
		t.Fatal("queried the same tool twice")
	}
}

func TestListAlwaysShowsLatest(t *testing.T) {
	m, _ := toolsView(t)
	if !strings.Contains(screen(m), "LATEST") || !strings.Contains(screen(m), "...") {
		t.Fatal(screen(m))
	}
	cmd := m.fetchLatest()
	if cmd != nil {
		t.Fatal("the list already asked for the latest on load; it does not ask again")
	}
	for _, r := range m.rows[viewTools] {
		m.Update(latestMsg{r.Key, tools.Upstream{Latest: "2.0.0"}})
	}
	if strings.Contains(screen(m), "...") || strings.Count(screen(m), "2.0.0") < 3 {
		t.Fatal(screen(m))
	}
}

func TestEnterChoosesAnotherVersion(t *testing.T) {
	m, f := toolsView(t)
	cmd := key(m, "enter")
	if m.mode != modeVersions || cmd == nil {
		t.Fatal("Enter in Tools opens the versions")
	}
	m.Update(cmd())
	view := screen(m)
	for _, want := range []string{"versions(kubectl)", "2.0.0", "latest", "1.0.0", "installed"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
	key(m, "down", "enter")
	if m.mode != modeConfirm || !strings.Contains(screen(m), "Install kubectl 1.9.0?") {
		t.Fatal(screen(m))
	}
	run := key(m, "y")
	run()
	if got := strings.Join(f.commands[0], " "); got != "tools install kubectl --version v1.9.0 --yes" {
		t.Fatal(got)
	}
}

func TestToolWithoutOtherVersions(t *testing.T) {
	m, _ := toolsView(t)
	key(m, "G")
	cmd := key(m, "enter")
	m.Update(cmd())
	if !strings.Contains(screen(m), "only installs the current version") {
		t.Fatal(screen(m))
	}
	if key(m, "enter"); m.mode != modeVersions {
		t.Fatal("without versions, Enter does nothing")
	}
	key(m, "esc")
	if m.mode != modeTable {
		t.Fatal(m.mode)
	}
}

func TestQueriesOnlyWhenCursorSettles(t *testing.T) {
	m, f := toolsView(t)
	key(m, "down", "down")
	if len(f.upstreams) != 0 {
		t.Fatal("moving the cursor must not hit the network")
	}
	m.Update(settleMsg{"helm"})
	if len(f.upstreams) != 0 {
		t.Fatal("helm is no longer selected")
	}
	cmd := func() tea.Cmd { _, c := m.Update(settleMsg{"psql"}); return c }()
	if cmd == nil {
		t.Fatal("psql is selected and should be queried")
	}
}

func TestInstallShowsVersionAndSource(t *testing.T) {
	m, f := toolsView(t)
	key(m, "down")
	cmd := key(m, "i")
	if m.mode != modeConfirm || cmd == nil {
		t.Fatal("install asks for confirmation and queries the version")
	}
	m.Update(cmd())
	view := screen(m)
	for _, want := range []string{"Install helm?", "2.0.0", "GitHub releases"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
	run := key(m, "y")
	run()
	if got := strings.Join(f.commands[0], " "); got != "tools install helm --yes" {
		t.Fatal(got)
	}
}

func TestToolDetailsHaveReadme(t *testing.T) {
	m, _ := toolsView(t)
	cmd := key(m, "d")
	m.Update(cmd())
	view := screen(m)
	for _, want := range []string{"README", "Usage guide for kubectl", "github.com/x/kubectl", "update available"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
}

func TestStoppingDockerWarnsAndUsesTerminal(t *testing.T) {
	m, f := setup(t)
	key(m, "2")
	load(m, viewServices)
	key(m, "G")
	if r, _ := m.selected(); r.Key != "docker" {
		t.Fatal(r.Key)
	}
	if !hasHint(screen(m), "x", "stop") {
		t.Fatal(screen(m))
	}
	if cmd := key(m, "x"); cmd != nil || m.mode != modeConfirm {
		t.Fatal("stopping Docker needs confirmation")
	}
	if !strings.Contains(screen(m), "every container") {
		t.Fatal(screen(m))
	}
	cmd := key(m, "y")
	if cmd == nil || len(f.commands) != 1 || strings.Join(f.commands[0], " ") != "stop docker --yes" {
		t.Fatalf("should take over the terminal with stop docker --yes: %v", f.commands)
	}
}
