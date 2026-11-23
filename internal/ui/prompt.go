package ui

import (
	"errors"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"golang.org/x/term"
	"os"
	"romanctl/internal/i18n"
)

var errBack = errors.New("back")

var errCancel = errors.New("canceled")

// ErrCanceled is a "no" answer to a confirmation: it cancels without being a failure.
var ErrCanceled = errCancel

func IsAbort(err error) bool { return errors.Is(err, errBack) || errors.Is(err, errCancel) }

var NavHint = i18n.T("Esc/← goes back · Ctrl+C cancels")

// Interactive reports whether someone is there to answer a menu.
func Interactive() bool {
	return !JSONOutput && term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// Ask runs a single-field prompt. list=true enables "←" to go back
// (only in lists; in text fields the arrow must move the cursor).
func Ask(list bool, field huh.Field) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return Usagef("this command needs a terminal; pass everything as arguments (see: romanctl help)")
	}
	keys := []string{"ctrl+c", "esc"}
	if list {
		keys = append(keys, "left")
	}
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys(keys...), key.WithHelp("esc", "back"))

	var last string
	err := huh.NewForm(huh.NewGroup(field)).WithKeyMap(km).
		WithProgramOptions(tea.WithFilter(func(_ tea.Model, msg tea.Msg) tea.Msg {
			if k, ok := msg.(tea.KeyMsg); ok {
				last = k.String()
			}
			return msg
		})).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		if last == "ctrl+c" {
			return errCancel
		}
		return errBack
	}
	return err
}

// Wizard runs steps in sequence; Esc/← goes back to the previous step.
func Wizard(steps ...func() error) error {
	dir := 1
	for i := 0; i < len(steps); {
		if i < 0 {
			return errBack
		}
		err := steps[i]()
		switch {
		case errors.Is(err, ErrSkip):
			// Keep moving in the same direction: going backwards, a skipped step does not send
			// the user back to the step they just left.
		case errors.Is(err, errBack):
			dir = -1
		case err != nil:
			return err
		default:
			dir = 1
		}
		i += dir
		if i < 0 {
			return errBack
		}
	}
	return nil
}

// ErrSkip is returned by a Wizard step that does not apply to this run.
var ErrSkip = errors.New("step skipped")

// Choose shows a filterable list and returns the chosen value.
func Choose(title string, options ...Option) (string, error) { return PickOne(title, options) }

// Confirm asks a yes-or-no question.
func Confirm(title string) (bool, error) {
	var ok bool
	err := Ask(false, huh.NewConfirm().Title(title).Description(i18n.T("Esc cancels")).
		Affirmative(i18n.T("Yes")).Negative(i18n.T("No")).Value(&ok))
	return ok, err
}
