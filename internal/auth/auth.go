// Package auth faz o login do romanctl no Keycloak (realm private, authorization code com PKCE)
// e guarda a sessão. Sem sessão válida a CLI não roda.
package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"romanctl/internal/config"
	"romanctl/internal/i18n"
)

const (
	DefaultIssuer   = "https://auth.marceloroman.com.br/realms/private"
	DefaultClientID = "romanctl"
	DefaultAPI      = "https://iam-api.marceloroman.com.br"
	// DefaultIDP é o Google Workspace da criarch, o único jeito de entrar no realm private.
	DefaultIDP = "google"
	// revalidate é de quanto em quanto tempo a sessão é conferida no Keycloak, quando há rede.
	revalidate = 12 * time.Hour
)

// Session é o que fica em disco depois do login. O refresh token é offline: vale enquanto for
// usado pelo menos uma vez a cada 30 dias, mesmo sem rede no meio.
type Session struct {
	Issuer        string    `json:"issuer"`
	ClientID      string    `json:"clientId"`
	API           string    `json:"api"`
	Username      string    `json:"username"`
	Name          string    `json:"name,omitempty"`
	Email         string    `json:"email,omitempty"`
	Roles         []string  `json:"roles"`
	AccessToken   string    `json:"accessToken"`
	AccessExpiry  time.Time `json:"accessExpiry"`
	RefreshToken  string    `json:"refreshToken"`
	RefreshExpiry time.Time `json:"refreshExpiry,omitempty"`
	ValidatedAt   time.Time `json:"validatedAt"`
}

var (
	ErrNotLoggedIn = errors.New("not logged in")
	ErrExpired     = errors.New("session expired")
)

func envOr(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return def
}

// Endpoints de onde fazer login; flags do login e variáveis de ambiente mudam o padrão.
type Endpoints struct {
	Issuer   string
	ClientID string
	API      string
	IDP      string
}

func DefaultEndpoints() Endpoints {
	return Endpoints{
		Issuer:   envOr("ROMANCTL_AUTH_ISSUER", DefaultIssuer),
		ClientID: envOr("ROMANCTL_AUTH_CLIENT", DefaultClientID),
		API:      envOr("ROMANCTL_IAM_API", DefaultAPI),
		IDP:      envOr("ROMANCTL_AUTH_IDP", DefaultIDP),
	}
}

func (e Endpoints) oidc(path string) string { return e.Issuer + "/protocol/openid-connect/" + path }

func SessionPath() string {
	if p := os.Getenv("ROMANCTL_SESSION"); p != "" {
		return p
	}
	return filepath.Join(config.Home, ".config", "romanctl", "session.json")
}

func Load() (*Session, error) {
	data, err := os.ReadFile(SessionPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotLoggedIn
	}
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil || s.RefreshToken == "" {
		return nil, ErrNotLoggedIn
	}
	return &s, nil
}

// Save grava com 0600 e troca o arquivo de uma vez: a web-ui roda vários romanctl ao mesmo tempo.
func Save(s *Session) error {
	path := SessionPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".session-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func Remove() error {
	err := os.Remove(SessionPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Session) endpoints() Endpoints {
	return Endpoints{Issuer: s.Issuer, ClientID: s.ClientID, API: s.API}
}

func (s *Session) RefreshExpired(now time.Time) bool {
	return !s.RefreshExpiry.IsZero() && now.After(s.RefreshExpiry)
}

type claims struct {
	Username       string `json:"preferred_username"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Expiry         int64  `json:"exp"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
}

// managerAudience é o client do IAM Manager; os papéis de quem usa o romanctl são client roles dele.
const managerAudience = "iam-manager-api"

// readClaims lê o corpo do JWT sem conferir a assinatura: o token veio do Keycloak por TLS, e quem
// confere de verdade é a API.
func readClaims(token string) (claims, error) {
	var c claims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c, i18n.Errorf("token em formato inesperado")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return c, err
	}
	return c, json.Unmarshal(payload, &c)
}

func expiry(token string) time.Time {
	c, err := readClaims(token)
	if err != nil || c.Expiry == 0 {
		return time.Time{}
	}
	return time.Unix(c.Expiry, 0)
}

func (s *Session) apply(t tokenResponse, now time.Time) error {
	c, err := readClaims(t.AccessToken)
	if err != nil {
		return err
	}
	s.AccessToken = t.AccessToken
	s.AccessExpiry = now.Add(time.Duration(t.ExpiresIn) * time.Second)
	if t.RefreshToken != "" {
		s.RefreshToken = t.RefreshToken
		s.RefreshExpiry = expiry(t.RefreshToken)
	}
	s.Username, s.Name, s.Email = c.Username, c.Name, c.Email
	s.Roles = c.ResourceAccess[managerAudience].Roles
	s.ValidatedAt = now
	return nil
}
