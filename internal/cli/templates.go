package cli

import (
	"romanctl/internal/i18n"
	"romanctl/internal/scaffold"
	"romanctl/internal/ui"
)

// templatesCommand shows and updates the templates, which live outside the binary.
func templatesCommand(args []string) error {
	switch arg(args, 0) {
	case "", "status", "path":
		status := map[string]any{"path": scaffold.Root(), "installed": scaffold.Installed(), "ref": scaffold.InstalledRef(),
			"count": len(scaffold.LoadTemplates())}
		if ui.JSONOutput {
			return ui.PrintJSON(status)
		}
		if !scaffold.Installed() {
			ui.Warnf("The templates have not been downloaded yet. With internet, run: romanctl templates update")
			return nil
		}
		i18n.Printf("Templates in %s\n", scaffold.Root())
		if ref := scaffold.InstalledRef(); ref != "" {
			i18n.Printf("Version: %s\n", ref)
		}
		i18n.Printf("%d project templates. To see them: romanctl catalog\n", status["count"])
		return nil
	case "update", "atualizar":
		var result scaffold.UpdateResult
		var err error
		ui.Spin(i18n.T("Downloading the templates..."), func(func(string)) { result, err = scaffold.Update(flagVals["ref"]) })
		if err != nil {
			return err
		}
		if ui.JSONOutput {
			return ui.PrintJSON(result)
		}
		if result.Fallback {
			ui.Warnf("Version %s has no published templates yet; the ones from main were used", version)
		}
		ui.Okf("Templates updated (%s) in %s", result.Ref, result.Path)
		return nil
	}
	return ui.Usagef("usage: romanctl templates [status|update] [--ref <tag or branch>] [--json]")
}
