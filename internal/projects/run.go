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

// RunCommand descobre como rodar o projeto, instalando as dependências de Node se faltarem.
func RunCommand(p, name string) ([]string, error) {
	project, hasManifest := manifest.Read(p)
	switch t := DetectType(p); {
	case hasManifest && len(project.Processes) > 0:
		self, err := os.Executable()
		if err != nil {
			return nil, err
		}
		return []string{self, SuperviseArg}, nil
	case hasManifest:
		return []string{"sh", "-c", project.Run}, nil
	case t == "dotnet":
		return []string{"dotnet", "run"}, nil
	case t == "go":
		return []string{"go", "run", "."}, nil
	case t != "?":
		if !sys.Exists(filepath.Join(p, "node_modules")) {
			ui.Infof("Instalando dependências (npm install)...")
			offlineMode = !sys.IsOnline()
			if err := npmInstall(p); err != nil {
				return nil, err
			}
		}
		if b, _ := os.ReadFile(filepath.Join(p, "package.json")); !strings.Contains(string(b), `"dev"`) {
			return []string{"npm", "start"}, nil
		}
		return []string{"npm", "run", "dev"}, nil
	}
	return nil, i18n.Errorf("não sei rodar %q (tipo não reconhecido)", name)
}

func display(cmd []string) string {
	if len(cmd) == 3 && cmd[0] == "sh" && cmd[1] == "-c" {
		return strings.Join(strings.Fields(strings.ReplaceAll(strings.TrimSpace(cmd[2]), "\n", "; ")), " ")
	}
	return strings.Join(cmd, " ")
}

// RunProject roda o projeto em primeiro plano; Ctrl+C encerra só o projeto.
func RunProject(name string, open bool) error {
	p, err := ProjectPath(name)
	if err != nil {
		return err
	}
	if project, ok := manifest.Read(p); ok {
		for _, key := range project.Services {
			if s := services.FindService(key); s != nil && !s.Running() {
				ui.Warnf("%s não está rodando e o projeto depende dele: romanctl start %s", s.Name, s.Key)
			}
		}
	}
	return RunAt(p, name, nil, open)
}

// RunAt roda em primeiro plano, com variáveis extras.
func RunAt(p, name string, env []string, open bool) error {
	cmd, err := RunCommand(p, name)
	if err != nil {
		return err
	}
	state.MarkProject(name, true)
	defer state.MarkProject(name, false)
	ui.Infof("Rodando %s em %s  (Ctrl+C para parar)", display(cmd), p)
	if open {
		openWhenUp(p)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	err = sys.Run(p, env, cmd[0], cmd[1:]...)
	select {
	case <-sig:
		fmt.Println()
		ui.Okf("Projeto parado")
		return nil
	default:
		return err
	}
}
