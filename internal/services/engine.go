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

// dockerEngine is the Docker daemon. It shows up as a service because the databases and queues
// depend on it, but stopping the daemon takes down every container on the machine, not only romanctl's.
func dockerEngine() *Service {
	s := &Service{Key: "docker", Name: "Docker", Aliases: []string{"dockerd", "docker-engine"}, Engine: true,
		Warning: i18n.T("stopping Docker brings down every container on the machine, not only romanctl's")}
	if platform.Current.Proot {
		s.Unsupported = i18n.T("proot cannot run containers; use a machine with Docker")
	}
	s.URL = dockerHost
	s.Installed = func() bool {
		if s.Unsupported != "" || !sys.Have("docker") {
			return false
		}
		// Docker Desktop does not put dockerd on this machine: there, existing means answering.
		return sys.Have("dockerd") || readEngine().up
	}
	s.Running = func() bool { return sys.Have("docker") && readEngine().up }
	s.before = sudoFirst
	s.start = func() error { defer forgetEngine(); return dockerDaemon("start") }
	s.stop = func() error { defer forgetEngine(); return dockerDaemon("stop") }
	s.install = func() error {
		return i18n.Errorf("install with: romanctl tools install docker")
	}
	if hasSystemd() {
		s.LogCmd = func() []string { return []string{"journalctl", "-u", "docker", "-n", "50", "-f", "--no-pager"} }
		s.LogTail = func() []string { return []string{"journalctl", "-u", "docker", "-n", "300", "--no-pager"} }
	}
	return s
}

// sudoFirst asks for the password before the spinner, which would erase sudo's prompt.
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

// dockerHost says how the client talks to the daemon: the local socket or the active context's.
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

// dockerDaemon starts or stops the daemon through systemd or the service script. Without a terminal,
// sudo cannot ask for a password: then the failure says to run the command in a terminal.
func dockerDaemon(action string) error {
	if !sys.Have("dockerd") {
		return i18n.Errorf("this Docker comes from Docker Desktop: %s it from Docker Desktop itself", map[string]string{"start": i18n.T("start"), "stop": i18n.T("stop")}[action])
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
			return i18n.Errorf("needs the sudo password: run romanctl %s docker in a terminal", action)
		}
		return i18n.Errorf("%v: %s", err, sys.Tail(out, 4))
	}
	return nil
}
