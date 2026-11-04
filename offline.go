package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
)

// Pacotes populares guardados no cache para instalar sem internet.
var (
	offlineNpm = []string{
		"express", "cors", "dotenv", "zod", "axios", "nodemon", "typescript",
		"react-router-dom", "zustand", "@tanstack/react-query", "lucide-react", "recharts",
		"@dnd-kit/core", "@dnd-kit/sortable", "tailwindcss", "@tailwindcss/vite",
		"pg", "mongoose", "mongodb", "redis", "jsonwebtoken", "bcryptjs", "geist",
	}
	offlineGo = []string{
		"github.com/gin-gonic/gin", "github.com/go-chi/chi/v5", "github.com/jackc/pgx/v5",
		"go.mongodb.org/mongo-driver/v2/mongo", "github.com/redis/go-redis/v9", "github.com/joho/godotenv",
		"gorm.io/gorm", "gorm.io/driver/postgres", "github.com/spf13/cobra", "github.com/charmbracelet/huh",
	}
	offlineNuget = []string{
		"Npgsql.EntityFrameworkCore.PostgreSQL", "Microsoft.EntityFrameworkCore.Design",
		"MongoDB.Driver", "StackExchange.Redis", "Swashbuckle.AspNetCore", "Serilog.AspNetCore",
	}
	offlineExtensions = []string{
		"esbenp.prettier-vscode", "dbaeumer.vscode-eslint", "bradlc.vscode-tailwindcss",
		"golang.Go", "muhammad-sammy.csharp",
	}
	// Tipos cujo gerador precisa de internet; sem rede, são copiados de um template.
	offlineTemplates = []string{"react", "react-ts", "next"}
)

var (
	offlineRoot  = filepath.Join(dataDir, "offline")
	templatesDir = filepath.Join(offlineRoot, "templates")
	prepStamp    = filepath.Join(offlineRoot, "preparado-em")
)

// offlineMode é definido por "new": faz o npm usar só o cache.
var offlineMode bool

func isOnline() bool {
	if os.Getenv("ROMANCTL_OFFLINE") != "" {
		return false
	}
	c, err := net.DialTimeout("tcp", "registry.npmjs.org:443", 2*time.Second)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

func npmInstall(dir string, pkgs ...string) error {
	args := append([]string{"install"}, pkgs...)
	if offlineMode {
		args = append(args, "--offline")
	}
	return run(dir, nil, "npm", args...)
}

func hasTemplate(key string) bool {
	return exists(filepath.Join(templatesDir, key, "package.json"))
}

func isTemplated(key string) bool {
	for _, k := range offlineTemplates {
		if k == key {
			return true
		}
	}
	return false
}

// createFromTemplate copia um template pronto, com node_modules, e ajusta o nome.
func createFromTemplate(key, name, path string) error {
	if !hasTemplate(key) {
		return fmt.Errorf("sem internet e sem template de %q — rode com internet: romanctl offline prep", key)
	}
	infof("Usando o template offline de %s", key)
	if out, err := output("cp", "-a", filepath.Join(templatesDir, key), path); err != nil {
		return fmt.Errorf("cópia do template: %v %s", err, out)
	}
	if err := renameNpmProject(path, key, name); err != nil {
		return err
	}
	now := time.Now()
	return os.Chtimes(path, now, now)
}

func renameNpmProject(path, from, to string) error {
	re := regexp.MustCompile(`"name": "` + regexp.QuoteMeta(from) + `"`)
	for file, count := range map[string]int{"package.json": 1, "package-lock.json": 2} {
		p := filepath.Join(path, file)
		b, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		n := 0
		b = re.ReplaceAllFunc(b, func(m []byte) []byte {
			if n++; n > count {
				return m
			}
			return []byte(`"name": "` + to + `"`)
		})
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return err
		}
	}
	html := filepath.Join(path, "index.html")
	if b, err := os.ReadFile(html); err == nil {
		b = regexp.MustCompile(`<title>.*</title>`).ReplaceAll(b, []byte("<title>"+to+"</title>"))
		os.WriteFile(html, b, 0o644)
	}
	return nil
}

