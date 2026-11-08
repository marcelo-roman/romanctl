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

// Row é uma linha de tabela. Status vira cor; Data guarda o que as ações precisam.
type Row struct {
	Key     string
	Columns []string
	Status  string
	Group   string
	Detail  [][2]string
}

// Backend isola o painel do sistema, para os testes dirigirem a tela sem efeitos colaterais.
type Backend interface {
	Services() []Row
	Projects() []Row
	Tools() []Row
	System() sysinfo.Snapshot
	Logs(view, key string) string
	Command(args ...string) *exec.Cmd
	Header() string
}

type realBackend struct{ self string }

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
		if s.Unsupported != "" {
			continue
		}
		status := "parado"
		switch {
		case !s.Installed():
			status = "nao-instalado"
		case s.Running():
			status = "rodando"
		case state.WantsRunning(s.Key):
			status = "caiu"
		}
		rows = append(rows, Row{Key: s.Key, Status: status,
			Columns: []string{s.Key, s.Name, statusText(status), s.URL()},
			Detail:  [][2]string{{i18n.T("Serviço"), s.Name}, {i18n.T("Chave"), s.Key}, {i18n.T("Estado"), statusText(status)}, {i18n.T("Endereço"), s.URL()}, {i18n.T("Log"), s.LogFile}}})
	}
	return rows
}

func (b *realBackend) Projects() []Row {
	rows := []Row{}
	ups := state.ReadState().Ups
	for _, p := range projects.ListProjects() {
		status := "parado"
		if e, ok := ups[p.Name]; ok {
			switch {
			case e.Alive():
				status = "rodando"
			case e.PID != 0:
				status = "caiu"
			default:
				status = "preparado"
			}
		}
		url := projects.AppURL(p.Path)
		rows = append(rows, Row{Key: p.Name, Status: status,
			Columns: []string{p.Name, p.Type, statusText(status), url, p.Mod.Format("02/01/06 15:04")},
			Detail: [][2]string{{i18n.T("Projeto"), p.Name}, {i18n.T("Tipo"), p.Type}, {i18n.T("Estado"), statusText(status)}, {i18n.T("Pasta"), p.Path},
				{i18n.T("Endereço"), url}, {i18n.T("Geradores"), generators(p.Path)}}})
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
		status, version := "disponivel", ""
		if s.Installed {
			status, version = "instalado", s.Version
		}
		rows = append(rows, Row{Key: s.Key, Status: status, Group: s.Group,
			Columns: []string{s.Key, s.Group, statusText(status), version, s.Description},
			Detail: [][2]string{{i18n.T("Ferramenta"), s.Name}, {i18n.T("Categoria"), tools.GroupTitle(s.Group)}, {i18n.T("Estado"), statusText(status)},
				{i18n.T("Versão"), version}, {i18n.T("Descrição"), s.Description}, {i18n.T("Ao remover"), warningOf(s.Key)}}})
	}
	return rows
}

func warningOf(key string) string {
	if t := tools.Find(key); t != nil && t.Warning != "" {
		return t.Warning
	}
	return "-"
}

func (b *realBackend) System() sysinfo.Snapshot { return sysinfo.Take(time.Second) }

// Logs lê só o final: o painel atualiza isso de tempos em tempos.
func (b *realBackend) Logs(view, key string) string {
	var cmd []string
	switch view {
	case viewServices:
		s := services.FindService(key)
		if s == nil {
			return ""
		}
		if s.LogCmd != nil {
			c := s.LogCmd()
			cmd = []string{c[0], "logs", "--tail", "300", c[len(c)-1]}
		} else {
			cmd = []string{"tail", "-n", "300", s.LogFile}
		}
	case viewProjects:
		e, ok := state.ReadState().Ups[key]
		if !ok || e.Log == "" {
			return i18n.T("Este projeto não está em segundo plano. Suba com a tecla u para ter log aqui.")
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

// Command roda o próprio romanctl em outro processo: a saída dele não atravessa o painel.
func (b *realBackend) Command(args ...string) *exec.Cmd {
	cmd := exec.Command(b.self, args...)
	cmd.Dir = filepath.Dir(b.self)
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}
	return cmd
}
