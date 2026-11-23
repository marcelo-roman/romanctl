package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"

	"romanctl/internal/i18n"
)

// Option is a row of a list. The label may be colored; the filter only sees the text.
type Option struct {
	Label    string
	Value    string
	Category string // category key, for categorized lists
	Search   string // extra text for the filter, not shown on screen
	Section  string // group title, like the sections of "git help"
}

// Category is a tab of a list.
type Category struct {
	Key   string
	Label string
}

// Opt creates an option.
func Opt(label, value string) Option { return Option{Label: label, Value: value} }

// PickConfig describes a filterable list.
type PickConfig struct {
	Title      string
	Options    []Option
	Multi      bool
	All        bool   // shows the "All" row at the top of a multiple-choice list
	Initial    string // value selected at start
	NoLeft     bool   // the left arrow does not go back, as in the main menu
	Subtitle   string
	Categories []Category // with this, ←/→ switch categories
	AllowNone  bool       // Enter with nothing marked is a valid answer, such as "no services"
}

// matches reports whether the text contains every word of the filter, ignoring case and accents.
func matches(text, filter string) bool {
	text = fold(ansi.Strip(text))
	for _, word := range strings.Fields(fold(filter)) {
		if !strings.Contains(text, word) {
			return false
		}
	}
	return true
}

// Matches is the same filter the lists use, for callers filtering by --filter.
func Matches(text, filter string) bool { return matches(text, filter) }

var foldAccents = strings.NewReplacer("\u00e1", "a", "\u00e0", "a", "\u00e2", "a", "\u00e3", "a", "\u00e9", "e", "\u00ea", "e", "\u00ed", "i",
	"\u00f3", "o", "\u00f4", "o", "\u00f5", "o", "\u00fa", "u", "\u00fc", "u", "\u00e7", "c")

func fold(s string) string { return foldAccents.Replace(strings.ToLower(s)) }

// Fold makes a text comparable, without case or accents.
func Fold(s string) string { return fold(s) }

const allRow = -1

type pickModel struct {
	cfg      PickConfig
	filter   string
	visible  []int // indexes into Options; the "All" row is allRow
	cursor   int
	offset   int
	marked   map[string]bool
	width    int
	height   int
	warn     string
	result   []string
	err      error
	finished bool
	category int // 0 is "All"; the others are Categories[category-1]
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

// switchCategory cycles through "All" and the categories; the typed filter still applies.
func (m *pickModel) switchCategory(delta int) {
	total := len(m.cfg.Categories) + 1
	m.category = (m.category + delta + total) % total
	m.refilter()
}

func (m *pickModel) categoryLabel() string {
	if m.category == 0 {
		return i18n.T("All")
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

// toggleVisible marks or unmarks what the filter is showing, and only that.
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
			// The first Esc only clears the filter, so the user does not lose the whole list.
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
	// The result follows the list order, not the order in which rows were marked.
	for _, o := range m.cfg.Options {
		if m.marked[o.Value] {
			m.result = append(m.result, o.Value)
		}
	}
	if len(m.result) == 0 && !m.cfg.AllowNone {
		m.warn = i18n.T("Mark at least one item with Space")
		return m, nil
	}
	return m.finish(nil)
}

func (m *pickModel) pageSize() int {
	// Sections and the footer take lines too: subtract them so the screen never scrolls.
	return max(3, min(m.height-10, 18))
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

// View follows terminal conventions: fzf's prompt and count, "git help" sections, a
// reverse-video row and htop's key bar. Text only, no icons.
func (m *pickModel) View() string {
	if m.finished {
		return ""
	}
	var b strings.Builder
	b.WriteString(SBoldTxt.Render(m.cfg.Title) + "\n")
	if m.cfg.Subtitle != "" {
		b.WriteString(SGray.Render(m.cfg.Subtitle) + "\n")
	}
	if len(m.cfg.Categories) > 0 {
		b.WriteString(SGray.Render(i18n.T("Category: ")) + "< " + SBoldTxt.Render(m.categoryLabel()) + " >\n")
	}
	shown, total := len(m.visibleOptions()), len(m.cfg.Options)
	b.WriteString(SBlue.Render(i18n.T("Filter> ")) + m.filter + cursorBlock.Render(" ") + SGray.Render(fmt.Sprintf("  %d/%d", shown, total)) + "\n")

	page := m.pageSize()
	end := min(m.offset+page, len(m.visible))
	if len(m.visible) == 0 {
		b.WriteString(SYellow.Render(i18n.T("  Nothing found; Esc clears the filter")) + "\n")
	}
	section := ""
	for i := m.offset; i < end; i++ {
		row := m.visible[i]
		label, check := "", ""
		if row == allRow {
			label = i18n.Sprintf("All (%d)", shown)
			check = m.checkbox(m.allMarked())
		} else {
			o := m.cfg.Options[row]
			label = o.Label
			if m.cfg.Multi {
				check = m.checkbox(m.marked[o.Value])
			}
			if o.Section != "" && o.Section != section {
				section = o.Section
				b.WriteString(SBoldTxt.Render(section) + "\n")
			}
		}
		line := check + label
		if m.width > 4 {
			line = ansi.Truncate(line, m.width-3, "...")
		}
		if i == m.cursor {
			// Without the label colors, so the reverse video covers the whole line.
			b.WriteString(selectedRow.Render("  "+ansi.Strip(line)) + "\n")
		} else {
			b.WriteString("  " + line + "\n")
		}
	}
	if len(m.visible) > page {
		b.WriteString(SGray.Render(i18n.Sprintf("  -- %d-%d of %d --", m.offset+1, end, len(m.visible))) + "\n")
	}
	if m.warn != "" {
		b.WriteString(SYellow.Render(m.warn) + "\n")
	}
	b.WriteString(m.keyBar())
	return b.String()
}

var (
	selectedRow = lipgloss.NewStyle().Reverse(true)
	cursorBlock = lipgloss.NewStyle().Reverse(true)
)

// keyBar is the footer key bar, like htop's and nano's.
func (m *pickModel) keyBar() string {
	type key struct{ name, action string }
	keys := []key{{"Enter", i18n.T("select")}}
	if m.cfg.Multi {
		keys = []key{{i18n.T("Space"), i18n.T("mark")}, {"^A", i18n.T("mark visible")}, {"Enter", i18n.T("confirm")}}
	}
	if len(m.cfg.Categories) > 0 {
		keys = append(keys, key{"</>", i18n.T("category")})
	}
	keys = append(keys, key{"Esc", i18n.T("back")})
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = SBoldTxt.Render(k.name) + " " + SGray.Render(k.action)
	}
	return strings.Join(parts, "  ")
}

func (m *pickModel) checkbox(on bool) string {
	if on {
		return "[x] "
	}
	return "[ ] "
}

// Pick shows the list and returns what was chosen.
func Pick(cfg PickConfig) ([]string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, Usagef("this command needs a terminal; pass everything as arguments (see: romanctl help)")
	}
	final, err := tea.NewProgram(newPickModel(cfg)).Run()
	if err != nil {
		return nil, err
	}
	m := final.(*pickModel)
	return m.result, m.err
}

// PickOne shows a single-choice list.
func PickOne(title string, options []Option) (string, error) {
	got, err := Pick(PickConfig{Title: title, Options: options})
	if err != nil || len(got) == 0 {
		return "", err
	}
	return got[0], nil
}
