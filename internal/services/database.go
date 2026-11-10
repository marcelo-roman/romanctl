package services

import (
	"regexp"
	"time"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/sys"
)

var safeDBName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Conn é como um projeto alcança o serviço. Os templates montam as próprias variáveis com estes campos.
type Conn struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	URL      string
}

// Conn devolve o endereço e as credenciais do serviço para o banco db.
func (s *Service) Conn(db string) (Conn, bool) {
	switch s.Key {
	case "postgres":
		return Conn{"127.0.0.1", "5432", "postgres", "postgres", db, "postgres://postgres:postgres@127.0.0.1:5432/" + db}, true
	case "mongo":
		return Conn{"127.0.0.1", "27017", "", "", db, "mongodb://127.0.0.1:27017/" + db}, true
	case "redis":
		return Conn{"127.0.0.1", "6379", "", "", "", "redis://127.0.0.1:6379"}, true
	case "mysql":
		return Conn{"127.0.0.1", "3306", "root", "mysql", db, "mysql://root:mysql@127.0.0.1:3306/" + db}, true
	case "mariadb":
		return Conn{"127.0.0.1", "3307", "root", "mariadb", db, "mysql://root:mariadb@127.0.0.1:3307/" + db}, true
	case "sqlserver":
		return Conn{"127.0.0.1", "1433", "sa", "Romanctl!Dev1", db,
			"Server=127.0.0.1,1433;Database=" + db + ";User Id=sa;Password=Romanctl!Dev1;TrustServerCertificate=True"}, true
	case "rabbitmq":
		return Conn{"127.0.0.1", "5672", "guest", "guest", "", "amqp://guest:guest@127.0.0.1:5672"}, true
	case "kafka":
		return Conn{"127.0.0.1", "9092", "", "", "", "127.0.0.1:9092"}, true
	}
	return Conn{}, false
}

var envNames = map[string]string{
	"postgres": "DATABASE_URL", "mongo": "MONGODB_URL", "redis": "REDIS_URL", "mysql": "DATABASE_URL",
	"mariadb": "DATABASE_URL", "sqlserver": "SQLSERVER_CONNECTION_STRING", "rabbitmq": "RABBITMQ_URL", "kafka": "KAFKA_BROKERS",
}

// EnvVars são as variáveis que todo projeto recebe para falar com o serviço.
func (s *Service) EnvVars(db string) [][2]string {
	c, ok := s.Conn(db)
	if !ok {
		return nil
	}
	return [][2]string{{envNames[s.Key], c.URL}}
}

// HasDatabases diz se o banco precisa ser criado antes de o app conectar.
func (s *Service) HasDatabases() bool {
	return s.Key == "postgres" || s.Key == "mysql" || s.Key == "mariadb"
}

// databaseCommand monta o comando que roda SQL no serviço, no ambiente atual.
func (s *Service) databaseCommand(sql string) []string {
	switch s.Key {
	case "postgres":
		if platform.Current.Proot {
			return []string{PgBin() + "/psql", "-h", "127.0.0.1", "-U", "postgres", "-tAc", sql}
		}
		return []string{containerEngine(), "exec", containerName("postgres"), "psql", "-U", "postgres", "-tAc", sql}
	case "mysql":
		return []string{containerEngine(), "exec", containerName("mysql"), "mysql", "-uroot", "-pmysql", "-e", sql}
	case "mariadb":
		return []string{containerEngine(), "exec", containerName("mariadb"), "mariadb", "-uroot", "-pmariadb", "-e", sql}
	}
	return nil
}

// EnsureDatabase cria o banco se faltar. Repete porque o serviço pode abrir a porta antes de
// aceitar comandos, principalmente na primeira subida de um container.
func (s *Service) EnsureDatabase(name string) error {
	if !safeDBName.MatchString(name) {
		return i18n.Errorf("nome de banco inválido: %q", name)
	}
	var last error
	ready := sys.WaitUntil(60*time.Second, func() bool {
		last = s.createDatabase(name)
		return last == nil
	})
	if !ready {
		return last
	}
	return nil
}

func (s *Service) createDatabase(name string) error {
	run := func(sql string) (string, error) {
		c := s.databaseCommand(sql)
		return sys.Output(c[0], c[1:]...)
	}
	if s.Key != "postgres" {
		out, err := run("CREATE DATABASE IF NOT EXISTS " + name)
		if err != nil {
			return i18n.Errorf("%v: %s", err, sys.Tail(out, 3))
		}
		return nil
	}
	// O PostgreSQL não tem CREATE DATABASE IF NOT EXISTS.
	out, err := run("SELECT 1 FROM pg_database WHERE datname = '" + name + "'")
	if err != nil {
		return i18n.Errorf("%v: %s", err, sys.Tail(out, 3))
	}
	if out == "1" {
		return nil
	}
	if out, err := run("CREATE DATABASE " + name); err != nil {
		return i18n.Errorf("%v: %s", err, sys.Tail(out, 3))
	}
	return nil
}
