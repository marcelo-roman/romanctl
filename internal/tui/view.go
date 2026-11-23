package tui

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"romanctl/internal/i18n"
	"romanctl/internal/tools"
	"romanctl/internal/ui"
)

func adaptive(light, dark string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: light, Dark: dark}
}

var (
	cOrange  = adaptive("#C25E00", "#FFA500")
	cAqua    = adaptive("#00808C", "#00FFFF")
	cBlue    = adaptive("#0B5FBF", "#1E90FF")
	cFuchsia = adaptive("#A0119E", "#FF00FF")
	cText    = adaptive("#1C1C1C", "#FFFFFF")
	cMuted   = adaptive("#5F5F5F", "#A8A8A8")
	cDim     = adaptive("#9E9E9E", "#5F5F5F")
	cGreen   = adaptive("#00875F", "#00FF7F")
	cRed     = adaptive("#C2185B", "#FF4500")
	cYellow  = adaptive("#9E7C00", "#FFD700")
	cPapaya  = adaptive("#5F3F00", "#FFEFD5")
	cInk     = lipgloss.Color("#000000")
)

var (
	sLogo        = lipgloss.NewStyle().Bold(true).Foreground(cOrange)
	sLabel       = lipgloss.NewStyle().Bold(true).Foreground(cOrange)
	sValue       = lipgloss.NewStyle().Bold(true).Foreground(cText)
	sKey         = lipgloss.NewStyle().Bold(true).Foreground(cBlue)
	sViewKey     = lipgloss.NewStyle().Bold(true).Foreground(cFuchsia)
	sHintLabel   = lipgloss.NewStyle().Foreground(cMuted)
	sBorder      = lipgloss.NewStyle().Foreground(cBlue)
	sTitle       = lipgloss.NewStyle().Bold(true).Foreground(cAqua)
	sTitleScope  = lipgloss.NewStyle().Bold(true).Foreground(cFuchsia)
	sTitleCount  = lipgloss.NewStyle().Bold(true).Foreground(cPapaya)
	sHead        = lipgloss.NewStyle().Bold(true).Foreground(cText)
	sSection     = lipgloss.NewStyle().Bold(true).Foreground(cAqua)
	sCursor      = lipgloss.NewStyle().Bold(true).Foreground(cInk).Background(cAqua)
	sCrumb       = lipgloss.NewStyle().Foreground(cInk).Background(cAqua).Padding(0, 1)
	sCrumbActive = lipgloss.NewStyle().Bold(true).Foreground(cInk).Background(cOrange).Padding(0, 1)
	sDim         = lipgloss.NewStyle().Foreground(cDim)
	sMuted       = lipgloss.NewStyle().Foreground(cMuted)
	sOK          = lipgloss.NewStyle().Foreground(cGreen)
	sBad         = lipgloss.NewStyle().Foreground(cRed)
	sBusy        = lipgloss.NewStyle().Foreground(cYellow)
	sWarn        = lipgloss.NewStyle().Foreground(cOrange)
	sPrompt      = lipgloss.NewStyle().Foreground(cText)
	sCaret       = lipgloss.NewStyle().Background(cAqua)
)

var logo = []string{
	"                              _   _ ",
	" _ _ ___ _ __   __ _ _ _   __| |_| |",
	"| '_/ _ \\ '  \\ / _` | ' \\ / _|  _| |",
	"|_| \\___/_|_|_|\\__,_|_||_|\\__|\\__|_|",
}

// rowStyle colors the whole row by state, as k9s does with pods.
func rowStyle(status string) lipgloss.Style {
	switch status {
	case "running", "installed":
		return lipgloss.NewStyle().Foreground(cAqua)
	case "crashed":
		return sBad
	case "prepared":
		return sWarn
	case "not-installed":
		return sDim
	}
	return sMuted
}

func statusStyle(status string) lipgloss.Style {
	switch status {
	case "running", "installed":
		return sOK.Bold(true)
	case "crashed":
		return sBad.Bold(true)
	case "prepared":
		return sWarn.Bold(true)
	case "not-installed":
		return sDim
	}
	return sMuted
}

func levelStyle(ratio float64) lipgloss.Style {
	switch {
	case ratio >= 0.85:
		return sBad.Bold(true)
	case ratio >= 0.6:
		return sWarn.Bold(true)
	}
	return sOK.Bold(true)
}

// darkBackground is decided before the panel opens: querying the terminal while the full screen
// is up mixes the answer with the keystrokes.
var darkBackground = true

// ---------------------------------------------------------------- measurements

func (m *Model) bigHeader() bool { return m.height >= 22 && m.width >= 70 }

func (m *Model) headerHeight() int {
	if m.bigHeader() {
		return len(m.infoLines())
	}
	return 1
}

func (m *Model) promptHeight() int {
	if m.mode == modeCommand || m.mode == modeFilter {
		return 3
	}
	return 0
}

func (m *Model) boxHeight() int {
	return max(3, m.height-m.headerHeight()-m.promptHeight()-2)
}

func (m *Model) innerWidth() int { return max(1, m.width-2) }

func (m *Model) bodyHeight() int { return max(1, m.boxHeight()-2) }

