package projects

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"sort"
	"strings"
	"time"
)

// appURL descobre em que endereço o servidor de desenvolvimento do projeto escuta.
func AppURL(p string) string {
	if m, ok := manifest.Read(p); ok {
		return m.URL
	}
	switch DetectType(p) {
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

type Project struct {
	Name, Path, Type string
	Mod              time.Time
}

func DetectType(p string) string {
	if m, ok := manifest.Read(p); ok {
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
	if sys.Exists(filepath.Join(p, "go.mod")) {
		return "go"
	}
	return "?"
}

func ListProjects() []Project {
	entries, _ := os.ReadDir(config.ProjDir)
	var ps []Project
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		p := filepath.Join(config.ProjDir, e.Name())
		info, _ := e.Info()
		ps = append(ps, Project{Name: e.Name(), Path: p, Type: DetectType(p), Mod: info.ModTime()})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].Mod.After(ps[j].Mod) })
	return ps
}

func PrintProjects() { PrintProjectsFiltered("") }

// PrintProjectsFiltered imprime os projetos que contêm o texto.
func PrintProjectsFiltered(filter string) {
	ps := ListProjects()
	if len(ps) == 0 {
		ui.Infof("Nenhum projeto em %s. Crie um com: romanctl new", config.ProjDir)
		return
	}
	fmt.Println(ui.STitle.Render(i18n.Sprintf("Projetos em %s", config.ProjDir)))
	fmt.Println(ui.SGray.Render(i18n.Sprintf("  %-28s %-24s %s", i18n.T("NOME"), i18n.T("TIPO"), i18n.T("MODIFICADO"))))
	for _, p := range ps {
		if filter != "" && !ui.Matches(p.Name+" "+p.Type, filter) {
			continue
		}
		i18n.Printf("  %-28s %-24s %s\n", p.Name, p.Type, p.Mod.Format("02/01/2006 15:04"))
	}
}

func ListJSON() []map[string]any {
	out := []map[string]any{}
	for _, p := range ListProjects() {
		item := map[string]any{"name": p.Name, "path": p.Path, "type": p.Type,
			"modified": p.Mod.Format(time.RFC3339), "url": AppURL(p.Path)}
		if m, ok := manifest.Read(p.Path); ok {
			item["manifest"] = m
		}
		out = append(out, item)
	}
	return out
}

func ProjectPath(name string) (string, error) {
	if name == "" || strings.ContainsAny(name, "/\\") || name == "." || name == ".." {
		return "", i18n.Errorf("nome de projeto inválido: %q", name)
	}
	p := filepath.Join(config.ProjDir, name)
	if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
		return "", i18n.Errorf("projeto %q não encontrado em %s", name, config.ProjDir)
	}
	return p, nil
}

func RemoveProject(name string) error {
	p, err := ProjectPath(name)
	if err != nil {
		return err
	}
	ui.Infof("Removendo %s (%s)...", p, sys.Human(sys.DirSize(p)))
	if err := os.RemoveAll(p); err != nil {
		return err
	}
	ui.Okf("Projeto %s removido", name)
	return nil
}

// Locate acha o projeto pelo nome ou, sem nome, subindo a partir da pasta atual.
func Locate(name string) (string, error) {
	if name != "" {
		return ProjectPath(name)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		if _, ok := manifest.Read(dir); ok || DetectType(dir) != "?" {
			return dir, nil
		}
		if dir == filepath.Dir(dir) {
			return "", ui.Usagef("nenhum projeto nesta pasta: entre em um ou use --project <nome>")
		}
	}
}
