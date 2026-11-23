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

// The state stores what the user wants running. It only changes through start/stop, so
// it survives Android killing Termux or the tablet rebooting.
type RunState struct {
	Services []string           `json:"services"`
	Projects map[string]int     `json:"projects"` // project in "run" → romanctl PID
	Updated  string             `json:"updated"`
	Ups      map[string]UpEntry `json:"ups,omitempty"`
}

// UpEntry is a project brought up with romanctl up.
type UpEntry struct {
	Path    string   `json:"path"`
	URL     string   `json:"url,omitempty"`
	URLs    []string `json:"urls,omitempty"` // every URL, in a multi-process project
	Log     string   `json:"log,omitempty"`
	PID     int      `json:"pid,omitempty"` // zero: only the services are up
	Uses    []string `json:"uses,omitempty"`
	Started []string `json:"started,omitempty"` // services that up started and down may stop
}

// Alive says whether the app is still running.
func (e UpEntry) Alive() bool { return e.PID > 0 && syscall.Kill(-e.PID, 0) == nil }

// SetUp records a project as up.
func SetUp(name string, e UpEntry) {
	err := UpdateState(func(st *RunState) {
		if st.Ups == nil {
			st.Ups = map[string]UpEntry{}
		}
		st.Ups[name] = e
	})
	if err != nil {
		ui.Warnf("Could not save the state: %v", err)
	}
}

// TransferStarted removes the project from the state and hands its services to another project that uses them.
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
		ui.Warnf("Could not save the state: %v", err)
	}
}

var statePath = filepath.Join(config.DataDir, "estado.json")

func ReadState() RunState {
	var st RunState
	if b, err := os.ReadFile(statePath); err == nil {
		if err := json.Unmarshal(b, &st); err != nil {
			ui.Warnf("Unreadable state in %s: %v", statePath, err)
		}
	}
	return st
}

// updateState changes the state under a lock, because several sessions may use romanctl.
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
		ui.Warnf("Could not save the state: %v", err)
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
		ui.Warnf("Could not save the state: %v", err)
	}
}

func WantsRunning(key string) bool { return slices.Contains(ReadState().Services, key) }
