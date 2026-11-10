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

// containerSpec descreve um serviço que roda em Docker ou Podman. As portas ficam só em
// 127.0.0.1 e os dados, num volume nomeado, então sobrevivem a parar o serviço.
type containerSpec struct {
	Key     string
	Name    string
	Aliases []string
	Image   string
	URL     string
	Ports   []string // host:container; a primeira indica quando o serviço está pronto
	Env     []string
	Volume  string // pasta de dados dentro do container
	Args    []string
	Arches  []string // vazio significa todas
}

// containerEngine devolve o motor disponível, preferindo o Docker.
func containerEngine() string {
	for _, bin := range []string{"docker", "podman"} {
		if sys.Have(bin) {
			return bin
		}
	}
	return ""
}

func containerName(key string) string { return "romanctl-" + key }

// engineState é uma leitura do motor: o painel e o status perguntam por vários serviços de uma
// vez, e cada pergunta separada ao Docker custava uma chamada.
type engineState struct {
	at         time.Time
	up         bool              // o daemon respondeu
	images     map[string]bool   // repositório:tag baixados
	containers map[string]string // nome → estado (running, exited...)
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

// forgetEngine faz a próxima pergunta ler o motor de novo, depois de mudar algo nele.
func forgetEngine() {
	engineMu.Lock()
	engineSeen = engineState{}
	engineMu.Unlock()
}

// imageRef completa a referência como o Docker lista: "redis" vira "redis:latest".
func imageRef(image string) string {
	if strings.Contains(image[strings.LastIndex(image, "/")+1:], ":") {
		return image
	}
	return image + ":latest"
}

func engineOutput(args ...string) (string, error) {
	engine := containerEngine()
	if engine == "" {
		return "", i18n.Errorf("instale um motor de containers: romanctl tools install docker")
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

// createArgs são os mesmos de runArgs, para criar o container sem ligá-lo.
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
		s.Unsupported = i18n.T("o proot não roda containers; use uma máquina com Docker")
	case !supportedArch(c, platform.Current.Arch):
		s.Unsupported = i18n.Sprintf("a imagem %s não existe para %s", c.Image, platform.Current.Arch)
	}
	s.LogCmd = func() []string { return []string{containerEngine(), "logs", "-f", "--tail", "50", name} }
	s.LogTail = func() []string { return []string{containerEngine(), "logs", "--tail", "300", name} }
	// Instalado é o container do romanctl existir. Ter o Docker não basta, e nem a imagem: ela
	// pode ter sido baixada por outro projeto da máquina.
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
		return waitPort(hostPort(c.Ports[0]), 90*time.Second)
	}
	s.Conflict = func() string {
		for _, mapping := range c.Ports {
			port := hostPort(mapping)
			if owner := portOwner(port, name); owner != "" {
				return i18n.Sprintf("o container %s também publica a porta %s: o app pode estar falando com ele, e não com o %s", owner, port, c.Name)
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
		// Cria o container parado: é ele que marca o serviço como instalado nesta máquina.
		if out, err := engineOutput(createArgs(c)...); err != nil {
			return i18n.Errorf("%v: %s", err, sys.Tail(out, 4))
		}
		return nil
	}
	return s
}

var checkPorts = portsFree

// portsFree impede ligar o serviço com a porta já ocupada: o app falaria com outro programa, e esperar
// a porta abrir daria certo na hora, sem o container ter subido.
func portsFree(c containerSpec, name string) error {
	for _, mapping := range c.Ports {
		port := hostPort(mapping)
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, time.Second)
		if err != nil {
			continue
		}
		conn.Close()
		if owner := portOwner(port, name); owner != "" {
			return i18n.Errorf("a porta %s já está em uso pelo container %s; pare-o antes de ligar o %s", port, owner, c.Name)
		}
		return i18n.Errorf("a porta %s já está em uso por outro programa; libere-a antes de ligar o %s", port, c.Name)
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
		return i18n.Errorf("a porta %s não abriu a tempo", port)
	}
	return nil
}

var (
	postgresContainer = containerSpec{Key: "postgres", Name: "PostgreSQL", Aliases: []string{"pg", "postgresql", "psql"}, Image: "postgres:17",
		URL: "postgres://postgres:postgres@127.0.0.1:5432/postgres", Ports: []string{"5432:5432"},
		Env: []string{"POSTGRES_PASSWORD=postgres"}, Volume: "/var/lib/postgresql/data"}
	mongoContainer = containerSpec{Key: "mongo", Name: "MongoDB", Aliases: []string{"mongodb", "mongod"}, Image: "mongo:8",
		URL: "mongodb://127.0.0.1:27017", Ports: []string{"27017:27017"}, Volume: "/data/db"}
	redisContainer = containerSpec{Key: "redis", Name: "Redis", Aliases: []string{"redis-server"}, Image: "redis:8",
		URL: "redis://127.0.0.1:6379", Ports: []string{"6379:6379"}, Volume: "/data"}
)

// Serviços que só existem em container.
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
			URL: i18n.T("amqp://guest:guest@127.0.0.1:5672 (painel: http://127.0.0.1:15672)"), Ports: []string{"5672:5672", "15672:15672"}, Volume: "/var/lib/rabbitmq"},
		{Key: "kafka", Name: "Kafka", Image: "apache/kafka:4.1.0", URL: "127.0.0.1:9092", Ports: []string{"9092:9092"}, Volume: "/var/lib/kafka/data"},
	}
	var out []*Service
	for _, c := range specs {
		out = append(out, containerService(c))
	}
	return out
}
