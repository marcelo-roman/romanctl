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
		// Fora de um projeto: pergunta qual, em vez de falhar.
		var name string
		if err = pickProject(i18n.T("Publicar qual projeto?"), &name); err == nil {
			path, err = projects.ProjectPath(name)
		}
	}
	if err != nil {
		return err
	}
	opts := publish.Options{Public: flags["public"], NoCI: flags["no-ci"], DryRun: flags["dry-run"], Description: flagVals["description"]}
	var plan publish.Result
	ui.Spin(i18n.T("Conferindo Git e GitHub..."), func(func(string)) { plan, err = publish.Plan(path, opts) })
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
		ui.Infof("Simulação: nada foi feito. O publish faria:")
		printPublishPlan(plan)
		return nil
	}
	if !flags["y"] && !flags["yes"] {
		if !ui.Interactive() {
			return ui.Usagef("publicar precisa de confirmação: rode em um terminal ou use --yes")
		}
		printPublishPlan(plan)
		title := i18n.Sprintf("Publicar %s como repositório privado?", plan.Name)
		if plan.Exists {
			title = i18n.Sprintf("Enviar %s para o repositório existente?", plan.Name)
		} else if opts.Public {
			title = i18n.Sprintf("Publicar %s como repositório PÚBLICO? Qualquer pessoa poderá ver o código.", plan.Name)
		}
		ok, err := ui.Confirm(title)
		if err != nil {
			return err
		}
		if !ok {
			ui.Infof("Cancelado")
			return nil
		}
	}
	var result publish.Result
	ui.Spin(i18n.Sprintf("Publicando %s...", plan.Name), func(func(string)) { result, err = publish.Run(path, opts) })
	if err != nil {
		return err
	}
	if ui.JSONOutput {
		return ui.PrintJSON(result)
	}
	ui.Okf("%s publicado: %s", result.Name, result.URL)
	return nil
}

func printPublishPlan(p publish.Result) {
	visibility := i18n.T("privado")
	if p.Public {
		visibility = i18n.T("PÚBLICO")
	}
	if p.Exists {
		i18n.Printf("  Repositório  já existe, só envia as alterações (%s)\n", p.Remote)
	} else {
		i18n.Printf("  Repositório  %s, %s\n", p.Name, visibility)
	}
	for _, f := range p.Generated {
		i18n.Printf("  Gera         %s\n", f)
	}
}

// askPublishOptions pergunta a visibilidade e a descrição, salvo se vieram por argumento.
func askPublishOptions(opts *publish.Options, publicGiven bool) error {
	var steps []func() error
	if !publicGiven {
		steps = append(steps, func() error {
			choice, err := ui.Choose(i18n.T("Visibilidade do repositório"),
				ui.Opt(i18n.T("Privado"), "private"),
				ui.Opt(i18n.T("Público: qualquer pessoa vê o código"), "public"))
			opts.Public = choice == "public"
			return err
		})
	}
	if opts.Description == "" {
		steps = append(steps, func() error {
			return ui.Ask(false, huh.NewInput().Title(i18n.T("Descrição (opcional)")).Description(i18n.T("Esc volta · Ctrl+C cancela")).Value(&opts.Description))
		})
	}
	return ui.Wizard(steps...)
}
