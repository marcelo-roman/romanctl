package services

import (
	"net"
	"strings"
	"sync"
	"time"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/sys"
)

// containerSpec describes a service that runs in Docker or Podman. Ports bind only to
// 127.0.0.1 and data lives in a named volume, so it survives stopping the service.
type containerSpec struct {
	Key     string
	Name    string
	Aliases []string
	Image   string
	URL     string
	Ports   []string // host:container; the first one signals when the service is ready
	Env     []string
	Volume  string // data directory inside the container
	Args    []string
	Arches  []string  // empty means all
	Ready   readiness // how to tell it is ready, beyond the port opening
}

// containerEngine returns the available engine, preferring Docker.
func containerEngine() string {
	for _, bin := range []string{"docker", "podman"} {
		if sys.Have(bin) {
			return bin
		}
	}
	return ""
}

func containerName(key string) string { return "romanctl-" + key }

// engineState is one reading of the engine: the dashboard and status ask about several services at
// once, and each separate question to Docker cost a call.
type engineState struct {
	at         time.Time
	up         bool              // the daemon answered
	images     map[string]bool   // pulled repository:tag
	containers map[string]string // name → state (running, exited...)
}

var (
	engineMu   sync.Mutex
	engineSeen engineState
)

const engineTTL = time.Second

func readEngine() engineState {
	engineMu.Lock()
	defer engineMu.Unlock()
	if time.Since(engineSeen.at) < engineTTL {
		return engineSeen
	}
	st := engineState{at: time.Now(), images: map[string]bool{}, containers: map[string]string{}}
	if out, err := engineOutput("ps", "-a", "--format", "{{.Names}}\t{{.State}}"); err == nil {
		st.up = true
		for _, line := range strings.Split(out, "\n") {
			if name, state, ok := strings.Cut(strings.TrimSpace(line), "\t"); ok {
				st.containers[name] = state
			}
		}
		if out, err := engineOutput("image", "ls", "--format", "{{.Repository}}:{{.Tag}}"); err == nil {
			for _, line := range strings.Fields(out) {
				st.images[line] = true
			}
		}
	}
	engineSeen = st
	return st
}

// forgetEngine makes the next question read the engine again, after something changed in it.
func forgetEngine() {
	engineMu.Lock()
	engineSeen = engineState{}
	engineMu.Unlock()
}

// imageRef completes the reference the way Docker lists it: "redis" becomes "redis:latest".
func imageRef(image string) string {
	if strings.Contains(image[strings.LastIndex(image, "/")+1:], ":") {
		return image
	}
	return image + ":latest"
}

func engineOutput(args ...string) (string, error) {
	engine := containerEngine()
	if engine == "" {
		return "", i18n.Errorf("install a container engine: romanctl tools install docker")
	}
	return sys.Output(engine, args...)
}

func hostPort(mapping string) string {
	for i := 0; i < len(mapping); i++ {
		if mapping[i] == ':' {
			return mapping[:i]
		}
	}
	return mapping
}

func runArgs(c containerSpec) []string {
	args := []string{"run", "-d", "--name", containerName(c.Key)}
	for _, p := range c.Ports {
		args = append(args, "-p", "127.0.0.1:"+p)
	}
	for _, e := range c.Env {
		args = append(args, "-e", e)
	}
	if c.Volume != "" {
		args = append(args, "-v", containerName(c.Key)+":"+c.Volume)
	}
	args = append(args, c.Image)
	return append(args, c.Args...)
}

// createArgs are the same as runArgs, to create the container without starting it.
func createArgs(c containerSpec) []string { return append([]string{"create"}, runArgs(c)[2:]...) }

func supportedArch(c containerSpec, arch string) bool {
	if len(c.Arches) == 0 {
		return true
	}
	for _, a := range c.Arches {
		if a == arch {
			return true
		}
	}
	return false
}

func containerService(c containerSpec) *Service {
	name := containerName(c.Key)
	s := &Service{Key: c.Key, Name: c.Name, Aliases: c.Aliases, URL: func() string { return c.URL }}
	switch {
	case platform.Current.Proot:
		s.Unsupported = i18n.T("proot cannot run containers; use a machine with Docker")
	case !supportedArch(c, platform.Current.Arch):
		s.Unsupported = i18n.Sprintf("image %s does not exist for %s", c.Image, platform.Current.Arch)
	}
	s.LogCmd = func() []string { return []string{containerEngine(), "logs", "-f", "--tail", "50", name} }
	s.LogTail = func() []string { return []string{containerEngine(), "logs", "--tail", "300", name} }
	// Installed means romanctl's container exists. Having Docker is not enough, nor is the image: it
	// may have been pulled by another project on the machine.
	s.Installed = func() bool {
		if s.Unsupported != "" || containerEngine() == "" {
			return false
		}
		_, created := readEngine().containers[name]
		return created
	}
	s.Running = func() bool {
		if containerEngine() == "" {
			return false
		}
		return readEngine().containers[name] == "running"
	}
	s.start = func() error {
		defer forgetEngine()
		if err := checkPorts(c, name); err != nil {
			return err
		}
		_, created := readEngine().containers[name]
		var out string
		var err error
		if created {
			out, err = engineOutput("start", name)
		} else {
			out, err = engineOutput(runArgs(c)...)
		}
		if err != nil {
			return i18n.Errorf("%v: %s", err, sys.Tail(out, 6))
		}
		if err := waitPort(hostPort(c.Ports[0]), 90*time.Second); err != nil {
			return err
		}
		if c.Ready == nil {
			return nil
		}
		if !sys.WaitUntil(90*time.Second, func() bool { return c.Ready(hostPort(c.Ports[0])) }) {
			return i18n.Errorf("%s opened its port but did not answer in time", c.Name)
		}
		return nil
	}
	s.Conflict = func() string {
		for _, mapping := range c.Ports {
			port := hostPort(mapping)
			if owner := portOwner(port, name); owner != "" {
				return i18n.Sprintf("container %s also publishes port %s: the app may be talking to it instead of %s", owner, port, c.Name)
			}
		}
		return ""
	}
	s.stop = func() error {
		defer forgetEngine()
		if out, err := engineOutput("stop", name); err != nil {
			return i18n.Errorf("%v: %s", err, sys.Tail(out, 4))
		}
		return nil
	}
	s.install = func() error {
		if s.Unsupported != "" {
			return i18n.Errorf("%s", s.Unsupported)
		}
		defer forgetEngine()
		if !readEngine().images[imageRef(c.Image)] {
			if out, err := engineOutput("pull", c.Image); err != nil {
				return i18n.Errorf("%v: %s", err, sys.Tail(out, 4))
			}
		}
		// Creates the container stopped: it is what marks the service as installed on this machine.
		if out, err := engineOutput(createArgs(c)...); err != nil {
			return i18n.Errorf("%v: %s", err, sys.Tail(out, 4))
		}
		return nil
	}
	return s
}

