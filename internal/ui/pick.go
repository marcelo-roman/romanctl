package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"

	"romanctl/internal/i18n"
)

// Option é uma linha de uma lista. O rótulo pode ter cor, e o filtro enxerga só o texto.
type Option struct {
	Label    string
	Value    string
	Category string // chave da categoria, para listas com categorias
	Search   string // texto a mais para o filtro, que não aparece na tela
}

// Category é uma aba de uma lista.
type Category struct {
	Key   string
	Label string
}

// Opt cria uma opção.
func Opt(label, value string) Option { return Option{Label: label, Value: value} }

// PickConfig descreve uma lista com filtro.
type PickConfig struct {
	Title      string
	Options    []Option
	Multi      bool
	All        bool   // mostra a linha "Todos" no topo de uma lista de múltipla escolha
	Initial    string // valor que começa selecionado
	NoLeft     bool   // a seta esquerda não volta, como no menu principal
	Subtitle   string
	Categories []Category // com isto, ←/→ trocam de categoria
	AllowNone  bool       // Enter sem nada marcado é uma resposta válida, como "sem serviços"
}

// matches diz se o texto contém todas as palavras do filtro, sem diferenciar maiúsculas e acentos.
func matches(text, filter string) bool {
	text = fold(ansi.Strip(text))
	for _, word := range strings.Fields(fold(filter)) {
		if !strings.Contains(text, word) {
			return false
		}
	}
	return true
}

// Matches é o mesmo filtro das listas, para quem filtra por --filter.
func Matches(text, filter string) bool { return matches(text, filter) }

var foldAccents = strings.NewReplacer("á", "a", "à", "a", "â", "a", "ã", "a", "é", "e", "ê", "e", "í", "i",
	"ó", "o", "ô", "o", "õ", "o", "ú", "u", "ü", "u", "ç", "c")

func fold(s string) string { return foldAccents.Replace(strings.ToLower(s)) }

// Fold deixa um texto comparável, sem maiúsculas nem acentos.
func Fold(s string) string { return fold(s) }

const allRow = -1

type pickModel struct {
	cfg      PickConfig
	filter   string
	visible  []int // índices em Options; a linha "Todos" é allRow
	cursor   int
	offset   int
	marked   map[string]bool
	width    int
	height   int
	warn     string
	result   []string
	err      error
	finished bool
	category int // 0 é "Todas"; os demais são Categories[category-1]
}

func newPickModel(cfg PickConfig) *pickModel {
	m := &pickModel{cfg: cfg, marked: map[string]bool{}, width: 100, height: 24}
	m.refilter()
	for i, row := range m.visible {
		if row != allRow && cfg.Options[row].Value == cfg.Initial {
			m.cursor = i
		}
	}
	return m
}

func (m *pickModel) refilter() {
	m.visible = m.visible[:0]
	matched := 0
	for i, o := range m.cfg.Options {
		if m.inCategory(o) && matches(o.Label+" "+o.Search, m.filter) {
			m.visible = append(m.visible, i)
			matched++
		}
	}
	if m.cfg.Multi && m.cfg.All && matched > 0 {
		m.visible = append([]int{allRow}, m.visible...)
	}
	m.cursor, m.offset = 0, 0
}

func (m *pickModel) inCategory(o Option) bool {
	return m.category == 0 || o.Category == m.cfg.Categories[m.category-1].Key
}

// switchCategory gira entre "Todas" e as categorias; o filtro digitado continua valendo.
func (m *pickModel) switchCategory(delta int) {
	total := len(m.cfg.Categories) + 1
	m.category = (m.category + delta + total) % total
	m.refilter()
}

func (m *pickModel) categoryLabel() string {
	if m.category == 0 {
		return i18n.T("Todas")
	}
	return m.cfg.Categories[m.category-1].Label
}

func (m *pickModel) Init() tea.Cmd { return nil }

