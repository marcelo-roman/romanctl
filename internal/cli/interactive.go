package cli

import (
	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/services"
	"romanctl/internal/ui"
)

// serviceCommand abre a lista de serviços quando nenhum foi informado em um terminal.
func serviceCommand(action, name string, f func(*services.Service) error) error {
	if name == "" && ui.Interactive() {
		return menuServices(action)
	}
	return eachService(name, f)
}

func menuRestore() error {
	sub, err := ui.Choose(i18n.T("Restaurar"),
		ui.Opt(i18n.T("Sobe de novo o que estava ligado"), "run"),
		ui.Opt(i18n.T("Ver o estado salvo"), "status"),
		ui.Opt(i18n.T("Salvar o que está rodando agora"), "save"),
		ui.Opt(i18n.T("Limpar o estado salvo"), "clear"))
	if err != nil {
		return err
	}
	if sub == "run" {
		sub = ""
	}
	return services.RestoreCommand(sub)
}

// newCommand cria direto quando tudo foi informado; em um terminal, pergunta o que falta.
func newCommand(args []string, flags map[string]bool) error {
	a, b := arg(args, 0), arg(args, 1)
	stack, archFlag := flagVals["stack"], flagVals["arch"]
	git := !flags["sem-git"]

	if stack != "" || archFlag != "" {
		if a != "" && stack != "" && archFlag != "" {
			return scaffold.NewFromTemplate(a, stack, archFlag, flagVals["with"], git, flags["dry-run"])
		}
		if !ui.Interactive() {
			if a == "" {
				return ui.Usagef("informe o nome: romanctl new <nome> --stack <stack> --arch <arquiteturas>")
			}
			return scaffold.NewFromTemplate(a, stack, archFlag, flagVals["with"], git, flags["dry-run"])
		}
		return runNewWizard(newRequest{name: a, stack: stack, arch: archFlag, with: scaffold.SplitServices(flagVals["with"]),
			noGit: !git, dryRun: flags["dry-run"]})
	}
	if a != "" && b != "" {
		return projects.NewProject(a, b, git, flags["offline"])
	}
	req := newRequest{typ: a, noGit: !git}
	// "new loja" informa o nome; "new express" informa o tipo.
	if a != "" && projects.FindType(a) == nil {
		req.typ, req.name = "", a
	}
	return runNewWizard(req)
}
