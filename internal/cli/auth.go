package cli

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"romanctl/internal/auth"
	"romanctl/internal/browser"
	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/ui"
)

// Comandos que rodam sem login: entrar, sair, ajuda, versão e o que o instalador chama.
var withoutLogin = []string{"login", "entrar", "logout", "sair", "whoami", "help", "ajuda", "version", "versao",
	"self-update", "update", "atualizar", "templates", "template", projects.SuperviseArg}

func needsLogin(args []string) bool {
	return len(args) == 0 || !slices.Contains(withoutLogin, args[0])
}

func requireLogin() error {
	err := auth.Require(context.Background())
	switch {
	case err == nil:
		return nil
	case errors.Is(err, auth.ErrNotLoggedIn):
		return i18n.Errorf("o romanctl só funciona com login. Entre com: romanctl login")
	case errors.Is(err, auth.ErrExpired):
		return i18n.Errorf("a sessão expirou ou foi encerrada. Entre de novo com: romanctl login")
	}
	return err
}

func loginCommand() error {
	ep := auth.DefaultEndpoints()
	if v := flagVals["issuer"]; v != "" {
		ep.Issuer = v
	}
	if v := flagVals["api"]; v != "" {
		ep.API = v
	}
	if v, ok := flagVals["idp"]; ok {
		ep.IDP = v
	}
	session, err := auth.Login(context.Background(), auth.LoginOptions{Endpoints: ep, Open: func(u string) {
		ui.Infof("Abra no navegador para entrar: %s", u)
		if ui.Interactive() {
			_ = browser.OpenURL(u)
		}
	}})
	if err != nil {
		return err
	}
	if err := auth.Save(session); err != nil {
		return err
	}
	if ui.JSONOutput {
		return ui.PrintJSON(sessionInfo(session))
	}
	ui.Okf("Logado como %s em %s", session.Username, session.Issuer)
	return nil
}

func logoutCommand() error {
	session, err := auth.Load()
	if errors.Is(err, auth.ErrNotLoggedIn) {
		ui.Infof("Nenhuma sessão aberta")
		return nil
	}
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := session.Revoke(ctx); err != nil {
		ui.Warnf("Não consegui encerrar a sessão no Keycloak (%v); ela some daqui mesmo assim", err)
	}
	if err := auth.Remove(); err != nil {
		return err
	}
	if ui.JSONOutput {
		return ui.PrintJSON(map[string]any{"ok": true})
	}
	ui.Okf("Sessão de %s encerrada", session.Username)
	return nil
}

func sessionInfo(s *auth.Session) map[string]any {
	info := map[string]any{"loggedIn": true, "username": s.Username, "name": s.Name, "email": s.Email,
		"issuer": s.Issuer, "client": s.ClientID, "api": s.API, "roles": s.Roles}
	if !s.RefreshExpiry.IsZero() {
		info["expires"] = s.RefreshExpiry
	}
	return info
}

func whoamiCommand() error {
	session, err := auth.Current(context.Background())
	if errors.Is(err, auth.ErrNotLoggedIn) {
		if ui.JSONOutput {
			return ui.PrintJSON(map[string]any{"loggedIn": false})
		}
		ui.Warnf("Sem login. Entre com: romanctl login")
		return ui.ErrReported
	}
	if err != nil {
		return err
	}
	if ui.JSONOutput {
		return ui.PrintJSON(sessionInfo(session))
	}
	i18n.Printf("Usuário: %s\n", session.Username)
	if session.Name != "" {
		i18n.Printf("Nome:    %s\n", session.Name)
	}
	i18n.Printf("Realm:   %s\n", session.Issuer)
	i18n.Printf("API:     %s\n", session.API)
	if len(session.Roles) > 0 {
		i18n.Printf("Papéis:  %s\n", strings.Join(session.Roles, ", "))
	}
	if !session.RefreshExpiry.IsZero() {
		i18n.Printf("Sessão vale até %s (renova a cada uso)\n", session.RefreshExpiry.Local().Format("02/01/2006 15:04"))
	}
	return nil
}
