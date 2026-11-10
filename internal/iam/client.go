// Package iam chama a API do IAM Manager com o token do login do romanctl.
package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"romanctl/internal/auth"
	"romanctl/internal/i18n"
)

var httpClient = &http.Client{Timeout: 3 * time.Minute}

// Error é a resposta de erro da API ({"error": "..."}) com o status HTTP.
type Error struct {
	Status  int
	Message string
}

func (e Error) Error() string {
	switch e.Status {
	case http.StatusUnauthorized:
		return i18n.Sprintf("a API recusou o token (%s); faça login de novo: romanctl login", e.Message)
	case http.StatusForbidden:
		return i18n.Sprintf("sem permissão: %s", e.Message)
	}
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message)
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
	token, session, err := auth.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	switch b := body.(type) {
	case nil:
	case json.RawMessage:
		reader = bytes.NewReader(b)
	default:
		data, err := json.Marshal(b)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(session.API, "/")+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if reader != nil {
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
		return nil, Error{Status: resp.StatusCode, Message: e.Error}
	}
	return data, nil
}
