package projects

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"romanctl/internal/browser"
	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/services"
	"romanctl/internal/state"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
	"time"
)

// openWhenUp espera o app do projeto responder e então abre o navegador.
func openWhenUp(p string) {
	u := AppURL(p)
	if u == "" {
		return
	}
	go func() {
		if sys.WaitUntil(3*time.Minute, func() bool { return sys.HttpUp(u) }) {
			browser.OpenURL(u)
		}
	}()
}

// runProject roda o projeto em primeiro plano; Ctrl+C encerra só o projeto.
func RunProject(name string, open bool) error {
	p, err := ProjectPath(name)
	if err != nil {
		return err
	}
	var cmd []string
	project, hasManifest := manifest.Read(p)
	switch t := DetectType(p); {
	case hasManifest:
		cmd = strings.Fields(project.Run)
		for _, key := range project.Services {
			if s := services.FindService(key); s != nil && !s.Running() {
				ui.Warnf("%s não está rodando e o projeto depende dele: romanctl start %s", s.Name, s.Key)
			}
		}
	case t == "dotnet":
		cmd = []string{"dotnet", "run"}
	case t == "go":
		cmd = []string{"go", "run", "."}
	case t != "?":
		if !sys.Exists(filepath.Join(p, "node_modules")) {
			ui.Infof("Instalando dependências (npm install)...")
			offlineMode = !sys.IsOnline()
			if err := npmInstall(p); err != nil {
				return err
			}
		}
		cmd = []string{"npm", "run", "dev"}
		if b, _ := os.ReadFile(filepath.Join(p, "package.json")); !strings.Contains(string(b), `"dev"`) {
			cmd = []string{"npm", "start"}
		}
	default:
		return i18n.Errorf("não sei rodar %q (tipo não reconhecido)", name)
	}
	state.MarkProject(name, true)
	defer state.MarkProject(name, false)
	ui.Infof("Rodando %s em %s  (Ctrl+C para parar)", strings.Join(cmd, " "), p)
	if open {
		openWhenUp(p)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	err = sys.Run(p, nil, cmd[0], cmd[1:]...)
	select {
	case <-sig:
		fmt.Println()
		ui.Okf("Projeto parado")
		return nil
	default:
		return err
	}
}
