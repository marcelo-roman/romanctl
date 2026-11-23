// Package tui is romanctl's full-screen panel, modeled after k9s.
package tui

import (
	"slices"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"romanctl/internal/i18n"
	"romanctl/internal/sysinfo"
	"romanctl/internal/tools"
	"romanctl/internal/ui"
)

const (
	viewServices = "services"
	viewProjects = "projects"
	viewTools    = "tools"
	viewSystem   = "system"
)

var viewOrder = []string{viewProjects, viewServices, viewTools, viewSystem}

func viewTitle(v string) string {
	switch v {
	case viewServices:
		return i18n.T("Services")
	case viewProjects:
		return i18n.T("Projects")
	case viewTools:
		return i18n.T("Tools")
	}
	return i18n.T("System")
}

func statusText(s string) string {
	switch s {
	case "running":
		return i18n.T("running")
	case "stopped":
		return i18n.T("stopped")
	case "crashed":
		return i18n.T("crashed")
	case "not-installed":
		return i18n.T("not installed")
	case "installed":
		return i18n.T("installed")
	case "available":
		return i18n.T("available")
	case "prepared":
		return i18n.T("services only")
	}
	return s
}

// action is something a key does to the selected row.
type action struct {
	key     string
	label   string
	args    func(key string) []string // romanctl command
	exec    bool                      // needs the terminal: the panel steps aside
	confirm string                    // asks first, with y/n
	when    func(r Row) bool

	question string // ready-made question, when the text depends on the row
	version  string // version chosen in the version list
}

// engineGroup marks the Docker row: starting and stopping the daemon needs sudo, so the action takes
// over the terminal, and stopping warns that it brings down every container on the machine.
const engineGroup = "engine"

const engineWarning = "Stopping Docker brings down every container on the machine, not only romanctl's. Stop %s?"

var actions = map[string][]action{
	viewServices: {
		{key: "s", label: "start", args: func(k string) []string { return []string{"start", k} }, when: both(container, not("running", "not-installed"))},
		{key: "x", label: "stop", args: func(k string) []string { return []string{"stop", k} }, when: both(container, is("running"))},
		{key: "r", label: "restart", args: func(k string) []string { return []string{"restart", k} }, when: both(container, is("running"))},
		{key: "s", label: "start", args: func(k string) []string { return []string{"start", k} }, exec: true, when: both(engine, not("running", "not-installed"))},
		{key: "x", label: "stop", args: func(k string) []string { return []string{"stop", k, "--yes"} }, exec: true, when: both(engine, is("running")), confirm: engineWarning},
		{key: "r", label: "restart", args: func(k string) []string { return []string{"restart", k, "--yes"} }, exec: true, when: both(engine, is("running")), confirm: engineWarning},
	},
	viewProjects: {
		{key: "u", label: "bring up (up -d)", args: func(k string) []string { return []string{"up", k, "-d"} }, when: not("running")},
		{key: "x", label: "bring down (down)", args: func(k string) []string { return []string{"down", k} }, when: is("running", "crashed", "prepared")},
		{key: "r", label: "run here", args: func(k string) []string { return []string{"run", k} }, exec: true},
		{key: "a", label: "add code", args: func(k string) []string { return []string{"add", "--project", k} }, exec: true},
		{key: "e", label: "edit (code-server)", args: func(k string) []string { return []string{"code", k} }},
		{key: "o", label: "open the app", args: func(k string) []string { return []string{"open", k, "--app"} }, when: is("running")},
		{key: "p", label: "publish", args: func(k string) []string { return []string{"publish", k} }, exec: true},
		{key: "ctrl+d", label: "remove", args: func(k string) []string { return []string{"rm", k, "--yes"} }, confirm: "Remove project %s and all its files?"},
	},
	viewTools: {
		{key: "i", label: "install", args: func(k string) []string { return []string{"tools", "install", k, "--yes"} }, when: is("available"), confirm: "Install %s?"},
		{key: "r", label: "reinstall", args: func(k string) []string { return []string{"tools", "install", k, "--yes", "--force"} }, when: is("installed"), confirm: "Reinstall %s?"},
		{key: "ctrl+d", label: "remove", args: func(k string) []string { return []string{"tools", "uninstall", k, "--yes"} }, when: is("installed"), confirm: "Remove tool %s?"},
	},
}

