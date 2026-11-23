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
	fmt.Print(ui.SGray.Render(i18n.T("\nPress Enter to return to the menu...")))
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func menu() error {
	choice := "start"
	for {
		fmt.Print("\033[H\033[2J")
		fmt.Println(ui.SBoldTxt.Render("romanctl "+version) + "\n")
		services.PrintStatus(false)
		fmt.Println()

		picked, err := ui.Pick(ui.PickConfig{Title: i18n.T("What do you want to do?"), Initial: choice, NoLeft: true, Options: mainMenu()})
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
			if err = pickProject(i18n.T("Which project to open in code-server?"), &name); err == nil {
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
			continue // Esc/Ctrl+C in a submenu: go straight back to the menu
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
		if s.Unsupported != "" || s.Engine && action != "start" {
			continue
		}
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
		ui.Infof("Nothing to do: no service available for this action.")
		return nil
	}
	titles := map[string]string{"start": i18n.T("Start which ones?"), "stop": i18n.T("Stop which ones?"),
		"restart": i18n.T("Restart which ones?"), "install": i18n.T("Install which ones?")}
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

// newRequest is what was already given as arguments: only the rest is asked.
type newRequest struct {
	typ, name, stack, arch string
	with                   []string
	noGit, dryRun, offline bool
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
			return ui.Usagef("there is no template yet for --stack %s --arch %s. Available combinations:\n%s", req.stack, req.arch, scaffold.AvailableCombinations())
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
			typ = opts[0].Value // only one option: nothing to ask
		} else {
			steps = append(steps, func() error {
				v, err := ui.PickOne(i18n.T("Project type"), opts)
				if err == nil {
					typ = v
				}
				return err
			})
		}
	}
	if name == "" || projects.ValidateName(name) != nil {
		steps = append(steps, func() error {
			return ui.Ask(false, huh.NewInput().Title(i18n.T("Project name")).Placeholder("meu-app").
				Description(i18n.T("Esc goes back · Ctrl+C cancels")).Validate(projects.ValidateName).Value(&name))
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
			v, err := ui.Pick(ui.PickConfig{Title: i18n.T("Which services does the project use?"), Options: opts, Multi: true, AllowNone: true})
			if err == nil {
				chosen = v
			}
			return err
		})
	}
	if !req.noGit {
		steps = append(steps, func() error {
			return ui.Ask(false, huh.NewConfirm().Title(i18n.T("Initialize a git repository?")).
				Description(i18n.T("←/→ toggles · Esc goes back")).Affirmative(i18n.T("Yes")).Negative(i18n.T("No")).Value(&git))
		})
	}
	if err := ui.Wizard(steps...); err != nil {
		return err
	}
	if key, ok := strings.CutPrefix(typ, templatePrefix); ok {
		t := templateByKey(key)
		if t == nil {
			return i18n.Errorf("template %s not found", key)
		}
		sort.Strings(chosen)
		return scaffold.ScaffoldProject(t, name, chosen, scaffold.Options{Git: git, DryRun: req.dryRun, Offline: req.offline})
	}
	return projects.NewProject(typ, name, git, req.offline)
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
		return i18n.Errorf("no projects in %s", config.ProjDir)
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
		func() error { return pickProject(i18n.T("Which project to run?"), &name) },
		func() error {
			return ui.Ask(false, huh.NewConfirm().Title(i18n.T("Open in the browser when it is up?")).
				Description(i18n.T("←/→ toggles · Esc goes back")).Affirmative(i18n.T("Yes")).Negative(i18n.T("No")).Value(&open))
		},
	)
	if err != nil {
		return err
	}
	return projects.RunProject(name, open)
}

func menuRemoveProject() error {
	var name string
	if err := pickProject(i18n.T("Which project to remove?"), &name); err != nil {
		return err
	}
	return confirmRemoveProject(name)
}

func confirmRemoveProject(name string) error {
	var typed string
	if err := ui.Ask(false, huh.NewInput().
		Title(i18n.Sprintf("Type %q to confirm the removal", name)).
		Description(i18n.T("All files will be permanently deleted. Esc goes back.")).
		Value(&typed)); err != nil {
		return err
	}
	if typed != name {
		ui.Infof("Name does not match — nothing was removed.")
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
		ui.Infof("No logs yet.")
		return nil
	}
	k, err := ui.PickOne(i18n.T("Logs of which service?"), opts)
	if err != nil {
		return err
	}
	return services.TailLog(services.FindService(k))
}

// mainMenu follows "git help": each row shows the real subcommand and what it does, in
// sections. This way the menu also teaches how to use the command line.
func mainMenu() []ui.Option {
	entries := []struct{ section, command, value, desc string }{
		{i18n.T("Projects"), "new", "new", i18n.T("Create a project")},
		{i18n.T("Projects"), "add", "add", i18n.T("Add code to a project")},
		{i18n.T("Projects"), "up", "up", i18n.T("Bring up services, database and app")},
		{i18n.T("Projects"), "down", "down", i18n.T("Bring down what up brought up")},
		{i18n.T("Projects"), "run", "run", i18n.T("Run a project")},
		{i18n.T("Projects"), "code", "code", i18n.T("Open a project in code-server")},
		{i18n.T("Projects"), "open", "open", i18n.T("Open in browser")},
		{i18n.T("Projects"), "publish", "publish", i18n.T("Publish to GitHub")},
		{i18n.T("Projects"), "list", "list", i18n.T("List the projects")},
		{i18n.T("Projects"), "rm", "rm", i18n.T("Remove a project")},
		{i18n.T("Services"), "start", "start", i18n.T("Start services")},
		{i18n.T("Services"), "stop", "stop", i18n.T("Stop services")},
		{i18n.T("Services"), "restart", "restart", i18n.T("Restart services")},
		{i18n.T("Services"), "restore", "restore", i18n.T("Turn back on what was running")},
		{i18n.T("Services"), "logs", "logs", i18n.T("View a service's logs")},
		{i18n.T("Services"), "install", "install", i18n.T("Prepare a service")},
		{i18n.T("Tools"), "tools list", "tools-list", i18n.T("List the tools")},
		{i18n.T("Tools"), "tools install", "tools", i18n.T("Install tools")},
		{i18n.T("Tools"), "tools uninstall", "tools-rm", i18n.T("Remove tools")},
		{i18n.T("System"), "system", "system", i18n.T("CPU, memory, disk and battery")},
		{i18n.T("System"), "doctor", "doctor", i18n.T("Versions of the main tools")},
		{i18n.T("System"), "offline", "offline", i18n.T("Prepare to work without internet")},
		{i18n.T("System"), "config", "config", i18n.T("Language")},
		{"", "quit", "quit", i18n.T("Quit")},
	}
	options := make([]ui.Option, len(entries))
	for i, e := range entries {
		options[i] = ui.Option{Label: fmt.Sprintf("%-16s %s", e.command, ui.SGray.Render(e.desc)), Value: e.value, Section: e.section}
	}
	return options
}
