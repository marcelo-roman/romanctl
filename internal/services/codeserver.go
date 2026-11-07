package services

import (
	"bufio"
	"os"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/sys"
	"romanctl/internal/tools"
	"strings"
)

func CodeServerURL() string {
	addr := "127.0.0.1:8080"
	if f, err := os.Open(filepath.Join(config.Home, ".config", "code-server", "config.yaml")); err == nil {
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
	const pat = "/code-serve[r]/"
	s := &Service{Key: "code", Name: "code-server", Aliases: []string{"code-server", "vscode", "cs"},
		LogFile: filepath.Join(config.LogDir, "code-server.log"), URL: CodeServerURL}
	s.Installed = func() bool { return sys.Have("code-server") }
	s.Running = func() bool { return sys.Pgrep(pat) && sys.HttpUp(CodeServerURL()) }
	s.start = func() error {
		os.MkdirAll(config.ProjDir, 0o755)
		return sys.StartDetached(s.LogFile, config.ProjDir, "code-server", config.ProjDir)
	}
	s.stop = func() error { _, err := sys.Output("pkill", "-f", pat); return err }
	s.install = func() error {
		if r := tools.InstallOne(tools.Find("code-server"), false); r.Action == "failed" {
			return i18n.Errorf("%s", r.Error)
		}
		return nil
	}
	return s
}