// Actions that apply to the screen, without needing a selected row.
var globalActions = map[string][]action{
	viewProjects: {{key: "n", label: "new project", args: func(string) []string { return []string{"new"} }, exec: true}},
	viewServices: {{key: "n", label: "install service", args: func(string) []string { return []string{"install"} }, exec: true}},
}

func engine(r Row) bool    { return r.Group == engineGroup }
func container(r Row) bool { return r.Group != engineGroup }

func both(a, b func(Row) bool) func(Row) bool { return func(r Row) bool { return a(r) && b(r) } }

func is(states ...string) func(Row) bool {
	return func(r Row) bool { return slices.Contains(states, r.Status) }
}

func not(states ...string) func(Row) bool {
	return func(r Row) bool { return !slices.Contains(states, r.Status) }
}

type mode int

const (
	modeTable mode = iota
	modeFilter
	modeCommand
	modeHelp
	modeDetail
	modeLogs
	modeActions
	modeConfirm
	modeVersions
)

// Model is the panel state.
type Model struct {
	backend Backend
	version string
	view    string
	mode    mode
	width   int
	height  int

	rows     map[string][]Row
	loaded   map[string]bool
	cursor   map[string]int
	offset   map[string]int
	filter   string
	command  string
	category int
	flash    string
	flashErr bool
	flashAt  time.Time
	running  map[string]string // view/key → what is in progress
	loading  bool
	spin     int

	system   *sysinfo.Snapshot
	logText  string
	logTitle string
	logView  string
	logKey   string
	logTop   int
	follow   bool

	pendingCategory string

	upstream  map[string]tools.Upstream
	latest    map[string]tools.Upstream
	asked     map[string]bool
	versions  map[string][]tools.Version
	verErr    map[string]string
	verKey    string
	verPos    int
	fetching  map[string]bool
	watched   string
	detailTop int
	readme    map[string][]string

	pending *action
	pendKey string
	menuPos int
	menu    []action
	quit    bool
}

// New creates the panel.
func New(b Backend) *Model {
	return &Model{backend: b, view: viewProjects, width: 100, height: 30,
		rows: map[string][]Row{}, loaded: map[string]bool{}, cursor: map[string]int{}, offset: map[string]int{}, running: map[string]string{},
		upstream: map[string]tools.Upstream{}, fetching: map[string]bool{}, readme: map[string][]string{},
		latest: map[string]tools.Upstream{}, asked: map[string]bool{}, versions: map[string][]tools.Version{}, verErr: map[string]string{}}
}

// ---------------------------------------------------------------- messages

type rowsMsg struct {
	view string
	rows []Row
}
type systemMsg struct{ snap sysinfo.Snapshot }
type logsMsg struct{ view, key, text string }
type tickMsg time.Time
type spinMsg time.Time
type doneMsg struct {
	id, label string
	err       error
	output    string
}
type execDoneMsg struct{ err error }
type upstreamMsg struct {
	key string
	up  tools.Upstream
}
type settleMsg struct{ key string }
type latestMsg struct {
	key string
	up  tools.Upstream
}
type versionsMsg struct {
	key  string
	list []tools.Version
	err  error
}

// fetchLatest asks for the latest version of each tool in the list, once per session; the backend
// limits how many hit the network at once and caches the result for a few hours.
func (m *Model) fetchLatest() tea.Cmd {
	var cmds []tea.Cmd
	b := m.backend
	for _, r := range m.rows[viewTools] {
		if m.asked[r.Key] {
			continue
		}
		m.asked[r.Key] = true
		key := r.Key
		cmds = append(cmds, func() tea.Msg { return latestMsg{key, b.Latest(key)} })
	}
	return tea.Batch(cmds...)
}

