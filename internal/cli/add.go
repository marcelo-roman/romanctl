package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/ui"
)

func addCommand(args []string, flags map[string]bool) error {
	path, err := projects.Locate(flagVals["project"])
	if err != nil {
		return err
	}
	target := scaffold.TargetOf(path)
	if target == "" {
		return ui.Usagef("there are no generators for this type of project: %s", path)
	}
	if flags["list"] {
		return addList(target, path)
	}
	kind, name, in := arg(args, 0), arg(args, 1), flagVals["in"]
	if ui.Interactive() && addIsIncomplete(target, kind, name, in) {
		return addInteractive(path, target, kind, name, in, flags)
	}
	if kind == "" || name == "" {
		return ui.Usagef("specify the type and name: romanctl add <type> <name>; see with: romanctl add --list")
	}
	return runAdd(path, scaffold.AddOptions{Kind: kind, Name: name, In: in, Plural: flagVals["plural"], Force: flags["force"], DryRun: flags["dry-run"]})
}

func addIsIncomplete(target, kind, name, in string) bool {
	if kind == "" || name == "" {
		return true
	}
	g := scaffold.FindGenerator(scaffold.LoadGenerators(target), kind)
	return g != nil && g.NeedsIn && in == ""
}

func addList(target, path string) error {
	list := scaffold.LoadGenerators(target)
	if ui.JSONOutput {
		return ui.PrintJSON(map[string]any{"project": path, "target": target, "generators": list})
	}
	i18n.Printf("Generators for %s (%s):\n", path, target)
	for _, g := range list {
		suffix := ""
		if g.NeedsIn {
			suffix = " --in <agregado>"
		}
		fmt.Printf("  %-14s %s%s\n", g.Kind, i18n.T(g.Description), ui.SGray.Render(suffix))
	}
	return nil
}

func addInteractive(path, target, kind, name, in string, flags map[string]bool) error {
	list := scaffold.LoadGenerators(target)
	// A step already answered by an argument is left out: otherwise Esc would go back to a
	// step that skips itself, and the user would be stuck on the next one.
	var steps []func() error
	if kind == "" {
		steps = append(steps, func() error {
			var opts []ui.Option
			for _, g := range list {
				opts = append(opts, ui.Opt(fmt.Sprintf("%-14s %s", g.Kind, ui.SGray.Render(i18n.T(g.Description))), g.Kind))
			}
			v, err := ui.PickOne(i18n.Sprintf("What to add to %s?", target), opts)
			if err == nil {
				kind = v
			}
			return err
		})
	}
	if name == "" {
		steps = append(steps, func() error {
			return ui.Ask(false, huh.NewInput().Title(i18n.T("Name")).Placeholder("Product").Description(i18n.T("Esc goes back · Ctrl+C cancels")).
				Validate(func(v string) error {
					if scaffold.Pascal(v) == "" {
						return i18n.Errorf("enter a name")
					}
					return nil
				}).Value(&name))
		})
	}
	if in == "" {
		// Whether the aggregate is needed is only known after the type is chosen.
		steps = append(steps, func() error {
			g := scaffold.FindGenerator(list, kind)
			if g == nil || !g.NeedsIn {
				return ui.ErrSkip
			}
			return ui.Ask(false, huh.NewInput().Title(i18n.T("In which aggregate?")).Placeholder("Product").Description(i18n.T("Esc goes back · Ctrl+C cancels")).Value(&in))
		})
	}
	if err := ui.Wizard(steps...); err != nil {
		return err
	}
	return runAdd(path, scaffold.AddOptions{Kind: kind, Name: name, In: in, Plural: flagVals["plural"], Force: flags["force"], DryRun: flags["dry-run"]})
}

func runAdd(path string, opts scaffold.AddOptions) error {
	result, err := scaffold.Add(path, opts)
	if err != nil {
		return err
	}
	if ui.JSONOutput {
		return ui.PrintJSON(result)
	}
	if result.DryRun {
		ui.Infof("Dry run: nothing was written. It would create %d files:", len(result.Files))
	} else {
		ui.Okf("%s %s: %d file(s) created in %s", result.Kind, result.Name, len(result.Files), path)
	}
	for _, f := range result.Files {
		fmt.Println("  + " + f)
	}
	for _, f := range result.Patched {
		fmt.Println("  ~ " + f)
	}
	if len(result.Manual) > 0 {
		ui.Warnf("Could not find where to register it; do it by hand:")
		for _, m := range result.Manual {
			fmt.Println("    " + m)
		}
	}
	for _, n := range result.Next {
		ui.Infof("%s", n)
	}
	return nil
}

func menuAdd() error {
	var name string
	if err := pickProject(i18n.T("Add to which project?"), &name); err != nil {
		return err
	}
	path, err := projects.ProjectPath(name)
	if err != nil {
		return err
	}
	target := scaffold.TargetOf(path)
	if target == "" {
		return ui.Usagef("there are no generators for this type of project: %s", path)
	}
	return addInteractive(path, target, "", "", "", map[string]bool{})
}

func menuPublish() error {
	var name string
	if err := pickProject(i18n.T("Which project to publish?"), &name); err != nil {
		return err
	}
	return publishCommand([]string{name}, map[string]bool{})
}
