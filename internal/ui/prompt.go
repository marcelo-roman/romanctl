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

var errBack = errors.New("voltar")

var errCancel = errors.New("cancelado")

// ErrCanceled é a resposta "não" de quem confirma: cancela sem ser falha.
var ErrCanceled = errCancel

func IsAbort(err error) bool { return errors.Is(err, errBack) || errors.Is(err, errCancel) }

var NavHint = i18n.T("Esc/← volta · Ctrl+C cancela")

// ask roda um prompt de um campo. list=true habilita "←" para voltar
// (só em listas; em campos de texto a seta precisa mover o cursor).
// Interactive diz se há alguém para responder a um menu.
func Interactive() bool {
	return !JSONOutput && term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func Ask(list bool, field huh.Field) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return Usagef("este comando precisa de um terminal; informe tudo por argumentos (veja: romanctl help)")
	}
	keys := []string{"ctrl+c", "esc"}
	if list {
		keys = append(keys, "left")
	}
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys(keys...), key.WithHelp("esc", "voltar"))

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

// wizard executa passos em sequência; Esc/← volta ao passo anterior.
func Wizard(steps ...func() error) error {
	dir := 1
	for i := 0; i < len(steps); {
		if i < 0 {
			return errBack
		}
		err := steps[i]()
		switch {
		case errors.Is(err, ErrSkip):
			// Segue na direção em que vinha: indo para trás, um passo pulado não devolve a
			// pessoa para o passo de onde ela acabou de sair.
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

// ErrSkip é o retorno de um passo do Wizard que não se aplica nesta execução.
var ErrSkip = errors.New("passo pulado")

// Choose mostra uma lista com filtro e devolve o valor escolhido.
func Choose(title string, options ...Option) (string, error) { return PickOne(title, options) }

// Confirm faz uma pergunta de sim ou não.
func Confirm(title string) (bool, error) {
	var ok bool
	err := Ask(false, huh.NewConfirm().Title(title).Description(i18n.T("Esc cancela")).
		Affirmative(i18n.T("Sim")).Negative(i18n.T("Não")).Value(&ok))
	return ok, err
}