var checkPorts = portsFree

// portsFree prevents starting the service with its port already taken: the app would talk to another
// program, and waiting for the port to open would succeed at once, without the container being up.
func portsFree(c containerSpec, name string) error {
	for _, mapping := range c.Ports {
		port := hostPort(mapping)
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, time.Second)
		if err != nil {
			continue
		}
		conn.Close()
		if owner := portOwner(port, name); owner != "" {
			return i18n.Errorf("port %s is already used by container %s; stop it before starting %s", port, owner, c.Name)
		}
		return i18n.Errorf("port %s is already used by another program; free it before starting %s", port, c.Name)
	}
	return nil
}

func portOwner(port, self string) string {
	out, err := engineOutput("ps", "--filter", "publish="+port, "--format", "{{.Names}}")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" && line != self {
			return line
		}
	}
	return ""
}

func waitPort(port string, timeout time.Duration) error {
	ok := sys.WaitUntil(timeout, func() bool {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, time.Second)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	})
	if !ok {
		return i18n.Errorf("port %s did not open in time", port)
	}
	return nil
}

var (
	postgresContainer = containerSpec{Key: "postgres", Name: "PostgreSQL", Aliases: []string{"pg", "postgresql", "psql"}, Image: "postgres:17",
		URL: "postgres://postgres:postgres@127.0.0.1:5432/postgres", Ports: []string{"5432:5432"},
		Env: []string{"POSTGRES_PASSWORD=postgres"}, Volume: "/var/lib/postgresql/data", Ready: postgresReady}
	mongoContainer = containerSpec{Key: "mongo", Name: "MongoDB", Aliases: []string{"mongodb", "mongod"}, Image: "mongo:8",
		URL: "mongodb://127.0.0.1:27017", Ports: []string{"27017:27017"}, Volume: "/data/db"}
	redisContainer = containerSpec{Key: "redis", Name: "Redis", Aliases: []string{"redis-server"}, Image: "redis:8",
		URL: "redis://127.0.0.1:6379", Ports: []string{"6379:6379"}, Volume: "/data"}
)

// Services that only exist as containers.
func containerOnly() []*Service {
	specs := []containerSpec{
		{Key: "mysql", Name: "MySQL", Image: "mysql:8.4", URL: "mysql://root:mysql@127.0.0.1:3306",
			Ports: []string{"3306:3306"}, Env: []string{"MYSQL_ROOT_PASSWORD=mysql"}, Volume: "/var/lib/mysql"},
		{Key: "mariadb", Name: "MariaDB", Image: "mariadb:11", URL: "mysql://root:mariadb@127.0.0.1:3307",
			Ports: []string{"3307:3306"}, Env: []string{"MARIADB_ROOT_PASSWORD=mariadb"}, Volume: "/var/lib/mysql"},
		{Key: "sqlserver", Name: "SQL Server", Aliases: []string{"mssql", "sql"}, Image: "mcr.microsoft.com/mssql/server:2022-latest",
			URL: "Server=127.0.0.1,1433;User Id=sa;Password=Romanctl!Dev1;TrustServerCertificate=True", Ports: []string{"1433:1433"},
			Env: []string{"ACCEPT_EULA=Y", "MSSQL_SA_PASSWORD=Romanctl!Dev1", "MSSQL_PID=Developer"}, Volume: "/var/opt/mssql", Arches: []string{"amd64"}},
		{Key: "rabbitmq", Name: "RabbitMQ", Aliases: []string{"rabbit"}, Image: "rabbitmq:4-management",
			URL: i18n.T("amqp://guest:guest@127.0.0.1:5672 (dashboard: http://127.0.0.1:15672)"), Ports: []string{"5672:5672", "15672:15672"}, Volume: "/var/lib/rabbitmq", Ready: amqpReady},
		{Key: "kafka", Name: "Kafka", Image: "apache/kafka:4.1.0", URL: "127.0.0.1:9092", Ports: []string{"9092:9092"}, Volume: "/var/lib/kafka/data"},
	}
	var out []*Service
	for _, c := range specs {
		out = append(out, containerService(c))
	}
	return out
}
