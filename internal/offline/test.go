package offline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
	"time"
)

type offlineCheck struct {
	typ    string
	args   []string // argumentos de "new" quando o tipo vem de um template de arquitetura
	checks [][]string
}

func offlineTest() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	os.MkdirAll(projects.OfflineRoot, 0o755)
	tmp, err := os.MkdirTemp(projects.OfflineRoot, "teste-")
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
		{typ: "node", checks: [][]string{{"node", "index.js"}}},
		{typ: "express", checks: [][]string{{"node", "-e", "require('express')"}}},
		{typ: "react", checks: [][]string{{"npm", "run", "build"}, {"npm", "install", "zustand", "--offline"}}},
		{typ: "react-ts", checks: [][]string{{"npm", "run", "build"}}},
		{typ: "next", checks: [][]string{{"npx", "--no-install", "next", "build"}}},
		{typ: "go", checks: [][]string{{"go", "build", "./..."}, {"go", "get", "github.com/go-chi/chi/v5"}}},
		{typ: "dotnet-console", checks: [][]string{dotnetBuild}},
		{typ: "dotnet-webapi", checks: [][]string{dotnetBuild, {"dotnet", "add", "package", "StackExchange.Redis"}}},
		{typ: "dotnet-mvc", checks: [][]string{dotnetBuild}},
		{typ: "dotnet-blazor", checks: [][]string{dotnetBuild}},
		{typ: "clean-ddd-cqrs", args: []string{"--stack", "dotnet", "--arch", "clean,ddd,cqrs"}, checks: [][]string{dotnetBuild, {"dotnet", "test"}}},
		{typ: "clean-ddd-cqrs+pg", args: []string{"--stack", "dotnet", "--arch", "clean,ddd,cqrs", "--with", "postgres"}, checks: [][]string{dotnetBuild, {"dotnet", "test"}}},
	}
	fmt.Println(ui.STitle.Render(i18n.T("Criando e compilando cada tipo com a rede bloqueada")))
	failures := 0
	for _, c := range cases {
		start := time.Now()
		name := "t-" + strings.ReplaceAll(c.typ, "+", "-")
		command := []string{self, "new", c.typ, name, "--sem-git", "--offline"}
		if c.args != nil {
			command = append([]string{self, "new", name, "--sem-git"}, c.args...)
		}
		err := testStep(tmp, env, command)
		for i := 0; err == nil && i < len(c.checks); i++ {
			err = testStep(filepath.Join(tmp, name), env, c.checks[i])
		}
		if err != nil {
			failures++
			ui.Failf("%-18s %v", c.typ, err)
			continue
		}
		ui.Okf("%-18s %s", c.typ, ui.SGray.Render(time.Since(start).Round(time.Second).String()))
		os.RemoveAll(filepath.Join(tmp, name))
	}
	fmt.Println()
	if failures > 0 {
		return i18n.Errorf("%d de %d tipos falharam. Rode com internet: romanctl offline prep --force", failures, len(cases))
	}
	ui.Okf("Todos os %d tipos funcionaram com a rede bloqueada", len(cases))
	ui.Infof("O teste só bloqueia a rede das ferramentas. Repita no modo avião antes do voo.")
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
		return i18n.Errorf("%s: %v\n%s", strings.Join(command[min(1, len(command)-1):], " "), err, sys.Tail(string(out), 10))
	}
	return nil
}
