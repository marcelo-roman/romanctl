package cli

import (
	"net/url"
	"os"
	"path/filepath"
	"romanctl/internal/browser"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/services"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
)

func codeFolderURL(p string) string {
	return services.CodeServerURL() + "/?folder=" + url.QueryEscape(p)
}

func ensureCodeServer() error {
	cs := services.FindService("code")
	if !cs.Running() {
		return cs.Start()
	}
	return nil
}

// openTarget: "code", um projeto, ou uma URL. app=true abre o app rodando do projeto.
func openTarget(target string, app bool) error {
	switch {
	case strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://"):
		return browser.OpenURL(target)
	case !app && services.FindService(target) == services.FindService("code"):
		if err := ensureCodeServer(); err != nil {
			return err
		}
		return browser.OpenURL(services.CodeServerURL())
	}
	p, err := projects.ProjectPath(target)
	if err != nil {
		return err
	}
	if app {
		u := projects.AppURL(p)
		if u == "" {
			return i18n.Errorf("não sei em que porta %q roda", target)
		}
		if !sys.HttpUp(u) {
			ui.Warnf("%s não está respondendo em %s — rode antes: romanctl run %s", target, u, target)
			return nil
		}
		return browser.OpenURL(u)
	}
	if err := ensureCodeServer(); err != nil {
		return err
	}
	return browser.OpenURL(codeFolderURL(p))
}

func menuOpen() error {
	opts := []ui.Option{ui.Opt("code-server", "code|")}
	for _, p := range projects.ListProjects() {
		opts = append(opts, ui.Opt("📁 "+p.Name+ui.SGray.Render(i18n.T("  no code-server")), p.Name+"|"))
		if u := projects.AppURL(p.Path); u != "" && sys.HttpUp(u) {
			opts = append(opts, ui.Opt("🌐 "+p.Name+ui.SGreen.Render(i18n.T("  app rodando em ")+u), p.Name+"|app"))
		}
	}
	sel, err := ui.PickOne(i18n.T("Abrir no navegador"), opts)
	if err != nil {
		return err
	}
	target, mode, _ := strings.Cut(sel, "|")
	return openTarget(target, mode == "app")
}

// openInCode abre uma pasta no code-server. Aceita nome de projeto ou caminho.
// Sem alvo: usa o projeto da pasta atual ou, fora de um projeto, mostra a lista.
func openInCode(target string) error {
	var dir string
	switch {
	case target == "":
		cwd, _ := os.Getwd()
		if rel, err := filepath.Rel(config.ProjDir, cwd); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			dir = filepath.Join(config.ProjDir, strings.Split(rel, string(filepath.Separator))[0])
		} else {
			var name string
			if err := pickProject(i18n.T("Abrir qual projeto no code-server?"), &name); err != nil {
				return err
			}
			dir = filepath.Join(config.ProjDir, name)
		}
	default:
		if p, err := projects.ProjectPath(target); err == nil {
			dir = p
		} else if abs, absErr := filepath.Abs(target); absErr == nil {
			if fi, statErr := os.Stat(abs); statErr == nil && fi.IsDir() {
				dir = abs
			}
		}
		if dir == "" {
			return i18n.Errorf("%q não é um projeto em %s nem uma pasta existente", target, config.ProjDir)
		}
	}
	if err := ensureCodeServer(); err != nil {
		return err
	}
	return browser.OpenURL(codeFolderURL(dir))
}
