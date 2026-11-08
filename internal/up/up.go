// Package up sobe e derruba o ambiente completo de um projeto.
package up

import (
	"net"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"romanctl/internal/browser"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/projects"
	"romanctl/internal/services"
	"romanctl/internal/state"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

// Options são as escolhas do comando up.
type Options struct {
	Detach, NoRun, Open bool
}

// Result descreve o ambiente que ficou no ar.
type Result struct {
	OK       bool     `json:"ok"`
	Project  string   `json:"project"`
	Path     string   `json:"path"`
	Services []string `json:"services"`
	Started  []string `json:"startedNow"`
	Database string   `json:"database,omitempty"`
	EnvAdded []string `json:"envAdded"`
	URL      string   `json:"url,omitempty"`
	PID      int      `json:"pid,omitempty"`
	Log      string   `json:"log,omitempty"`
	Already  bool     `json:"alreadyUp"`
}

func logPath(name string) string { return filepath.Join(config.LogDir, "projeto-"+name+".log") }

// DBName é o nome do banco do projeto: o mesmo que os templates usam.
func DBName(name string) string {
	return strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToLower(name))
}

func portOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Port()
}

func portBusy(port string) bool {
	c, err := net.DialTimeout("tcp", "127.0.0.1:"+port, time.Second)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

// ownerOfPort diz qual projeto no ar está usando a porta, para o erro ajudar de verdade.
func ownerOfPort(port, self string) string {
	for name, e := range state.ReadState().Ups {
		if name != self && portOf(e.URL) == port && e.Alive() {
			return name
		}
	}
	return ""
}

func checkPort(rawURL, self string) error {
	port := portOf(rawURL)
	if port == "" || !portBusy(port) {
		return nil
	}
	if owner := ownerOfPort(port, self); owner != "" {
		return i18n.Errorf("a porta %s já está em uso pelo projeto %s; derrube com: romanctl down %s", port, owner, owner)
	}
	return i18n.Errorf("a porta %s já está em uso por outro programa", port)
}

// serviceEnv junta as variáveis dos serviços do projeto.
func serviceEnv(m *manifest.Project, db string) []KV {
	var vars []KV
	for _, key := range m.Services {
		if s := services.FindService(key); s != nil {
			for _, kv := range s.EnvVars(db) {
				vars = append(vars, KV{kv[0], kv[1]})
			}
		}
	}
	// O ASP.NET lê a connection string desta variável, por cima do appsettings.
	if m.Stack == "dotnet" && slices.Contains(m.Services, "postgres") {
		vars = append(vars, KV{"ConnectionStrings__Default", "Host=127.0.0.1;Port=5432;Database=" + db + ";Username=postgres;Password=postgres"})
	}
	return vars
}

func ensureService(key string) (*services.Service, error) {
	s := services.FindService(key)
	if s == nil {
		return nil, i18n.Errorf("o projeto usa um serviço desconhecido: %s", key)
	}
	if s.Unsupported != "" {
		return nil, i18n.Errorf("%s: %s", s.Name, s.Unsupported)
	}
	if !s.Installed() {
		if !ui.Interactive() {
			return nil, i18n.Errorf("%s não está instalado — rode: romanctl install %s", s.Name, s.Key)
		}
		ok, err := ui.Confirm(i18n.Sprintf("%s não está instalado. Instalar agora?", s.Name))
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, i18n.Errorf("%s não está instalado — rode: romanctl install %s", s.Name, s.Key)
		}
		if err := s.Install(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Run sobe o ambiente. Em primeiro plano, só volta quando o app para.
func Run(name, path string, opts Options) (Result, error) {
	m, hasManifest := manifest.Read(path)
	result := Result{Project: name, Path: path, Services: []string{}, Started: []string{}, EnvAdded: []string{}}
	previous, wasKnown := state.ReadState().Ups[name]
	if wasKnown && previous.PID != 0 && previous.Alive() {
		result.OK, result.Already, result.URL, result.PID, result.Log = true, true, previous.URL, previous.PID, previous.Log
		ui.Okf("%s já está no ar → %s", name, previous.URL)
		return result, nil
	}
	appURL := projects.AppURL(path)
	result.URL = appURL
	if !opts.NoRun && appURL != "" {
		if err := checkPort(appURL, name); err != nil {
			return result, err
		}
	}

	started := []string{}
	if wasKnown {
		started = append(started, previous.Started...)
	}
	if hasManifest {
		result.Services = append(result.Services, m.Services...)
		for _, key := range m.Services {
			s, err := ensureService(key)
			if err != nil {
				return result, err
			}
			wasRunning := s.Running()
			if err := s.Start(); err != nil {
				return result, err
			}
			if !wasRunning {
				started = append(started, s.Key)
				result.Started = append(result.Started, s.Key)
			}
		}
		db := DBName(m.Name)
		for _, key := range m.Services {
			s := services.FindService(key)
			if !s.HasDatabases() {
				continue
			}
			var dbErr error
			ui.Spin(i18n.Sprintf("Criando o banco %s...", db), func(func(string)) { dbErr = s.EnsureDatabase(db) })
			if dbErr != nil {
				return result, i18n.Errorf("banco %s em %s: %w", db, s.Name, dbErr)
			}
			result.Database = db
		}
		if vars := serviceEnv(m, db); len(vars) > 0 {
			added, err := MergeEnv(filepath.Join(path, ".env"), vars)
			if err != nil {
				return result, err
			}
			result.EnvAdded = append(result.EnvAdded, added...)
			if _, err := EnsureIgnored(path); err != nil {
				return result, err
			}
			if len(added) > 0 {
				ui.Okf("Variáveis em .env: %s", strings.Join(added, ", "))
			}
		}
	}

	entry := state.UpEntry{Path: path, URL: appURL, Uses: result.Services, Started: dedupe(started)}
	if opts.NoRun {
		state.SetUp(name, entry)
		result.OK = true
		return result, nil
	}
	env := ProcessEnv(filepath.Join(path, ".env"), os.LookupEnv)
	if !opts.Detach {
		state.SetUp(name, entry)
		err := projects.RunAt(path, name, env, opts.Open)
		if len(entry.Started) > 0 {
			ui.Infof("Os serviços continuam ligados. Para desligar: romanctl down %s", name)
		}
		result.OK = err == nil
		return result, err
	}

	command, err := projects.RunCommand(path, name)
	if err != nil {
		return result, err
	}
	entry.Log = logPath(name)
	pid, err := sys.StartDetachedEnv(entry.Log, path, env, command[0], command[1:]...)
	if err != nil {
		return result, err
	}
	entry.PID = pid
	state.SetUp(name, entry)
	result.PID, result.Log = pid, entry.Log

	var failed error
	ui.Spin(i18n.Sprintf("Subindo %s...", name), func(func(string)) { failed = waitForApp(pid, appURL, entry.Log) })
	if failed != nil {
		sys.StopGroup(pid, 3*time.Second)
		entry.PID = 0
		state.SetUp(name, entry)
		return result, failed
	}
	if opts.Open && appURL != "" {
		browser.OpenURL(appURL)
	}
	result.OK = true
	return result, nil
}

var (
	appTimeout = 3 * time.Minute
	noURLWatch = 8 * time.Second
)

// waitForApp só dá o app por pronto quando ele responde. Sem URL, basta continuar vivo.
func waitForApp(pid int, appURL, log string) error {
	if appURL == "" {
		// Sem endereço, só dá para ver se continua vivo. Compilar pode levar mais que este prazo.
		if sys.WaitUntil(noURLWatch, func() bool { return !sys.GroupAlive(pid) }) {
			return i18n.Errorf("o app parou logo depois de subir:\n%s", tailLog(log))
		}
		return nil
	}
	port := portOf(appURL)
	deadline := time.Now().Add(appTimeout)
	for time.Now().Before(deadline) {
		if !sys.GroupAlive(pid) {
			return i18n.Errorf("o app parou logo depois de subir:\n%s", tailLog(log))
		}
		if sys.HttpUp(appURL) || (port != "" && portBusy(port)) {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return i18n.Errorf("o app não respondeu em %s a tempo; veja: %s", appURL, log)
}

func tailLog(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return sys.Tail(string(b), 12)
}

func dedupe(list []string) []string {
	out := []string{}
	for _, v := range list {
		if !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	return out
}

// DownResult descreve o que foi parado.
type DownResult struct {
	OK              bool     `json:"ok"`
	Project         string   `json:"project"`
	WasUp           bool     `json:"wasUp"`
	AppStopped      bool     `json:"appStopped"`
	ServicesStopped []string `json:"servicesStopped"`
	ServicesKept    []string `json:"servicesKept"`
}

// splitServices separa o que pode ser desligado. Um serviço que o up não ligou nunca é tocado,
// e um que outro projeto usa passa a ser responsabilidade desse projeto.
func splitServices(started []string, others map[string]state.UpEntry, self string) (stop, handOff []string) {
	for _, key := range started {
		inUse := false
		for name, e := range others {
			if name != self && slices.Contains(e.Uses, key) {
				inUse = true
			}
		}
		if inUse {
			handOff = append(handOff, key)
		} else {
			stop = append(stop, key)
		}
	}
	return stop, handOff
}

// Down para o app e os serviços que só este projeto estava usando.
func Down(name string, keepServices bool) (DownResult, error) {
	result := DownResult{Project: name, ServicesStopped: []string{}, ServicesKept: []string{}}
	ups := state.ReadState().Ups
	entry, ok := ups[name]
	if !ok {
		ui.Infof("%s não está no ar", name)
		result.OK = true
		return result, nil
	}
	result.WasUp = true
	if entry.PID != 0 && sys.GroupAlive(entry.PID) {
		ui.Spin(i18n.Sprintf("Parando %s...", name), func(func(string)) { sys.StopGroup(entry.PID, 8*time.Second) })
		result.AppStopped = true
		ui.Okf("%s parado", name)
	}
	stop, handOff := splitServices(entry.Started, ups, name)
	if keepServices {
		handOff, stop = nil, nil
		result.ServicesKept = append(result.ServicesKept, entry.Started...)
	}
	var failed []string
	for _, key := range stop {
		s := services.FindService(key)
		if s == nil {
			continue
		}
		if err := s.Stop(); err != nil {
			ui.Failf("%v", err)
			failed = append(failed, key)
			continue
		}
		result.ServicesStopped = append(result.ServicesStopped, key)
	}
	for _, key := range handOff {
		result.ServicesKept = append(result.ServicesKept, key)
		ui.Infof("%s continua ligado, porque outro projeto usa", key)
	}
	state.TransferStarted(name, handOff)
	if len(failed) > 0 {
		return result, i18n.Errorf("não pararam: %s", strings.Join(failed, ", "))
	}
	result.OK = true
	return result, nil
}

// Names lista os projetos no ar, para o menu do down.
func Names() []string {
	var out []string
	for name := range state.ReadState().Ups {
		out = append(out, name)
	}
	slices.Sort(out)
	return out
}

// Describe resume a linha de um projeto no ar.
func Describe(e state.UpEntry) string {
	if e.PID == 0 {
		return strconv.Itoa(len(e.Uses)) + " serv."
	}
	return e.URL
}
