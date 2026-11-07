package cli

import (
	"errors"
	"fmt"
	"slices"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/tools"
	"romanctl/internal/ui"
)

// toolsBrowse é a lista navegável de tools list: filtra ao digitar, e Enter mostra as ações.
func toolsBrowse(showAll bool, categories []string) error {
	cursor := ""
	for {
		var opts []ui.Option
		for _, s := range loadStatuses() {
			if s.Unavailable != "" && !s.Installed && !showAll {
				continue
			}
			if len(categories) > 0 && !slices.Contains(categories, s.Group) {
				continue
			}
			state := ui.SGray.Render("○ " + s.Description)
			switch {
			case s.Installed:
				state = ui.SGreen.Render("● ") + s.Version
			case s.Unavailable != "":
				state = ui.SGray.Render("✕ " + s.Unavailable)
			}
			opts = append(opts, toolOption(fmt.Sprintf("%-15s %s %s", s.Key, ui.SGray.Render(fmt.Sprintf("%-13s", s.Group)), state), s))
		}
		got, err := ui.Pick(ui.PickConfig{Title: i18n.T("Ferramentas"), Subtitle: platform.Current.Describe(), Options: opts, Initial: cursor, NoLeft: true, Categories: toolCategories(opts)})
		if err != nil {
			if ui.IsAbort(err) {
				return nil
			}
			return err
		}
		cursor = got[0]
		if err := toolActions(cursor); err != nil && !ui.IsAbort(err) {
			ui.Failf("%v", err)
			pause()
		}
	}
}

func toolActions(key string) error {
	t := tools.Find(key)
	if t == nil {
		return nil
	}
	if why := t.Unavailable(); why != "" {
		if _, installed := t.Detected(); !installed {
			ui.Warnf("%s não roda nesta máquina: %s", t.Name, why)
			pause()
			return nil
		}
	}
	var opts []ui.Option
	if _, installed := t.Detected(); installed {
		opts = []ui.Option{ui.Opt(i18n.Sprintf("Remover %s", t.Name), "uninstall"), ui.Opt(i18n.Sprintf("Reinstalar %s", t.Name), "reinstall")}
	} else {
		opts = []ui.Option{ui.Opt(i18n.Sprintf("Instalar %s", t.Name), "install")}
	}
	action, err := ui.PickOne(t.Name+" — "+i18n.T(t.Description), opts)
	if err != nil {
		return err
	}
	switch action {
	case "install":
		err = toolsInstall([]string{key}, installOptions{})
	case "reinstall":
		err = toolsInstall([]string{key}, installOptions{force: true})
	case "uninstall":
		err = toolsUninstall([]string{key}, installOptions{})
	}
	if err != nil && !ui.IsAbort(err) {
		return err
	}
	pause()
	return nil
}

// projectList é navegável em um terminal; com --plain, --filter, --json ou um pipe, imprime.
func projectList(plain bool) error {
	filter := flagVals["filter"]
	if !plain && filter == "" && ui.Interactive() {
		return projectBrowse()
	}
	if ui.JSONOutput {
		list := projects.ListJSON()
		if filter != "" {
			kept := list[:0]
			for _, p := range list {
				if ui.Matches(fmt.Sprint(p["name"], " ", p["type"]), filter) {
					kept = append(kept, p)
				}
			}
			list = kept
		}
		return ui.PrintJSON(list)
	}
	projects.PrintProjectsFiltered(filter)
	return nil
}

func projectBrowse() error {
	cursor := ""
	for {
		list := projects.ListProjects()
		if len(list) == 0 {
			projects.PrintProjects()
			return nil
		}
		var opts []ui.Option
		for _, p := range list {
			opts = append(opts, ui.Opt(fmt.Sprintf("%-28s %-24s %s", p.Name, ui.SGray.Render(p.Type), ui.SGray.Render(p.Mod.Format("02/01/2006 15:04"))), p.Name))
		}
		got, err := ui.Pick(ui.PickConfig{Title: i18n.T("Projetos"), Options: opts, Initial: cursor, NoLeft: true})
		if err != nil {
			if ui.IsAbort(err) {
				return nil
			}
			return err
		}
		cursor = got[0]
		if err := projectActions(cursor); err != nil && !errors.Is(err, nil) {
			if ui.IsAbort(err) {
				continue
			}
			ui.Failf("%v", err)
			pause()
		}
	}
}

func projectActions(name string) error {
	path, err := projects.ProjectPath(name)
	if err != nil {
		return err
	}
	opts := []ui.Option{ui.Opt(i18n.T("Rodar"), "run"), ui.Opt(i18n.T("Abrir no code-server"), "code")}
	target := scaffold.TargetOf(path)
	if target != "" {
		opts = append(opts, ui.Opt(i18n.T("Adicionar código"), "add"))
	}
	opts = append(opts, ui.Opt(i18n.T("Publicar no GitHub"), "publish"), ui.Opt(i18n.T("Remover o projeto"), "rm"))
	action, err := ui.PickOne(name, opts)
	if err != nil {
		return err
	}
	switch action {
	case "run":
		err = projects.RunProject(name, true)
	case "code":
		err = openInCode(name)
	case "add":
		err = addInteractive(path, target, "", "", "", map[string]bool{})
	case "publish":
		err = publishCommand([]string{name}, map[string]bool{})
	case "rm":
		err = confirmRemoveProject(name)
	}
	if err != nil {
		return err
	}
	pause()
	return nil
}
