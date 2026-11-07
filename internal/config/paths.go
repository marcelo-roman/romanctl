package config

import (
	"os"
	"path/filepath"
)

var Home, _ = os.UserHomeDir()

var DataDir = filepath.Join(Home, ".local", "share", "romanctl")

var LogDir = filepath.Join(DataDir, "logs")

var ProjDir = envOr("ROMANCTL_PROJETOS", filepath.Join(Home, "projetos"))

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
