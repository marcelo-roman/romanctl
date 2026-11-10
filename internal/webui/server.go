// Package webui serve a interface web do romanctl, só em 127.0.0.1.
package webui

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"romanctl/internal/browser"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/services"
	"romanctl/internal/state"
	"romanctl/internal/sys"
	"romanctl/internal/sysinfo"
	"romanctl/internal/tools"
	"romanctl/internal/ui"
)

//go:embed all:dist
var dist embed.FS

// DefaultPort é a porta da interface quando --port não é informado.
const DefaultPort = 7777

const tokenHeader = "X-Romanctl-Token"

// Options são as escolhas do comando web-ui.
type Options struct {
	Port    int
	Open    bool
	Version string
}

type server struct {
	port    int
	token   string
	version string
	jobs    *jobStore
	assets  fs.FS
	index   []byte
}

// Serve sobe a interface e só volta com Ctrl+C ou SIGTERM.
func Serve(opts Options) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(opts.Port))
	if err != nil {
		return i18n.Errorf("não consegui usar a porta %d: %v (escolha outra com --port)", opts.Port, err)
	}
	s, err := newServer(opts.Port, opts.Version, exe)
	if err != nil {
		listener.Close()
		return err
	}
	srv := &http.Server{Handler: s.routes(), ReadHeaderTimeout: 10 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdown)
	}()
	url := "http://localhost:" + strconv.Itoa(opts.Port)
	ui.Okf("Interface web em %s  (Ctrl+C para parar)", url)
	if opts.Open {
		browser.OpenURL(url)
	}
	if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func newServer(port int, version, exe string) (*server, error) {
	assets, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, err
	}
	index, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		return nil, err
	}
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(b)
	meta := []byte(`<meta name="romanctl-token" content="` + token + `">`)
	index = bytes.Replace(index, []byte("</head>"), append(meta, []byte("</head>")...), 1)
	return &server{port: port, token: token, version: version, jobs: newJobStore(exe), assets: assets, index: index}, nil
}

func (s *server) routes() http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/info", s.info)
	api.HandleFunc("GET /api/system", s.system)
	api.HandleFunc("GET /api/projects", s.projectList)
	api.HandleFunc("GET /api/projects/{name}/logs", s.projectLogs)
	api.HandleFunc("GET /api/catalog", s.catalog)
	api.HandleFunc("GET /api/services", s.serviceList)
	api.HandleFunc("GET /api/services/{key}/logs", s.serviceLogs)
	api.HandleFunc("GET /api/tools", s.toolList)
	api.HandleFunc("GET /api/tools/{key}/versions", s.toolVersions)
	api.HandleFunc("GET /api/jobs", s.jobList)
	api.HandleFunc("GET /api/jobs/{id}", s.jobGet)
	api.HandleFunc("POST /api/jobs", s.jobStart)

	mux := http.NewServeMux()
	mux.Handle("/api/", s.requireToken(api))
	mux.HandleFunc("/", s.static)
	return s.localOnly(mux)
}

