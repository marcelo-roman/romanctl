package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"romanctl/internal/i18n"
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// ErrRejected é a recusa do Keycloak (sessão revogada, usuário bloqueado), diferente de falta de rede.
type ErrRejected struct{ Code, Description string }

func (e ErrRejected) Error() string {
	if e.Description != "" {
		return e.Code + ": " + e.Description
	}
	return e.Code
}

func postToken(ctx context.Context, endpoint string, form url.Values) (tokenResponse, error) {
	var t tokenResponse
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return t, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return t, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return t, fmt.Errorf("%s: HTTP %d", endpoint, resp.StatusCode)
	}
	if t.Error != "" {
		return t, ErrRejected{t.Error, t.ErrorDescription}
	}
	if resp.StatusCode != http.StatusOK || t.AccessToken == "" {
		return t, fmt.Errorf("%s: HTTP %d", endpoint, resp.StatusCode)
	}
	return t, nil
}

// Token é um access token e o momento em que ele deixa de valer.
type Token struct {
	AccessToken string
	Expiry      time.Time
}

// ClientCredentials troca client_id e client_secret por um access token (sem refresh token).
func ClientCredentials(ctx context.Context, endpoint, clientID, clientSecret string) (Token, error) {
	t, err := postToken(ctx, endpoint, url.Values{"grant_type": {"client_credentials"}, "client_id": {clientID}, "client_secret": {clientSecret}})
	if err != nil {
		return Token{}, err
	}
	return Token{AccessToken: t.AccessToken, Expiry: time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)}, nil
}

func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// LoginOptions: Open recebe o endereço de login e tenta abrir o navegador.
type LoginOptions struct {
	Endpoints Endpoints
	Open      func(url string)
	Timeout   time.Duration
}

type callback struct {
	code string
	err  error
}

// Login faz o authorization code com PKCE (S256) num redirect de loopback, como manda a RFC 8252:
// uma porta livre em 127.0.0.1 recebe o código, e só quem tem o verifier troca por token.
func Login(ctx context.Context, opts LoginOptions) (*Session, error) {
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Minute
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	defer listener.Close()
	redirect := fmt.Sprintf("http://127.0.0.1:%d/callback", listener.Addr().(*net.TCPAddr).Port)
	verifier, state := randomString(48), randomString(24)
	ep := opts.Endpoints

	done := make(chan callback, 1)
	server := &http.Server{ReadHeaderTimeout: 10 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/callback" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		var result callback
		switch {
		case q.Get("state") != state:
			result.err = i18n.Errorf("resposta de login com state diferente; tente de novo")
		case q.Get("error") != "":
			result.err = ErrRejected{q.Get("error"), q.Get("error_description")}
		case q.Get("code") == "":
			result.err = i18n.Errorf("o Keycloak não devolveu o código de login")
		default:
			result.code = q.Get("code")
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		message := i18n.T("Login feito. Pode fechar esta aba e voltar ao terminal.")
		if result.err != nil {
			message = i18n.T("O login falhou") + ": " + result.err.Error()
		}
		fmt.Fprintf(w, `<!doctype html><meta charset="utf-8"><title>romanctl</title><body style="font-family:sans-serif;padding:3rem"><h2>romanctl</h2><p>%s</p>`, html.EscapeString(message))
		select {
		case done <- result:
		default:
		}
	})}
	go server.Serve(listener)
	defer server.Close()

	query := url.Values{
		"client_id":             {ep.ClientID},
		"response_type":         {"code"},
		"scope":                 {"openid offline_access"},
		"redirect_uri":          {redirect},
		"state":                 {state},
		"code_challenge":        {challenge(verifier)},
		"code_challenge_method": {"S256"},
	}
	// Com o hint, o Keycloak pula a tela de login e vai direto ao Google.
	if ep.IDP != "" && ep.IDP != "-" {
		query.Set("kc_idp_hint", ep.IDP)
	}
	auth := ep.oidc("auth") + "?" + query.Encode()
	if opts.Open != nil {
		opts.Open(auth)
	}

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	var result callback
	select {
	case result = <-done:
	case <-ctx.Done():
		return nil, i18n.Errorf("o login não foi concluído a tempo")
	}
	if result.err != nil {
		return nil, result.err
	}
	t, err := postToken(ctx, ep.oidc("token"), url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {ep.ClientID},
		"code":          {result.code},
		"redirect_uri":  {redirect},
		"code_verifier": {verifier},
	})
	if err != nil {
		return nil, err
	}
	s := &Session{Issuer: ep.Issuer, ClientID: ep.ClientID, API: ep.API}
	if err := s.apply(t, time.Now()); err != nil {
		return nil, err
	}
	return s, nil
}

// Refresh troca o refresh token por um access token novo e grava a sessão.
func (s *Session) Refresh(ctx context.Context) error {
	t, err := postToken(ctx, s.endpoints().oidc("token"), url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {s.ClientID},
		"refresh_token": {s.RefreshToken},
	})
	if err != nil {
		return err
	}
	if err := s.apply(t, time.Now()); err != nil {
		return err
	}
	return Save(s)
}

// Revoke encerra a sessão offline no Keycloak; o logout local acontece mesmo se falhar.
func (s *Session) Revoke(ctx context.Context) error {
	form := url.Values{"client_id": {s.ClientID}, "token": {s.RefreshToken}, "token_type_hint": {"refresh_token"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoints().oidc("revoke"), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke: HTTP %d", resp.StatusCode)
	}
	return nil
}

func IsRejected(err error) bool {
	var r ErrRejected
	return errors.As(err, &r)
}
