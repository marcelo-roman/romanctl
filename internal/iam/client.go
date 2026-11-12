// Package iam chama a API do IAM Manager com um cliente da API (client_credentials).
package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"romanctl/internal/auth"
	"romanctl/internal/i18n"
)

var httpClient = &http.Client{Timeout: 3 * time.Minute}

// Error é a resposta de erro da API ({"error": "..."}) com o status HTTP.
type Error struct {
	Status   int
	Message  string
	ClientID string
}

func (e Error) Error() string {
	switch e.Status {
	case http.StatusUnauthorized:
		return i18n.Sprintf("o IAM Manager recusou o token do cliente %s (%s); confira com: romanctl iam credentials", e.ClientID, e.Message)
	case http.StatusForbidden:
		return i18n.Sprintf("sem permissão: %s", e.Message)
	}
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message)
}

var cache struct {
	sync.Mutex
	key   string
	token auth.Token
}

// Token devolve um access token do cliente, reaproveitado até faltar menos de 30 s para expirar.
func Token(ctx context.Context, c Credentials, renew bool) (string, error) {
	cache.Lock()
	defer cache.Unlock()
	key := c.TokenURL + "\x00" + c.ClientID + "\x00" + c.ClientSecret
	if !renew && cache.key == key && time.Until(cache.token.Expiry) > 30*time.Second {
		return cache.token.AccessToken, nil
	}
	t, err := auth.ClientCredentials(ctx, c.TokenURL, c.ClientID, c.ClientSecret)
	if err != nil {
		cache.key = ""
		if auth.IsRejected(err) {
			return "", i18n.Errorf("o Keycloak recusou o cliente %s (%v); confira com: romanctl iam credentials", c.ClientID, err)
		}
		return "", err
	}
	cache.key, cache.token = key, t
	return t.AccessToken, nil
}

func credentials() (Credentials, error) {
	c, err := LoadCredentials()
	if errors.Is(err, ErrNoCredentials) {
		return c, i18n.Errorf("sem cliente da API do IAM Manager. Defina IAM_CLIENT_ID e IAM_CLIENT_SECRET ou rode: romanctl iam credentials set")
	}
	return c, err
}

// Do faz a chamada e decodifica o JSON em out (se não for nil). path começa em /api.
func Do(ctx context.Context, method, path string, body, out any) error {
	raw, err := Raw(ctx, method, path, body)
	if err != nil {
		return err
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

// Raw devolve o corpo da resposta sem interpretar; serve ao "romanctl iam api".
func Raw(ctx context.Context, method, path string, body any) ([]byte, error) {
	c, err := credentials()
	if err != nil {
		return nil, err
	}
	return With(ctx, c, method, path, body)
}

// With chama a API com uma credencial explícita. Em 401 pede outro token uma vez.
func With(ctx context.Context, c Credentials, method, path string, body any) ([]byte, error) {
	payload, err := encode(body)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	data, err := send(ctx, c, method, path, payload, false)
	var apiErr Error
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
		return send(ctx, c, method, path, payload, true)
	}
	return data, err
}

func encode(body any) ([]byte, error) {
	switch b := body.(type) {
	case nil:
		return nil, nil
	case json.RawMessage:
		return b, nil
	default:
		return json.Marshal(b)
	}
}

func send(ctx context.Context, c Credentials, method, path string, payload []byte, renew bool) ([]byte, error) {
	token, err := Token(ctx, c, renew)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.API+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(data, &e)
		if e.Error == "" {
			e.Error = strings.TrimSpace(string(data))
		}
		return nil, Error{Status: resp.StatusCode, Message: e.Error, ClientID: c.ClientID}
	}
	return data, nil
}