func (m *Model) openVersions(r Row) tea.Cmd {
	m.mode, m.verKey, m.verPos = modeVersions, r.Key, 0
	if _, ok := m.versions[r.Key]; ok {
		return nil
	}
	if _, failed := m.verErr[r.Key]; failed {
		return nil
	}
	b, key := m.backend, r.Key
	return func() tea.Msg {
		list, err := b.Versions(key)
		return versionsMsg{key, list, err}
	}
}

// fetchUpstream queries the tool's upstream once per session, in the background.
func (m *Model) fetchUpstream(key string) tea.Cmd {
	if _, ok := m.upstream[key]; ok || m.fetching[key] || key == "" {
		return nil
	}
	m.fetching[key] = true
	b := m.backend
	return func() tea.Msg { return upstreamMsg{key, b.Upstream(key)} }
}

// watchSelection waits for the cursor to settle on a tool before querying the network: moving
// down the list with j must not fire one query per row.
func (m *Model) watchSelection() tea.Cmd {
	if m.view != viewTools {
		return nil
	}
	r, ok := m.selected()
	if !ok || r.Key == m.watched {
		return nil
	}
	m.watched = r.Key
	key := r.Key
	return tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg { return settleMsg{key} })
}

func (m *Model) load(view string) tea.Cmd {
	b := m.backend
	switch view {
	case viewServices:
		return func() tea.Msg { return rowsMsg{viewServices, b.Services()} }
	case viewProjects:
		return func() tea.Msg { return rowsMsg{viewProjects, b.Projects()} }
	case viewTools:
		return func() tea.Msg { return rowsMsg{viewTools, b.Tools()} }
	case viewSystem:
		return func() tea.Msg { return systemMsg{b.System()} }
	}
	return nil
}

func tick() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func spinTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return spinMsg(t) })
}

// Init loads the first screen and turns on automatic refresh.
func (m *Model) Init() tea.Cmd {
	m.loading = true
	return tea.Batch(m.load(m.view), m.load(viewServices), m.load(viewSystem), tick(), spinTick())
}

// ---------------------------------------------------------------- filter and selection

func (m *Model) categories() []string {
	var out []string
	for _, g := range tools.Groups {
		if slices.ContainsFunc(m.rows[viewTools], func(r Row) bool { return r.Group == g.Key }) {
			out = append(out, g.Key)
		}
	}
	return out
}

func (m *Model) visible() []Row {
	out := []Row{}
	cats := m.categories()
	for _, r := range m.rows[m.view] {
		if m.view == viewTools && m.category > 0 && m.category <= len(cats) && r.Group != cats[m.category-1] {
			continue
		}
		text := strings.Join(r.Columns, " ")
		if m.view == viewTools {
			text += " " + tools.GroupTitle(r.Group)
		}
		if ui.Matches(text, m.filter) {
			out = append(out, r)
		}
	}
	return out
}

func (m *Model) selected() (Row, bool) {
	rows := m.visible()
	c := m.cursor[m.view]
	if c < 0 || c >= len(rows) {
		return Row{}, false
	}
	return rows[c], true
}

func (m *Model) moveCursor(delta int) {
	rows := m.visible()
	if len(rows) == 0 {
		m.cursor[m.view] = 0
		return
	}
	c := m.cursor[m.view] + delta
	m.cursor[m.view] = max(0, min(c, len(rows)-1))
}

func (m *Model) clampCursor() { m.moveCursor(0) }

func (m *Model) rowActions(r Row) []action {
	out := []action{}
	for _, a := range actions[m.view] {
		if a.when == nil || a.when(r) {
			out = append(out, a)
		}
	}
	return out
}

func (m *Model) setFlash(text string, isErr bool) {
	m.flash, m.flashErr, m.flashAt = text, isErr, time.Now()
}

