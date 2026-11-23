package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/services"
	"romanctl/internal/state"
	"romanctl/internal/sys"
	"romanctl/internal/sysinfo"
	"romanctl/internal/tools"
)

// Row is a table row. Status becomes a color; Detail holds what the details view shows.
type Row struct {
	Key     string
	Columns []string
	Status  string
	Group   string
	Detail  [][2]string
}

// Backend isolates the panel from the system, so tests can drive the screen without side effects.
type Backend interface {
	Services() []Row
	Projects() []Row
	Tools() []Row
	System() sysinfo.Snapshot
	Logs(view, key string) string
	Command(args ...string) *exec.Cmd
	Header() string
	Upstream(key string) tools.Upstream
	Latest(key string) tools.Upstream
	Versions(key string) ([]tools.Version, error)
}

type realBackend struct{ self string }

// latestSlots limits concurrent queries: the whole list is about 170 network round trips.
var latestSlots = make(chan struct{}, 8)

func (b *realBackend) Latest(key string) tools.Upstream {
	t := tools.Find(key)
	if t == nil {
		return tools.Upstream{}
	}
	if u, ok := tools.CachedLatest(key); ok {
		return u
	}
	latestSlots <- struct{}{}
	defer func() { <-latestSlots }()
	return t.Latest()
}

func (b *realBackend) Versions(key string) ([]tools.Version, error) {
	t := tools.Find(key)
	if t == nil {
		return nil, nil
	}
	return t.Versions()
}

func newRealBackend() *realBackend {
	self, err := os.Executable()
	if err != nil {
		self = "romanctl"
	}
	return &realBackend{self: self}
}

func (b *realBackend) Header() string { return platform.Current.Describe() }

func (b *realBackend) Services() []Row {
	rows := []Row{}
	for _, s := range services.All {
		// Only what exists here: a database that was never pulled is not a service of this machine.
		if s.Unsupported != "" || !s.Installed() && !state.WantsRunning(s.Key) {
			continue
		}
		group := ""
		if s.Engine {
			group = engineGroup
		}
		status := "stopped"
		switch {
		case !s.Installed():
			status = "not-installed"
		case s.Running():
			status = "running"
		case state.WantsRunning(s.Key):
			status = "crashed"
		}
		rows = append(rows, Row{Key: s.Key, Status: status, Group: group,
			Columns: []string{s.Key, s.Name, statusText(status), s.URL()},
			Detail:  [][2]string{{i18n.T("Service"), s.Name}, {i18n.T("Key"), s.Key}, {i18n.T("State"), statusText(status)}, {i18n.T("Address"), s.URL()}, {i18n.T("Log"), s.LogFile}}})
	}
	return rows
}

func (b *realBackend) Projects() []Row {
	rows := []Row{}
	ups := state.ReadState().Ups
	for _, p := range projects.ListProjects() {
		status := "stopped"
		if e, ok := ups[p.Name]; ok {
			switch {
			case e.Alive():
				status = "running"
			case e.PID != 0:
				status = "crashed"
			default:
				status = "prepared"
			}
		}
		url := projects.AppURL(p.Path)
		rows = append(rows, Row{Key: p.Name, Status: status,
			Columns: []string{p.Name, p.Type, statusText(status), url, p.Mod.Format("02/01/06 15:04")},
			Detail: [][2]string{{i18n.T("Project"), p.Name}, {i18n.T("Type"), p.Type}, {i18n.T("State"), statusText(status)}, {i18n.T("Folder"), p.Path},
				{i18n.T("Address"), url}, {i18n.T("Generators"), generators(p.Path)}}})
	}
	return rows
}

func generators(path string) string {
	target := scaffold.TargetOf(path)
	if target == "" {
		return "-"
	}
	var names []string
	for _, g := range scaffold.LoadGenerators(target) {
		names = append(names, g.Kind)
	}
	return strings.Join(names, ", ")
}

func (b *realBackend) Tools() []Row {
	rows := []Row{}
	for _, s := range tools.Statuses(nil) {
		if s.Unavailable != "" && !s.Installed {
			continue
		}
		status, version := "available", ""
		if s.Installed {
			status, version = "installed", s.Version
		}
		rows = append(rows, Row{Key: s.Key, Status: status, Group: s.Group,
			Columns: []string{s.Key, s.Group, statusText(status), version, "", s.Description},
			Detail: [][2]string{{i18n.T("Tool"), s.Name}, {i18n.T("Category"), tools.GroupTitle(s.Group)}, {i18n.T("State"), statusText(status)},
				{i18n.T("Version"), version}, {i18n.T("Description"), s.Description}, {i18n.T("When removing"), warningOf(s.Key)}}})
	}
	return rows
}

func warningOf(key string) string {
	if t := tools.Find(key); t != nil && t.Warning != "" {
		return t.Warning
	}
	return "-"
}

func (b *realBackend) Upstream(key string) tools.Upstream {
	if t := tools.Find(key); t != nil {
		return t.Upstream(true)
	}
	return tools.Upstream{}
}

func (b *realBackend) System() sysinfo.Snapshot { return sysinfo.Take(time.Second) }

// Logs reads only the tail: the panel refreshes it from time to time.
func (b *realBackend) Logs(view, key string) string {
	var cmd []string
	switch view {
	case viewServices:
		s := services.FindService(key)
		if s == nil {
			return ""
		}
		switch {
		case s.LogTail != nil:
			cmd = s.LogTail()
		case s.LogFile == "":
			return i18n.T("This service has no log to show here.")
		default:
			cmd = []string{"tail", "-n", "300", s.LogFile}
		}
	case viewProjects:
		e, ok := state.ReadState().Ups[key]
		if !ok || e.Log == "" {
			return i18n.T("This project is not in the background. Bring it up with the u key to have logs here.")
		}
		cmd = []string{"tail", "-n", "300", e.Log}
	default:
		return ""
	}
	out, err := sys.Output(cmd[0], cmd[1:]...)
	if err != nil && out == "" {
		return fmt.Sprint(err)
	}
	return out
}

// Command runs romanctl itself in another process: its output does not cross the panel.
func (b *realBackend) Command(args ...string) *exec.Cmd {
	cmd := exec.Command(b.self, args...)
	cmd.Dir = filepath.Dir(b.self)
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}
	return cmd
}
