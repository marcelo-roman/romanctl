package projects

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"romanctl/internal/browser"
	"romanctl/internal/dotenv"
	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/services"
	"romanctl/internal/state"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
	"time"
)

// openWhenUp waits for the project's app to respond and then opens the browser.
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

// RunCommand works out how to run the project, installing Node dependencies if missing.
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
			ui.Infof("Installing dependencies (npm install)...")
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
	return nil, i18n.Errorf("I don't know how to run %q (unrecognized type)", name)
}

func display(cmd []string) string {
	if len(cmd) == 3 && cmd[0] == "sh" && cmd[1] == "-c" {
		return strings.Join(strings.Fields(strings.ReplaceAll(strings.TrimSpace(cmd[2]), "\n", "; ")), " ")
	}
	return strings.Join(cmd, " ")
}

// RunProject runs the project in the foreground; Ctrl+C stops only the project.
func RunProject(name string, open bool) error {
	p, err := ProjectPath(name)
	if err != nil {
		return err
	}
	if project, ok := manifest.Read(p); ok {
		for _, key := range project.Services {
			if s := services.FindService(key); s != nil && !s.Running() {
				ui.Warnf("%s is not running and the project depends on it: romanctl start %s", s.Name, s.Key)
			}
		}
	}
	return RunAt(p, name, dotenv.Load(filepath.Join(p, ".env"), os.LookupEnv), open)
}

// RunAt runs in the foreground, with extra variables.
func RunAt(p, name string, env []string, open bool) error {
	cmd, err := RunCommand(p, name)
	if err != nil {
		return err
	}
	state.MarkProject(name, true)
	defer state.MarkProject(name, false)
	ui.Infof("Running %s in %s  (Ctrl+C to stop)", display(cmd), p)
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
		ui.Okf("Project stopped")
		return nil
	default:
		return err
	}
}