// ---------------------------------------------------------------- commands

var commandViews = map[string]string{
	"svc": viewServices, "services": viewServices, "servicos": viewServices, "s": viewServices,
	"prj": viewProjects, "projects": viewProjects, "projetos": viewProjects, "p": viewProjects, "project": viewProjects,
	"tl": viewTools, "tools": viewTools, "ferramentas": viewTools, "t": viewTools,
	"sys": viewSystem, "system": viewSystem, "sistema": viewSystem,
}

var commandNames = []string{"projects", "services", "tools", "system", "help", "quit"}

func completeCommand(prefix string) string {
	if prefix == "" {
		return ""
	}
	for _, c := range commandNames {
		if strings.HasPrefix(c, strings.ToLower(prefix)) {
			return c
		}
	}
	return ""
}

func (m *Model) runCommand(text string) tea.Cmd {
	text = strings.TrimSpace(strings.ToLower(text))
	switch text {
	case "":
		return nil
	case "q", "q!", "quit", "exit", "sair":
		m.quit = true
		return tea.Quit
	case "help", "h", "?", "ajuda":
		m.mode = modeHelp
		return nil
	}
	if v, ok := commandViews[text]; ok {
		return m.switchView(v)
	}
	// A tool group is also a destination, such as :db or :k8s.
	if key, ok := tools.FindGroup(text); ok {
		cmd := m.switchView(viewTools)
		for i, c := range m.categories() {
			if c == key {
				m.category = i + 1
			}
		}
		m.pendingCategory = key
		return cmd
	}
	m.setFlash(i18n.Sprintf("Unknown command: %s. Use ? to see the options.", text), true)
	return nil
}

func (m *Model) switchView(v string) tea.Cmd {
	m.view, m.mode, m.filter = v, modeTable, ""
	m.category = 0
	m.pendingCategory = ""
	if !m.loaded[v] {
		m.loading = true
	}
	return m.load(v)
}

