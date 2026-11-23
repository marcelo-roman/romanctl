package cli

import (
	"fmt"

	"romanctl/internal/claude"
	"romanctl/internal/i18n"
	"romanctl/internal/ui"
)

// claudeCommand installs the skills that teach Claude Code to use romanctl.
func claudeCommand(args []string, flags map[string]bool) error {
	dir := claude.Dir()
	switch arg(args, 0) {
	case "", "status":
		list := claude.Status(dir)
		if ui.JSONOutput {
			return ui.PrintJSON(map[string]any{"dir": dir, "skills": list})
		}
		i18n.Printf("Claude skills in %s\n", dir)
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
	return ui.Usagef("usage: romanctl claude [status|install|uninstall] [--force] [--dry-run] [--json]")
}

func claudeState(state string) string {
	switch state {
	case claude.StateInstalled:
		return ui.SGreen.Render(i18n.T("installed"))
	case claude.StateOutdated:
		return ui.SYellow.Render(i18n.T("outdated")) + ui.SGray.Render(i18n.T("  (romanctl claude install)"))
	case claude.StateForeign:
		return ui.SRed.Render(i18n.T("foreign")) + ui.SGray.Render(i18n.T("  (created outside romanctl; --force replaces it)"))
	}
	return ui.SGray.Render(i18n.T("missing"))
}

func reportClaude(dir string, results []claude.Result, err error, dryRun bool) error {
	if ui.JSONOutput {
		if err != nil {
			return err
		}
		return ui.PrintJSON(map[string]any{"ok": true, "dir": dir, "dryRun": dryRun, "results": results})
	}
	if dryRun {
		ui.Infof("Dry run: nothing was written.")
	}
	labels := map[string]string{"installed": i18n.T("installed"), "updated": i18n.T("updated"), "unchanged": i18n.T("unchanged"),
		"skipped": i18n.T("kept: created outside romanctl (use --force)"), "removed": i18n.T("removed"),
		"kept": i18n.T("kept: created outside romanctl"), "absent": i18n.T("was not installed")}
	for _, r := range results {
		fmt.Printf("  %-22s %s\n", r.Name, labels[r.Action])
	}
	if err != nil {
		return err
	}
	ui.Okf("Skills in %s", dir)
	return nil
}
