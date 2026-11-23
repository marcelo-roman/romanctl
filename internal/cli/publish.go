package cli

import (
	"github.com/charmbracelet/huh"

	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/publish"
	"romanctl/internal/ui"
)

func publishCommand(args []string, flags map[string]bool) error {
	path, err := projects.Locate(arg(args, 0))
	if err != nil && arg(args, 0) == "" && ui.Interactive() {
		// Outside a project: ask which one instead of failing.
		var name string
		if err = pickProject(i18n.T("Which project to publish?"), &name); err == nil {
			path, err = projects.ProjectPath(name)
		}
	}
	if err != nil {
		return err
	}
	opts := publish.Options{Public: flags["public"], NoCI: flags["no-ci"], DryRun: flags["dry-run"], Description: flagVals["description"]}
	var plan publish.Result
	ui.Spin(i18n.T("Checking Git and GitHub..."), func(func(string)) { plan, err = publish.Plan(path, opts) })
	if err != nil {
		return err
	}
	if ui.Interactive() && !plan.Exists && !opts.DryRun && !flags["y"] && !flags["yes"] {
		if err := askPublishOptions(&opts, flags["public"]); err != nil {
			return err
		}
		plan.Public = opts.Public
	}
	if opts.DryRun {
		if ui.JSONOutput {
			return ui.PrintJSON(plan)
		}
		ui.Infof("Dry run: nothing was done. Publish would:")
		printPublishPlan(plan)
		return nil
	}
	if !flags["y"] && !flags["yes"] {
		if !ui.Interactive() {
			return ui.Usagef("publishing needs confirmation: run it in a terminal or use --yes")
		}
		printPublishPlan(plan)
		title := i18n.Sprintf("Publish %s as a private repository?", plan.Name)
		if plan.Exists {
			title = i18n.Sprintf("Push %s to the existing repository?", plan.Name)
		} else if opts.Public {
			title = i18n.Sprintf("Publish %s as a PUBLIC repository? Anyone will be able to see the code.", plan.Name)
		}
		ok, err := ui.Confirm(title)
		if err != nil {
			return err
		}
		if !ok {
			ui.Infof("Cancelled")
			return nil
		}
	}
	var result publish.Result
	ui.Spin(i18n.Sprintf("Publishing %s...", plan.Name), func(func(string)) { result, err = publish.Run(path, opts) })
	if err != nil {
		return err
	}
	if ui.JSONOutput {
		return ui.PrintJSON(result)
	}
	ui.Okf("%s published: %s", result.Name, result.URL)
	return nil
}

func printPublishPlan(p publish.Result) {
	visibility := i18n.T("private")
	if p.Public {
		visibility = i18n.T("PUBLIC")
	}
	if p.Exists {
		i18n.Printf("  Repository   already exists, only pushes the changes (%s)\n", p.Remote)
	} else {
		i18n.Printf("  Repository   %s, %s\n", p.Name, visibility)
	}
	for _, f := range p.Generated {
		i18n.Printf("  Generates    %s\n", f)
	}
}

// askPublishOptions asks for the visibility and the description, unless they came as arguments.
func askPublishOptions(opts *publish.Options, publicGiven bool) error {
	var steps []func() error
	if !publicGiven {
		steps = append(steps, func() error {
			choice, err := ui.Choose(i18n.T("Repository visibility"),
				ui.Opt(i18n.T("Private"), "private"),
				ui.Opt(i18n.T("Public: anyone can see the code"), "public"))
			opts.Public = choice == "public"
			return err
		})
	}
	if opts.Description == "" {
		steps = append(steps, func() error {
			return ui.Ask(false, huh.NewInput().Title(i18n.T("Description (optional)")).Description(i18n.T("Esc goes back · Ctrl+C cancels")).Value(&opts.Description))
		})
	}
	return ui.Wizard(steps...)
}