func (m *Model) tableHeight() int { return max(1, m.bodyHeight()-1) }

// ---------------------------------------------------------------- primitives

func fit(s string, w int) string {
	if lipgloss.Width(s) > w {
		s = ansi.Truncate(s, w, "") + "\x1b[0m"
	}
	return s + strings.Repeat(" ", max(0, w-lipgloss.Width(s)))
}

func pad(s string, w int) string {
	s = ansi.Truncate(s, w, "~")
	return s + strings.Repeat(" ", max(0, w-lipgloss.Width(s)))
}

func widest(lines []string) int {
	w := 0
	for _, l := range lines {
		w = max(w, lipgloss.Width(l))
	}
	return w
}

// box draws the frame with the title in the middle of the top border.
func box(title string, inner []string, w, h int, border lipgloss.Style) []string {
	iw := max(0, w-2)
	if title != "" {
		title = " " + title + " "
	}
	if lipgloss.Width(title) > iw {
		title = ansi.Truncate(title, iw, "") + "\x1b[0m"
	}
	tw := lipgloss.Width(title)
	left := (iw - tw) / 2
	out := []string{border.Render("┌"+strings.Repeat("─", left)) + title + border.Render(strings.Repeat("─", iw-tw-left)+"┐")}
	for i := 0; i < h-2; i++ {
		line := ""
		if i < len(inner) {
			line = inner[i]
		}
		out = append(out, border.Render("│")+fit(line, iw)+border.Render("│"))
	}
	return append(out, border.Render("└"+strings.Repeat("─", iw)+"┘"))
}

// overlay puts a dialog in the center of the lines below, without erasing the rest.
func overlay(base, dialog []string, width int) []string {
	dw := widest(dialog)
	top := max(0, (len(base)-len(dialog))/2)
	x := max(0, (width-dw)/2)
	out := slices.Clone(base)
	for i, d := range dialog {
		r := top + i
		if r >= len(out) {
			break
		}
		left := fit(ansi.Truncate(out[r], x, ""), x)
		right := ansi.TruncateLeft(out[r], x+dw, "")
		out[r] = left + "\x1b[0m" + fit(d, dw) + right
	}
	return out
}

func badge(text string, active bool) string {
	if active {
		return sCrumbActive.Render("<" + text + ">")
	}
	return sCrumb.Render("<" + text + ">")
}

func keyName(k string) string { return strings.ReplaceAll(k, "+", "-") }

// ---------------------------------------------------------------- screen

