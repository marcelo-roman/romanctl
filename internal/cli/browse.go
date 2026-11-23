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

// toolsBrowse is the browsable list of tools list: it filters as you type, and Enter shows the actions.
func toolsBrowse(showAll bool, categories []string) error {
	cursor := ""
	for {
		var opts []ui.Option
		statuses := loadStatuses()
		fillLatest(statuses)
		for _, s := range statuses {
			if s.Unavailable != "" && !s.Installed && !showAll {
				continue
			}
			if len(categories) > 0 && !slices.Contains(categories, s.Group) {
				continue
			}
			state := toolState(s)
			opts = append(opts, toolOption(fmt.Sprintf("%-15s %s %s", s.Key, ui.SGray.Render(fmt.Sprintf("%-13s", s.Group)), state), s))
		}
		got, err := ui.Pick(ui.PickConfig{Title: i18n.T("Tools"), Subtitle: platform.Current.Describe(), Options: opts, Initial: cursor, NoLeft: true, Categories: toolCategories(opts)})
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
			ui.Warnf("%s does not run on this machine: %s", t.Name, why)
			pause()
			return nil
		}
	}
	var opts []ui.Option
	if _, installed := t.Detected(); installed {
		opts = []ui.Option{ui.Opt(i18n.Sprintf("Remove %s", t.Name), "uninstall"), ui.Opt(i18n.Sprintf("Reinstall %s", t.Name), "reinstall")}
	} else {
		opts = []ui.Option{ui.Opt(i18n.Sprintf("Install %s, latest version", t.Name), "install")}
	}
	opts = append(opts, ui.Opt(i18n.T("Choose another version"), "version"))
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
	case "version":
		var version string
		if version, err = pickVersion(t); err == nil {
			err = toolsInstall([]string{key}, installOptions{version: version, yes: true})
		}
	}
	if err != nil && !ui.IsAbort(err) {
		return err
	}
	pause()
	return nil
}

// projectList is browsable in a terminal; with --plain, --filter, --json or a pipe, it prints.
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
		got, err := ui.Pick(ui.PickConfig{Title: i18n.T("Projects"), Options: opts, Initial: cursor, NoLeft: true})
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
	opts := []ui.Option{ui.Opt(i18n.T("Bring everything up in the background (up -d)"), "up"), ui.Opt(i18n.T("Bring down (down)"), "down"),
		ui.Opt(i18n.T("Run"), "run"), ui.Opt(i18n.T("Open in code-server"), "code")}
	target := scaffold.TargetOf(path)
	if target != "" {
		opts = append(opts, ui.Opt(i18n.T("Add code"), "add"))
	}
	opts = append(opts, ui.Opt(i18n.T("Publish to GitHub"), "publish"), ui.Opt(i18n.T("Remove the project"), "rm"))
	action, err := ui.PickOne(name, opts)
	if err != nil {
		return err
	}
	switch action {
	case "up":
		err = upCommand([]string{name}, map[string]bool{"d": true})
	case "down":
		err = downCommand([]string{name}, map[string]bool{})
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

// fillLatest fills in the latest version of the tools that run here; it only hits the network for
// what was not queried in the last few hours.
func fillLatest(statuses []tools.Status) {
	var keys []string
	for _, s := range statuses {
		if s.Latest == "" && (s.Unavailable == "" || s.Installed) {
			keys = append(keys, s.Key)
		}
	}
	if len(keys) == 0 {
		return
	}
	var latest map[string]tools.Upstream
	ui.Spin(i18n.T("Checking the latest versions..."), func(update func(string)) {
		latest = tools.RefreshLatest(keys, func(done, total int) {
			update(i18n.Sprintf("Checking the latest versions %d/%d", done, total))
		})
	})
	for i := range statuses {
		if u, ok := latest[statuses[i].Key]; ok {
			statuses[i].Latest = u.Latest
		}
	}
}
