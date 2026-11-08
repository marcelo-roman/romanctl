// Package tui é o painel de tela cheia do romanctl, nos moldes do k9s.
package tui

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		return i18n.T("Serviços")
	case viewProjects:
		return i18n.T("Projetos")
	case viewTools:
		return i18n.T("Ferramentas")
	}
	return i18n.T("Sistema")
}

func statusText(s string) string {
	switch s {
	case "rodando":
		return i18n.T("rodando")
	case "parado":
		return i18n.T("parado")
	case "caiu":
		return i18n.T("caiu")
	case "nao-instalado":
		return i18n.T("não instalado")
	case "instalado":
		return i18n.T("instalado")
	case "disponivel":
		return i18n.T("disponível")
	case "preparado":
		return i18n.T("só os serviços")
	}
	return s
}

// action é algo que a tecla faz sobre a linha selecionada.
type action struct {
	key     string
	label   string
	args    func(key string) []string // comando do romanctl
	exec    bool                      // precisa do terminal: o painel sai da frente
	confirm string                    // pergunta antes, com y/n
	when    func(r Row) bool
}

var actions = map[string][]action{
	viewServices: {
		{key: "s", label: "iniciar", args: func(k string) []string { return []string{"start", k} }, when: not("rodando", "nao-instalado")},
		{key: "x", label: "parar", args: func(k string) []string { return []string{"stop", k} }, when: is("rodando")},
		{key: "r", label: "reiniciar", args: func(k string) []string { return []string{"restart", k} }, when: is("rodando")},
		{key: "i", label: "instalar", args: func(k string) []string { return []string{"install", k} }, when: is("nao-instalado")},
	},
	viewProjects: {
		{key: "u", label: "subir (up -d)", args: func(k string) []string { return []string{"up", k, "-d"} }, when: not("rodando")},
		{key: "x", label: "derrubar (down)", args: func(k string) []string { return []string{"down", k} }, when: is("rodando", "caiu", "preparado")},
		{key: "r", label: "rodar aqui", args: func(k string) []string { return []string{"run", k} }, exec: true},
		{key: "a", label: "adicionar código", args: func(k string) []string { return []string{"add", "--project", k} }, exec: true},
		{key: "e", label: "editar (code-server)", args: func(k string) []string { return []string{"code", k} }},
		{key: "o", label: "abrir o app", args: func(k string) []string { return []string{"open", k, "--app"} }, when: is("rodando")},
		{key: "p", label: "publicar", args: func(k string) []string { return []string{"publish", k} }, exec: true},
		{key: "ctrl+d", label: "remover", args: func(k string) []string { return []string{"rm", k, "--yes"} }, confirm: "Remover o projeto %s e todos os arquivos dele?"},
	},
	viewTools: {
		{key: "i", label: "instalar", args: func(k string) []string { return []string{"tools", "install", k, "--yes"} }, when: is("disponivel")},
		{key: "r", label: "reinstalar", args: func(k string) []string { return []string{"tools", "install", k, "--yes", "--force"} }, when: is("instalado")},
		{key: "ctrl+d", label: "remover", args: func(k string) []string { return []string{"tools", "uninstall", k, "--yes"} }, when: is("instalado"), confirm: "Remover a ferramenta %s?"},
	},
}

// Ações que valem para a tela, sem precisar de uma linha selecionada.
var globalActions = map[string][]action{
	viewProjects: {{key: "n", label: "novo projeto", args: func(string) []string { return []string{"new"} }, exec: true}},
}

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
)

// Model é o estado do painel.
type Model struct {
	backend Backend
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
	running  map[string]string // view/key → o que está em andamento
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

	pending *action
	pendKey string
	menuPos int
	menu    []action
	quit    bool
}

// New cria o painel.
func New(b Backend) *Model {
	return &Model{backend: b, view: viewProjects, width: 100, height: 30,
		rows: map[string][]Row{}, loaded: map[string]bool{}, cursor: map[string]int{}, offset: map[string]int{}, running: map[string]string{}}
}

// ---------------------------------------------------------------- mensagens

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

// Init carrega a primeira tela e liga a atualização automática.
func (m *Model) Init() tea.Cmd {
	m.loading = true
	return tea.Batch(m.load(m.view), tick(), spinTick())
}

// ---------------------------------------------------------------- filtro e seleção

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