func (m *Model) start(a action, r Row) tea.Cmd {
	args := a.args(r.Key)
	if a.exec {
		c := m.backend.Command(args...)
		return tea.ExecProcess(c, func(err error) tea.Msg { return execDoneMsg{err} })
	}
	id := m.view + "/" + r.Key
	if _, busy := m.running[id]; busy {
		m.setFlash(i18n.Sprintf("%s is still in progress", r.Key), true)
		return nil
	}
	label := i18n.T(a.label) + " " + r.Key
	m.running[id] = i18n.T(a.label)
	if a.version != "" {
		label = i18n.Sprintf("install %s %s", r.Key, a.version)
		m.running[id] = i18n.Sprintf("installing %s", a.version)
	}
	b := m.backend
	return func() tea.Msg {
		c := b.Command(args...)
		out, err := c.CombinedOutput()
		return doneMsg{id: id, label: label, err: err, output: string(out)}
	}
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(ansi.Strip(s)), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// ---------------------------------------------------------------- update

// Update handles messages and keys.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case rowsMsg:
		m.rows[msg.view] = msg.rows
		m.loaded[msg.view] = true
		if msg.view == m.view {
			m.loading = false
		}
		if m.pendingCategory != "" && msg.view == viewTools {
			for i, c := range m.categories() {
				if c == m.pendingCategory {
					m.category = i + 1
				}
			}
			m.pendingCategory = ""
		}
		m.clampCursor()
		if msg.view == viewTools {
			return m, tea.Batch(m.watchSelection(), m.fetchLatest())
		}
		return m, m.watchSelection()
	case systemMsg:
		m.system = &msg.snap
		m.loaded[viewSystem] = true
		if m.view == viewSystem {
			m.loading = false
		}
	case logsMsg:
		if m.mode == modeLogs && msg.view == m.logView && msg.key == m.logKey {
			m.logText = msg.text
			if m.follow {
				m.logTop = max(0, len(m.logLines())-m.bodyHeight())
			}
		}
	case tickMsg:
		cmds := []tea.Cmd{tick()}
		if m.view != viewSystem {
			cmds = append(cmds, m.load(viewSystem))
		}
		switch {
		case m.mode == modeLogs:
			cmds = append(cmds, m.loadLogs())
		// The tool list rarely changes and is expensive to query: it only refreshes after an action.
		case m.view != viewTools:
			cmds = append(cmds, m.load(m.view))
		}
		return m, tea.Batch(cmds...)
	case spinMsg:
		m.spin++
		return m, spinTick()
	case doneMsg:
		delete(m.running, msg.id)
		if msg.err != nil {
			m.setFlash(msg.label+": "+lastLine(msg.output), true)
		} else {
			m.setFlash(msg.label+": "+i18n.T("done"), false)
		}
		view := strings.SplitN(msg.id, "/", 2)[0]
		cmds := []tea.Cmd{m.load(view)}
		if view == viewProjects {
			cmds = append(cmds, m.load(viewServices))
		}
		return m, tea.Batch(cmds...)
	case execDoneMsg:
		if msg.err != nil {
			m.setFlash(msg.err.Error(), true)
		}
		return m, m.load(m.view)
	case upstreamMsg:
		delete(m.fetching, msg.key)
		m.upstream[msg.key] = msg.up
		if msg.up.Latest != "" {
			m.latest[msg.key] = msg.up
		}
	case latestMsg:
		m.latest[msg.key] = msg.up
	case versionsMsg:
		if msg.err != nil {
			m.verErr[msg.key] = msg.err.Error()
		} else {
			m.versions[msg.key] = msg.list
		}
	case settleMsg:
		if r, ok := m.selected(); ok && m.view == viewTools && r.Key == msg.key {
			return m, m.fetchUpstream(msg.key)
		}
	case tea.KeyMsg:
		model, cmd := m.handleKey(msg)
		return model, tea.Batch(cmd, m.watchSelection())
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Fast typing or pasting arrives as several letters in a single event: "/redis" must open
	// the filter and then type, not be ignored as an unknown key.
	if msg.Type == tea.KeyRunes && len(msg.Runes) > 1 && !msg.Paste {
		var cmds []tea.Cmd
		for _, r := range msg.Runes {
			_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}
	key := msg.String()
	if key == "ctrl+c" {
		m.quit = true
		return m, tea.Quit
	}
	switch m.mode {
	case modeFilter:
		return m.keyFilter(msg)
	case modeCommand:
		return m.keyCommand(msg)
	case modeHelp:
		if key == "esc" || key == "q" || key == "?" || key == "enter" {
			m.mode = modeTable
		}
		return m, nil
	case modeDetail:
		return m.keyDetail(key)
	case modeLogs:
		return m.keyLogs(key)
	case modeActions:
		return m.keyActions(key)
	case modeVersions:
		return m.keyVersions(key)
	case modeConfirm:
		return m.keyConfirm(key)
	}
	return m.keyTable(key)
}

func (m *Model) keyFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.filter, m.mode = "", modeTable
	case tea.KeyEnter, tea.KeyDown, tea.KeyUp:
		m.mode = modeTable
		if msg.Type == tea.KeyDown {
			m.moveCursor(1)
		}
	case tea.KeyBackspace:
		if r := []rune(m.filter); len(r) > 0 {
			m.filter = string(r[:len(r)-1])
		}
	case tea.KeySpace:
		m.filter += " "
	case tea.KeyRunes:
		m.filter += string(msg.Runes)
	}
	m.cursor[m.view] = 0
	m.offset[m.view] = 0
	return m, nil
}

