package cli

import (
	"fmt"
	"path/filepath"

	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/state"
	"romanctl/internal/ui"
	"romanctl/internal/up"
)

func upCommand(args []string, flags map[string]bool) error {
	opts := up.Options{Detach: flags["d"] || flags["detach"], NoRun: flags["no-run"], Open: flags["open"]}
	if ui.JSONOutput && !opts.Detach && !opts.NoRun {
		return ui.Usagef("with --json, use -d or --no-run: in the foreground the command only ends when the app stops")
	}
	path, err := projects.Locate(arg(args, 0))
	if err != nil && arg(args, 0) == "" && ui.Interactive() {
		var name string
		if err = pickProject(i18n.T("Which project to bring up?"), &name); err == nil {
			path, err = projects.ProjectPath(name)
		}
	}
	if err != nil {
		return err
	}
	result, err := up.Run(filepath.Base(path), path, opts)
	if err != nil {
		return err
	}
	if ui.JSONOutput {
		return ui.PrintJSON(result)
	}
	if result.Already {
		return nil
	}
	if opts.NoRun {
		ui.Okf("Environment for %s is ready. To run it: romanctl run %s", result.Project, result.Project)
		return nil
	}
	if opts.Detach {
		if result.URL != "" {
			ui.Okf("%s is up → %s", result.Project, result.URL)
		} else {
			ui.Okf("%s started; no address to check whether it is responding", result.Project)
		}
		i18n.Printf("   Log:    romanctl logs %s\n   Stop:   romanctl down %s\n", result.Project, result.Project)
	}
	return nil
}

func downCommand(args []string, flags map[string]bool) error {
	keep := flags["keep-services"]
	var names []string
	switch {
	case flags["all"]:
		names = up.Names()
		if len(names) == 0 {
			ui.Infof("No projects are up.")
		}
	case arg(args, 0) != "":
		names = []string{arg(args, 0)}
	default:
		name, err := pickDown()
		if err != nil {
			return err
		}
		if name == "" {
			return nil
		}
		names = []string{name}
	}
	results := []up.DownResult{}
	var firstErr error
	for _, name := range names {
		r, err := up.Down(name, keep)
		results = append(results, r)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if ui.JSONOutput && firstErr == nil {
		return ui.PrintJSON(results)
	}
	return firstErr
}

// pickDown uses the project of the current folder if it is up; otherwise, shows the ones that are.
func pickDown() (string, error) {
	ups := state.ReadState().Ups
	if path, err := projects.Locate(""); err == nil {
		if _, ok := ups[filepath.Base(path)]; ok {
			return filepath.Base(path), nil
		}
	}
	if !ui.Interactive() {
		return "", ui.Usagef("specify the project: romanctl down <project>, or use --all")
	}
	var opts []ui.Option
	for _, name := range up.Names() {
		opts = append(opts, ui.Opt(fmt.Sprintf("%-28s %s", name, ui.SGray.Render(up.Describe(ups[name]))), name))
	}
	if len(opts) == 0 {
		ui.Infof("No projects are up.")
		return "", nil
	}
	return ui.PickOne(i18n.T("Which project to bring down?"), opts)
}
