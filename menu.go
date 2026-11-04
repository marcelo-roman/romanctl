package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// Navegação nos prompts:
//
//	Esc        volta um passo (no primeiro passo, cancela)
//	←          volta um passo em listas (em texto/Sim-Não a seta move o cursor)
//	Ctrl+C     cancela tudo e volta ao menu (ou sai)
var (
	errBack   = errors.New("voltar")
	errCancel = errors.New("cancelado")
)

func isAbort(err error) bool { return errors.Is(err, errBack) || errors.Is(err, errCancel) }

const navHint = "Esc/← volta · Ctrl+C cancela"

// ask roda um prompt de um campo. list=true habilita "←" para voltar
// (só em listas; em campos de texto a seta precisa mover o cursor).
func ask(list bool, field huh.Field) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return usagef("este comando precisa de um terminal; informe tudo por argumentos (veja: romanctl help)")
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
func wizard(steps ...func() error) error {
	for i := 0; i < len(steps); {
		err := steps[i]()
		switch {
		case errors.Is(err, errBack):
			if i == 0 {
				return errBack
			}
			i--
		case err != nil:
			return err
		default:
			i++
		}
	}
	return nil
}

func confirm(title string) (bool, error) {
	var ok bool
	err := ask(false, huh.NewConfirm().Title(title).Description("Esc cancela").
		Affirmative("Sim").Negative("Não").Value(&ok))
	return ok, err
}

