package services

import (
	_ "embed"
	"os"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"sort"
	"strconv"
	"strings"
)

//go:embed pgshim.c.txt
var pgshimSrc []byte

const shimDir = "/usr/local/lib/romanctl"

const shimSO = shimDir + "/pgshim.so"

const pgSock = "/var/run/postgresql"

func PgBin() string {
	dirs, _ := filepath.Glob("/usr/lib/postgresql/*/bin")
	sort.Slice(dirs, func(i, j int) bool {
		vi, _ := strconv.Atoi(filepath.Base(filepath.Dir(dirs[i])))
		vj, _ := strconv.Atoi(filepath.Base(filepath.Dir(dirs[j])))
		return vi < vj
	})
	if len(dirs) == 0 {
		return ""
	}
	return dirs[len(dirs)-1]
}

func asPostgres(cmdline string) error {
	return sys.Run("", nil, "su", "postgres", "-s", "/bin/sh", "-c", "LD_PRELOAD="+shimSO+" "+cmdline)
}

func buildShim() error {
	os.MkdirAll(shimDir, 0o755)
	src := filepath.Join(shimDir, "pgshim.c")
	if err := os.WriteFile(src, pgshimSrc, 0o644); err != nil {
		return err
	}
	if out, err := sys.Output("gcc", "-O2", "-shared", "-fPIC", "-o", shimSO, src, "-ldl"); err != nil {
		return i18n.Errorf("gcc: %v\n%s", err, out)
	}
	return nil
}

func postgres() *Service {
	if !platform.Current.Proot {
		return containerService(postgresContainer)
	}
	data := filepath.Join(config.DataDir, "postgres")
	s := &Service{Key: "postgres", Name: "PostgreSQL", Aliases: []string{"pg", "postgresql", "psql"},
		LogFile: filepath.Join(config.LogDir, "postgres.log"),
		URL:     func() string { return "postgres://postgres:postgres@127.0.0.1:5432/postgres" }}
	s.Installed = func() bool { return PgBin() != "" && sys.Exists(shimSO) }
	s.Running = func() bool {
		b, err := os.ReadFile(filepath.Join(data, "postmaster.pid"))
		if err != nil {
			return false
		}
		pid, _ := strconv.Atoi(strings.SplitN(string(b), "\n", 2)[0])
		return sys.ProcessIs(pid, "postgres")
	}
	s.start = func() error {
		bin := PgBin()
		// Sobrou de uma queda: se o PID foi reaproveitado, o postgres se recusaria a subir.
		os.Remove(filepath.Join(data, "postmaster.pid"))
		os.Remove(filepath.Join(pgSock, ".s.PGSQL.5432"))
		os.Remove(filepath.Join(pgSock, ".s.PGSQL.5432.lock"))
		os.MkdirAll(pgSock, 0o777)
		os.MkdirAll(config.LogDir, 0o777)
		os.Chmod(config.LogDir, 0o777)
		if !sys.Exists(filepath.Join(data, "PG_VERSION")) {
			ui.Infof("Primeira execução: criando o cluster em %s", data)
			if err := os.MkdirAll(data, 0o700); err != nil {
				return err
			}
			os.Chmod(data, 0o700)
			if err := asPostgres(bin + "/initdb -D " + data +
				" -U postgres --auth=trust -E UTF8 --locale=C.UTF-8 >/dev/null"); err != nil {
				return i18n.Errorf("initdb falhou: %w", err)
			}
		}
		return asPostgres(i18n.Sprintf("%s/pg_ctl -D %s -l %s -o '-c listen_addresses=127.0.0.1 -c unix_socket_directories=%s' -w start >/dev/null",
			bin, data, s.LogFile, pgSock))
	}
	s.stop = func() error { return asPostgres(PgBin() + "/pg_ctl -D " + data + " -m fast -w stop >/dev/null") }
	s.install = func() error {
		if PgBin() == "" {
			if err := aptInstall("postgresql"); err != nil {
				return err
			}
		}
		return buildShim()
	}
	return s
}