// ---------------------------------------------------------------- comandos

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
	// Um grupo de ferramentas também é um destino, como :db ou :k8s.
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
	m.setFlash(i18n.Sprintf("Comando desconhecido: %s. Use ? para ver as opções.", text), true)
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
		m.setFlash(i18n.Sprintf("%s ainda está em andamento", r.Key), true)
		return nil
	}
	label := i18n.T(a.label) + " " + r.Key
	m.running[id] = i18n.T(a.label)
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

// ---------------------------------------------------------------- atualização

// Update trata mensagens e teclas.
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
		switch {
		case m.mode == modeLogs:
			cmds = append(cmds, m.loadLogs())
		// A lista de ferramentas muda pouco, e consultá-la custa caro: só atualiza por ação.
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
			m.setFlash(msg.label+": "+i18n.T("concluído"), false)
		}
		return m, m.load(strings.SplitN(msg.id, "/", 2)[0])
	case execDoneMsg:
		return m, m.load(m.view)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Digitação rápida ou colar chega como várias letras num evento só: "/redis" precisa abrir
	// o filtro e depois escrever, e não ser ignorado como uma tecla desconhecida.
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
	case modeHelp, modeDetail:
		if key == "esc" || key == "q" || key == "?" || key == "enter" || key == "d" {
			m.mode = modeTable
		}
		return m, nil
	case modeLogs:
		return m.keyLogs(key)
	case modeActions:
		return m.keyActions(key)
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
		m.setFlash(i18n.T("Atualizando..."), false)
		return m, m.load(m.view)
	case "d":
		if _, ok := m.selected(); ok {
			m.mode = modeDetail
		}
	case "l":
		if r, ok := m.selected(); ok && (m.view == viewServices || m.view == viewProjects) {
			return m, m.openLogs(r)
		}
	case "enter":
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

func (m *Model) openLogs(r Row) tea.Cmd {
	m.mode, m.logTitle, m.logView, m.logKey, m.logText, m.logTop, m.follow = modeLogs, r.Key, m.view, r.Key, i18n.T("Carregando..."), 0, true
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

// ---------------------------------------------------------------- desenho

var (
	sLogo     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	sTabOn    = lipgloss.NewStyle().Bold(true).Reverse(true).Padding(0, 1)
	sTabOff   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 1)
	sHead     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	sSelected = lipgloss.NewStyle().Reverse(true)
	sKey      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	sDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	sOK       = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	sBad      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	sBusy     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	sBar      = lipgloss.NewStyle().Background(lipgloss.Color("237"))
)

func statusStyle(status string) lipgloss.Style {
	switch status {
	case "rodando", "instalado":
		return sOK
	case "caiu":
		return sBad
	case "preparado":
		return sBusy
	}
	return sDim
}

func (m *Model) bodyHeight() int { return max(3, m.height-4) }

func (m *Model) tableHeight() int { return max(1, m.bodyHeight()-1) }

// View desenha a tela: cabeçalho, abas, corpo, barra de comando e rodapé.
func (m *Model) View() string {
	if m.quit {
		return ""
	}
	lines := []string{m.headerLine(), m.tabsLine()}
	lines = append(lines, m.body()...)
	lines = append(lines, m.promptLine(), m.footerLine())
	for i, l := range lines {
		lines[i] = ansi.Truncate(l, m.width, "")
	}
	return strings.Join(lines, "\n")
}