// ---------------------------------------------------------------- prep

func offlinePrep(force bool) error {
	if !isOnline() {
		return fmt.Errorf("a preparação precisa de internet")
	}
	os.MkdirAll(templatesDir, 0o755)
	var failed []string
	step := func(title string, f func() error) {
		fmt.Println("\n" + sTitle.Render(title))
		if err := f(); err != nil {
			failf("%v", err)
			failed = append(failed, title)
			return
		}
		okf("Concluído")
	}
	step("Templates de projeto", func() error { return prepTemplates(force) })
	step("Pacotes npm", prepNpm)
	step("Módulos Go", prepGo)
	step("Pacotes NuGet", prepNuget)
	step("Extensões do code-server", prepExtensions)
	if len(failed) > 0 {
		return fmt.Errorf("etapas com falha: %s. Rode de novo com internet", strings.Join(failed, ", "))
	}
	os.WriteFile(prepStamp, []byte(time.Now().Format("02/01/2006 15:04")), 0o644)
	fmt.Println()
	okf("Tudo preparado. Confira com: romanctl offline test")
	return nil
}

func prepTemplates(force bool) error {
	for _, key := range offlineTemplates {
		dst := filepath.Join(templatesDir, key)
		if hasTemplate(key) && !force {
			infof("%s já existe (use --force para refazer)", key)
			continue
		}
		infof("Gerando %s...", key)
		tmp, err := os.MkdirTemp(offlineRoot, "gerando-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmp)
		path := filepath.Join(tmp, key)
		if err := findType(key).create(key, path); err != nil {
			return fmt.Errorf("template %s: %w", key, err)
		}
		os.RemoveAll(filepath.Join(path, ".next"))
		os.RemoveAll(filepath.Join(path, ".git"))
		os.RemoveAll(dst)
		if err := os.Rename(path, dst); err != nil {
			return err
		}
	}
	return nil
}

func prepNpm() error {
	tmp, err := os.MkdirTemp("", "romanctl-npm-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if out, err := outputIn(tmp, "npm", "init", "-y"); err != nil {
		return fmt.Errorf("npm init: %v %s", err, out)
	}
	infof("Baixando %d pacotes...", len(offlineNpm))
	args := append([]string{"install", "--legacy-peer-deps", "--ignore-scripts", "--no-audit", "--no-fund"}, offlineNpm...)
	return run(tmp, nil, "npm", args...)
}

func prepGo() error {
	// Tenta a rede e, se falhar por qualquer motivo, usa o próprio cache como proxy.
	proxy := "https://proxy.golang.org|file://" + goModCache() + "/cache/download|direct"
	if out, err := output("go", "env", "-w", "GOPROXY="+proxy); err != nil {
		return fmt.Errorf("go env: %v %s", err, out)
	}
	infof("GOPROXY configurado com fallback para o cache local")
	tmp, err := os.MkdirTemp("", "romanctl-go-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if out, err := outputIn(tmp, "go", "mod", "init", "cache"); err != nil {
		return fmt.Errorf("go mod init: %v %s", err, out)
	}
	infof("Baixando %d módulos...", len(offlineGo))
	if out, err := outputIn(tmp, "go", append([]string{"get"}, offlineGo...)...); err != nil {
		return fmt.Errorf("go get: %v %s", err, out)
	}
	if out, err := outputIn(tmp, "go", "mod", "download", "all"); err != nil {
		return fmt.Errorf("go mod download: %v %s", err, out)
	}
	infof("Instalando gopls (usado pela extensão de Go; pode demorar)...")
	if out, err := outputIn(tmp, "go", "install", "golang.org/x/tools/gopls@latest"); err != nil {
		return fmt.Errorf("gopls: %v %s", err, out)
	}
	return nil
}

func goModCache() string {
	out, err := output("go", "env", "GOMODCACHE")
	if err != nil || out == "" {
		return filepath.Join(home, "go", "pkg", "mod")
	}
	return out
}

func prepNuget() error {
	tmp, err := os.MkdirTemp("", "romanctl-nuget-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	env := []string{"DOTNET_NOLOGO=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1"}
	for _, tpl := range []string{"webapi", "console", "mvc", "blazor"} {
		infof("Restaurando o template %s...", tpl)
		if err := quiet(tmp, env, "dotnet", "new", tpl, "-o", filepath.Join(tmp, tpl)); err != nil {
			return fmt.Errorf("dotnet new %s: %w", tpl, err)
		}
	}
	infof("Baixando %d pacotes...", len(offlineNuget))
	for _, pkg := range offlineNuget {
		if err := quiet(filepath.Join(tmp, "webapi"), env, "dotnet", "add", "package", pkg); err != nil {
			return fmt.Errorf("%s: %w", pkg, err)
		}
	}
	return nil
}

func prepExtensions() error {
	if !have("code-server") {
		return fmt.Errorf("code-server não instalado")
	}
	var failed []string
	for _, ext := range offlineExtensions {
		infof("Instalando %s...", ext)
		if err := quiet("", nil, "code-server", "--install-extension", ext); err != nil {
			warnf("%s: %v", ext, err)
			failed = append(failed, ext)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("não instaladas: %s", strings.Join(failed, ", "))
	}
	return nil
}

// quiet esconde a saída do comando e só a mostra em caso de erro.
func quiet(dir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%s", err, tail(string(out), 12))
	}
	return nil
}

func tail(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------- status

func offlineStatus() {
	fmt.Println(sTitle.Render("Preparação offline"))
	if b, err := os.ReadFile(prepStamp); err == nil {
		fmt.Printf("  %-14s %s\n", "Preparado em", sGreen.Render(string(b)))
	} else {
		fmt.Printf("  %-14s %s\n", "Preparado em", sYellow.Render("nunca — rode: romanctl offline prep"))
	}
	conn := sRed.Render("sem conexão")
	if isOnline() {
		conn = sGreen.Render("conectado")
	}
	fmt.Printf("  %-14s %s\n", "Internet", conn)
	var tpls []string
	for _, k := range offlineTemplates {
		if hasTemplate(k) {
			tpls = append(tpls, k)
		}
	}
	fmt.Printf("  %-14s %s\n", "Templates", strings.Join(tpls, ", "))
	for _, c := range [][2]string{
		{"Templates", templatesDir},
		{"Cache npm", filepath.Join(home, ".npm")},
		{"Cache Go", goModCache()},
		{"Cache NuGet", filepath.Join(home, ".nuget", "packages")},
	} {
		fmt.Printf("  %-14s %s\n", c[0], human(dirSize(c[1])))
	}
	if out, err := output("code-server", "--list-extensions"); err == nil {
		fmt.Printf("  %-14s %d instaladas\n", "Extensões", len(strings.Fields(out)))
	}

	fmt.Println("\n" + sTitle.Render("Sem internet"))
	fmt.Println("  romanctl new <tipo> <nome>             usa templates e cache sozinho")
	fmt.Println("  npm install <pacote> --offline         instala do cache")
	fmt.Println("  go get <módulo>                        usa o cache sozinho")
	fmt.Println("  dotnet add package <pacote>            usa o cache se o pacote foi baixado")

	fmt.Println("\n" + sTitle.Render("O que você precisa fazer manualmente"))
	fmt.Println("  • Abrir devdocs.io e ativar as documentações para uso offline")
	fmt.Println("  • Testar de verdade: ligar o modo avião e rodar romanctl offline test")
	fmt.Println("  • Sem internet, o Claude Code não funciona")
}

// ---------------------------------------------------------------- test

type offlineCheck struct {
	typ    string
	checks [][]string
}

func offlineTest() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	os.MkdirAll(offlineRoot, 0o755)
	tmp, err := os.MkdirTemp(offlineRoot, "teste-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	// Um proxy morto bloqueia a rede de npm, Go e .NET sem desligar o Wi-Fi.
	dead := "http://127.0.0.1:9"
	env := []string{"ROMANCTL_OFFLINE=1", "ROMANCTL_PROJETOS=" + tmp,
		"HTTPS_PROXY=" + dead, "HTTP_PROXY=" + dead, "https_proxy=" + dead, "http_proxy=" + dead,
		"NEXT_TELEMETRY_DISABLED=1", "DOTNET_NOLOGO=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1", "MSBUILDDISABLENODEREUSE=1"}
	dotnetBuild := []string{"dotnet", "build", "-nodeReuse:false", "-p:UseSharedCompilation=false"}

	cases := []offlineCheck{
		{"node", [][]string{{"node", "index.js"}}},
		{"express", [][]string{{"node", "-e", "require('express')"}}},
		{"react", [][]string{{"npm", "run", "build"}, {"npm", "install", "zustand", "--offline"}}},
		{"react-ts", [][]string{{"npm", "run", "build"}}},
		{"next", [][]string{{"npx", "--no-install", "next", "build"}}},
		{"go", [][]string{{"go", "build", "./..."}, {"go", "get", "github.com/go-chi/chi/v5"}}},
		{"dotnet-console", [][]string{dotnetBuild}},
		{"dotnet-webapi", [][]string{dotnetBuild, {"dotnet", "add", "package", "StackExchange.Redis"}}},
		{"dotnet-mvc", [][]string{dotnetBuild}},
		{"dotnet-blazor", [][]string{dotnetBuild}},
	}
	fmt.Println(sTitle.Render("Criando e compilando cada tipo com a rede bloqueada"))
	failures := 0
	for _, c := range cases {
		start := time.Now()
		name := "t-" + c.typ
		err := testStep(tmp, env, []string{self, "new", c.typ, name, "--sem-git", "--offline"})
		for i := 0; err == nil && i < len(c.checks); i++ {
			err = testStep(filepath.Join(tmp, name), env, c.checks[i])
		}
		if err != nil {
			failures++
			failf("%-16s %v", c.typ, err)
			continue
		}
		okf("%-16s %s", c.typ, sGray.Render(time.Since(start).Round(time.Second).String()))
		os.RemoveAll(filepath.Join(tmp, name))
	}
	fmt.Println()
	if failures > 0 {
		return fmt.Errorf("%d de %d tipos falharam. Rode com internet: romanctl offline prep --force", failures, len(cases))
	}
	okf("Todos os %d tipos funcionaram com a rede bloqueada", len(cases))
	infof("O teste só bloqueia a rede das ferramentas. Repita no modo avião antes do voo.")
	return nil
}

func testStep(dir string, env, command []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %v\n%s", strings.Join(command[min(1, len(command)-1):], " "), err, tail(string(out), 10))
	}
	return nil
}

// ---------------------------------------------------------------- CLI / menu

func offlineCommand(sub string, flags map[string]bool) error {
	switch sub {
	case "", "status":
		offlineStatus()
		return nil
	case "prep", "preparar":
		return offlinePrep(flags["force"])
	case "test", "testar":
		return offlineTest()
	}
	return fmt.Errorf("use: romanctl offline [status|prep|test]")
}

func menuOffline() error {
	var sub string
	if err := ask(true, huh.NewSelect[string]().Title("Modo offline").Description(navHint).
		Options(
			huh.NewOption("Ver o estado da preparação", "status"),
			huh.NewOption("Preparar agora (baixa tudo, precisa de internet)", "prep"),
			huh.NewOption("Testar com a rede bloqueada", "test"),
		).Value(&sub)); err != nil {
		return err
	}
	return offlineCommand(sub, nil)
}
