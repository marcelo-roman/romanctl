package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"io"
	"os"
	"romanctl/internal/i18n"
)

var JSONOutput bool

func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// usageError signals incorrect usage: exits with code 2, distinct from an execution failure.
type usageError struct{ msg string }

func (e usageError) Error() string { return e.msg }

func Usagef(f string, a ...any) error { return usageError{i18n.Sprintf(f, a...)} }

func IsUsage(err error) bool {
	var u usageError
	return errors.As(err, &u)
}

var SGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

var SRed = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

var SYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

var SGray = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

var SBlue = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))

var STitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))

var SBoldTxt = lipgloss.NewStyle().Bold(true)

// messages is where messages go: with --json, stdout carries only the JSON.
func messages() io.Writer {
	if JSONOutput {
		return os.Stderr
	}
	return os.Stdout
}

// ErrReported signals a failure that was already reported on the output.
var ErrReported = errors.New(i18n.T("failure already reported"))

func Okf(f string, a ...any) { emit(SGreen.Render("✔ ") + i18n.Sprintf(f, a...)) }

func Failf(f string, a ...any) { emit(SRed.Render("✘ ") + i18n.Sprintf(f, a...)) }

func Infof(f string, a ...any) { emit(SBlue.Render("• ") + i18n.Sprintf(f, a...)) }

func Warnf(f string, a ...any) { emit(SYellow.Render("! ") + i18n.Sprintf(f, a...)) }

// emit writes a message without mixing it with the Spin animation.
func emit(line string) {
	if silent {
		return
	}
	spinMu.Lock()
	defer spinMu.Unlock()
	clearSpinnerLine()
	fmt.Fprintln(messages(), line)
}

var silent bool

// SetSilent discards messages: a full-screen panel must not have its screen scribbled on.
func SetSilent(on bool) { silent = on }