func (m *pickModel) visibleOptions() []string {
	var out []string
	for _, row := range m.visible {
		if row != allRow {
			out = append(out, m.cfg.Options[row].Value)
		}
	}
	return out
}

func (m *pickModel) allMarked() bool {
	opts := m.visibleOptions()
	if len(opts) == 0 {
		return false
	}
	for _, v := range opts {
		if !m.marked[v] {
			return false
		}
	}
	return true
}

// toggleVisible marca ou desmarca o que o filtro está mostrando, e só isso.
func (m *pickModel) toggleVisible() {
	mark := !m.allMarked()
	for _, v := range m.visibleOptions() {
		m.marked[v] = mark
	}
}

func (m *pickModel) move(delta int) {
	if len(m.visible) == 0 {
		return
	}
	m.cursor = (m.cursor + delta + len(m.visible)) % len(m.visible)
}

func (m *pickModel) finish(err error) (tea.Model, tea.Cmd) {
	m.err, m.finished = err, true
	return m, tea.Quit
}

func (m *pickModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		m.warn = ""
		switch msg.Type {
		case tea.KeyCtrlC:
			return m.finish(errCancel)
		case tea.KeyEsc:
			// O primeiro Esc só limpa o filtro: a pessoa não perde a lista inteira.
			if m.filter != "" {
				m.filter = ""
				m.refilter()
				return m, nil
			}
			return m.finish(errBack)
		case tea.KeyLeft:
			if len(m.cfg.Categories) > 0 {
				m.switchCategory(-1)
			} else if m.filter == "" && !m.cfg.NoLeft {
				return m.finish(errBack)
			}
		case tea.KeyRight:
			if len(m.cfg.Categories) > 0 {
				m.switchCategory(1)
			}
		case tea.KeyUp, tea.KeyCtrlP:
			m.move(-1)
		case tea.KeyDown, tea.KeyCtrlN, tea.KeyTab:
			m.move(1)
		case tea.KeyPgUp:
			m.move(-m.pageSize())
		case tea.KeyPgDown:
			m.move(m.pageSize())
		case tea.KeyHome:
			m.cursor = 0
		case tea.KeyEnd:
			m.cursor = max(0, len(m.visible)-1)
		case tea.KeyCtrlU:
			m.filter = ""
			m.refilter()
		case tea.KeyBackspace:
			if r := []rune(m.filter); len(r) > 0 {
				m.filter = string(r[:len(r)-1])
				m.refilter()
			}
		case tea.KeyCtrlA:
			if m.cfg.Multi {
				m.toggleVisible()
			}
		case tea.KeySpace:
			if !m.cfg.Multi {
				m.filter += " "
				m.refilter()
			} else if len(m.visible) > 0 {
				if row := m.visible[m.cursor]; row == allRow {
					m.toggleVisible()
				} else {
					v := m.cfg.Options[row].Value
					m.marked[v] = !m.marked[v]
				}
			}
		case tea.KeyEnter:
			return m.submit()
		case tea.KeyRunes:
			m.filter += string(msg.Runes)
			m.refilter()
		}
	}
	m.keepCursorVisible()
	return m, nil
}

func (m *pickModel) submit() (tea.Model, tea.Cmd) {
	if len(m.visible) == 0 {
		return m, nil
	}
	if !m.cfg.Multi {
		m.result = []string{m.cfg.Options[m.visible[m.cursor]].Value}
		return m.finish(nil)
	}
	// A ordem do resultado é a da lista, não a ordem em que as linhas foram marcadas.
	for _, o := range m.cfg.Options {
		if m.marked[o.Value] {
			m.result = append(m.result, o.Value)
		}
	}
	if len(m.result) == 0 && !m.cfg.AllowNone {
		m.warn = i18n.T("Marque ao menos um item com Espaço")
		return m, nil
	}
	return m.finish(nil)
}

func (m *pickModel) pageSize() int {
	return max(3, min(m.height-8, 18))
}

