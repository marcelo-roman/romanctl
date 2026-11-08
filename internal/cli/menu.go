package cli

import (
	"bufio"
	"fmt"
	"github.com/charmbracelet/huh"
	"os"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/services"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"sort"
	"strings"
)

func pause() {
	fmt.Print(ui.SGray.Render(i18n.T("\nEnter para voltar ao menu...")))
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func menu() error {
	choice := "start"
	for {
		fmt.Print("\033[H\033[2J")
		fmt.Println(ui.SBoldTxt.Render("romanctl "+version) + "\n")
		services.PrintStatus()
		fmt.Println()

		picked, err := ui.Pick(ui.PickConfig{Title: i18n.T("O que você quer fazer?"), Initial: choice, NoLeft: true, Options: mainMenu()})
		if len(picked) > 0 {
			choice = picked[0]
		}
		if ui.IsAbort(err) || choice == "quit" {
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
			err = services.RestoreRun()
		case "new":
			err = menuNewProject("", "")
		case "run":
			err = menuRunProject()
		case "code":
			var name string
			if err = pickProject(i18n.T("Abrir qual projeto no code-server?"), &name); err == nil {
				err = openInCode(name)
			}
		case "open":
			err = menuOpen()
		case "up":
			err = upCommand(nil, map[string]bool{"d": true})
		case "down":
			err = downCommand(nil, map[string]bool{})
		case "add":
			err = menuAdd()
		case "publish":
			err = menuPublish()
		case "list":
			projects.PrintProjects()
		case "rm":
			err = menuRemoveProject()
		case "logs":
			err = menuLogs()
		case "offline":
			err = menuOffline()
		case "tools-list":
			err = toolsList(false, false)
		case "config":
			err = configCommand(nil)
		case "tools":
			err = toolsInstall(nil, installOptions{})
		case "tools-rm":
			err = toolsUninstall(nil, installOptions{})
		case "system":
			err = systemStatus()
		case "doctor":
			doctor()
		}
		if ui.IsAbort(err) {
			continue // Esc/Ctrl+C num submenu: volta direto ao menu
		}
		if err != nil {
			ui.Failf("%v", err)
		}
		pause()
	}
}

func menuServices(action string) error {
	var opts []ui.Option
	for _, s := range services.All {
		installed := s.Installed()
		if (action == "install") == installed {
			continue
		}
		if action == "start" && s.Running() || action == "stop" && !s.Running() {
			continue
		}
		opts = append(opts, ui.Opt(s.Name, s.Key))
	}
	if len(opts) == 0 {
		ui.Infof("Nada a fazer: nenhum serviço disponível para essa ação.")
		return nil
	}
	titles := map[string]string{"start": i18n.T("Iniciar quais?"), "stop": i18n.T("Parar quais?"),
		"restart": i18n.T("Reiniciar quais?"), "install": i18n.T("Instalar quais?")}
	sel, err := ui.Pick(ui.PickConfig{Title: titles[action], Options: opts, Multi: true, All: len(opts) > 1})
	if err != nil {
		return err
	}
	for _, k := range sel {
		s := services.FindService(k)
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
			ui.Failf("%v", err)
		}
	}
	return nil
}

const templatePrefix = "template:"

// newRequest é o que já foi informado por argumentos: só o resto é perguntado.
type newRequest struct {
	typ, name, stack, arch string
	with                   []string
	noGit, dryRun          bool
}

func menuNewProject(typ, name string) error {
	return runNewWizard(newRequest{typ: typ, name: name})
}

