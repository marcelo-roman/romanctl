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
	ui.Spin(i18n.T("Looking for the latest version..."), func(func(string)) {
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
		ui.Okf("You are already on the latest version: %s", version)
		return nil
	}
	if flags["check"] {
		ui.Infof("New version available: %s (installed: %s). To update: romanctl self-update", want, version)
		return nil
	}
	path, err := selfupdate.Executable()
	if err != nil {
		return err
	}
	if ui.Interactive() && !flags["y"] && !flags["yes"] {
		ok, err := ui.Confirm(i18n.Sprintf("Update romanctl from %s to %s?", version, want))
		if err != nil {
			return err
		}
		if !ok {
			ui.Infof("Cancelled")
			return nil
		}
	}
	ui.Spin(i18n.Sprintf("Downloading version %s...", want), func(func(string)) { err = selfupdate.Install(want, path) })
	if err != nil {
		return err
	}
	templatesErr := updateTemplatesFor(want)
	if ui.JSONOutput {
		return ui.PrintJSON(map[string]any{"ok": true, "from": version, "to": want, "path": path, "templates": templatesErr == nil})
	}
	ui.Okf("romanctl updated from %s to %s (%s)", version, want, path)
	return nil
}

// updateTemplatesFor fetches the new version's templates, so they do not fall behind the binary.
func updateTemplatesFor(v string) error {
	if os.Getenv("ROMANCTL_TEMPLATES") != "" {
		return nil
	}
	scaffold.Version = v
	var err error
	ui.Spin(i18n.T("Downloading the templates..."), func(func(string)) { _, err = scaffold.Update("") })
	if err != nil {
		ui.Warnf("Could not download the templates: %v", err)
	}
	return err
}
