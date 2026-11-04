package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type ProjectType struct {
	Key    string
	Label  string
	create func(name, path string) error
}

var projectTypes = []ProjectType{
	{"node", "Node.js (básico)", createNode(false)},
	{"express", "API Node.js + Express", createNode(true)},
	{"react", "React + Vite (JavaScript)", createVite("react")},
	{"react-ts", "React + Vite (TypeScript)", createVite("react-ts")},
	{"next", "Next.js (TypeScript, App Router, Tailwind)", createNext},
	{"dotnet-webapi", ".NET Web API", createDotnet("webapi")},
	{"dotnet-console", ".NET Console", createDotnet("console")},
	{"dotnet-mvc", ".NET MVC", createDotnet("mvc")},
	{"dotnet-blazor", ".NET Blazor", createDotnet("blazor")},
	{"go", "Go (módulo + main.go)", createGo},
}

func findType(k string) *ProjectType {
	for i := range projectTypes {
		if projectTypes[i].Key == strings.ToLower(k) {
			return &projectTypes[i]
		}
	}
	return nil
}

var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

func validateName(n string) error {
	if !nameRe.MatchString(n) {
		return fmt.Errorf("use letras minúsculas, números, '-', '_' ou '.' (ex.: meu-app)")
	}
	if exists(filepath.Join(projDir, n)) {
		return fmt.Errorf("já existe um projeto chamado %q", n)
	}
	return nil
}

func newProject(typ, name string, git, forceOffline bool) error {
	t := findType(typ)
	if t == nil {
		return fmt.Errorf("tipo desconhecido %q (tipos: %s)", typ, typeKeys())
	}
	if err := validateName(name); err != nil {
		return err
	}
	os.MkdirAll(projDir, 0o755)
	path := filepath.Join(projDir, name)
	infof("Criando %s em %s ...", t.Label, path)
	offlineMode = forceOffline || !isOnline()
	if offlineMode {
		warnf("Sem internet: usando templates e cache locais")
	}
	var err error
	if offlineMode && isTemplated(t.Key) {
		err = createFromTemplate(t.Key, name, path)
	} else {
		err = t.create(name, path)
		if err != nil && !offlineMode && isTemplated(t.Key) && hasTemplate(t.Key) {
			warnf("O gerador falhou (%v). Tentando pelo template local...", err)
			os.RemoveAll(path)
			err = createFromTemplate(t.Key, name, path)
		}
	}
	if err != nil {
		os.RemoveAll(path) // não deixa projeto pela metade
		return fmt.Errorf("falha ao criar %s: %w", name, err)
	}
	if git && have("git") && !exists(filepath.Join(path, ".git")) {
		output("git", "-C", path, "init", "-q")
	}
	okf("Projeto %s criado!", sBoldTxt.Render(name))
	fmt.Printf("   Pasta:   %s\n   Rodar:   romanctl run %s\n   Abrir:   %s/?folder=%s\n",
		path, name, codeServerURL(), path)
	return nil
}

func typeKeys() string {
	var ks []string
	for _, t := range projectTypes {
		ks = append(ks, t.Key)
	}
	return strings.Join(ks, ", ")
}

func createNode(express bool) func(string, string) error {
	return func(name, path string) error {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
		if out, err := outputIn(path, "npm", "init", "-y"); err != nil {
			return fmt.Errorf("npm init: %v: %s", err, out)
		}
		index := "console.log('Olá do " + name + "!');\n"
		if express {
			if err := npmInstall(path, "express"); err != nil {
				return err
			}
			index = `const express = require('express');

const app = express();
app.use(express.json());

app.get('/', (req, res) => {
  res.json({ mensagem: 'Olá do ` + name + `!' });
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log('API rodando em http://localhost:' + PORT));
`
		}
		os.WriteFile(filepath.Join(path, "index.js"), []byte(index), 0o644)
		os.WriteFile(filepath.Join(path, ".gitignore"), []byte("node_modules/\n.env\n"), 0o644)
		return setScripts(filepath.Join(path, "package.json"), map[string]string{
			"dev": "node --watch index.js", "start": "node index.js"})
	}
}