func runNewWizard(req newRequest) error {
	typ, name, chosen := req.typ, req.name, req.with
	git := !req.noGit
	templateOnly := req.stack != "" || req.arch != ""
	var steps []func() error
	if projects.FindType(typ) == nil {
		templates, err := scaffold.MatchTemplates(req.stack, req.arch)
		if err != nil {
			return err
		}
		if templateOnly && len(templates) == 0 {
			return ui.Usagef("ainda não existe template para --stack %s --arch %s. Combinações disponíveis:\n%s", req.stack, req.arch, scaffold.AvailableCombinations())
		}
		var opts []ui.Option
		for _, t := range templates {
			opts = append(opts, ui.Opt(i18n.T(t.Description), templatePrefix+t.Key()))
		}
		if !templateOnly {
			for _, t := range projects.ProjectTypes {
				opts = append(opts, ui.Opt(t.Label, t.Key))
			}
		}
		if len(opts) == 1 {
			typ = opts[0].Value // só há uma opção: não há o que perguntar
		} else {
			steps = append(steps, func() error {
				v, err := ui.PickOne(i18n.T("Tipo de projeto"), opts)
				if err == nil {
					typ = v
				}
				return err
			})
		}
	}
	if name == "" || projects.ValidateName(name) != nil {
		steps = append(steps, func() error {
			return ui.Ask(false, huh.NewInput().Title(i18n.T("Nome do projeto")).Placeholder("meu-app").
				Description(i18n.T("Esc volta · Ctrl+C cancela")).Validate(projects.ValidateName).Value(&name))
		})
	}
	if len(req.with) == 0 {
		steps = append(steps, func() error {
			t := templateByKey(strings.TrimPrefix(typ, templatePrefix))
			if !strings.HasPrefix(typ, templatePrefix) || t == nil || len(t.Services) == 0 {
				return ui.ErrSkip
			}
			var opts []ui.Option
			for _, key := range t.Services {
				opts = append(opts, ui.Opt(key, key))
			}
			v, err := ui.Pick(ui.PickConfig{Title: i18n.T("Quais serviços o projeto usa?"), Options: opts, Multi: true, AllowNone: true})
			if err == nil {
				chosen = v
			}
			return err
		})
	}
	if !req.noGit {
		steps = append(steps, func() error {
			return ui.Ask(false, huh.NewConfirm().Title(i18n.T("Iniciar repositório git?")).
				Description(i18n.T("←/→ alterna · Esc volta")).Affirmative(i18n.T("Sim")).Negative(i18n.T("Não")).Value(&git))
		})
	}
	if err := ui.Wizard(steps...); err != nil {
		return err
	}
	if key, ok := strings.CutPrefix(typ, templatePrefix); ok {
		t := templateByKey(key)
		if t == nil {
			return i18n.Errorf("template %s não encontrado", key)
		}
		sort.Strings(chosen)
		return scaffold.ScaffoldProject(t, name, chosen, git, req.dryRun)
	}
	return projects.NewProject(typ, name, git, false)
}

func templateByKey(key string) *scaffold.Template {
	for _, t := range scaffold.LoadTemplates() {
		if t.Key() == key {
			return t
		}
	}
	return nil
}

func pickProject(title string, name *string) error {
	ps := projects.ListProjects()
	if len(ps) == 0 {
		return i18n.Errorf("nenhum projeto em %s", config.ProjDir)
	}
	var opts []ui.Option
	for _, p := range ps {
		opts = append(opts, ui.Opt(i18n.Sprintf("%-28s %s", p.Name, ui.SGray.Render(p.Type)), p.Name))
	}
	v, err := ui.PickOne(title, opts)
	if err == nil {
		*name = v
	}
	return err
}

func menuRunProject() error {
	var name string
	open := true
	err := ui.Wizard(
		func() error { return pickProject(i18n.T("Rodar qual projeto?"), &name) },
		func() error {
			return ui.Ask(false, huh.NewConfirm().Title(i18n.T("Abrir no navegador quando subir?")).
				Description(i18n.T("←/→ alterna · Esc volta")).Affirmative(i18n.T("Sim")).Negative(i18n.T("Não")).Value(&open))
		},
	)
	if err != nil {
		return err
	}
	return projects.RunProject(name, open)
}

func menuRemoveProject() error {
	var name string
	if err := pickProject(i18n.T("Remover qual projeto?"), &name); err != nil {
		return err
	}
	return confirmRemoveProject(name)
}

