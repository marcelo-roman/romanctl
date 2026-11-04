package main

import (
	"bufio"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//go:embed pgshim.c.txt
var pgshimSrc []byte

type Service struct {
	Key       string
	Name      string
	Aliases   []string
	URL       func() string
	LogFile   string
	Installed func() bool
	Running   func() bool
	start     func() error
	stop      func() error
	install   func() error
}

var services = []*Service{codeServer(), postgres(), mongo(), redis()}

func findService(name string) *Service {
	name = strings.ToLower(name)
	for _, s := range services {
		if s.Key == name {
			return s
		}
		for _, a := range s.Aliases {
			if a == name {
				return s
			}
		}
	}
	return nil
}

func (s *Service) Start() error {
	if !s.Installed() {
		return fmt.Errorf("%s não está instalado — rode: romanctl install %s", s.Name, s.Key)
	}
	if s.Running() {
		okf("%s já está rodando → %s", s.Name, s.URL())
		return nil
	}
	infof("Iniciando %s...", s.Name)
	if err := s.start(); err != nil {
		return fmt.Errorf("%s: %w (veja: romanctl logs %s)", s.Name, err, s.Key)
	}
	if !waitUntil(30*time.Second, s.Running) {
		return fmt.Errorf("%s não subiu a tempo (veja: romanctl logs %s)", s.Name, s.Key)
	}
	okf("%s rodando → %s", s.Name, s.URL())
	return nil
}

func (s *Service) Stop() error {
	if !s.Running() {
		infof("%s não estava rodando", s.Name)
		return nil
	}
	if err := s.stop(); err != nil {
		return fmt.Errorf("%s: %w", s.Name, err)
	}
	if !waitUntil(20*time.Second, func() bool { return !s.Running() }) {
		return fmt.Errorf("%s não parou a tempo", s.Name)
	}
	okf("%s parado", s.Name)
	return nil
}

func (s *Service) Install() error {
	if s.Installed() {
		okf("%s já está instalado", s.Name)
		return nil
	}
	infof("Instalando %s (precisa de internet)...", s.Name)
	if err := s.install(); err != nil {
		return fmt.Errorf("instalação de %s falhou: %w", s.Name, err)
	}
	okf("%s instalado", s.Name)
	return nil
}

func aptInstall(pkgs ...string) error {
	args := append([]string{"install", "-y"}, pkgs...)
	return run("", []string{"DEBIAN_FRONTEND=noninteractive"}, "apt-get", args...)
}

// ---------------------------------------------------------------- code-server

func codeServerURL() string {
	addr := "127.0.0.1:8080"
	if f, err := os.Open(filepath.Join(home, ".config", "code-server", "config.yaml")); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if v, ok := strings.CutPrefix(sc.Text(), "bind-addr:"); ok {
				addr = strings.TrimSpace(v)
			}
		}
	}
	return "http://" + addr
}

func codeServer() *Service {
	const pat = "lib/code-serve[r]"
	s := &Service{Key: "code", Name: "code-server", Aliases: []string{"code-server", "vscode", "cs"},
		LogFile: filepath.Join(logDir, "code-server.log"), URL: codeServerURL}
	s.Installed = func() bool { return have("code-server") }
	s.Running = func() bool { return pgrep(pat) && httpUp(codeServerURL()) }
	s.start = func() error {
		os.MkdirAll(projDir, 0o755)
		return startDetached(s.LogFile, projDir, "code-server", projDir)
	}
	s.stop = func() error { _, err := output("pkill", "-f", pat); return err }
	s.install = func() error {
		return fmt.Errorf("instale manualmente o .deb arm64 de https://github.com/coder/code-server/releases")
	}
	return s
}

// ---------------------------------------------------------------- PostgreSQL
// No proot o postgres não roda direto: ele recusa root e confere o dono do
// diretório de dados. Rodamos como usuário "postgres" com um shim LD_PRELOAD
// (pgshim.so) que corrige o dono reportado por stat() e emula SysV shm.

const (
	shimDir = "/usr/local/lib/romanctl"
	shimSO  = shimDir + "/pgshim.so"
	pgSock  = "/var/run/postgresql"
)

func pgBin() string {
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
	return run("", nil, "su", "postgres", "-s", "/bin/sh", "-c", "LD_PRELOAD="+shimSO+" "+cmdline)
}

func buildShim() error {
	os.MkdirAll(shimDir, 0o755)
	src := filepath.Join(shimDir, "pgshim.c")
	if err := os.WriteFile(src, pgshimSrc, 0o644); err != nil {
		return err
	}
	if out, err := output("gcc", "-O2", "-shared", "-fPIC", "-o", shimSO, src, "-ldl"); err != nil {
		return fmt.Errorf("gcc: %v\n%s", err, out)
	}
	return nil
}