func (m *pickModel) keepCursorVisible() {
	page := m.pageSize()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+page {
		m.offset = m.cursor - page + 1
	}
}

func (m *pickModel) View() string {
	if m.finished {
		return ""
	}
	var b strings.Builder
	b.WriteString(STitle.Render(m.cfg.Title) + "\n")
	if m.cfg.Subtitle != "" {
		b.WriteString(SGray.Render(m.cfg.Subtitle) + "\n")
	}
	if len(m.cfg.Categories) > 0 {
		b.WriteString(i18n.T("Categoria: ") + SBlue.Render("◀ "+m.categoryLabel()+" ▶") + SGray.Render(fmt.Sprintf(" (%d/%d)", m.category, len(m.cfg.Categories))) + "\n")
	}
	total, shown := len(m.cfg.Options), len(m.visibleOptions())
	count := SGray.Render(fmt.Sprintf(" (%d/%d)", shown, total))
	if m.filter == "" {
		b.WriteString(SGray.Render(i18n.T("Filtro: digite para buscar")) + count + "\n")
	} else {
		b.WriteString(i18n.T("Filtro: ") + SBlue.Render(m.filter) + SBlue.Render("▏") + count + "\n")
	}
	page := m.pageSize()
	end := min(m.offset+page, len(m.visible))
	if len(m.visible) == 0 {
		b.WriteString("  " + SYellow.Render(i18n.T("Nada encontrado; Esc limpa o filtro")) + "\n")
	}
	for i := m.offset; i < end; i++ {
		row := m.visible[i]
		cursor := "  "
		if i == m.cursor {
			cursor = SBlue.Render("> ")
		}
		label, check := "", ""
		if row == allRow {
			label = i18n.Sprintf("Todos (%d)", shown)
			check = m.checkbox(m.allMarked())
		} else {
			label = m.cfg.Options[row].Label
			if m.cfg.Multi {
				check = m.checkbox(m.marked[m.cfg.Options[row].Value])
			}
		}
		line := check + label
		if m.width > 4 {
			line = ansi.Truncate(line, m.width-3, "…")
		}
		if i == m.cursor {
			line = sBoldLine(line)
		}
		b.WriteString(cursor + line + "\n")
	}
	if len(m.visible) > page {
		b.WriteString(SGray.Render(fmt.Sprintf("  %d-%d/%d", m.offset+1, end, len(m.visible))) + "\n")
	}
	if m.warn != "" {
		b.WriteString(SYellow.Render(m.warn) + "\n")
	}
	if len(m.cfg.Categories) > 0 {
		b.WriteString(SGray.Render(i18n.T("←/→ categoria · ")))
	}
	if m.cfg.Multi {
		b.WriteString(SGray.Render(i18n.T("↑/↓ move · Espaço marca · Ctrl+A marca os visíveis · Enter confirma · Esc volta")))
	} else {
		b.WriteString(SGray.Render(i18n.T("↑/↓ move · Enter escolhe · Esc volta")))
	}
	return b.String()
}

func (m *pickModel) checkbox(on bool) string {
	if on {
		return SGreen.Render("[x] ")
	}
	return "[ ] "
}

func sBoldLine(s string) string { return SBoldTxt.Render(s) }

// Pick mostra a lista e devolve o que foi escolhido.
func Pick(cfg PickConfig) ([]string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, Usagef("este comando precisa de um terminal; informe tudo por argumentos (veja: romanctl help)")
	}
	final, err := tea.NewProgram(newPickModel(cfg)).Run()
	if err != nil {
		return nil, err
	}
	m := final.(*pickModel)
	return m.result, m.err
}

// PickOne mostra uma lista de escolha única.
func PickOne(title string, options []Option) (string, error) {
	got, err := Pick(PickConfig{Title: title, Options: options})
	if err != nil || len(got) == 0 {
		return "", err
	}
	return got[0], nil
}
