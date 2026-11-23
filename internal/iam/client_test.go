package iam

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func isolate(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "credentials.env")
	t.Setenv("IAM_CREDENTIALS_FILE", path)
	for _, k := range []string{envAPI, envTokenURL, envClientID, envSecret} {
		t.Setenv(k, "")
	}
	cache.Lock()
	cache.key = ""
	cache.Unlock()
	return path
}

func TestNoCredentials(t *testing.T) {
	isolate(t)
	if _, err := LoadCredentials(); !errors.Is(err, ErrNoCredentials) {
		t.Fatalf("no file and no environment: %v", err)
	}
}

func TestSavesWith0600AndReadsBack(t *testing.T) {
	path := isolate(t)
	want := Credentials{API: "https://api.test", TokenURL: "https://kc.test/token", ClientID: "iam-api-x", ClientSecret: "s3cr=t"}
	if err := SaveCredentials(want); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("permission %v (%v)", info.Mode().Perm(), err)
	}
	got, err := LoadCredentials()
	if err != nil {
		t.Fatal(err)
	}
	want.Source = path
	if got != want {
		t.Fatalf("read %+v, want %+v", got, want)
	}
	if SaveCredentials(Credentials{ClientID: "a", ClientSecret: "b\nIAM_API_URL=evil"}) == nil {
		t.Fatal("a newline in the secret must be rejected")
	}
}

func TestEnvironmentOverridesFile(t *testing.T) {
	isolate(t)
	if err := SaveCredentials(Credentials{API: "https://file", TokenURL: "https://file/token", ClientID: "file-id", ClientSecret: "file-secret"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envClientID, "env-id")
	t.Setenv(envSecret, "env-secret")
	got, err := LoadCredentials()
	if err != nil {
		t.Fatal(err)
	}
	if got.ClientID != "env-id" || got.ClientSecret != "env-secret" || got.API != "https://file" || got.Source != "env" {
		t.Fatalf("%+v", got)
	}
}

type fakeIAM struct {
	tokens   atomic.Int32
	rejectAt int32
	server   *httptest.Server
}

func newFakeIAM(t *testing.T) *fakeIAM {
	f := &fakeIAM{rejectAt: -1}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("client_secret") != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"unauthorized_client","error_description":"Invalid client credentials"}`)
			return
		}
		n := f.tokens.Add(1)
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{"exp":9999999999}`))
		fmt.Fprintf(w, `{"access_token":"h.%s.%d","expires_in":300}`, payload, n)
	})
	mux.HandleFunc("GET /api/me", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if strings.HasSuffix(token, fmt.Sprintf(".%d", f.rejectAt)) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"token expired"}`)
			return
		}
		fmt.Fprint(w, `{"username":"service-account-iam-api-x"}`)
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeIAM) credentials(secret string) Credentials {
	return Credentials{API: f.server.URL, TokenURL: f.server.URL + "/token", ClientID: "iam-api-x", ClientSecret: secret}
}

func TestReusesTokenAcrossCalls(t *testing.T) {
	isolate(t)
	f := newFakeIAM(t)
	c := f.credentials("secret")
	for range 3 {
		if _, err := With(context.Background(), c, "GET", "/api/me", nil); err != nil {
			t.Fatal(err)
		}
	}
	if n := f.tokens.Load(); n != 1 {
		t.Fatalf("requested %d tokens; one is enough while valid", n)
	}
}

func TestOn401RequestsNewTokenOnce(t *testing.T) {
	isolate(t)
	f := newFakeIAM(t)
	f.rejectAt = 1
	raw, err := With(context.Background(), f.credentials("secret"), "GET", "/api/me", nil)
	if err != nil || !strings.Contains(string(raw), "service-account") {
		t.Fatalf("%s %v", raw, err)
	}
	if n := f.tokens.Load(); n != 2 {
		t.Fatalf("requested %d tokens; want 2", n)
	}
}

func TestWrongSecretNeverReachesAPI(t *testing.T) {
	isolate(t)
	f := newFakeIAM(t)
	_, err := With(context.Background(), f.credentials("wrong"), "GET", "/api/me", nil)
	if err == nil || !strings.Contains(err.Error(), "iam-api-x") || strings.Contains(err.Error(), "wrong") {
		t.Fatalf("the error should name the client, never the secret: %v", err)
	}
}