func postgres() *Service {
	data := filepath.Join(dataDir, "postgres")
	s := &Service{Key: "postgres", Name: "PostgreSQL", Aliases: []string{"pg", "postgresql", "psql"},
		LogFile: filepath.Join(logDir, "postgres.log"),
		URL:     func() string { return "postgres://postgres:postgres@127.0.0.1:5432/postgres" }}
	s.Installed = func() bool { return pgBin() != "" && exists(shimSO) }
	s.Running = func() bool {
		b, err := os.ReadFile(filepath.Join(data, "postmaster.pid"))
		if err != nil {
			return false
		}
		pid, _ := strconv.Atoi(strings.SplitN(string(b), "\n", 2)[0])
		return pid > 0 && syscall.Kill(pid, 0) == nil
	}
	s.start = func() error {
		bin := pgBin()
		os.MkdirAll(pgSock, 0o777)
		os.MkdirAll(logDir, 0o777)
		os.Chmod(logDir, 0o777)
		if !exists(filepath.Join(data, "PG_VERSION")) {
			infof("Primeira execução: criando o cluster em %s", data)
			if err := os.MkdirAll(data, 0o700); err != nil {
				return err
			}
			os.Chmod(data, 0o700)
			if err := asPostgres(bin + "/initdb -D " + data +
				" -U postgres --auth=trust -E UTF8 --locale=C.UTF-8 >/dev/null"); err != nil {
				return fmt.Errorf("initdb falhou: %w", err)
			}
		}
		return asPostgres(fmt.Sprintf("%s/pg_ctl -D %s -l %s -o '-c listen_addresses=127.0.0.1 -c unix_socket_directories=%s' -w start >/dev/null",
			bin, data, s.LogFile, pgSock))
	}
	s.stop = func() error { return asPostgres(pgBin() + "/pg_ctl -D " + data + " -m fast -w stop >/dev/null") }
	s.install = func() error {
		if pgBin() == "" {
			if err := aptInstall("postgresql"); err != nil {
				return err
			}
		}
		return buildShim()
	}
	return s
}

// ---------------------------------------------------------------- MongoDB
// MongoDB 8+ usa um tcmalloc que falha no espaço de endereços do kernel Android;
// por isso fixamos a série 7.0 (build ubuntu2204 aarch64, roda no 24.04).

const mongoHome = "/opt/mongodb"

func mongo() *Service {
	data := filepath.Join(dataDir, "mongo")
	bin := mongoHome + "/bin/mongod"
	s := &Service{Key: "mongo", Name: "MongoDB", Aliases: []string{"mongodb", "mongod"},
		LogFile: filepath.Join(logDir, "mongo.log"),
		URL:     func() string { return "mongodb://127.0.0.1:27017" }}
	s.Installed = func() bool { return exists(bin) }
	s.Running = func() bool { return pgrep("bin/mongo[d] ") }
	s.start = func() error {
		os.MkdirAll(data, 0o755)
		os.MkdirAll(logDir, 0o777)
		_, err := output(bin, "--dbpath", data, "--bind_ip", "127.0.0.1", "--port", "27017",
			"--fork", "--logpath", s.LogFile, "--logappend")
		return err
	}
	s.stop = func() error { _, err := output(bin, "--shutdown", "--dbpath", data); return err }
	s.install = installMongo
	return s
}

func installMongo() error {
	infof("Procurando a última versão 7.0.x...")
	resp, err := http.Get("https://downloads.mongodb.org/full.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var meta struct {
		Versions []struct {
			Version    string `json:"version"`
			Production bool   `json:"production_release"`
			Downloads  []struct {
				Target, Arch, Edition string
				Archive               struct{ URL, Sha256 string } `json:"archive"`
			} `json:"downloads"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return err
	}
	var url, sum, ver string
	for _, v := range meta.Versions {
		if !v.Production || !strings.HasPrefix(v.Version, "7.0.") {
			continue
		}
		for _, d := range v.Downloads {
			if d.Target == "ubuntu2204" && d.Arch == "aarch64" && d.Edition == "targeted" {
				url, sum, ver = d.Archive.URL, d.Archive.Sha256, v.Version
				break
			}
		}
		if url != "" {
			break
		}
	}
	if url == "" {
		return fmt.Errorf("não achei build 7.0 aarch64")
	}
	infof("Baixando MongoDB %s...", ver)
	tmp, err := os.CreateTemp("", "mongodb-*.tgz")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, h), r.Body); err != nil {
		return err
	}
	tmp.Close()
	if got := hex.EncodeToString(h.Sum(nil)); got != sum {
		return fmt.Errorf("checksum não confere (esperado %s, obtido %s)", sum, got)
	}
	okf("Checksum SHA-256 conferido")
	os.RemoveAll(mongoHome)
	os.MkdirAll(mongoHome, 0o755)
	if out, err := output("tar", "-xzf", tmp.Name(), "-C", mongoHome, "--strip-components=1"); err != nil {
		return fmt.Errorf("tar: %v %s", err, out)
	}
	if !have("mongosh") {
		infof("Instalando mongosh (npm)...")
		if err := run("", nil, "npm", "install", "-g", "mongosh"); err != nil {
			warnf("mongosh não instalado: %v", err)
		}
	}
	return nil
}

// ---------------------------------------------------------------- Redis

func redis() *Service {
	data := filepath.Join(dataDir, "redis")
	s := &Service{Key: "redis", Name: "Redis", Aliases: []string{"redis-server"},
		LogFile: filepath.Join(logDir, "redis.log"),
		URL:     func() string { return "redis://127.0.0.1:6379" }}
	s.Installed = func() bool { return have("redis-server") }
	s.Running = func() bool { out, _ := output("redis-cli", "-h", "127.0.0.1", "ping"); return out == "PONG" }
	s.start = func() error {
		os.MkdirAll(data, 0o755)
		os.MkdirAll(logDir, 0o777)
		_, err := output("redis-server", "--daemonize", "yes", "--bind", "127.0.0.1", "--port", "6379",
			"--dir", data, "--logfile", s.LogFile, "--pidfile", filepath.Join(data, "redis.pid"))
		return err
	}
	s.stop = func() error { _, err := output("redis-cli", "-h", "127.0.0.1", "shutdown"); return err }
	s.install = func() error { return aptInstall("redis-server") }
	return s
}
