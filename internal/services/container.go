package services

import (
	"net"
	"romanctl/internal/i18n"
	"time"

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
	s.Installed = func() bool { return s.Unsupported == "" && containerEngine() != "" }
	s.Running = func() bool {
		if containerEngine() == "" {
			return false
		}
		out, err := engineOutput("inspect", "-f", "{{.State.Running}}", name)
		return err == nil && out == "true"
	}
	s.start = func() error {
		_, inspectErr := engineOutput("inspect", name)
		var out string
		var err error
		if inspectErr == nil {
			out, err = engineOutput("start", name)
		} else {
			out, err = engineOutput(runArgs(c)...)
		}
		if err != nil {
			return i18n.Errorf("%v: %s", err, sys.Tail(out, 6))
		}
		return waitPort(hostPort(c.Ports[0]), 90*time.Second)
	}
	s.stop = func() error {
		if out, err := engineOutput("stop", name); err != nil {
			return i18n.Errorf("%v: %s", err, sys.Tail(out, 4))
		}
		return nil
	}
	s.install = func() error {
		if s.Unsupported != "" {
			return i18n.Errorf("%s", s.Unsupported)
		}
		if out, err := engineOutput("pull", c.Image); err != nil {
			return i18n.Errorf("%v: %s", err, sys.Tail(out, 4))
		}
		return nil
	}
	s.alreadyInstalled = func() bool {
		_, err := engineOutput("image", "inspect", c.Image)
		return err == nil
	}
	return s
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
		{Key: "minio", Name: "MinIO", Aliases: []string{"s3"}, Image: "minio/minio",
			URL: i18n.T("http://127.0.0.1:9000 (painel: http://127.0.0.1:9001, minioadmin/minioadmin)"), Ports: []string{"9000:9000", "9001:9001"},
			Volume: "/data", Args: []string{"server", "/data", "--console-address", ":9001"}},
	}
	var out []*Service
	for _, c := range specs {
		out = append(out, containerService(c))
	}
	return out
}
