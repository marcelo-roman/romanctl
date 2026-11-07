package cli

import (
	"romanctl/internal/i18n"
	"romanctl/internal/offline"
	"romanctl/internal/ui"
)

func menuOffline() error {
	sub, err := ui.Choose(i18n.T("Modo offline"),
		ui.Opt(i18n.T("Ver o estado da preparação"), "status"),
		ui.Opt(i18n.T("Preparar agora (baixa tudo, precisa de internet)"), "prep"),
		ui.Opt(i18n.T("Testar com a rede bloqueada"), "test"))
	if err != nil {
		return err
	}
	return offline.Command(sub, nil)
}
