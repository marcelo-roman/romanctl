package cli

import (
	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/services"
	"romanctl/internal/ui"
)

// serviceCommand opens the service list when none was given in a terminal.
func serviceCommand(action, name string, f func(*services.Service) error) error {
	if name == "" && ui.Interactive() {
		return menuServices(action)
	}
	return eachService(name, f)
}

func menuRestore() error {
	sub, err := ui.Choose(i18n.T("Restore"),
		ui.Opt(i18n.T("Starts again what was running"), "run"),
		ui.Opt(i18n.T("View the saved state"), "status"),
		ui.Opt(i18n.T("Save what is running now"), "save"),
		ui.Opt(i18n.T("Clear the saved state"), "clear"))
	if err != nil {
		return err
	}
	if sub == "run" {
		sub = ""
	}
	return services.RestoreCommand(sub)
}

// newCommand creates right away when everything was given; in a terminal, it asks for what is missing.
func newCommand(args []string, flags map[string]bool) error {
	a, b := arg(args, 0), arg(args, 1)
	stack, archFlag := flagVals["stack"], flagVals["arch"]
	git := !flags["no-git"]
	opts := scaffold.Options{Git: git, DryRun: flags["dry-run"], Offline: flags["offline"]}

	if stack != "" || archFlag != "" {
		if a != "" && stack != "" && archFlag != "" {
			return scaffold.NewFromTemplate(a, stack, archFlag, flagVals["with"], opts)
		}
		if !ui.Interactive() {
			if a == "" {
				return ui.Usagef("specify the name: romanctl new <name> --stack <stack> --arch <architectures>")
			}
			return scaffold.NewFromTemplate(a, stack, archFlag, flagVals["with"], opts)
		}
		return runNewWizard(newRequest{name: a, stack: stack, arch: archFlag, with: scaffold.SplitServices(flagVals["with"]),
			noGit: !git, dryRun: flags["dry-run"], offline: flags["offline"]})
	}
	if a != "" && b != "" {
		return projects.NewProject(a, b, git, flags["offline"])
	}
	req := newRequest{typ: a, noGit: !git, offline: flags["offline"]}
	// "new shop" gives the name; "new express" gives the type.
	if a != "" && projects.FindType(a) == nil {
		req.typ, req.name = "", a
	}
	return runNewWizard(req)
}