func confirmRemoveProject(name string) error {
	var typed string
	if err := ui.Ask(false, huh.NewInput().
		Title(i18n.Sprintf("Digite %q para confirmar a remoção", name)).
		Description(i18n.T("Todos os arquivos serão apagados, sem volta. Esc volta.")).
		Value(&typed)); err != nil {
		return err
	}
	if typed != name {
		ui.Infof("Nome não confere — nada foi removido.")
		return nil
	}
	return projects.RemoveProject(name)
}

func menuLogs() error {
	var opts []ui.Option
	for _, s := range services.All {
		if sys.Exists(s.LogFile) {
			opts = append(opts, ui.Opt(s.Name, s.Key))
		}
	}
	if len(opts) == 0 {
		ui.Infof("Nenhum log ainda.")
		return nil
	}
	k, err := ui.PickOne(i18n.T("Log de qual serviço?"), opts)
	if err != nil {
		return err
	}
	return services.TailLog(services.FindService(k))
}

// mainMenu segue o "git help": cada linha mostra o subcomando real, com o que ele faz, em
// seções. Assim o menu também ensina a usar a linha de comando.
func mainMenu() []ui.Option {
	entries := []struct{ section, command, value, desc string }{
		{i18n.T("Projetos"), "new", "new", i18n.T("Criar um projeto")},
		{i18n.T("Projetos"), "add", "add", i18n.T("Adicionar código a um projeto")},
		{i18n.T("Projetos"), "up", "up", i18n.T("Subir serviços, banco e app")},
		{i18n.T("Projetos"), "down", "down", i18n.T("Derrubar o que o up subiu")},
		{i18n.T("Projetos"), "run", "run", i18n.T("Rodar um projeto")},
		{i18n.T("Projetos"), "code", "code", i18n.T("Abrir um projeto no code-server")},
		{i18n.T("Projetos"), "open", "open", i18n.T("Abrir no navegador")},
		{i18n.T("Projetos"), "publish", "publish", i18n.T("Publicar no GitHub")},
		{i18n.T("Projetos"), "list", "list", i18n.T("Listar os projetos")},
		{i18n.T("Projetos"), "rm", "rm", i18n.T("Remover um projeto")},
		{i18n.T("Serviços"), "start", "start", i18n.T("Iniciar serviços")},
		{i18n.T("Serviços"), "stop", "stop", i18n.T("Parar serviços")},
		{i18n.T("Serviços"), "restart", "restart", i18n.T("Reiniciar serviços")},
		{i18n.T("Serviços"), "restore", "restore", i18n.T("Religar o que estava ligado")},
		{i18n.T("Serviços"), "logs", "logs", i18n.T("Ver os logs de um serviço")},
		{i18n.T("Serviços"), "install", "install", i18n.T("Preparar um serviço")},
		{i18n.T("Ferramentas"), "tools list", "tools-list", i18n.T("Listar as ferramentas")},
		{i18n.T("Ferramentas"), "tools install", "tools", i18n.T("Instalar ferramentas")},
		{i18n.T("Ferramentas"), "tools uninstall", "tools-rm", i18n.T("Remover ferramentas")},
		{i18n.T("Sistema"), "system", "system", i18n.T("CPU, memória, disco e bateria")},
		{i18n.T("Sistema"), "doctor", "doctor", i18n.T("Versões das ferramentas principais")},
		{i18n.T("Sistema"), "offline", "offline", i18n.T("Preparar para ficar sem internet")},
		{i18n.T("Sistema"), "config", "config", i18n.T("Idioma")},
		{"", "quit", "quit", i18n.T("Sair")},
	}
	options := make([]ui.Option, len(entries))
	for i, e := range entries {
		options[i] = ui.Option{Label: fmt.Sprintf("%-16s %s", e.command, ui.SGray.Render(e.desc)), Value: e.value, Section: e.section}
	}
	return options
}
