package cli

import (
	"romanctl/internal/i18n"
	"romanctl/internal/scaffold"
	"romanctl/internal/ui"
)

// templatesCommand mostra e atualiza os templates, que ficam fora do binário.
func templatesCommand(args []string) error {
	switch arg(args, 0) {
	case "", "status", "path":
		status := map[string]any{"path": scaffold.Root(), "installed": scaffold.Installed(), "ref": scaffold.InstalledRef(),
			"count": len(scaffold.LoadTemplates())}
		if ui.JSONOutput {
			return ui.PrintJSON(status)
		}
		if !scaffold.Installed() {
			ui.Warnf("Os templates ainda não foram baixados. Com internet, rode: romanctl templates update")
			return nil
		}
		i18n.Printf("Templates em %s\n", scaffold.Root())
		if ref := scaffold.InstalledRef(); ref != "" {
			i18n.Printf("Versão: %s\n", ref)
		}
		i18n.Printf("%d templates de projeto. Para ver: romanctl catalog\n", status["count"])
		return nil
	case "update", "atualizar":
		var result scaffold.UpdateResult
		var err error
		ui.Spin(i18n.T("Baixando os templates..."), func(func(string)) { result, err = scaffold.Update(flagVals["ref"]) })
		if err != nil {
			return err
		}
		if ui.JSONOutput {
			return ui.PrintJSON(result)
		}
		if result.Fallback {
			ui.Warnf("A versão %s ainda não tem templates publicados; vieram os da main", version)
		}
		ui.Okf("Templates atualizados (%s) em %s", result.Ref, result.Path)
		return nil
	}
	return ui.Usagef("use: romanctl templates [status|update] [--ref <tag ou branch>] [--json]")
}
