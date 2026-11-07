package cli

import (
	"romanctl/internal/i18n"
	"romanctl/internal/ui"
)

// configCommand mostra ou grava o idioma: pt, en ou auto, que volta a seguir o sistema.
func configCommand(args []string) error {
	if arg(args, 0) != "" && arg(args, 0) != "lang" {
		return ui.Usagef("use: romanctl config lang [pt|en|auto]")
	}
	value := arg(args, 1)
	if value == "" && ui.Interactive() {
		var err error
		value, err = ui.Choose(i18n.Sprintf("Idioma: %s (origem: %s)\n", i18n.Lang(), i18n.Source()),
			ui.Opt("Português", "pt"), ui.Opt("English", "en"),
			ui.Opt(i18n.T("Seguir o sistema"), "auto"))
		if err != nil {
			return err
		}
	}
	switch {
	case value == "":
		if ui.JSONOutput {
			return ui.PrintJSON(map[string]string{"lang": i18n.Lang(), "source": i18n.Source()})
		}
		i18n.Printf("Idioma: %s (origem: %s)\n", i18n.Lang(), i18n.Source())
		return nil
	case value == "auto":
		if err := i18n.Save(""); err != nil {
			return err
		}
		ui.Okf("Idioma salvo apagado: volta a seguir o sistema")
		return nil
	}
	lang := i18n.Parse(value)
	if lang == "" {
		return ui.Usagef("use: romanctl config lang [pt|en|auto]")
	}
	if err := i18n.Save(lang); err != nil {
		return err
	}
	i18n.Set(lang)
	ui.Okf("Idioma salvo: %s", lang)
	return nil
}