// View draws the screen: header, prompt, main frame and the trail at the bottom.
func (m *Model) View() string {
	if m.quit {
		return ""
	}
	lines := m.header()
	lines = append(lines, m.promptBox()...)
	lines = append(lines, m.mainBox()...)
	lines = append(lines, m.crumbsLine(), m.bottomLine())
	for i, l := range lines {
		lines[i] = ansi.Truncate(l, m.width, "")
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------- header

func (m *Model) counts(view string) (total, on int, ok bool) {
	if !m.loaded[view] {
		return 0, 0, false
	}
	for _, r := range m.rows[view] {
		if r.Status == "not-installed" {
			continue
		}
		total++
		if r.Status == "running" || r.Status == "installed" {
			on++
		}
	}
	return total, on, true
}

func (m *Model) infoLines() []string {
	version := m.version
	if version == "" {
		version = "dev"
	}
	type info struct{ label, value string }
	items := []info{
		{i18n.T("Machine:"), sValue.Render(m.backend.Header())},
		{"romanctl:", sValue.Render("v" + strings.TrimPrefix(version, "v"))},
	}
	value := func(view, format string) string {
		total, on, ok := m.counts(view)
		if !ok {
			return sDim.Render("-")
		}
		return sValue.Render(i18n.Sprintf(format, on, total))
	}
	items = append(items,
		info{i18n.T("Projects:"), value(viewProjects, "%d/%d up")},
		info{i18n.T("Services:"), value(viewServices, "%d/%d running")},
		info{i18n.T("Tools:"), value(viewTools, "%d/%d installed")})
	cpu, mem := sDim.Render("-"), sDim.Render("-")
	if s := m.system; s != nil {
		if s.SystemCPU != nil {
			cpu = levelStyle(*s.SystemCPU / 100).Render(fmt.Sprintf("%.0f%%", *s.SystemCPU))
		} else {
			cpu = sValue.Render(i18n.Sprintf("%.2f cores", s.OwnCPUCores))
		}
		if s.Memory.Total > 0 {
			used := s.Memory.Total - s.Memory.Available
			ratio := float64(used) / float64(s.Memory.Total)
			mem = levelStyle(ratio).Render(fmt.Sprintf("%.0f%%", ratio*100)) + sMuted.Render(fmt.Sprintf(" %.1f/%s", float64(used)/(1<<30), gb(s.Memory.Total)))
		}
	}
	items = append(items, info{"CPU:", cpu}, info{"MEM:", mem})
	width := 0
	for _, it := range items {
		width = max(width, lipgloss.Width(it.label))
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = " " + sLabel.Render(pad(it.label, width)) + " " + it.value
	}
	return out
}

type hint struct {
	key, label string
	view       bool
}

func (m *Model) hints() []hint {
	switch m.mode {
	case modeFilter:
		return []hint{{key: "enter", label: i18n.T("apply")}, {key: "esc", label: i18n.T("clear")}}
	case modeCommand:
		return []hint{{key: "tab", label: i18n.T("complete")}, {key: "enter", label: i18n.T("go")}, {key: "esc", label: i18n.T("cancel")}}
	case modeHelp, modeDetail:
		return []hint{{key: "esc", label: i18n.T("back")}}
	case modeLogs:
		return []hint{{key: "j/k", label: i18n.T("scroll")}, {key: "f", label: i18n.T("follow")}, {key: "g", label: i18n.T("start")}, {key: "esc", label: i18n.T("back")}}
	case modeActions:
		return []hint{{key: "enter", label: i18n.T("select")}, {key: "esc", label: i18n.T("back")}}
	case modeVersions:
		return []hint{{key: "j/k", label: i18n.T("move")}, {key: "enter", label: i18n.T("install this one")}, {key: "esc", label: i18n.T("back")}}
	case modeConfirm:
		return []hint{{key: "y", label: i18n.T("yes")}, {key: "n", label: i18n.T("no")}}
	}
	var out []hint
	for _, a := range globalActions[m.view] {
		out = append(out, hint{key: keyName(a.key), label: i18n.T(a.label)})
	}
	if r, ok := m.selected(); ok {
		for _, a := range m.rowActions(r) {
			out = append(out, hint{key: keyName(a.key), label: i18n.T(a.label)})
		}
		enter := i18n.T("actions")
		if m.view == viewTools {
			enter = i18n.T("versions")
		}
		out = append(out, hint{key: "enter", label: enter}, hint{key: "d", label: i18n.T("details")})
		if m.view == viewServices || m.view == viewProjects {
			out = append(out, hint{key: "l", label: i18n.T("logs")})
		}
	}
	if m.view == viewTools {
		out = append(out, hint{key: "< >", label: i18n.T("category")})
	}
	return append(out, hint{key: "/", label: i18n.T("filter")}, hint{key: ":", label: i18n.T("go to")}, hint{key: "?", label: i18n.T("help")}, hint{key: "q", label: i18n.T("quit")})
}

func (m *Model) viewHints() []hint {
	out := make([]hint, len(viewOrder))
	for i, v := range viewOrder {
		out[i] = hint{key: fmt.Sprint(i + 1), label: strings.ToLower(viewTitle(v)), view: true}
	}
	return out
}

// hintColumns arranges the keys in top-to-bottom columns, as far as the width allows.
func hintColumns(groups [][]hint, rows, avail int) ([]string, int) {
	out := make([]string, rows)
	used := 0
	for _, group := range groups {
		for start := 0; start < len(group); start += rows {
			col := group[start:min(start+rows, len(group))]
			kw, lw := 0, 0
			for _, h := range col {
				kw = max(kw, lipgloss.Width(h.key)+2)
				lw = max(lw, lipgloss.Width(h.label))
			}
			cw := kw + 1 + lw + 3
			if used+cw-3 > avail {
				return out, used
			}
			for i := range rows {
				cell := ""
				if i < len(col) {
					style := sKey
					if col[i].view {
						style = sViewKey
					}
					cell = style.Render(pad("<"+col[i].key+">", kw)) + " " + sHintLabel.Render(col[i].label)
				}
				out[i] += fit(cell, cw)
			}
			used += cw
		}
	}
	return out, used
}

func (m *Model) header() []string {
	if !m.bigHeader() {
		return []string{m.compactHeader()}
	}
	info := m.infoLines()
	rows := len(info)
	infoW := widest(info) + 3
	logoW := widest(logo)
	avail := m.width - infoW
	groups := [][]hint{m.hints()}
	if m.mode == modeTable {
		groups = [][]hint{m.viewHints(), m.hints()}
	}
	cols, used := hintColumns(groups, rows, 1<<20)
	showLogo := avail-used >= logoW+2
	if showLogo {
		avail -= logoW + 2
	} else if used > avail && len(groups) > 1 {
		cols, _ = hintColumns(groups[1:], rows, avail)
	} else {
		cols, _ = hintColumns(groups, rows, avail)
	}
	out := make([]string, rows)
	for i := range rows {
		line := fit(info[i], infoW) + fit(cols[i], avail)
		if showLogo {
			l := ""
			if j := i - 1; j >= 0 && j < len(logo) {
				l = sLogo.Render(logo[j])
			}
			line += "  " + fit(l, logoW)
		}
		out[i] = line
	}
	return out
}

func (m *Model) compactHeader() string {
	parts := []string{sLogo.Render("romanctl")}
	for i, v := range viewOrder {
		label := fmt.Sprintf("%d %s", i+1, strings.ToLower(viewTitle(v)))
		if v == m.view {
			parts = append(parts, sCrumbActive.Render(label))
		} else {
			parts = append(parts, sViewKey.Render(fmt.Sprint(i+1))+" "+sHintLabel.Render(strings.ToLower(viewTitle(v))))
		}
	}
	left := strings.Join(parts, " ")
	right := sDim.Render(m.backend.Header())
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 2 {
		return left
	}
	return left + strings.Repeat(" ", gap) + right
}

// ---------------------------------------------------------------- prompt

func (m *Model) promptBox() []string {
	var text string
	border := sBorder
	switch m.mode {
	case modeCommand:
		hint := ""
		if c := completeCommand(m.command); c != "" && c != strings.ToLower(m.command) {
			hint = sDim.Render(c[len(m.command):])
		}
		text = sViewKey.Render(":") + sPrompt.Render(m.command) + sCaret.Render(" ") + hint
		border = lipgloss.NewStyle().Foreground(cAqua)
	case modeFilter:
		text = sViewKey.Render("/") + sPrompt.Render(m.filter) + sCaret.Render(" ")
		border = lipgloss.NewStyle().Foreground(cFuchsia)
	default:
		return nil
	}
	return box("", []string{" " + text}, m.width, 3, border)
}

// ---------------------------------------------------------------- main frame

func (m *Model) mainBox() []string {
	h, w := m.boxHeight(), m.width
	var title string
	var inner []string
	switch m.mode {
	case modeHelp:
		title, inner = sTitle.Render(i18n.T("help")), m.helpLines()
	case modeDetail:
		r, _ := m.selected()
		lines := m.detailLines()
		top := max(0, min(m.detailTop, len(lines)-m.bodyHeight()))
		title, inner = sTitle.Render(i18n.T("details"))+sTitleScope.Render("("+r.Key+")"), lines[top:]
		if len(lines) > m.bodyHeight() {
			title += sTitleCount.Render(fmt.Sprintf("[%d%%]", 100*(top+m.bodyHeight())/len(lines)))
		}
	case modeLogs:
		state := i18n.T("paused")
		if m.follow {
			state = i18n.T("following")
		}
		title, inner = sTitle.Render("logs")+sTitleScope.Render("("+m.logTitle+")")+sTitleCount.Render("["+state+"]"), m.logWindow()
	default:
		if m.view == viewSystem {
			title, inner = sTitle.Render(strings.ToLower(viewTitle(viewSystem))), m.systemLines()
		} else {
			title, inner = m.tableTitle(), m.tableLines()
		}
	}
	out := box(title, inner, w, h, sBorder)
	switch m.mode {
	case modeActions:
		out = overlay(out, m.actionDialog(), w)
	case modeVersions:
		out = overlay(out, m.versionDialog(), w)
	case modeConfirm:
		out = overlay(out, m.confirmDialog(), w)
	}
	return out
}

func (m *Model) tableTitle() string {
	title := sTitle.Render(strings.ToLower(viewTitle(m.view)))
	if m.view == viewTools {
		scope := i18n.T("all")
		if cats := m.categories(); m.category > 0 && m.category <= len(cats) {
			scope = cats[m.category-1]
		}
		title += sTitleScope.Render("(" + scope + ")")
	}
	title += sTitleCount.Render(fmt.Sprintf("[%d]", len(m.visible())))
	if m.filter != "" {
		title += " " + sTitleScope.Render("</"+m.filter+">")
	}
	return title
}

var columns = map[string][]string{
	viewServices: {"NAME", "SERVICE", "STATE", "ADDRESS"},
	viewProjects: {"NAME", "TYPE", "STATE", "ADDRESS", "CHANGED"},
	viewTools:    {"NAME", "CATEGORY", "STATE", "INSTALLED", "LATEST", "DESCRIPTION"},
}

const statusColumn = 2

const (
	versionColumn = 3
	latestColumn  = 4
)

// latestOf is the latest known version: the list's, or the one that came with the details.
func (m *Model) latestOf(key string) (tools.Upstream, bool) {
	if u, ok := m.upstream[key]; ok && u.Latest != "" {
		return u, true
	}
	u, ok := m.latest[key]
	return u, ok
}

// cell is the column text; in Tools, LATEST comes from the query running in the background.
func (m *Model) cell(r Row, i int) string {
	if i >= len(r.Columns) {
		return ""
	}
	if m.view != viewTools || i != latestColumn {
		return r.Columns[i]
	}
	u, ok := m.latestOf(r.Key)
	switch {
	case !ok:
		return "..."
	case u.Latest == "":
		return "-"
	}
	return u.Latest
}

func (m *Model) updateAvailable(r Row) bool {
	u, ok := m.latestOf(r.Key)
	return m.view == viewTools && ok && r.Status == "installed" && u.Latest != "" && !tools.UpToDate(r.Columns[versionColumn], u.Latest)
}

func (m *Model) columnWidths(rows []Row) []int {
	head := columns[m.view]
	widths := make([]int, len(head))
	for i, h := range head {
		widths[i] = lipgloss.Width(i18n.T(h))
	}
	for _, r := range rows {
		for i := range r.Columns {
			if i < len(widths) {
				widths[i] = max(widths[i], lipgloss.Width(m.cell(r, i)))
			}
		}
		if busy, ok := m.running[m.view+"/"+r.Key]; ok && statusColumn < len(widths) {
			widths[statusColumn] = max(widths[statusColumn], lipgloss.Width(busy)+2)
		}
	}
	// The last column gets whatever is left; the one before it must not push it out.
	total := m.innerWidth() - 2
	for i := 0; i < len(widths)-1; i++ {
		widths[i] = min(widths[i], max(8, total/3))
		total -= widths[i] + 3
	}
	if total > 0 {
		widths[len(widths)-1] = min(widths[len(widths)-1], total)
	}
	return widths
}

func (m *Model) tableLines() []string {
	rows := m.visible()
	if m.loading && !m.loaded[m.view] {
		return []string{"", "  " + sBusy.Render(spinnerFrame(m.spin)+" "+i18n.T("Loading..."))}
	}
	widths := m.columnWidths(rows)
	var heads []string
	for i, h := range columns[m.view] {
		heads = append(heads, pad(i18n.T(h), widths[i]))
	}
	out := []string{" " + sHead.Render(strings.Join(heads, "   "))}
	if len(rows) == 0 {
		text := i18n.T("Nothing here.")
		if m.filter != "" {
			text = i18n.T("Nothing found. Esc clears the filter.")
		} else if m.view == viewProjects {
			text = i18n.T("No projects yet. Press n to create the first one.")
		}
		return append(out, "", "  "+sWarn.Render(text))
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
	base := rowStyle(r.Status)
	for i := range widths {
		text := m.cell(r, i)
		if i == statusColumn && isBusy {
			text = spinnerFrame(m.spin) + " " + busy
		}
		cells[i] = pad(text, widths[i])
		if selected {
			continue
		}
		switch {
		case i == statusColumn && isBusy:
			cells[i] = sBusy.Bold(true).Render(cells[i])
		case i == statusColumn:
			cells[i] = statusStyle(r.Status).Render(cells[i])
		case i == latestColumn && m.updateAvailable(r):
			cells[i] = sWarn.Bold(true).Render(cells[i])
		case i == latestColumn && cells[i] != "" && strings.TrimSpace(cells[i]) == "...":
			cells[i] = sDim.Render(cells[i])
		case i == 0:
			cells[i] = base.Bold(true).Render(cells[i])
		default:
			cells[i] = base.Render(cells[i])
		}
	}
	line := " " + strings.Join(cells, "   ")
	if selected {
		return sCursor.Render(fit(line, m.innerWidth()))
	}
	return line
}

func spinnerFrame(i int) string { return string("|/-\\"[i%4]) }

// ---------------------------------------------------------------- system

func gauge(used, total int64, width int) string {
	if total <= 0 {
		return sDim.Render(strings.Repeat("-", width))
	}
	ratio := float64(used) / float64(total)
	filled := max(0, min(int(ratio*float64(width)), width))
	return levelStyle(ratio).Render(strings.Repeat("#", filled)) + sDim.Render(strings.Repeat(".", width-filled))
}

func gb(n int64) string { return fmt.Sprintf("%.1f GB", float64(n)/(1<<30)) }

func (m *Model) systemLines() []string {
	if m.system == nil {
		return []string{"", "  " + sBusy.Render(spinnerFrame(m.spin)+" "+i18n.T("Measuring..."))}
	}
	s := m.system
	bar := max(10, min(m.innerWidth()-46, 50))
	row := func(label string, used, total int64, tail string) string {
		return " " + sLabel.Render(pad(label, 10)) + " " + sDim.Render("[") + gauge(used, total, bar) + sDim.Render("]") + " " + sValue.Render(tail)
	}
	out := []string{"", " " + sSection.Render(i18n.T("RESOURCES"))}
	if s.SystemCPU != nil {
		out = append(out, row("CPU", int64(*s.SystemCPU), 100, fmt.Sprintf("%.0f%%", *s.SystemCPU)))
	} else {
		out = append(out, " "+sLabel.Render(pad("CPU", 10))+" "+sValue.Render(i18n.Sprintf("%.2f cores in your processes", s.OwnCPUCores)))
	}
	used := s.Memory.Total - s.Memory.Available
	out = append(out, row(i18n.T("Memory"), used, s.Memory.Total, i18n.Sprintf("%s of %s", gb(used), gb(s.Memory.Total))))
	if s.Memory.SwapTotal > 0 {
		swap := s.Memory.SwapTotal - s.Memory.SwapFree
		out = append(out, row("Swap", swap, s.Memory.SwapTotal, i18n.Sprintf("%s of %s", gb(swap), gb(s.Memory.SwapTotal))))
	}
	disk := s.Disk.Size - s.Disk.Free
	out = append(out, row(i18n.T("Disk"), disk, s.Disk.Size, i18n.Sprintf("%s of %s", gb(disk), gb(s.Disk.Size))))
	if b := s.Battery; b != nil {
		tail := fmt.Sprintf("%.0f%%", b.Percent)
		if b.Power > 0 {
			tail += fmt.Sprintf(", %.1f W", b.Power)
		}
		if b.HoursLeft > 0 {
			tail += fmt.Sprintf(", %dh%02d", int(b.HoursLeft), int(b.HoursLeft*60)%60)
		}
		out = append(out, row(i18n.T("Battery"), 100-int64(b.Percent), 100, tail))
	}
	out = append(out, "", " "+sSection.Render(i18n.T("PROCESSES")))
	out = append(out, " "+sHead.Render(fmt.Sprintf("%-8s %6s %8s   %s", "PID", "CPU", "MEM", i18n.T("NAME"))))
	top := slices.Clone(s.Top)
	sort.SliceStable(top, func(i, j int) bool { return top[i].CPU > top[j].CPU })
	for i, p := range top {
		if i >= max(3, m.bodyHeight()-len(out)) {
			break
		}
		cpu := levelStyle(p.CPU / 100).Render(fmt.Sprintf("%5.1f%%", p.CPU))
		out = append(out, " "+sMuted.Render(fmt.Sprintf("%-8d", p.PID))+" "+cpu+" "+sMuted.Render(fmt.Sprintf("%5.0f MB", float64(p.Memory)/(1<<20)))+"   "+lipgloss.NewStyle().Foreground(cAqua).Render(p.Name))
	}
	return out
}

// ---------------------------------------------------------------- details, help and logs

func (m *Model) detailLines() []string {
	r, ok := m.selected()
	if !ok {
		return nil
	}
	out := []string{""}
	for _, d := range r.Detail {
		if d[1] == "" {
			continue
		}
		value := sValue.Render(d[1])
		if d[0] == i18n.T("State") {
			value = statusStyle(r.Status).Render(d[1])
		}
		out = append(out, " "+sSection.Render(pad(d[0]+":", 14))+" "+value)
	}
	if m.view == viewTools {
		out = append(out, m.upstreamLines(r)...)
	}
	out = append(out, "", " "+sLabel.Render(i18n.T("What you can do:")))
	for _, a := range m.rowActions(r) {
		out = append(out, " "+sKey.Render(pad("<"+keyName(a.key)+">", 9))+" "+sHintLabel.Render(i18n.T(a.label)))
	}
	if m.view == viewTools {
		out = append(out, m.readmeLines(r.Key)...)
	}
	return out
}

func (m *Model) upstreamLines(r Row) []string {
	field := func(label, value string) string { return " " + sSection.Render(pad(label+":", 14)) + " " + value }
	u, ok := m.upstream[r.Key]
	if !ok {
		u, ok = m.latest[r.Key]
	}
	if !ok {
		return []string{field(i18n.T("Available"), sBusy.Render(spinnerFrame(m.spin)+" "+i18n.T("checking...")))}
	}
	latest := sDim.Render(i18n.T("not reported"))
	if u.Latest != "" {
		latest = sValue.Render(u.Latest)
		installed := r.Columns[3]
		switch {
		case installed == "":
		case tools.UpToDate(installed, u.Latest):
			latest += " " + sOK.Render(i18n.T("(the installed one)"))
		default:
			latest += " " + sWarn.Bold(true).Render(i18n.T("(update available)"))
		}
	}
	out := []string{field(i18n.T("Available"), latest), field(i18n.T("Source"), sValue.Render(u.Source))}
	if u.URL != "" {
		out = append(out, field(i18n.T("Project"), lipgloss.NewStyle().Foreground(cBlue).Underline(true).Render(u.URL)))
	}
	return out
}

// readmeLines renders the README once per width: the markdown does not change every frame.
func (m *Model) readmeLines(key string) []string {
	u, ok := m.upstream[key]
	if !ok || u.Readme == "" {
		return nil
	}
	width := max(20, m.innerWidth()-2)
	cacheKey := fmt.Sprintf("%s/%d", key, width)
	if lines, ok := m.readme[cacheKey]; ok {
		return lines
	}
	lines := append([]string{"", " " + sLabel.Render("README")}, strings.Split(ui.Markdown(u.Readme, width, darkBackground), "\n")...)
	m.readme[cacheKey] = lines
	return lines
}

func (m *Model) helpLines() []string {
	section := func(title string, items ...[2]string) []string {
		out := []string{" " + sLabel.Render(title)}
		for _, it := range items {
			out = append(out, " "+sKey.Render(pad(it[0], 14))+" "+sHintLabel.Render(it[1]))
		}
		return out
	}
	blocks := [][]string{
		section(i18n.T("NAVIGATE"),
			[2]string{"<tab> <1-4>", i18n.T("change screen")},
			[2]string{"<j/k>", i18n.T("move in the list")},
			[2]string{"<g> <G>", i18n.T("start and end of the list")},
			[2]string{"<enter>", i18n.T("see the item's actions")},
			[2]string{"<d>", i18n.T("item details")},
			[2]string{"<l>", i18n.T("item logs")},
			[2]string{"<esc>", i18n.T("go back and clear the filter")},
			[2]string{"<q>", i18n.T("quit")}),
		section(i18n.T("SEARCH"),
			[2]string{"</>", i18n.T("filter the list")},
			[2]string{"<< >>", i18n.T("change category, in Tools")},
			[2]string{"<:>", i18n.T("jump: :tools, :svc, :prj, :sys, :db, :k8s")},
			[2]string{"<ctrl-r>", i18n.T("refresh now")}),
	}
	var items [][2]string
	for _, a := range globalActions[m.view] {
		items = append(items, [2]string{"<" + keyName(a.key) + ">", i18n.T(a.label)})
	}
	for _, a := range actions[m.view] {
		item := [2]string{"<" + keyName(a.key) + ">", i18n.T(a.label)}
		if !slices.Contains(items, item) {
			items = append(items, item)
		}
	}
	if len(items) > 0 {
		blocks = append(blocks, section(strings.ToUpper(viewTitle(m.view)), items...))
	}
	return layoutBlocks(blocks, m.innerWidth())
}

// layoutBlocks places the sections side by side while they fit, like k9s's help.
func layoutBlocks(blocks [][]string, width int) []string {
	out := []string{""}
	var row [][]string
	used := 0
	flush := func() {
		if len(row) == 0 {
			return
		}
		height := 0
		for _, b := range row {
			height = max(height, len(b))
		}
		for i := range height {
			line := ""
			for _, b := range row {
				cell := ""
				if i < len(b) {
					cell = b[i]
				}
				line += fit(cell, widest(b)+4)
			}
			out = append(out, line)
		}
		out = append(out, "")
		row, used = nil, 0
	}
	for _, b := range blocks {
		w := widest(b) + 4
		if used > 0 && used+w > width {
			flush()
		}
		row = append(row, b)
		used += w
	}
	flush()
	return out
}

var (
	reLogError = regexp.MustCompile(`(?i)\b(error|err|fatal|panic|fail(ed|ure)?|exception)\b`)
	reLogWarn  = regexp.MustCompile(`(?i)\b(warn(ing)?|deprecated)\b`)
	reLogDebug = regexp.MustCompile(`(?i)\b(debug|trace)\b`)
)

func logStyle(line string) lipgloss.Style {
	switch {
	case reLogError.MatchString(line):
		return sBad
	case reLogWarn.MatchString(line):
		return sWarn
	case reLogDebug.MatchString(line):
		return sDim
	}
	return lipgloss.NewStyle().Foreground(cAqua)
}

func (m *Model) logWindow() []string {
	lines := m.logLines()
	height := m.bodyHeight()
	top := max(0, min(m.logTop, max(0, len(lines)-height)))
	out := make([]string, 0, height)
	for i := top; i < min(top+height, len(lines)); i++ {
		out = append(out, " "+logStyle(lines[i]).Render(lines[i]))
	}
	return out
}

// ---------------------------------------------------------------- dialogs

func (m *Model) actionDialog() []string {
	r, _ := m.selected()
	var labels []string
	for _, a := range m.menu {
		labels = append(labels, i18n.T(a.label))
	}
	lw := widest(labels)
	inner := []string{""}
	for i, a := range m.menu {
		if i == m.menuPos {
			inner = append(inner, sCursor.Render(fit(" "+pad("<"+keyName(a.key)+">", 9)+" "+labels[i]+" ", lw+12)))
			continue
		}
		inner = append(inner, " "+sKey.Render(pad("<"+keyName(a.key)+">", 9))+" "+sHintLabel.Render(labels[i]))
	}
	inner = append(inner, "")
	title := sTitle.Render(i18n.T("actions")) + sTitleScope.Render("("+r.Key+")")
	w := max(lw+14, lipgloss.Width(title)+6)
	return box(title, inner, min(w, m.width), len(inner)+2, lipgloss.NewStyle().Foreground(cAqua))
}

func (m *Model) confirmDialog() []string {
	question := m.pending.question
	if question == "" {
		question = i18n.Sprintf(m.pending.confirm, m.pendKey)
	}
	yes := " <y> " + i18n.T("yes") + " "
	no := " <n> " + i18n.T("no") + " "
	buttons := sKey.Render(yes) + "   " + sCursor.Render(no)
	var info []string
	if m.view == viewTools {
		info = m.installInfo(m.pendKey, m.pending.version)
	}
	w := min(max(lipgloss.Width(question), lipgloss.Width(yes+no)+3, widest(info))+6, m.width)
	center := func(s string) string {
		return strings.Repeat(" ", max(0, (w-2-lipgloss.Width(s))/2)) + s
	}
	inner := []string{"", center(sValue.Render(question)), ""}
	if len(info) > 0 {
		for _, l := range info {
			inner = append(inner, "  "+l)
		}
		inner = append(inner, "")
	}
	inner = append(inner, center(buttons), "")
	return box(sBad.Bold(true).Render(i18n.T("confirm")), inner, w, len(inner)+2, sBad)
}

// installInfo tells what is on the machine and what the install will bring, and from where.
func (m *Model) installInfo(key, chosen string) []string {
	r := Row{Key: key, Columns: make([]string, versionColumn+1)}
	for _, row := range m.rows[viewTools] {
		if row.Key == key && len(row.Columns) > versionColumn {
			r = row
		}
	}
	var out []string
	if v := r.Columns[versionColumn]; v != "" {
		out = append(out, sSection.Render(pad(i18n.T("Installed")+":", 14))+" "+sValue.Render(v))
	}
	if chosen != "" {
		out = append(out, sSection.Render(pad(i18n.T("Will install")+":", 14))+" "+sValue.Bold(true).Render(chosen))
		if u, ok := m.latestOf(key); ok && u.Source != "" {
			out = append(out, sSection.Render(pad(i18n.T("Source")+":", 14))+" "+sValue.Render(u.Source))
		}
		return out
	}
	for _, l := range m.upstreamLines(r) {
		out = append(out, strings.TrimPrefix(l, " "))
	}
	return out
}

// versionDialog lists the tool's versions: the latest first, the installed one marked.
func (m *Model) versionDialog() []string {
	key := m.verKey
	r, _ := m.selected()
	installed := ""
	if len(r.Columns) > versionColumn {
		installed = r.Columns[versionColumn]
	}
	title := sTitle.Render(i18n.T("versions")) + sTitleScope.Render("("+key+")")
	list, ok := m.versions[key]
	var inner []string
	switch {
	case m.verErr[key] != "":
		inner = []string{"", " " + sWarn.Render(m.verErr[key]), "", " " + sHintLabel.Render(i18n.T("Latest version only: use <i> to install or <r> to reinstall.")), ""}
	case !ok:
		inner = []string{"", " " + sBusy.Render(spinnerFrame(m.spin)+" "+i18n.T("checking versions...")), ""}
	case len(list) == 0:
		inner = []string{"", " " + sDim.Render(i18n.T("No version found.")), ""}
	default:
		lw := 0
		for _, v := range list {
			lw = max(lw, lipgloss.Width(v.Label))
		}
		height := max(3, m.bodyHeight()-4)
		top := max(0, min(m.verPos-height/2, len(list)-height))
		inner = []string{""}
		for i := top; i < min(top+height, len(list)); i++ {
			v := list[i]
			var notes []string
			if i == 0 {
				notes = append(notes, i18n.T("latest"))
			}
			if installed != "" && tools.UpToDate(installed, v.Label) {
				notes = append(notes, i18n.T("installed"))
			}
			if v.Note != "" {
				notes = append(notes, v.Note)
			}
			line := " " + pad(v.Label, lw) + "  " + strings.Join(notes, " · ") + " "
			switch {
			case i == m.verPos:
				inner = append(inner, sCursor.Render(fit(line, lw+40)))
			case installed != "" && tools.UpToDate(installed, v.Label):
				inner = append(inner, sOK.Bold(true).Render(line))
			case i == 0:
				inner = append(inner, sValue.Render(line))
			default:
				inner = append(inner, sMuted.Render(line))
			}
		}
		inner = append(inner, "")
	}
	w := min(max(widest(inner)+4, lipgloss.Width(title)+6, 36), m.width)
	return box(title, inner, w, len(inner)+2, lipgloss.NewStyle().Foreground(cAqua))
}

// ---------------------------------------------------------------- footer

func (m *Model) crumbs() []string {
	out := []string{strings.ToLower(viewTitle(m.view))}
	r, ok := m.selected()
	switch m.mode {
	case modeDetail:
		if ok {
			out = append(out, r.Key, i18n.T("details"))
		}
	case modeLogs:
		out = append(out, m.logTitle, "logs")
	case modeActions, modeConfirm:
		if ok {
			out = append(out, r.Key)
		}
	case modeVersions:
		if ok {
			out = append(out, r.Key, i18n.T("versions"))
		}
	case modeHelp:
		out = append(out, i18n.T("help"))
	}
	return out
}

func (m *Model) crumbsLine() string {
	list := m.crumbs()
	parts := make([]string, len(list))
	for i, c := range list {
		parts[i] = badge(c, i == len(list)-1)
	}
	left := " " + strings.Join(parts, " ")
	var right string
	if n := len(m.running); n > 0 {
		right = sBusy.Render(spinnerFrame(m.spin) + " " + i18n.Sprintf("%d in progress", n))
	} else if m.view != viewSystem && m.mode == modeTable {
		right = sDim.Render(fmt.Sprintf("%d/%d", len(m.visible()), len(m.rows[m.view])))
	}
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 1
	if right == "" || gap < 2 {
		return left
	}
	return left + strings.Repeat(" ", gap) + right
}

// bottomLine is the message line, centered; on a small screen it shows the keys.
func (m *Model) bottomLine() string {
	if m.flash != "" && time.Since(m.flashAt) < 8*time.Second {
		style := sOK.Bold(true)
		if m.flashErr {
			style = sBad.Bold(true)
		}
		text := style.Render(m.flash)
		return strings.Repeat(" ", max(0, (m.width-lipgloss.Width(text))/2)) + text
	}
	if m.bigHeader() {
		return ""
	}
	var parts []string
	for _, h := range m.hints() {
		parts = append(parts, sKey.Render("<"+h.key+">")+" "+sHintLabel.Render(h.label))
	}
	return " " + strings.Join(parts, "  ")
}

// Run opens the panel. While it is on screen, nothing else writes to the terminal.
func Run(version string) error {
	darkBackground = ui.DarkBackground()
	ui.SetSilent(true)
	defer ui.SetSilent(false)
	m := New(newRealBackend())
	m.version = version
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}