func pause() {
	fmt.Print(sGray.Render("\nEnter para voltar ao menu..."))
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func menu() error {
	choice := "start"
	for {
		fmt.Print("\033[H\033[2J")
		fmt.Println(sTitle.Render("romanctl") + sGray.Render(" v"+version+" — utilitário de desenvolvimento\n"))
		printStatus()
		fmt.Println()

		err := ask(false, huh.NewSelect[string]().
			Title("O que você quer fazer?").Description("Enter escolhe · Esc sai").
			Options(
				huh.NewOption("▶  Iniciar serviços", "start"),
				huh.NewOption("■  Parar serviços", "stop"),
				huh.NewOption("↻  Reiniciar serviços", "restart"),
				huh.NewOption("⟲  Restaurar o que estava ligado", "restore"),
				huh.NewOption("✚  Criar projeto", "new"),
				huh.NewOption("⚡ Rodar projeto", "run"),
				huh.NewOption("📝 Abrir projeto no code-server", "code"),
				huh.NewOption("🌐 Abrir no navegador", "open"),
				huh.NewOption("☰  Listar projetos", "list"),
				huh.NewOption("✖  Remover projeto", "rm"),
				huh.NewOption("📜 Ver logs de um serviço", "logs"),
				huh.NewOption("⬇  Instalar serviço", "install"),
				huh.NewOption("✈  Modo offline", "offline"),
				huh.NewOption("🩺 Ferramentas instaladas", "doctor"),
				huh.NewOption("⏻  Sair", "quit"),
			).Value(&choice))
		if isAbort(err) || choice == "quit" {
			fmt.Print("\033[H\033[2J")
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Println()
		switch choice {
		case "start", "stop", "restart", "install":
			err = menuServices(choice)
		case "restore":
			err = restoreRun()
		case "new":
			err = menuNewProject("", "")
		case "run":
			err = menuRunProject()
		case "code":
			var name string
			if err = pickProject("Abrir qual projeto no code-server?", &name); err == nil {
				err = openInCode(name)
			}
		case "open":
			err = menuOpen()
		case "list":
			printProjects()
		case "rm":
			err = menuRemoveProject()
		case "logs":
			err = menuLogs()
		case "offline":
			err = menuOffline()
		case "doctor":
			doctor()
		}
		if isAbort(err) {
			continue // Esc/Ctrl+C num submenu: volta direto ao menu
		}
		if err != nil {
			failf("%v", err)
		}
		pause()
	}
}

func menuServices(action string) error {
	var opts []huh.Option[string]
	for _, s := range services {
		installed := s.Installed()
		if (action == "install") == installed {
			continue
		}
		if action == "start" && s.Running() || action == "stop" && !s.Running() {
			continue
		}
		opts = append(opts, huh.NewOption(s.Name, s.Key))
	}
	if len(opts) == 0 {
		infof("Nada a fazer: nenhum serviço disponível para essa ação.")
		return nil
	}
	titles := map[string]string{"start": "Iniciar quais?", "stop": "Parar quais?",
		"restart": "Reiniciar quais?", "install": "Instalar quais?"}
	var sel []string
	if err := ask(true, huh.NewMultiSelect[string]().
		Title(titles[action]).Description("Espaço marca · Enter confirma · "+navHint).
		Options(opts...).Value(&sel)); err != nil {
		return err
	}
	for _, k := range sel {
		s := findService(k)
		var err error
		switch action {
		case "start":
			err = s.Start()
		case "stop":
			err = s.Stop()
		case "restart":
			if err = s.Stop(); err == nil {
				err = s.Start()
			}
		case "install":
			err = s.Install()
		}
		if err != nil {
			failf("%v", err)
		}
	}
	return nil
}

const templatePrefix = "template:"

func menuNewProject(typ, name string) error {
	git := true
	var steps []func() error
	var chosen []string
	if findType(typ) == nil {
		var opts []huh.Option[string]
		for _, t := range loadTemplates() {
			opts = append(opts, huh.NewOption(t.Description, templatePrefix+t.Key()))
		}
		for _, t := range projectTypes {
			opts = append(opts, huh.NewOption(t.Label, t.Key))
		}
		steps = append(steps, func() error {
			return ask(true, huh.NewSelect[string]().Title("Tipo de projeto").Description(navHint).
				Options(opts...).Value(&typ))
		})
	}
	if name == "" || validateName(name) != nil {
		steps = append(steps, func() error {
			return ask(false, huh.NewInput().Title("Nome do projeto").Placeholder("meu-app").
				Description("Esc volta · Ctrl+C cancela").Validate(validateName).Value(&name))
		})
	}
	steps = append(steps, func() error {
		t := templateByKey(strings.TrimPrefix(typ, templatePrefix))
		if !strings.HasPrefix(typ, templatePrefix) || t == nil || len(t.Services) == 0 {
			return nil
		}
		var opts []huh.Option[string]
		for _, key := range t.Services {
			opts = append(opts, huh.NewOption(key, key))
		}
		return ask(true, huh.NewMultiSelect[string]().Title("Quais serviços o projeto usa?").
			Description("Espaço marca · Enter confirma · "+navHint).Options(opts...).Value(&chosen))
	}, func() error {
		return ask(false, huh.NewConfirm().Title("Iniciar repositório git?").
			Description("←/→ alterna · Esc volta").Affirmative("Sim").Negative("Não").Value(&git))
	})
	if err := wizard(steps...); err != nil {
		return err
	}
	if key, ok := strings.CutPrefix(typ, templatePrefix); ok {
		t := templateByKey(key)
		if t == nil {
			return fmt.Errorf("template %s não encontrado", key)
		}
		sort.Strings(chosen)
		return scaffoldProject(t, name, chosen, git, false)
	}
	return newProject(typ, name, git, false)
}

func templateByKey(key string) *Template {
	for _, t := range loadTemplates() {
		if t.Key() == key {
			return t
		}
	}
	return nil
}

func pickProject(title string, name *string) error {
	ps := listProjects()
	if len(ps) == 0 {
		return fmt.Errorf("nenhum projeto em %s", projDir)
	}
	var opts []huh.Option[string]
	for _, p := range ps {
		opts = append(opts, huh.NewOption(fmt.Sprintf("%-28s %s", p.Name, sGray.Render(p.Type)), p.Name))
	}
	return ask(true, huh.NewSelect[string]().Title(title).Description(navHint).Options(opts...).Value(name))
}

func menuRunProject() error {
	var name string
	open := true
	err := wizard(
		func() error { return pickProject("Rodar qual projeto?", &name) },
		func() error {
			return ask(false, huh.NewConfirm().Title("Abrir no navegador quando subir?").
				Description("←/→ alterna · Esc volta").Affirmative("Sim").Negative("Não").Value(&open))
		},
	)
	if err != nil {
		return err
	}
	return runProject(name, open)
}

func menuRemoveProject() error {
	var name, typed string
	err := wizard(
		func() error { return pickProject("Remover qual projeto?", &name) },
		func() error {
			typed = ""
			return ask(false, huh.NewInput().
				Title(fmt.Sprintf("Digite %q para confirmar a remoção", name)).
				Description("Todos os arquivos serão apagados, sem volta. Esc volta.").
				Value(&typed))
		},
	)
	if err != nil {
		return err
	}
	if typed != name {
		infof("Nome não confere — nada foi removido.")
		return nil
	}
	return removeProject(name)
}

func menuLogs() error {
	var opts []huh.Option[string]
	for _, s := range services {
		if exists(s.LogFile) {
			opts = append(opts, huh.NewOption(s.Name, s.Key))
		}
	}
	if len(opts) == 0 {
		infof("Nenhum log ainda.")
		return nil
	}
	var k string
	if err := ask(true, huh.NewSelect[string]().Title("Log de qual serviço?").Description(navHint).
		Options(opts...).Value(&k)); err != nil {
		return err
	}
	return tailLog(findService(k))
}
