package auth

import (
	"context"
	"net/url"
	"os"
	"sync"
	"time"

	"romanctl/internal/i18n"
)

// Pipelines have no browser: ROMANCTL_CLIENT_ID and ROMANCTL_CLIENT_SECRET log in with
// client_credentials using an API client of the same realm, created in IAM Manager.
func serviceCredentials() (id, secret string, ok bool) {
	id, secret = os.Getenv("ROMANCTL_CLIENT_ID"), os.Getenv("ROMANCTL_CLIENT_SECRET")
	return id, secret, id != "" && secret != ""
}

var service struct {
	sync.Mutex
	session *Session
}

func serviceSession(ctx context.Context) (*Session, error) {
	service.Lock()
	defer service.Unlock()
	if s := service.session; s != nil && time.Until(s.AccessExpiry) > 30*time.Second {
		return s, nil
	}
	id, secret, _ := serviceCredentials()
	ep := DefaultEndpoints()
	t, err := postToken(ctx, ep.oidc("token"), url.Values{"grant_type": {"client_credentials"}, "client_id": {id}, "client_secret": {secret}})
	if err != nil {
		return nil, i18n.Errorf("login with ROMANCTL_CLIENT_ID failed: %v", err)
	}
	s := &Session{Issuer: ep.Issuer, ClientID: id, API: ep.API}
	if err := s.apply(t, time.Now()); err != nil {
		return nil, err
	}
	service.session = s
	return s, nil
}

// Current returns the session in use: the service client's, if configured, or the login's.
func Current(ctx context.Context) (*Session, error) {
	if _, _, ok := serviceCredentials(); ok {
		return serviceSession(ctx)
	}
	return Load()
}

// Require is the CLI's gate. Offline, the saved session is enough while the refresh token is valid;
// online, it is checked against Keycloak every 12 h, so a revoked session or a blocked user is out at once.
func Require(ctx context.Context) error {
	if _, _, ok := serviceCredentials(); ok {
		_, err := serviceSession(ctx)
		return err
	}
	s, err := Load()
	if err != nil {
		return err
	}
	now := time.Now()
	if s.RefreshExpired(now) {
		_ = Remove()
		return ErrExpired
	}
	if now.Sub(s.ValidatedAt) < revalidate {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.Refresh(ctx); err != nil {
		if IsRejected(err) {
			_ = Remove()
			return ErrExpired
		}
	}
	return nil
}

// AccessToken returns a valid access token, refreshing it when less than 30 s remain.
func AccessToken(ctx context.Context) (string, *Session, error) {
	s, err := Current(ctx)
	if err != nil {
		return "", nil, err
	}
	if _, _, ok := serviceCredentials(); ok {
		return s.AccessToken, s, nil
	}
	if time.Until(s.AccessExpiry) < 30*time.Second {
		if err := s.Refresh(ctx); err != nil {
			if IsRejected(err) {
				_ = Remove()
				return "", nil, ErrExpired
			}
			return "", nil, err
		}
	}
	return s.AccessToken, s, nil
}
