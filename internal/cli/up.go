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
		return ui.Usagef("com --json, use -d ou --no-run: em primeiro plano o comando só termina quando o app para")
	}
	path, err := projects.Locate(arg(args, 0))
	if err != nil && arg(args, 0) == "" && ui.Interactive() {
		var name string
		if err = pickProject(i18n.T("Subir qual projeto?"), &name); err == nil {
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
		ui.Okf("Ambiente de %s pronto. Para rodar: romanctl run %s", result.Project, result.Project)
		return nil
	}
	if opts.Detach {
		if result.URL != "" {
			ui.Okf("%s no ar → %s", result.Project, result.URL)
		} else {
			ui.Okf("%s iniciado; sem endereço para conferir se está respondendo", result.Project)
		}
		i18n.Printf("   Log:    romanctl logs %s\n   Parar:  romanctl down %s\n", result.Project, result.Project)
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
			ui.Infof("Nenhum projeto no ar.")
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

// pickDown usa o projeto da pasta atual, se ele estiver no ar; senão, mostra os que estão.
func pickDown() (string, error) {
	ups := state.ReadState().Ups
	if path, err := projects.Locate(""); err == nil {
		if _, ok := ups[filepath.Base(path)]; ok {
			return filepath.Base(path), nil
		}
	}
	if !ui.Interactive() {
		return "", ui.Usagef("informe o projeto: romanctl down <projeto>, ou use --all")
	}
	var opts []ui.Option
	for _, name := range up.Names() {
		opts = append(opts, ui.Opt(fmt.Sprintf("%-28s %s", name, ui.SGray.Render(up.Describe(ups[name]))), name))
	}
	if len(opts) == 0 {
		ui.Infof("Nenhum projeto no ar.")
		return "", nil
	}
	return ui.PickOne(i18n.T("Derrubar qual projeto?"), opts)
}
