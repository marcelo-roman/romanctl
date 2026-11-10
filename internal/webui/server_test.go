package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func testServer(t *testing.T, exe string) (*server, http.Handler) {
	t.Helper()
	s, err := newServer(7777, "9.9.9", exe)
	if err != nil {
		t.Fatal(err)
	}
	return s, s.routes()
}

func call(h http.Handler, method, target, host, token, body string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Host = host
	if token != "" {
		req.Header.Set(tokenHeader, token)
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSoAceitaEsteComputadorEOToken(t *testing.T) {
	s, h := testServer(t, "/bin/true")
	if rec := call(h, "GET", "/", "evil.example:7777", "", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("Host de outro site: %d", rec.Code)
	}
	if rec := call(h, "GET", "/api/info", "localhost:7777", "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("sem token: %d", rec.Code)
	}
	if rec := call(h, "GET", "/api/info", "localhost:7777", "errado", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("token errado: %d", rec.Code)
	}
	if rec := call(h, "POST", "/api/jobs", "localhost:7777", s.token, `{"action":"templates.update"}`, "Origin", "http://evil.example"); rec.Code != http.StatusForbidden {
		t.Fatalf("Origin de outro site: %d", rec.Code)
	}
	rec := call(h, "GET", "/api/info", "127.0.0.1:7777", s.token, "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"version":"9.9.9"`) {
		t.Fatalf("info: %d %s", rec.Code, rec.Body)
	}
}

func TestAPaginaLevaOTokenEAsRotasDoAppVoltamParaEla(t *testing.T) {
	s, h := testServer(t, "/bin/true")
	for _, target := range []string{"/", "/projects/loja"} {
		rec := call(h, "GET", target, "localhost:7777", "", "")
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `content="`+s.token+`"`) {
			t.Fatalf("%s: %d sem o token na página", target, rec.Code)
		}
	}
}

func TestAcoesSoComValoresSeguros(t *testing.T) {
	cases := []ActionRequest{
		{Action: "project.up", Name: "-rf"},
		{Action: "project.remove", Name: "../x"},
		{Action: "service.start", Key: "--all"},
		{Action: "project.generate", Name: "loja", Kind: "resource", Entity: "--force"},
		{Action: "project.create", Name: "loja", Stack: "go"},
		{Action: "tool.install", Key: "jq", Version: "1.7; rm -rf /"},
	}
	for _, c := range cases {
		if _, err := actions[c.Action].args(c); err == nil {
			t.Errorf("%+v deveria ser recusada", c)
		}
	}
	args, err := actions["project.create"].args(ActionRequest{Name: "loja", Stack: "go", Arch: []string{"hexagonal", "ddd"}, With: []string{"postgres"}})
	if err != nil || !slices.Equal(args, []string{"new", "loja", "--stack", "go", "--arch", "hexagonal,ddd", "--with", "postgres"}) {
		t.Fatalf("%v %v", args, err)
	}
	_, h := testServer(t, "/bin/true")
	if rec := call(h, "POST", "/api/jobs", "localhost:7777", "", `{"action":"rm"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("sem token: %d", rec.Code)
	}
}

func TestTarefaRodaORomanctlECapturaASaida(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "romanctl")
	script := "#!/bin/sh\necho \"args: $*\" >&2\necho 'progresso' >&2\nif [ \"$1\" = up ]; then echo '{\"ok\": false, \"error\": \"porta ocupada\"}'; exit 1; fi\necho '{\"ok\": true}'\n"
	if err := os.WriteFile(exe, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	s, h := testServer(t, exe)
	rec := call(h, "POST", "/api/jobs", "localhost:7777", s.token, `{"action":"templates.update"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	var started Job
	json.Unmarshal(rec.Body.Bytes(), &started)
	done := waitJob(t, s, started.ID)
	if done.Status != jobSucceeded || !slices.Contains(done.Lines, "args: templates update --json") || string(done.Result) != `{"ok": true}` {
		t.Fatalf("%+v", done)
	}
	rec = call(h, "POST", "/api/jobs", "localhost:7777", s.token, `{"action":"project.up","name":"loja"}`)
	json.Unmarshal(rec.Body.Bytes(), &started)
	failed := waitJob(t, s, started.ID)
	if failed.Status != jobFailed || failed.Error != "porta ocupada" {
		t.Fatalf("%+v", failed)
	}
}

func waitJob(t *testing.T, s *server, id string) Job {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if j, ok := s.jobs.get(id); ok && j.Status != jobRunning {
			return j
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("a tarefa não terminou")
	return Job{}
}
