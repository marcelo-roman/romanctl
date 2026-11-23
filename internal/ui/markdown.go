package ui

import (
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Markdown renders the text for the terminal; if rendering fails, it returns the raw text.
func Markdown(text string, width int, dark bool) string {
	style := "dark"
	if !dark {
		style = "light"
	}
	r, err := glamour.NewTermRenderer(glamour.WithStandardStyle(style), glamour.WithWordWrap(width))
	if err != nil {
		return text
	}
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	lines := strings.Split(strings.Trim(out, "\n"), "\n")
	for i, l := range lines {
		lines[i] = trailingPad.ReplaceAllString(l, "") + "\x1b[0m"
	}
	return strings.Join(lines, "\n")
}

// glamour pads each line with colored spaces up to the full width.
var trailingPad = regexp.MustCompile(`(?:\x1b\[[0-9;]*m| )+$`)

// Width is the terminal width, with a default for when the output is not a terminal.
func Width() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 100
}

func init() { lipgloss.SetHasDarkBackground(DarkBackground()) }

// DarkBackground picks the theme without querying the terminal: the query (OSC 11) hangs on
// terminals that do not answer. Uses COLORFGBG when set; otherwise assumes a dark background.
func DarkBackground() bool {
	v := os.Getenv("COLORFGBG")
	if v == "" {
		return true
	}
	parts := strings.Split(v, ";")
	switch parts[len(parts)-1] {
	case "7", "15":
		return false
	}
	return true
}
