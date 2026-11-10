package services

import (
	"os"
	"strings"
	"sync"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

// dockerEngine é o daemon do Docker. Ele aparece como serviço porque os bancos e filas dependem
// dele, mas parar o daemon derruba todos os containers da máquina, não só os do romanctl.
func dockerEngine() *Service {
	s := &Service{Key: "docker", Name: "Docker", Aliases: []string{"dockerd", "docker-engine"}, Engine: true,
		Warning: i18n.T("parar o Docker derruba todos os containers da máquina, não só os do romanctl")}
	if platform.Current.Proot {
		s.Unsupported = i18n.T("o proot não roda containers; use uma máquina com Docker")
	}
	s.URL = dockerHost
	s.Installed = func() bool {
		if s.Unsupported != "" || !sys.Have("docker") {
			return false
		}
		// O Docker Desktop não põe o dockerd nesta máquina: aí, existir é responder.
		return sys.Have("dockerd") || readEngine().up
	}
	s.Running = func() bool { return sys.Have("docker") && readEngine().up }
	s.before = sudoFirst
	s.start = func() error { defer forgetEngine(); return dockerDaemon("start") }
	s.stop = func() error { defer forgetEngine(); return dockerDaemon("stop") }
	s.install = func() error {
		return i18n.Errorf("instale com: romanctl tools install docker")
	}
	if hasSystemd() {
		s.LogCmd = func() []string { return []string{"journalctl", "-u", "docker", "-n", "50", "-f", "--no-pager"} }
		s.LogTail = func() []string { return []string{"journalctl", "-u", "docker", "-n", "300", "--no-pager"} }
	}
	return s
}

// sudoFirst pede a senha antes do spinner, que apagaria a pergunta do sudo.
func sudoFirst() error {
	if os.Geteuid() == 0 || !ui.Interactive() || !sys.Have("dockerd") {
		return nil
	}
	return sys.Run("", nil, "sudo", "-v")
}

var (
	hostOnce sync.Once
	hostSeen string
)

// dockerHost diz por onde o cliente fala com o daemon: o socket local ou o do contexto ativo.
func dockerHost() string {
	if h := os.Getenv("DOCKER_HOST"); h != "" {
		return h
	}
	hostOnce.Do(func() {
		hostSeen = "unix:///var/run/docker.sock"
		if out, err := sys.Output("docker", "context", "inspect", "--format", "{{.Endpoints.docker.Host}}"); err == nil && out != "" {
			hostSeen = strings.TrimSpace(out)
		}
	})
	return hostSeen
}

func hasSystemd() bool { return sys.Exists("/run/systemd/system") && sys.Have("systemctl") }

// dockerDaemon liga ou desliga o daemon pelo systemd ou pelo script de serviço. Sem terminal,
// o sudo não pode pedir senha: aí a falha diz para rodar o comando num terminal.
func dockerDaemon(action string) error {
	if !sys.Have("dockerd") {
		return i18n.Errorf("este Docker vem do Docker Desktop: %s pelo próprio Docker Desktop", map[string]string{"start": i18n.T("inicie"), "stop": i18n.T("pare")}[action])
	}
	command := []string{"service", "docker", action}
	if hasSystemd() {
		command = []string{"systemctl", action, "docker"}
	}
	if os.Geteuid() != 0 {
		command = append([]string{"sudo", "-n"}, command...)
	}
	if out, err := sys.Output(command[0], command[1:]...); err != nil {
		if strings.Contains(out, "password") || strings.Contains(out, "senha") {
			return i18n.Errorf("precisa da senha do sudo: rode romanctl %s docker em um terminal", action)
		}
		return i18n.Errorf("%v: %s", err, sys.Tail(out, 4))
	}
	return nil
}