func setScripts(pkgPath string, scripts map[string]string) error {
	b, err := os.ReadFile(pkgPath)
	if err != nil {
		return err
	}
	var pkg map[string]any
	if err := json.Unmarshal(b, &pkg); err != nil {
		return err
	}
	sc, _ := pkg["scripts"].(map[string]any)
	if sc == nil {
		sc = map[string]any{}
	}
	delete(sc, "test")
	for k, v := range scripts {
		sc[k] = v
	}
	pkg["scripts"] = sc
	out, _ := json.MarshalIndent(pkg, "", "  ")
	return os.WriteFile(pkgPath, append(out, '\n'), 0o644)
}

func createVite(tpl string) func(string, string) error {
	return func(name, path string) error {
		if err := run(filepath.Dir(path), nil, "npm", "create", "vite@latest", filepath.Base(path), "--", "--template", tpl, "--no-interactive"); err != nil {
			return err
		}
		return run(path, nil, "npm", "install")
	}
}

// O layout padrão baixa fontes do Google ao compilar, o que quebra o "next build"
// sem internet. O pacote geist traz a mesma fonte em arquivos locais.
const nextLayout = `import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";
import "./globals.css";

export const metadata: Metadata = {
  title: "Create Next App",
  description: "Generated by create next app",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html
      lang="pt-BR"
      className={` + "`${GeistSans.variable} ${GeistMono.variable} h-full antialiased`" + `}
    >
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
`

func createNext(name, path string) error {
	if err := run(filepath.Dir(path), []string{"NEXT_TELEMETRY_DISABLED=1"}, "npx", "--yes", "create-next-app@latest", filepath.Base(path),
		"--ts", "--app", "--tailwind", "--eslint", "--src-dir", "--use-npm", "--disable-git", "--import-alias", "@/*", "--yes"); err != nil {
		return err
	}
	if err := npmInstall(path, "geist"); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, "src", "app", "layout.tsx"), []byte(nextLayout), 0o644)
}

func createDotnet(tpl string) func(string, string) error {
	return func(name, path string) error {
		if err := run(filepath.Dir(path), []string{"DOTNET_NOLOGO=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1"},
			"dotnet", "new", tpl, "-o", path, "-n", dotnetName(name)); err != nil {
			return err
		}
		if !exists(filepath.Join(path, ".gitignore")) {
			output("dotnet", "new", "gitignore", "-o", path)
		}
		return nil
	}
}

// Nomes .NET não aceitam '-' no namespace; meu-app vira MeuApp.
func dotnetName(n string) string {
	parts := strings.FieldsFunc(n, func(r rune) bool { return r == '-' || r == '_' || r == '.' })
	for i, p := range parts {
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func createGo(name, path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	if out, err := outputIn(path, "go", "mod", "init", name); err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}
	main := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Olá do " + name + "!\")\n}\n"
	return os.WriteFile(filepath.Join(path, "main.go"), []byte(main), 0o644)
}

// ---------------------------------------------------------------- listar / remover / rodar

type Project struct {
	Name, Path, Type string
	Mod              time.Time
}

func detectType(p string) string {
	if m, ok := readManifest(p); ok {
		return m.Stack + "/" + m.Template
	}
	if b, err := os.ReadFile(filepath.Join(p, "package.json")); err == nil {
		s := string(b)
		switch {
		case strings.Contains(s, `"next"`):
			return "next"
		case strings.Contains(s, `"vite"`):
			return "react/vite"
		case strings.Contains(s, `"express"`):
			return "express"
		}
		return "node"
	}
	if m, _ := filepath.Glob(filepath.Join(p, "*.csproj")); len(m) > 0 {
		return "dotnet"
	}
	if exists(filepath.Join(p, "go.mod")) {
		return "go"
	}
	return "?"
}

func listProjects() []Project {
	entries, _ := os.ReadDir(projDir)
	var ps []Project
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		p := filepath.Join(projDir, e.Name())
		info, _ := e.Info()
		ps = append(ps, Project{Name: e.Name(), Path: p, Type: detectType(p), Mod: info.ModTime()})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].Mod.After(ps[j].Mod) })
	return ps
}