func (m *Model) headerLine() string {
	right := sDim.Render(m.backend.Header())
	left := sLogo.Render("romanctl")
	gap := max(1, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) tabsLine() string {
	var tabs []string
	for i, v := range viewOrder {
		label := fmt.Sprintf("%d %s", i+1, viewTitle(v))
		if v == m.view {
			tabs = append(tabs, sTabOn.Render(label))
		} else {
			tabs = append(tabs, sTabOff.Render(label))
		}
	}
	return strings.Join(tabs, "")
}

func (m *Model) body() []string {
	height := m.bodyHeight()
	var out []string
	switch m.mode {
	case modeHelp:
		out = m.helpLines()
	case modeDetail:
		out = m.detailLines()
	case modeLogs:
		out = m.logWindow()
	case modeActions:
		out = m.actionLines()
	default:
		if m.view == viewSystem {
			out = m.systemLines()
		} else {
			out = m.tableLines()
		}
	}
	for len(out) < height {
		out = append(out, "")
	}
	return out[:height]
}

var columns = map[string][]string{
	viewServices: {"NOME", "SERVIÇO", "ESTADO", "ENDEREÇO"},
	viewProjects: {"NOME", "TIPO", "ESTADO", "ENDEREÇO", "ALTERADO"},
	viewTools:    {"NOME", "CATEGORIA", "ESTADO", "VERSÃO", "DESCRIÇÃO"},
}

const statusColumn = 2

func (m *Model) columnWidths(rows []Row) []int {
	head := columns[m.view]
	widths := make([]int, len(head))
	for i, h := range head {
		widths[i] = lipgloss.Width(i18n.T(h))
	}
	for _, r := range rows {
		for i, c := range r.Columns {
			if i < len(widths) {
				widths[i] = max(widths[i], lipgloss.Width(c))
			}
		}
	}
	// A última coluna fica com o que sobrar; a penúltima não pode empurrá-la para fora.
	total := m.width - 2
	for i := 0; i < len(widths)-1; i++ {
		widths[i] = min(widths[i], max(8, total/3))
		total -= widths[i] + 2
	}
	if total > 0 {
		widths[len(widths)-1] = min(widths[len(widths)-1], total)
	}
	return widths
}

func pad(s string, w int) string {
	s = ansi.Truncate(s, w, "~")
	return s + strings.Repeat(" ", max(0, w-lipgloss.Width(s)))
}

func (m *Model) tableLines() []string {
	rows := m.visible()
	if m.loading && !m.loaded[m.view] {
		return []string{"", "  " + sBusy.Render(spinnerFrame(m.spin)+" "+i18n.T("Carregando..."))}
	}
	widths := m.columnWidths(rows)
	var heads []string
	for i, h := range columns[m.view] {
		heads = append(heads, pad(i18n.T(h), widths[i]))
	}
	out := []string{" " + sHead.Render(strings.Join(heads, "  "))}
	if len(rows) == 0 {
		text := i18n.T("Nada por aqui.")
		if m.filter != "" {
			text = i18n.T("Nada encontrado. Esc limpa o filtro.")
		} else if m.view == viewProjects {
			text = i18n.T("Nenhum projeto ainda. Aperte n para criar o primeiro.")
		}
		return append(out, "", "  "+sDim.Render(text))
	}
	height := m.tableHeight()
	cursor := m.cursor[m.view]
	offset := m.offset[m.view]
	if cursor < offset {
		offset = cursor
	}
	if cursor >= offset+height {
		offset = cursor - height + 1
	}
	m.offset[m.view] = offset
	for i := offset; i < min(offset+height, len(rows)); i++ {
		out = append(out, m.tableRow(rows[i], widths, i == cursor))
	}
	return out
}

func (m *Model) tableRow(r Row, widths []int, selected bool) string {
	cells := make([]string, len(widths))
	busy, isBusy := m.running[m.view+"/"+r.Key]
	for i := range widths {
		text := ""
		if i < len(r.Columns) {
			text = r.Columns[i]
		}
		if i == statusColumn && isBusy {
			text = spinnerFrame(m.spin) + " " + busy
		}
		cells[i] = pad(text, widths[i])
		if !selected {
			switch {
			case i == statusColumn && isBusy:
				cells[i] = sBusy.Render(cells[i])
			case i == statusColumn:
				cells[i] = statusStyle(r.Status).Render(cells[i])
			}
		}
	}
	line := " " + strings.Join(cells, "  ")
	if selected {
		return sSelected.Render(pad(line, m.width))
	}
	return line
}

func spinnerFrame(i int) string { return string("|/-\\"[i%4]) }

func gauge(used, total int64, width int) string {
	if total <= 0 {
		return strings.Repeat("-", width)
	}
	filled := int(float64(used) / float64(total) * float64(width))
	filled = max(0, min(filled, width))
	return strings.Repeat("#", filled) + strings.Repeat(".", width-filled)
}

func gb(n int64) string { return fmt.Sprintf("%.1f GB", float64(n)/(1<<30)) }

func (m *Model) systemLines() []string {
	if m.system == nil {
		return []string{"", "  " + sBusy.Render(spinnerFrame(m.spin)+" "+i18n.T("Medindo..."))}
	}
	s := m.system
	bar := max(10, min(m.width-46, 40))
	row := func(label string, used, total int64, tail string) string {
		return fmt.Sprintf(" %-10s [%s] %s", label, gauge(used, total, bar), tail)
	}
	out := []string{sHead.Render(" " + i18n.T("RECURSOS"))}
	if s.SystemCPU != nil {
		out = append(out, row("CPU", int64(*s.SystemCPU), 100, fmt.Sprintf("%.0f%%", *s.SystemCPU)))
	} else {
		out = append(out, fmt.Sprintf(" %-10s %s", "CPU", i18n.Sprintf("%.2f núcleos nos seus processos", s.OwnCPUCores)))
	}
	used := s.Memory.Total - s.Memory.Available
	out = append(out, row(i18n.T("Memória"), used, s.Memory.Total, i18n.Sprintf("%s de %s", gb(used), gb(s.Memory.Total))))
	if s.Memory.SwapTotal > 0 {
		swap := s.Memory.SwapTotal - s.Memory.SwapFree
		out = append(out, row("Swap", swap, s.Memory.SwapTotal, i18n.Sprintf("%s de %s", gb(swap), gb(s.Memory.SwapTotal))))
	}
	disk := s.Disk.Size - s.Disk.Free
	out = append(out, row(i18n.T("Disco"), disk, s.Disk.Size, i18n.Sprintf("%s de %s", gb(disk), gb(s.Disk.Size))))
	if b := s.Battery; b != nil {
		tail := fmt.Sprintf("%.0f%%", b.Percent)
		if b.Power > 0 {
			tail += fmt.Sprintf(", %.1f W", b.Power)
		}
		if b.HoursLeft > 0 {
			tail += fmt.Sprintf(", %dh%02d", int(b.HoursLeft), int(b.HoursLeft*60)%60)
		}
		out = append(out, row(i18n.T("Bateria"), int64(b.Percent), 100, tail))
	}
	out = append(out, "", sHead.Render(" "+i18n.T("PROCESSOS")))
	out = append(out, sDim.Render(fmt.Sprintf(" %-8s %6s %8s  %s", "PID", "CPU", "MEM", i18n.T("NOME"))))
	top := slices.Clone(s.Top)
	sort.SliceStable(top, func(i, j int) bool { return top[i].CPU > top[j].CPU })
	for i, p := range top {
		if i >= max(3, m.bodyHeight()-len(out)) {
			break
		}
		out = append(out, fmt.Sprintf(" %-8d %5.1f%% %5.0f MB  %s", p.PID, p.CPU, float64(p.Memory)/(1<<20), p.Name))
	}
	return out
}

func (m *Model) detailLines() []string {
	r, ok := m.selected()
	if !ok {
		return nil
	}
	out := []string{sHead.Render(" " + r.Key), ""}
	for _, d := range r.Detail {
		if d[1] != "" {
			out = append(out, fmt.Sprintf(" %-14s %s", d[0]+":", d[1]))
		}
	}
	out = append(out, "", sDim.Render(" "+i18n.T("O que dá para fazer:")))
	for _, a := range m.rowActions(r) {
		out = append(out, " "+sKey.Render(fmt.Sprintf("%-7s", keyName(a.key)))+" "+i18n.T(a.label))
	}
	return out
}

func (m *Model) actionLines() []string {
	r, _ := m.selected()
	out := []string{sHead.Render(" " + i18n.Sprintf("Ações para %s", r.Key)), ""}
	for i, a := range m.menu {
		line := fmt.Sprintf(" %-7s %s", keyName(a.key), i18n.T(a.label))
		if i == m.menuPos {
			out = append(out, sSelected.Render(pad(line, min(m.width, 44))))
		} else {
			out = append(out, " "+sKey.Render(fmt.Sprintf("%-7s", keyName(a.key)))+" "+i18n.T(a.label))
		}
	}
	return out
}

func (m *Model) logWindow() []string {
	lines := m.logLines()
	height := m.bodyHeight()
	top := max(0, min(m.logTop, max(0, len(lines)-height)))
	out := make([]string, 0, height)
	for i := top; i < min(top+height, len(lines)); i++ {
		out = append(out, lines[i])
	}
	return out
}

func keyName(k string) string {
	if k == "ctrl+d" {
		return "Ctrl+D"
	}
	return k
}

func (m *Model) helpLines() []string {
	section := func(title string, items ...[2]string) []string {
		out := []string{sHead.Render(" " + title)}
		for _, it := range items {
			out = append(out, " "+sKey.Render(fmt.Sprintf("%-14s", it[0]))+" "+it[1])
		}
		return append(out, "")
	}
	out := section(i18n.T("NAVEGAR"),
		[2]string{"Tab, 1-4", i18n.T("trocar de tela")},
		[2]string{i18n.T("setas, j/k"), i18n.T("mover na lista")},
		[2]string{"g / G", i18n.T("início e fim da lista")},
		[2]string{"Enter", i18n.T("ver as ações do item")},
		[2]string{"d", i18n.T("detalhes do item")},
		[2]string{"l", i18n.T("logs do item")},
		[2]string{"Esc", i18n.T("voltar e limpar o filtro")},
		[2]string{"q", i18n.T("sair")})
	out = append(out, section(i18n.T("BUSCAR"),
		[2]string{"/", i18n.T("filtrar a lista")},
		[2]string{"< >", i18n.T("trocar de categoria, em Ferramentas")},
		[2]string{":", i18n.T("ir direto: :tools, :svc, :prj, :sys, :db, :k8s")},
		[2]string{"Ctrl+R", i18n.T("atualizar agora")})...)
	var items [][2]string
	for _, a := range globalActions[m.view] {
		items = append(items, [2]string{keyName(a.key), i18n.T(a.label)})
	}
	for _, a := range actions[m.view] {
		items = append(items, [2]string{keyName(a.key), i18n.T(a.label)})
	}
	if len(items) > 0 {
		out = append(out, section(strings.ToUpper(viewTitle(m.view)), items...)...)
	}
	return out
}

func (m *Model) promptLine() string {
	switch m.mode {
	case modeCommand:
		hint := ""
		if c := completeCommand(m.command); c != "" && c != strings.ToLower(m.command) {
			hint = sDim.Render(c[len(m.command):])
		}
		return sKey.Render(":") + m.command + hint + sSelected.Render(" ")
	case modeFilter:
		return sKey.Render("/") + m.filter + sSelected.Render(" ")
	case modeConfirm:
		return sBad.Render(i18n.Sprintf(m.pending.confirm, m.pendKey)) + " " + sKey.Render(i18n.T("[y/N]"))
	}
	if m.flash != "" && time.Since(m.flashAt) < 8*time.Second {
		if m.flashErr {
			return sBad.Render(m.flash)
		}
		return sOK.Render(m.flash)
	}
	return m.statusSummary()
}

func (m *Model) statusSummary() string {
	if m.view == viewSystem {
		return ""
	}
	rows := m.visible()
	text := fmt.Sprintf("%s  %d/%d", viewTitle(m.view), len(rows), len(m.rows[m.view]))
	if m.view == viewTools {
		cats := m.categories()
		if m.category > 0 && m.category <= len(cats) {
			text += "  " + i18n.T("categoria:") + " " + tools.GroupTitle(cats[m.category-1])
		}
	}
	if m.filter != "" {
		text += "  " + i18n.T("filtro:") + " " + m.filter
	}
	return sDim.Render(text)
}

// footerLine mostra só as teclas que fazem sentido agora, como no k9s.
func (m *Model) footerLine() string {
	type hint struct{ key, label string }
	var hints []hint
	switch m.mode {
	case modeFilter:
		hints = []hint{{"Enter", i18n.T("aplicar")}, {"Esc", i18n.T("limpar")}}
	case modeCommand:
		hints = []hint{{"Tab", i18n.T("completar")}, {"Enter", i18n.T("ir")}, {"Esc", i18n.T("cancelar")}}
	case modeHelp, modeDetail:
		hints = []hint{{"Esc", i18n.T("voltar")}}
	case modeLogs:
		state := i18n.T("seguir")
		hints = []hint{{"j/k", i18n.T("rolar")}, {"f", state}, {"g", i18n.T("início")}, {"Esc", i18n.T("voltar")}}
	case modeActions:
		hints = []hint{{"Enter", i18n.T("escolher")}, {"Esc", i18n.T("voltar")}}
	case modeConfirm:
		hints = []hint{{"y", i18n.T("sim")}, {"n", i18n.T("não")}}
	default:
		for _, a := range globalActions[m.view] {
			hints = append(hints, hint{keyName(a.key), i18n.T(a.label)})
		}
		if r, ok := m.selected(); ok {
			for _, a := range m.rowActions(r) {
				hints = append(hints, hint{keyName(a.key), i18n.T(a.label)})
			}
			hints = append(hints, hint{"d", i18n.T("detalhes")})
			if m.view != viewTools {
				hints = append(hints, hint{"l", i18n.T("logs")})
			}
		}
		if m.view == viewTools {
			hints = append(hints, hint{"< >", i18n.T("categoria")})
		}
		hints = append(hints, hint{"/", i18n.T("filtrar")}, hint{":", i18n.T("ir para")}, hint{"?", i18n.T("ajuda")}, hint{"q", i18n.T("sair")})
	}
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = sKey.Render(h.key) + " " + h.label
	}
	return strings.Join(parts, "  ")
}

// Run abre o painel. Enquanto ele está na tela, nada mais escreve no terminal.
func Run() error {
	ui.SetSilent(true)
	defer ui.SetSilent(false)
	_, err := tea.NewProgram(New(newRealBackend()), tea.WithAltScreen()).Run()
	return err
}