func (m *Model) keyCommand(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.command, m.mode = "", modeTable
	case tea.KeyTab:
		if c := completeCommand(m.command); c != "" {
			m.command = c
		}
	case tea.KeyEnter:
		text := m.command
		if c := completeCommand(text); c != "" && commandViews[strings.ToLower(text)] == "" {
			if _, isGroup := tools.FindGroup(text); !isGroup {
				text = c
			}
		}
		m.command, m.mode = "", modeTable
		return m, m.runCommand(text)
	case tea.KeyBackspace:
		if r := []rune(m.command); len(r) > 0 {
			m.command = string(r[:len(r)-1])
		} else {
			m.mode = modeTable
		}
	case tea.KeySpace:
		m.command += " "
	case tea.KeyRunes:
		m.command += string(msg.Runes)
	}
	return m, nil
}

func (m *Model) keyTable(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		m.quit = true
		return m, tea.Quit
	case ":":
		m.mode, m.command = modeCommand, ""
	case "/":
		m.mode = modeFilter
	case "?":
		m.mode = modeHelp
	case "esc":
		m.filter, m.category = "", 0
		m.clampCursor()
	case "j", "down":
		m.moveCursor(1)
	case "k", "up":
		m.moveCursor(-1)
	case "g", "home":
		m.cursor[m.view] = 0
	case "G", "end":
		m.moveCursor(1 << 20)
	case "pgdown", "ctrl+f":
		m.moveCursor(m.tableHeight())
	case "pgup", "ctrl+b":
		m.moveCursor(-m.tableHeight())
	case "tab":
		return m, m.switchView(viewOrder[(slices.Index(viewOrder, m.view)+1)%len(viewOrder)])
	case "shift+tab":
		return m, m.switchView(viewOrder[(slices.Index(viewOrder, m.view)+len(viewOrder)-1)%len(viewOrder)])
	case "1", "2", "3", "4":
		i, _ := strconv.Atoi(key)
		return m, m.switchView(viewOrder[i-1])
	case "left", "<":
		if m.view == viewTools {
			total := len(m.categories()) + 1
			m.category = (m.category - 1 + total) % total
			m.cursor[m.view] = 0
		}
	case "right", ">":
		if m.view == viewTools {
			m.category = (m.category + 1) % (len(m.categories()) + 1)
			m.cursor[m.view] = 0
		}
	case "ctrl+r":
		m.setFlash(i18n.T("Refreshing..."), false)
		return m, m.load(m.view)
	case "d":
		if r, ok := m.selected(); ok {
			m.mode, m.detailTop = modeDetail, 0
			if m.view == viewTools {
				return m, m.fetchUpstream(r.Key)
			}
		}
	case "l":
		if r, ok := m.selected(); ok && (m.view == viewServices || m.view == viewProjects) {
			return m, m.openLogs(r)
		}
	case "enter":
		if r, ok := m.selected(); ok && m.view == viewTools {
			return m, m.openVersions(r)
		}
		if r, ok := m.selected(); ok {
			if list := m.rowActions(r); len(list) > 0 {
				m.menu, m.menuPos, m.mode = list, 0, modeActions
			}
		}
	default:
		for _, a := range globalActions[m.view] {
			if a.key == key {
				return m, m.start(a, Row{})
			}
		}
		if r, ok := m.selected(); ok {
			for _, a := range m.rowActions(r) {
				if a.key == key {
					return m, m.trigger(a, r)
				}
			}
		}
	}
	return m, nil
}

func (m *Model) trigger(a action, r Row) tea.Cmd {
	if a.confirm != "" {
		copied := a
		m.pending, m.pendKey, m.mode = &copied, r.Key, modeConfirm
		if m.view == viewTools {
			return m.fetchUpstream(r.Key)
		}
		return nil
	}
	m.mode = modeTable
	return m.start(a, r)
}

func (m *Model) keyActions(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "q":
		m.mode = modeTable
	case "j", "down":
		m.menuPos = min(m.menuPos+1, len(m.menu)-1)
	case "k", "up":
		m.menuPos = max(m.menuPos-1, 0)
	case "enter":
		if r, ok := m.selected(); ok {
			return m, m.trigger(m.menu[m.menuPos], r)
		}
	default:
		if r, ok := m.selected(); ok {
			for _, a := range m.menu {
				if a.key == key {
					return m, m.trigger(a, r)
				}
			}
		}
	}
	return m, nil
}

