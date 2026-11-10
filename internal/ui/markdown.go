package ui

import (
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Markdown desenha o texto no terminal; se não renderizar, devolve o texto cru.
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

// O glamour completa cada linha com espaços coloridos até a largura toda.
var trailingPad = regexp.MustCompile(`(?:\x1b\[[0-9;]*m| )+$`)

// Width é a largura do terminal, com um padrão para quando a saída não é um terminal.
func Width() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 100
}

func init() { lipgloss.SetHasDarkBackground(DarkBackground()) }

// DarkBackground decide o tema sem perguntar ao terminal: a pergunta (OSC 11) trava em
// terminais que não respondem. Usa o COLORFGBG quando existe; sem ele, fundo escuro.
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