// localOnly recusa um Host que não seja este computador: barra DNS rebinding, em que outro site
// faz o navegador chamar 127.0.0.1 com o nome dele.
func (s *server) localOnly(next http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, h := range []string{"localhost", "127.0.0.1", "[::1]"} {
		allowed[h+":"+strconv.Itoa(s.port)] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowed[r.Host] {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && !allowed[strings.TrimPrefix(origin, "http://")] {
			http.Error(w, "forbidden origin", http.StatusForbidden)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// requireToken exige o token da página: outro site não consegue lê-lo, então não chama a API.
func (s *server) requireToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(tokenHeader) != s.token {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": i18n.T("token ausente ou inválido; recarregue a página")})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (s *server) static(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name != "" && name != "index.html" {
		if f, err := fs.Stat(s.assets, name); err == nil && !f.IsDir() {
			if strings.HasPrefix(name, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.ServeFileFS(w, r, s.assets, name)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")
	w.Write(s.index)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *server) info(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version":     s.version,
		"lang":        i18n.Lang(),
		"platform":    platform.Current,
		"projectsDir": config.ProjDir,
		"templates":   map[string]any{"installed": scaffold.Installed(), "ref": scaffold.InstalledRef(), "path": scaffold.Root()},
	})
}

func (s *server) system(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, sysinfo.Take(500*time.Millisecond))
}

type generatorInfo struct {
	Kind        string   `json:"kind"`
	Description string   `json:"description"`
	Aliases     []string `json:"aliases,omitempty"`
	NeedsIn     bool     `json:"needsIn"`
}

func generatorsFor(target string) []generatorInfo {
	out := []generatorInfo{}
	if target == "" {
		return out
	}
	for _, g := range scaffold.LoadGenerators(target) {
		out = append(out, generatorInfo{Kind: g.Kind, Description: i18n.T(g.Description), Aliases: g.Aliases, NeedsIn: g.NeedsIn})
	}
	return out
}

func (s *server) projectList(w http.ResponseWriter, _ *http.Request) {
	list := projects.ListJSON()
	for _, item := range list {
		p, _ := item["path"].(string)
		target := scaffold.TargetOf(p)
		item["target"] = target
		item["generators"] = generatorsFor(target)
	}
	writeJSON(w, http.StatusOK, list)
}

func linesParam(r *http.Request) int {
	n, err := strconv.Atoi(r.URL.Query().Get("lines"))
	if err != nil || n <= 0 || n > 2000 {
		return 300
	}
	return n
}

func (s *server) projectLogs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	entry, ok := state.ReadState().Ups[name]
	if !ok || entry.Log == "" {
		writeJSON(w, http.StatusOK, map[string]any{"lines": []string{}, "running": false})
		return
	}
	b, err := os.ReadFile(entry.Log)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"lines": []string{}, "running": entry.Alive()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"lines": tail(string(b), linesParam(r)), "running": entry.Alive(), "path": entry.Log})
}

func tail(text string, n int) []string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}

func (s *server) catalog(w http.ResponseWriter, _ *http.Request) {
	c := scaffold.CatalogData()
	type templateView struct {
		*scaffold.Template
		Description string `json:"description"`
	}
	views := make([]templateView, 0, len(c.Templates))
	for _, t := range c.Templates {
		views = append(views, templateView{Template: t, Description: i18n.T(t.Description)})
	}
	generators := map[string][]generatorInfo{}
	for target := range c.Generators {
		generators[target] = generatorsFor(target)
	}
	writeJSON(w, http.StatusOK, map[string]any{"stacks": c.Stacks, "arch": c.Arch, "services": c.Services,
		"templates": views, "simpleTypes": c.Simple, "generators": generators})
}

func (s *server) serviceList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, services.StatusJSON(true))
}

func (s *server) serviceLogs(w http.ResponseWriter, r *http.Request) {
	svc := services.FindService(r.PathValue("key"))
	if svc == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": i18n.T("serviço desconhecido")})
		return
	}
	if svc.LogTail == nil {
		writeJSON(w, http.StatusOK, map[string]any{"lines": []string{}})
		return
	}
	command := svc.LogTail()
	if len(command) == 0 || !sys.Have(command[0]) {
		writeJSON(w, http.StatusOK, map[string]any{"lines": []string{}})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	out, _ := exec.CommandContext(ctx, command[0], command[1:]...).CombinedOutput()
	writeJSON(w, http.StatusOK, map[string]any{"lines": tail(string(out), linesParam(r))})
}

func (s *server) toolList(w http.ResponseWriter, _ *http.Request) {
	groups := make([]map[string]string, 0, len(tools.Groups))
	for _, g := range tools.Groups {
		groups = append(groups, map[string]string{"key": g.Key, "title": i18n.T(g.Title)})
	}
	writeJSON(w, http.StatusOK, map[string]any{"groups": groups, "tools": tools.Statuses(nil)})
}

func (s *server) toolVersions(w http.ResponseWriter, r *http.Request) {
	t := tools.Find(r.PathValue("key"))
	if t == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": i18n.T("ferramenta desconhecida")})
		return
	}
	versions, err := t.Versions()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"versions": []tools.Version{}, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"versions": versions})
}

func (s *server) jobList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.jobs.list())
}

func (s *server) jobGet(w http.ResponseWriter, r *http.Request) {
	j, ok := s.jobs.get(r.PathValue("id"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": i18n.T("tarefa não encontrada")})
		return
	}
	writeJSON(w, http.StatusOK, j)
}

func (s *server) jobStart(w http.ResponseWriter, r *http.Request) {
	var req ActionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": i18n.T("corpo inválido")})
		return
	}
	a, ok := actions[req.Action]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": i18n.Sprintf("ação desconhecida: %s", req.Action)})
		return
	}
	args, err := a.args(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, s.jobs.start(req.Action, a.title(req), args))
}
