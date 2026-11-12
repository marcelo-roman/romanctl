package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func jwt(claims map[string]any) string {
	body, _ := json.Marshal(claims)
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"none"}`)) + "." + enc(body) + ".sig"
}

// fakeKeycloak imita o que o romanctl usa do Keycloak: auth (já logado), token e revoke.
type fakeKeycloak struct {
	mu        sync.Mutex
	server    *httptest.Server
	challenge string
	revoked   bool
	refreshes int
	reject    bool
}

func newFakeKeycloak(t *testing.T) *fakeKeycloak {
	k := &fakeKeycloak{}
	mux := http.NewServeMux()
	mux.HandleFunc("/realms/private/protocol/openid-connect/auth", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("code_challenge_method") != "S256" || q.Get("client_id") != "romanctl" || !strings.Contains(q.Get("scope"), "offline_access") {
			http.Error(w, "bad request", 400)
			return
		}
		k.mu.Lock()
		k.challenge = q.Get("code_challenge")
		k.mu.Unlock()
		http.Redirect(w, r, q.Get("redirect_uri")+"?code=the-code&state="+url.QueryEscape(q.Get("state")), http.StatusFound)
	})
	mux.HandleFunc("/realms/private/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		k.mu.Lock()
		defer k.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		access := jwt(map[string]any{"preferred_username": "ana", "name": "Ana", "exp": time.Now().Add(5 * time.Minute).Unix(),
			"resource_access": map[string]any{"iam-manager-api": map[string]any{"roles": []string{"iam-admin", "public:viewer"}}}})
		refresh := jwt(map[string]any{"typ": "Offline", "exp": time.Now().Add(30 * 24 * time.Hour).Unix()})
		switch r.Form.Get("grant_type") {
		case "authorization_code":
			if challenge(r.Form.Get("code_verifier")) != k.challenge || r.Form.Get("code") != "the-code" {
				w.WriteHeader(400)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
				return
			}
		case "refresh_token":
			k.refreshes++
			if k.reject {
				w.WriteHeader(400)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant", "error_description": "Session not active"})
				return
			}
		case "client_credentials":
			if r.Form.Get("client_secret") != "s3cret" {
				w.WriteHeader(401)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized_client"})
				return
			}
			refresh = ""
		}
		json.NewEncoder(w).Encode(map[string]any{"access_token": access, "refresh_token": refresh, "expires_in": 300})
	})
	mux.HandleFunc("/realms/private/protocol/openid-connect/revoke", func(w http.ResponseWriter, r *http.Request) {
		k.mu.Lock()
		k.revoked = true
		k.mu.Unlock()
	})
	k.server = httptest.NewServer(mux)
	t.Cleanup(k.server.Close)
	return k
}

func (k *fakeKeycloak) endpoints() Endpoints {
	return Endpoints{Issuer: k.server.URL + "/realms/private", ClientID: "romanctl", API: "https://api.test"}
}

func isolate(t *testing.T) {
	t.Setenv("ROMANCTL_SESSION", filepath.Join(t.TempDir(), "session.json"))
	t.Setenv("ROMANCTL_CLIENT_ID", "")
	t.Setenv("ROMANCTL_CLIENT_SECRET", "")
}

func browse(t *testing.T) func(string) {
	return func(u string) {
		go func() {
			resp, err := http.Get(u)
			if err != nil {
				t.Error(err)
				return
			}
			resp.Body.Close()
		}()
	}
}

func login(t *testing.T, k *fakeKeycloak) *Session {
	t.Helper()
	s, err := Login(context.Background(), LoginOptions{Endpoints: k.endpoints(), Open: browse(t), Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(s); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestLoginComPKCEGuardaASessaoSoParaODono(t *testing.T) {
	isolate(t)
	k := newFakeKeycloak(t)
	s := login(t, k)
	if s.Username != "ana" || s.RefreshToken == "" || s.RefreshExpiry.IsZero() || len(s.Roles) != 2 {
		t.Fatalf("%+v", s)
	}
	info, err := os.Stat(SessionPath())
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("permissão da sessão: %v %v", info.Mode(), err)
	}
	if err := Require(context.Background()); err != nil {
		t.Fatalf("sessão nova deveria passar: %v", err)
	}
}

func TestSemSessaoACLINaoRoda(t *testing.T) {
	isolate(t)
	if err := Require(context.Background()); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("%v", err)
	}
}

func TestSessaoAntigaEConferidaERevogadaSai(t *testing.T) {
	isolate(t)
	k := newFakeKeycloak(t)
	s := login(t, k)
	s.ValidatedAt = time.Now().Add(-13 * time.Hour)
	Save(s)
	if err := Require(context.Background()); err != nil || k.refreshes != 1 {
		t.Fatalf("deveria renovar: %v (%d)", err, k.refreshes)
	}
	s, _ = Load()
	s.ValidatedAt = time.Now().Add(-13 * time.Hour)
	Save(s)
	k.reject = true
	if err := Require(context.Background()); !errors.Is(err, ErrExpired) {
		t.Fatalf("sessão revogada deveria sair: %v", err)
	}
	if _, err := Load(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatal("a sessão revogada deveria ser apagada")
	}
}

func TestSemRedeASessaoSalvaBasta(t *testing.T) {
	isolate(t)
	k := newFakeKeycloak(t)
	s := login(t, k)
	s.ValidatedAt = time.Now().Add(-48 * time.Hour)
	s.Issuer = "http://127.0.0.1:1/realms/private"
	Save(s)
	if err := Require(context.Background()); err != nil {
		t.Fatalf("sem rede, a sessão salva deveria bastar: %v", err)
	}
	s.RefreshExpiry = time.Now().Add(-time.Minute)
	Save(s)
	if err := Require(context.Background()); !errors.Is(err, ErrExpired) {
		t.Fatalf("refresh vencido deveria pedir login: %v", err)
	}
}

func TestAccessTokenRenovaPertoDeVencerELogoutRevoga(t *testing.T) {
	isolate(t)
	k := newFakeKeycloak(t)
	s := login(t, k)
	s.AccessExpiry = time.Now().Add(10 * time.Second)
	Save(s)
	token, _, err := AccessToken(context.Background())
	if err != nil || token == "" || k.refreshes != 1 {
		t.Fatalf("%v %d", err, k.refreshes)
	}
	if err := s.Revoke(context.Background()); err != nil || !k.revoked {
		t.Fatalf("revoke: %v", err)
	}
}

func TestClientDeServicoSubstituiOLogin(t *testing.T) {
	isolate(t)
	k := newFakeKeycloak(t)
	t.Setenv("ROMANCTL_AUTH_ISSUER", k.endpoints().Issuer)
	t.Setenv("ROMANCTL_CLIENT_ID", "iam-api-ci")
	t.Setenv("ROMANCTL_CLIENT_SECRET", "errado")
	service.session = nil
	if err := Require(context.Background()); err == nil {
		t.Fatal("secret errado deveria falhar")
	}
	t.Setenv("ROMANCTL_CLIENT_SECRET", "s3cret")
	if err := Require(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(SessionPath()); !os.IsNotExist(err) {
		t.Fatal("o client de serviço não grava sessão em disco")
	}
}

func TestStateDiferenteEhRecusado(t *testing.T) {
	isolate(t)
	k := newFakeKeycloak(t)
	_, err := Login(context.Background(), LoginOptions{Endpoints: k.endpoints(), Timeout: 5 * time.Second, Open: func(u string) {
		parsed, _ := url.Parse(u)
		redirect := parsed.Query().Get("redirect_uri")
		go http.Get(redirect + "?code=the-code&state=forjado")
	}})
	if err == nil || !strings.Contains(err.Error(), "state") {
		t.Fatalf("%v", err)
	}
}
