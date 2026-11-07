package services

import (
	"os"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/platform"
	"romanctl/internal/sys"
)

func redis() *Service {
	if !platform.Current.Proot {
		return containerService(redisContainer)
	}
	data := filepath.Join(config.DataDir, "redis")
	s := &Service{Key: "redis", Name: "Redis", Aliases: []string{"redis-server"},
		LogFile: filepath.Join(config.LogDir, "redis.log"),
		URL:     func() string { return "redis://127.0.0.1:6379" }}
	s.Installed = func() bool { return sys.Have("redis-server") }
	s.Running = func() bool { out, _ := sys.Output("redis-cli", "-h", "127.0.0.1", "ping"); return out == "PONG" }
	s.start = func() error {
		os.MkdirAll(data, 0o755)
		os.MkdirAll(config.LogDir, 0o777)
		_, err := sys.Output("redis-server", "--daemonize", "yes", "--bind", "127.0.0.1", "--port", "6379",
			"--dir", data, "--logfile", s.LogFile, "--pidfile", filepath.Join(data, "redis.pid"))
		return err
	}
	s.stop = func() error { _, err := sys.Output("redis-cli", "-h", "127.0.0.1", "shutdown"); return err }
	s.install = func() error { return aptInstall("redis-server") }
	return s
}
