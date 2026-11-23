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

// Commands that run without login: login, logout, help, version and what the installer calls.
var withoutLogin = []string{"login", "entrar", "logout", "quit", "whoami", "help", "help", "version", "versao",
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
		return i18n.Errorf("romanctl only works with a login. Sign in with: romanctl login")
	case errors.Is(err, auth.ErrExpired):
		return i18n.Errorf("the session expired or was ended. Sign in again with: romanctl login")
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
		ui.Infof("Open in the browser to sign in: %s", u)
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
	ui.Okf("Logged in as %s at %s", session.Username, session.Issuer)
	return nil
}

func logoutCommand() error {
	session, err := auth.Load()
	if errors.Is(err, auth.ErrNotLoggedIn) {
		ui.Infof("No open session")
		return nil
	}
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := session.Revoke(ctx); err != nil {
		ui.Warnf("Could not end the session in Keycloak (%v); it is removed from here anyway", err)
	}
	if err := auth.Remove(); err != nil {
		return err
	}
	if ui.JSONOutput {
		return ui.PrintJSON(map[string]any{"ok": true})
	}
	ui.Okf("Session of %s ended", session.Username)
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
		ui.Warnf("Not logged in. Sign in with: romanctl login")
		return ui.ErrReported
	}
	if err != nil {
		return err
	}
	if ui.JSONOutput {
		return ui.PrintJSON(sessionInfo(session))
	}
	i18n.Printf("User:    %s\n", session.Username)
	if session.Name != "" {
		i18n.Printf("Name:    %s\n", session.Name)
	}
	i18n.Printf("Realm:   %s\n", session.Issuer)
	i18n.Printf("API:     %s\n", session.API)
	if len(session.Roles) > 0 {
		i18n.Printf("Roles:   %s\n", strings.Join(session.Roles, ", "))
	}
	if !session.RefreshExpiry.IsZero() {
		i18n.Printf("Session valid until %s (renews on every use)\n", session.RefreshExpiry.Local().Format("02/01/2006 15:04"))
	}
	return nil
}
