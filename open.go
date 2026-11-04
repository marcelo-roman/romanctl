package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
)

const termuxPrefix = "/data/data/com.termux/files/usr"

// openURL abre a URL no navegador do Android via termux-open-url.
func openURL(u string) error {
	bin := termuxPrefix + "/bin/termux-open-url"
	if !exists(bin) {
		warnf("termux-open-url não encontrado; abra manualmente: %s", u)
		return nil
	}
	env := []string{"PATH=" + termuxPrefix + "/bin:" + os.Getenv("PATH"), "PREFIX=" + termuxPrefix, "LD_LIBRARY_PATH="}
	if out, err := outputEnv(env, bin, u); err != nil {
		return fmt.Errorf("não consegui abrir o navegador (%v) %s — abra manualmente: %s", err, out, u)
	}
	okf("Abrindo %s", sBlue.Render(u))
	return nil
}

// appURL descobre em que endereço o servidor de desenvolvimento do projeto escuta.
func appURL(p string) string {
	switch detectType(p) {
	case "react/vite":
		return "http://localhost:5173"
	case "next", "express", "node":
		return "http://localhost:3000"
	case "dotnet":
		b, _ := os.ReadFile(filepath.Join(p, "Properties", "launchSettings.json"))
		if m := regexp.MustCompile(`http://localhost:\d+`).Find(b); m != nil {
			return string(m)
		}
		return "http://localhost:5000"
	}
	return ""
}

func codeFolderURL(p string) string {
	return codeServerURL() + "/?folder=" + url.QueryEscape(p)
}

func ensureCodeServer() error {
	cs := findService("code")
	if !cs.Running() {
		return cs.Start()
	}
	return nil
}

// openTarget: "code", um projeto, ou uma URL. app=true abre o app rodando do projeto.
func openTarget(target string, app bool) error {
	switch {
	case strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://"):
		return openURL(target)
	case !app && findService(target) == findService("code"):
		if err := ensureCodeServer(); err != nil {
			return err
		}
		return openURL(codeServerURL())
	}
	p, err := projectPath(target)
	if err != nil {
		return err
	}
	if app {
		u := appURL(p)
		if u == "" {
			return fmt.Errorf("não sei em que porta %q roda", target)
		}
		if !httpUp(u) {
			warnf("%s não está respondendo em %s — rode antes: romanctl run %s", target, u, target)
			return nil
		}
		return openURL(u)
	}
	if err := ensureCodeServer(); err != nil {
		return err
	}
	return openURL(codeFolderURL(p))
}

// openWhenUp espera o app do projeto responder e então abre o navegador.
func openWhenUp(p string) {
	u := appURL(p)
	if u == "" {
		return
	}
	go func() {
		if waitUntil(3*time.Minute, func() bool { return httpUp(u) }) {
			openURL(u)
		}
	}()
}

func menuOpen() error {
	opts := []huh.Option[string]{huh.NewOption("code-server", "code|")}
	for _, p := range listProjects() {
		opts = append(opts, huh.NewOption("📁 "+p.Name+sGray.Render("  no code-server"), p.Name+"|"))
		if u := appURL(p.Path); u != "" && httpUp(u) {
			opts = append(opts, huh.NewOption("🌐 "+p.Name+sGreen.Render("  app rodando em "+u), p.Name+"|app"))
		}
	}
	var sel string
	if err := ask(true, huh.NewSelect[string]().Title("Abrir no navegador").Description(navHint).
		Options(opts...).Value(&sel)); err != nil {
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
		if rel, err := filepath.Rel(projDir, cwd); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			dir = filepath.Join(projDir, strings.Split(rel, string(filepath.Separator))[0])
		} else {
			var name string
			if err := pickProject("Abrir qual projeto no code-server?", &name); err != nil {
				return err
			}
			dir = filepath.Join(projDir, name)
		}
	default:
		if p, err := projectPath(target); err == nil {
			dir = p
		} else if abs, absErr := filepath.Abs(target); absErr == nil {
			if fi, statErr := os.Stat(abs); statErr == nil && fi.IsDir() {
				dir = abs
			}
		}
		if dir == "" {
			return fmt.Errorf("%q não é um projeto em %s nem uma pasta existente", target, projDir)
		}
	}
	if err := ensureCodeServer(); err != nil {
		return err
	}
	return openURL(codeFolderURL(dir))
}