func printProjects() {
	ps := listProjects()
	if len(ps) == 0 {
		infof("Nenhum projeto em %s. Crie um com: romanctl new", projDir)
		return
	}
	fmt.Println(sTitle.Render("Projetos em " + projDir))
	fmt.Println(sGray.Render(fmt.Sprintf("  %-28s %-24s %s", "NOME", "TIPO", "MODIFICADO")))
	for _, p := range ps {
		fmt.Printf("  %-28s %-24s %s\n", p.Name, p.Type, p.Mod.Format("02/01/2006 15:04"))
	}
}

func projectsJSON() []map[string]any {
	out := []map[string]any{}
	for _, p := range listProjects() {
		item := map[string]any{"name": p.Name, "path": p.Path, "type": p.Type,
			"modified": p.Mod.Format(time.RFC3339), "url": appURL(p.Path)}
		if m, ok := readManifest(p.Path); ok {
			item["manifest"] = m
		}
		out = append(out, item)
	}
	return out
}

func projectPath(name string) (string, error) {
	if name == "" || strings.ContainsAny(name, "/\\") || name == "." || name == ".." {
		return "", fmt.Errorf("nome de projeto inválido: %q", name)
	}
	p := filepath.Join(projDir, name)
	if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
		return "", fmt.Errorf("projeto %q não encontrado em %s", name, projDir)
	}
	return p, nil
}

func removeProject(name string) error {
	p, err := projectPath(name)
	if err != nil {
		return err
	}
	infof("Removendo %s (%s)...", p, human(dirSize(p)))
	if err := os.RemoveAll(p); err != nil {
		return err
	}
	okf("Projeto %s removido", name)
	return nil
}

// runProject roda o projeto em primeiro plano; Ctrl+C encerra só o projeto.
func runProject(name string, open bool) error {
	p, err := projectPath(name)
	if err != nil {
		return err
	}
	var cmd []string
	manifest, hasManifest := readManifest(p)
	switch t := detectType(p); {
	case hasManifest:
		cmd = strings.Fields(manifest.Run)
		for _, key := range manifest.Services {
			if s := findService(key); s != nil && !s.Running() {
				warnf("%s não está rodando e o projeto depende dele: romanctl start %s", s.Name, s.Key)
			}
		}
	case t == "dotnet":
		cmd = []string{"dotnet", "run"}
	case t == "go":
		cmd = []string{"go", "run", "."}
	case t != "?":
		if !exists(filepath.Join(p, "node_modules")) {
			infof("Instalando dependências (npm install)...")
			offlineMode = !isOnline()
			if err := npmInstall(p); err != nil {
				return err
			}
		}
		cmd = []string{"npm", "run", "dev"}
		if b, _ := os.ReadFile(filepath.Join(p, "package.json")); !strings.Contains(string(b), `"dev"`) {
			cmd = []string{"npm", "start"}
		}
	default:
		return fmt.Errorf("não sei rodar %q (tipo não reconhecido)", name)
	}
	markProject(name, true)
	defer markProject(name, false)
	infof("Rodando %s em %s  (Ctrl+C para parar)", strings.Join(cmd, " "), p)
	if open {
		openWhenUp(p)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	err = run(p, nil, cmd[0], cmd[1:]...)
	select {
	case <-sig:
		fmt.Println()
		okf("Projeto parado")
		return nil
	default:
		return err
	}
}