func (m *Model) keyVersions(key string) (tea.Model, tea.Cmd) {
	list := m.versions[m.verKey]
	switch key {
	case "esc", "q":
		m.mode = modeTable
	case "j", "down":
		m.verPos = min(m.verPos+1, max(0, len(list)-1))
	case "k", "up":
		m.verPos = max(m.verPos-1, 0)
	case "g", "home":
		m.verPos = 0
	case "G", "end":
		m.verPos = max(0, len(list)-1)
	case "enter":
		r, ok := m.selected()
		if !ok || m.verPos >= len(list) {
			return m, nil
		}
		return m, m.trigger(installAction(list[m.verPos], r), r)
	}
	return m, nil
}

// installAction installs the version chosen in the list. Changing the version of something installed
// is a reinstall, and romanctl works that out by itself when the requested version differs.
func installAction(v tools.Version, r Row) action {
	return action{key: "enter", label: i18n.Sprintf("install %s", v.Label), version: v.Label, confirm: "Install %s?",
		question: i18n.Sprintf("Install %s %s?", r.Key, v.Label),
		args:     func(k string) []string { return []string{"tools", "install", k, "--version", v.ID, "--yes"} }}
}

func (m *Model) keyConfirm(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "s", "Y", "S":
		a := *m.pending
		r := Row{Key: m.pendKey}
		m.pending, m.mode = nil, modeTable
		return m, m.start(a, r)
	case "n", "N", "esc", "q", "enter":
		m.pending, m.mode = nil, modeTable
	}
	return m, nil
}

func (m *Model) keyDetail(key string) (tea.Model, tea.Cmd) {
	last := max(0, len(m.detailLines())-m.bodyHeight())
	switch key {
	case "esc", "q", "enter", "d", "?":
		m.mode = modeTable
	case "j", "down":
		m.detailTop = min(m.detailTop+1, last)
	case "k", "up":
		m.detailTop = max(m.detailTop-1, 0)
	case "pgdown", " ", "ctrl+f":
		m.detailTop = min(m.detailTop+m.bodyHeight(), last)
	case "pgup", "ctrl+b":
		m.detailTop = max(m.detailTop-m.bodyHeight(), 0)
	case "g", "home":
		m.detailTop = 0
	case "G", "end":
		m.detailTop = last
	}
	return m, nil
}

func (m *Model) openLogs(r Row) tea.Cmd {
	m.mode, m.logTitle, m.logView, m.logKey, m.logText, m.logTop, m.follow = modeLogs, r.Key, m.view, r.Key, i18n.T("Loading..."), 0, true
	return m.loadLogs()
}

func (m *Model) loadLogs() tea.Cmd {
	b, view, key := m.backend, m.logView, m.logKey
	return func() tea.Msg { return logsMsg{view, key, b.Logs(view, key)} }
}

func (m *Model) logLines() []string { return strings.Split(ansi.Strip(m.logText), "\n") }

func (m *Model) keyLogs(key string) (tea.Model, tea.Cmd) {
	last := max(0, len(m.logLines())-m.bodyHeight())
	switch key {
	case "esc", "q", "l":
		m.mode = modeTable
	case "j", "down":
		m.logTop = min(m.logTop+1, last)
		m.follow = m.logTop == last
	case "k", "up":
		m.logTop, m.follow = max(m.logTop-1, 0), false
	case "pgdown", " ":
		m.logTop = min(m.logTop+m.bodyHeight(), last)
		m.follow = m.logTop == last
	case "pgup":
		m.logTop, m.follow = max(m.logTop-m.bodyHeight(), 0), false
	case "g":
		m.logTop, m.follow = 0, false
	case "G", "f":
		m.logTop, m.follow = last, true
	}
	return m, nil
}
