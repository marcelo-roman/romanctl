package cli

import (
	"romanctl/internal/i18n"
	"romanctl/internal/offline"
	"romanctl/internal/ui"
)

func menuOffline() error {
	sub, err := ui.Choose(i18n.T("Offline mode"),
		ui.Opt(i18n.T("View the preparation status"), "status"),
		ui.Opt(i18n.T("Prepare now (downloads everything, needs internet)"), "prep"),
		ui.Opt(i18n.T("Test with the network blocked"), "test"))
	if err != nil {
		return err
	}
	return offline.Command(sub, nil)
}
