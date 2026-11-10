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
		return ui.Usagef("não há geradores para este tipo de projeto: %s", path)
	}
	if flags["list"] {
		return addList(target, path)
	}
	kind, name, in := arg(args, 0), arg(args, 1), flagVals["in"]
	if ui.Interactive() && addIsIncomplete(target, kind, name, in) {
		return addInteractive(path, target, kind, name, in, flags)
	}
	if kind == "" || name == "" {
		return ui.Usagef("informe o tipo e o nome: romanctl add <tipo> <nome>; veja com: romanctl add --list")
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
	i18n.Printf("Geradores para %s (%s):\n", path, target)
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
	// Um passo já respondido por argumento não entra: se entrasse, o Esc voltaria para um
	// passo que se pula sozinho, e a pessoa ficaria presa no seguinte.
	var steps []func() error
	if kind == "" {
		steps = append(steps, func() error {
			var opts []ui.Option
			for _, g := range list {
				opts = append(opts, ui.Opt(fmt.Sprintf("%-14s %s", g.Kind, ui.SGray.Render(i18n.T(g.Description))), g.Kind))
			}
			v, err := ui.PickOne(i18n.Sprintf("O que adicionar em %s?", target), opts)
			if err == nil {
				kind = v
			}
			return err
		})
	}
	if name == "" {
		steps = append(steps, func() error {
			return ui.Ask(false, huh.NewInput().Title(i18n.T("Nome")).Placeholder("Product").Description(i18n.T("Esc volta · Ctrl+C cancela")).
				Validate(func(v string) error {
					if scaffold.Pascal(v) == "" {
						return i18n.Errorf("informe um nome")
					}
					return nil
				}).Value(&name))
		})
	}
	if in == "" {
		// Só se sabe se o agregado é necessário depois de escolhido o tipo.
		steps = append(steps, func() error {
			g := scaffold.FindGenerator(list, kind)
			if g == nil || !g.NeedsIn {
				return ui.ErrSkip
			}
			return ui.Ask(false, huh.NewInput().Title(i18n.T("Em qual agregado?")).Placeholder("Product").Description(i18n.T("Esc volta · Ctrl+C cancela")).Value(&in))
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
		ui.Infof("Simulação: nada foi gravado. Seriam criados %d arquivos:", len(result.Files))
	} else {
		ui.Okf("%s %s: %d arquivo(s) criado(s) em %s", result.Kind, result.Name, len(result.Files), path)
	}
	for _, f := range result.Files {
		fmt.Println("  + " + f)
	}
	for _, f := range result.Patched {
		fmt.Println("  ~ " + f)
	}
	if len(result.Manual) > 0 {
		ui.Warnf("Não achei onde registrar; faça à mão:")
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
	if err := pickProject(i18n.T("Adicionar em qual projeto?"), &name); err != nil {
		return err
	}
	path, err := projects.ProjectPath(name)
	if err != nil {
		return err
	}
	target := scaffold.TargetOf(path)
	if target == "" {
		return ui.Usagef("não há geradores para este tipo de projeto: %s", path)
	}
	return addInteractive(path, target, "", "", "", map[string]bool{})
}

func menuPublish() error {
	var name string
	if err := pickProject(i18n.T("Publicar qual projeto?"), &name); err != nil {
		return err
	}
	return publishCommand([]string{name}, map[string]bool{})
}
