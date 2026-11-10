package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/ui"
	"slices"
	"sort"
	"syscall"
	"time"
)

// O estado guarda o que o usuário quer ligado. Ele só muda por start/stop, então
// sobrevive quando o Android mata o Termux ou o tablet reinicia.
type RunState struct {
	Services []string           `json:"services"`
	Projects map[string]int     `json:"projects"` // projeto em "run" → PID do romanctl
	Updated  string             `json:"updated"`
	Ups      map[string]UpEntry `json:"ups,omitempty"`
}

// UpEntry é um projeto levantado com romanctl up.
type UpEntry struct {
	Path    string   `json:"path"`
	URL     string   `json:"url,omitempty"`
	URLs    []string `json:"urls,omitempty"` // todas as URLs, num projeto com vários processos
	Log     string   `json:"log,omitempty"`
	PID     int      `json:"pid,omitempty"` // zero: só os serviços estão no ar
	Uses    []string `json:"uses,omitempty"`
	Started []string `json:"started,omitempty"` // serviços que o up ligou, e o down pode desligar
}

// Alive diz se o app segue rodando.
func (e UpEntry) Alive() bool { return e.PID > 0 && syscall.Kill(-e.PID, 0) == nil }

// SetUp registra um projeto no ar.
func SetUp(name string, e UpEntry) {
	err := UpdateState(func(st *RunState) {
		if st.Ups == nil {
			st.Ups = map[string]UpEntry{}
		}
		st.Ups[name] = e
	})
	if err != nil {
		ui.Warnf("Não consegui salvar o estado: %v", err)
	}
}

// TransferStarted tira o projeto do estado e passa os serviços dele para outro que os usa.
func TransferStarted(name string, services []string) {
	err := UpdateState(func(st *RunState) {
		delete(st.Ups, name)
		for _, key := range services {
			for other, e := range st.Ups {
				if slices.Contains(e.Uses, key) {
					if !slices.Contains(e.Started, key) {
						e.Started = append(e.Started, key)
						st.Ups[other] = e
					}
					break
				}
			}
		}
	})
	if err != nil {
		ui.Warnf("Não consegui salvar o estado: %v", err)
	}
}

var statePath = filepath.Join(config.DataDir, "estado.json")

func ReadState() RunState {
	var st RunState
	if b, err := os.ReadFile(statePath); err == nil {
		if err := json.Unmarshal(b, &st); err != nil {
			ui.Warnf("Estado ilegível em %s: %v", statePath, err)
		}
	}
	return st
}

// updateState altera o estado com trava, porque várias sessões podem usar o romanctl.
func UpdateState(f func(*RunState)) error {
	if err := os.MkdirAll(config.DataDir, 0o755); err != nil {
		return err
	}
	lock, err := os.OpenFile(statePath+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	st := ReadState()
	f(&st)
	st.Updated = time.Now().Format("02/01/2006 15:04")
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := statePath + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, statePath)
}

func MarkService(key string, on bool) {
	err := UpdateState(func(st *RunState) {
		st.Services = slices.DeleteFunc(st.Services, func(k string) bool { return k == key })
		if on {
			st.Services = append(st.Services, key)
			sort.Strings(st.Services)
		}
	})
	if err != nil {
		ui.Warnf("Não consegui salvar o estado: %v", err)
	}
}

func MarkProject(name string, on bool) {
	err := UpdateState(func(st *RunState) {
		if st.Projects == nil {
			st.Projects = map[string]int{}
		}
		delete(st.Projects, name)
		if on {
			st.Projects[name] = os.Getpid()
		}
	})
	if err != nil {
		ui.Warnf("Não consegui salvar o estado: %v", err)
	}
}

func WantsRunning(key string) bool { return slices.Contains(ReadState().Services, key) }
