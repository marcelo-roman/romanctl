package cli

import (
	"os"

	"romanctl/internal/i18n"
	"romanctl/internal/scaffold"
	"romanctl/internal/selfupdate"
	"romanctl/internal/ui"
)

func selfUpdateCommand(flags map[string]bool) error {
	var latest string
	var err error
	target := flagVals["version"]
	ui.Spin(i18n.T("Procurando a última versão..."), func(func(string)) {
		if target == "" {
			latest, err = selfupdate.Latest()
		}
	})
	if err != nil {
		return err
	}
	want := selfupdate.Normalize(target)
	if want == "" {
		want = latest
	}
	newer := selfupdate.Compare(version, want) < 0
	if ui.JSONOutput && flags["check"] {
		return ui.PrintJSON(map[string]any{"current": version, "latest": want, "updateAvailable": newer})
	}
	if target == "" && !newer {
		ui.Okf("Você já está na última versão: %s", version)
		return nil
	}
	if flags["check"] {
		ui.Infof("Versão nova disponível: %s (instalada: %s). Para atualizar: romanctl self-update", want, version)
		return nil
	}
	path, err := selfupdate.Executable()
	if err != nil {
		return err
	}
	if ui.Interactive() && !flags["y"] && !flags["yes"] {
		ok, err := ui.Confirm(i18n.Sprintf("Atualizar o romanctl de %s para %s?", version, want))
		if err != nil {
			return err
		}
		if !ok {
			ui.Infof("Cancelado")
			return nil
		}
	}
	ui.Spin(i18n.Sprintf("Baixando a versão %s...", want), func(func(string)) { err = selfupdate.Install(want, path) })
	if err != nil {
		return err
	}
	templatesErr := updateTemplatesFor(want)
	if ui.JSONOutput {
		return ui.PrintJSON(map[string]any{"ok": true, "from": version, "to": want, "path": path, "templates": templatesErr == nil})
	}
	ui.Okf("romanctl atualizado de %s para %s (%s)", version, want, path)
	return nil
}

// updateTemplatesFor traz os templates da versão nova, para não ficarem para trás do binário.
func updateTemplatesFor(v string) error {
	if os.Getenv("ROMANCTL_TEMPLATES") != "" {
		return nil
	}
	scaffold.Version = v
	var err error
	ui.Spin(i18n.T("Baixando os templates..."), func(func(string)) { _, err = scaffold.Update("") })
	if err != nil {
		ui.Warnf("Não consegui baixar os templates: %v", err)
	}
	return err
}
