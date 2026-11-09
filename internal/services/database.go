package services

import (
	"regexp"
	"time"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/sys"
)

var safeDBName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// EnvVars são as variáveis que um projeto usa para falar com o serviço.
func (s *Service) EnvVars(db string) [][2]string {
	switch s.Key {
	case "postgres":
		return [][2]string{{"DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5432/" + db}}
	case "mongo":
		return [][2]string{{"MONGODB_URL", "mongodb://127.0.0.1:27017/" + db}}
	case "redis":
		return [][2]string{{"REDIS_URL", "redis://127.0.0.1:6379"}}
	case "mysql":
		return [][2]string{{"DATABASE_URL", "mysql://root:mysql@127.0.0.1:3306/" + db}}
	case "mariadb":
		return [][2]string{{"DATABASE_URL", "mysql://root:mariadb@127.0.0.1:3307/" + db}}
	case "sqlserver":
		return [][2]string{{"SQLSERVER_CONNECTION_STRING", "Server=127.0.0.1,1433;Database=" + db + ";User Id=sa;Password=Romanctl!Dev1;TrustServerCertificate=True"}}
	case "rabbitmq":
		return [][2]string{{"RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672"}}
	case "kafka":
		return [][2]string{{"KAFKA_BROKERS", "127.0.0.1:9092"}}
	}
	return nil
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
