package cli

import (
	"romanctl/internal/i18n"
	"romanctl/internal/ui"
)

// configCommand shows or saves the language: pt, en or auto, which goes back to following the system.
func configCommand(args []string) error {
	if arg(args, 0) != "" && arg(args, 0) != "lang" {
		return ui.Usagef("usage: romanctl config lang [pt|en|auto]")
	}
	value := arg(args, 1)
	if value == "" && ui.Interactive() {
		var err error
		value, err = ui.Choose(i18n.Sprintf("Language: %s (source: %s)\n", i18n.Lang(), i18n.Source()),
			ui.Opt("Portuguese", "pt"), ui.Opt("English", "en"),
			ui.Opt(i18n.T("Follow the system"), "auto"))
		if err != nil {
			return err
		}
	}
	switch {
	case value == "":
		if ui.JSONOutput {
			return ui.PrintJSON(map[string]string{"lang": i18n.Lang(), "source": i18n.Source()})
		}
		i18n.Printf("Language: %s (source: %s)\n", i18n.Lang(), i18n.Source())
		return nil
	case value == "auto":
		if err := i18n.Save(""); err != nil {
			return err
		}
		ui.Okf("Saved language cleared: it follows the system again")
		return nil
	}
	lang := i18n.Parse(value)
	if lang == "" {
		return ui.Usagef("usage: romanctl config lang [pt|en|auto]")
	}
	if err := i18n.Save(lang); err != nil {
		return err
	}
	i18n.Set(lang)
	ui.Okf("Language saved: %s", lang)
	return nil
}
