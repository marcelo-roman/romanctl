package cli

import (
	"fmt"

	"romanctl/internal/claude"
	"romanctl/internal/i18n"
	"romanctl/internal/ui"
)

// claudeCommand instala as skills que ensinam o Claude Code a usar o romanctl.
func claudeCommand(args []string, flags map[string]bool) error {
	dir := claude.Dir()
	switch arg(args, 0) {
	case "", "status":
		list := claude.Status(dir)
		if ui.JSONOutput {
			return ui.PrintJSON(map[string]any{"dir": dir, "skills": list})
		}
		i18n.Printf("Skills do Claude em %s\n", dir)
		for _, s := range list {
			fmt.Printf("  %-22s %s\n", s.Name, claudeState(s.State))
		}
		return nil
	case "install", "instalar", "update", "atualizar":
		results, err := claude.Install(dir, version, flags["force"], flags["dry-run"])
		return reportClaude(dir, results, err, flags["dry-run"])
	case "uninstall", "remove", "rm", "remover":
		results, err := claude.Uninstall(dir, flags["dry-run"])
		return reportClaude(dir, results, err, flags["dry-run"])
	}
	return ui.Usagef("use: romanctl claude [status|install|uninstall] [--force] [--dry-run] [--json]")
}

func claudeState(state string) string {
	switch state {
	case claude.StateInstalled:
		return ui.SGreen.Render(i18n.T("instalada"))
	case claude.StateOutdated:
		return ui.SYellow.Render(i18n.T("desatualizada")) + ui.SGray.Render(i18n.T("  (romanctl claude install)"))
	case claude.StateForeign:
		return ui.SRed.Render(i18n.T("alheia")) + ui.SGray.Render(i18n.T("  (criada fora do romanctl; --force troca)"))
	}
	return ui.SGray.Render(i18n.T("ausente"))
}

func reportClaude(dir string, results []claude.Result, err error, dryRun bool) error {
	if ui.JSONOutput {
		if err != nil {
			return err
		}
		return ui.PrintJSON(map[string]any{"ok": true, "dir": dir, "dryRun": dryRun, "results": results})
	}
	if dryRun {
		ui.Infof("Simulação: nada foi gravado.")
	}
	labels := map[string]string{"installed": i18n.T("instalada"), "updated": i18n.T("atualizada"), "unchanged": i18n.T("sem mudança"),
		"skipped": i18n.T("mantida: foi criada fora do romanctl (use --force)"), "removed": i18n.T("removida"),
		"kept": i18n.T("mantida: foi criada fora do romanctl"), "absent": i18n.T("não estava instalada")}
	for _, r := range results {
		fmt.Printf("  %-22s %s\n", r.Name, labels[r.Action])
	}
	if err != nil {
		return err
	}
	ui.Okf("Skills em %s", dir)
	return nil
}
