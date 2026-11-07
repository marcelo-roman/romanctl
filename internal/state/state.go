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
	Services []string       `json:"services"`
	Projects map[string]int `json:"projects"` // projeto em "run" → PID do romanctl
	Updated  string         `json:"updated"`
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
