// Package publish leva um projeto para o GitHub, com CI e Dockerfile.
package publish

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

//go:embed all:files
var embedded embed.FS

// Options são as escolhas do comando publish.
type Options struct {
	Public, NoCI, DryRun bool
	Description          string
}

// Result descreve o que foi feito ou, em uma simulação, o que seria feito.
type Result struct {
	OK        bool     `json:"ok"`
	DryRun    bool     `json:"dryRun"`
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Public    bool     `json:"public"`
	Exists    bool     `json:"remoteExists"`
	Remote    string   `json:"remote,omitempty"`
	URL       string   `json:"url,omitempty"`
	Generated []string `json:"generated"`
	Committed bool     `json:"committed"`
}

// fileData são os valores dos arquivos de CI e Dockerfile: Name, Lower e o que o template definir.
type fileData map[string]string

// Kinds são os conjuntos de CI e Dockerfile que existem.
func Kinds() []string {
	entries, _ := fs.ReadDir(embedded, "files")
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}

// KindOf diz qual conjunto de CI e Dockerfile serve para o projeto.
func KindOf(projectPath string) (string, fileData) {
	base := filepath.Base(projectPath)
	data := fileData{"Name": projects.DotnetName(base), "Lower": base, "Dll": "", "Main": "."}
	if m, ok := manifest.Read(projectPath); ok {
		scaffold.Complete(m)
		if m.Publish.Kind != "" {
			data["Name"], data["Lower"] = projects.DotnetName(m.Name), m.Name
			for k, v := range m.Publish.Vars {
				data[k] = v
			}
			return m.Publish.Kind, data
		}
	}
	switch projects.DetectType(projectPath) {
	case "next":
		return "next", data
	case "react/vite":
		return "react", data
	case "express", "node":
		return "node", data
	case "go":
		return "go", data
	case "dotnet":
		if files, _ := filepath.Glob(filepath.Join(projectPath, "*.csproj")); len(files) > 0 {
			data["Dll"] = strings.TrimSuffix(filepath.Base(files[0]), ".csproj")
		}
		return "dotnet", data
	}
	return "", data
}

func render(kind string, data fileData) (map[string][]byte, error) {
	out := map[string][]byte{}
	root, err := fs.Sub(embedded, path.Join("files", kind))
	if err != nil {
		return nil, err
	}
	err = fs.WalkDir(root, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(root, p)
		if err != nil {
			return err
		}
		// Os workflows usam ${{ }}, então aqui o delimitador é outro.
		tpl, err := template.New(p).Delims("[[", "]]").Option("missingkey=error").Parse(string(b))
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return err
		}
		out[strings.ReplaceAll(p, "dot-", ".")] = buf.Bytes()
		return nil
	})
	return out, err
}

func remoteURL(projectPath string) string {
	out, err := sys.OutputIn(projectPath, "git", "remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	return out
}

// Plan confere os pré-requisitos e descreve o que será feito, sem mudar nada.
func Plan(projectPath string, opts Options) (Result, error) {
	name := filepath.Base(projectPath)
	result := Result{Name: name, Path: projectPath, Public: opts.Public, DryRun: opts.DryRun, Generated: []string{}, Remote: remoteURL(projectPath)}
	result.Exists = result.Remote != ""
	kind, data := KindOf(projectPath)
	if kind == "" {
		return result, ui.Usagef("não sei publicar este tipo de projeto: %s", projectPath)
	}
	if !sys.Have("gh") {
		return result, i18n.Errorf("precisa do GitHub CLI: romanctl tools install gh")
	}
	if !sys.Have("git") {
		return result, i18n.Errorf("precisa do Git: romanctl tools install git")
	}
	if out, err := sys.Output("gh", "auth", "status"); err != nil {
		return result, i18n.Errorf("entre no GitHub antes: gh auth login (%s)", sys.Tail(out, 1))
	}
	if email, _ := sys.OutputIn(projectPath, "git", "config", "user.email"); email == "" {
		return result, i18n.Errorf("configure o Git antes: git config --global user.name \"Nome\" e git config --global user.email \"voce@exemplo.com\"")
	}
	if opts.NoCI {
		result.OK = opts.DryRun
		return result, nil
	}
	files, err := render(kind, data)
	if err != nil {
		return result, err
	}
	for p := range files {
		// O que já existe é da pessoa: nunca é sobrescrito.
		if _, err := os.Stat(filepath.Join(projectPath, filepath.FromSlash(p))); err != nil {
			result.Generated = append(result.Generated, p)
		}
	}
	sort.Strings(result.Generated)
	result.OK = opts.DryRun
	return result, nil
}

var sensitiveFile = regexp.MustCompile(`(^|/)(\.env(\.[^/]*)?|id_rsa|id_ed25519|credentials\.json|[^/]*\.(pem|key|p12|pfx))$`)

// sensitive lista arquivos que não devem ir para um repositório. Modelos de .env são a exceção.
func sensitive(names []string) []string {
	var found []string
	for _, n := range names {
		if sensitiveFile.MatchString(n) && !strings.HasSuffix(n, ".example") && !strings.HasSuffix(n, ".sample") && !strings.HasSuffix(n, ".template") {
			found = append(found, n)
		}
	}
	return found
}

// Run executa a publicação. Os segredos são conferidos antes do commit.
func Run(projectPath string, opts Options) (Result, error) {
	result, err := Plan(projectPath, opts)
	if err != nil {
		return result, err
	}
	result.OK = false
	kind, data := KindOf(projectPath)
	if !opts.NoCI {
		files, err := render(kind, data)
		if err != nil {
			return result, err
		}
		for _, p := range result.Generated {
			dest := filepath.Join(projectPath, filepath.FromSlash(p))
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return result, err
			}
			if err := os.WriteFile(dest, files[p], 0o644); err != nil {
				return result, err
			}
		}
	}
	git := func(args ...string) (string, error) { return sys.OutputIn(projectPath, "git", args...) }
	if _, err := os.Stat(filepath.Join(projectPath, ".git")); err != nil {
		if out, err := git("init", "-q", "-b", "main"); err != nil {
			return result, i18n.Errorf("git init: %v %s", err, out)
		}
	}
	if out, err := git("add", "-A"); err != nil {
		return result, i18n.Errorf("git add: %v %s", err, out)
	}
	staged, err := git("diff", "--cached", "--name-only")
	if err != nil {
		return result, i18n.Errorf("git diff: %v %s", err, staged)
	}
	if found := sensitive(strings.Fields(staged)); len(found) > 0 {
		git("reset", "-q")
		return result, i18n.Errorf("publicação cancelada: estes arquivos parecem ter segredos; coloque-os no .gitignore: %s", strings.Join(found, ", "))
	}
	if sys.Have("gitleaks") {
		if out, err := sys.OutputIn(projectPath, "gitleaks", "dir", ".", "--no-banner"); err != nil && strings.Contains(out, "leaks found") {
			git("reset", "-q")
			return result, i18n.Errorf("publicação cancelada: o gitleaks encontrou segredos no código:\n%s", sys.Tail(out, 6))
		}
	}
	if staged != "" {
		message := "Publica o projeto"
		if result.Exists {
			message = "Atualiza o projeto"
		}
		if out, err := git("commit", "-q", "-m", message); err != nil {
			return result, i18n.Errorf("git commit: %v %s", err, out)
		}
		result.Committed = true
	}
	if result.Exists {
		if out, err := git("push", "-u", "origin", "HEAD"); err != nil {
			return result, i18n.Errorf("git push: %v %s", err, sys.Tail(out, 4))
		}
		result.URL = result.Remote
		result.OK = true
		return result, nil
	}
	visibility := "--private"
	if opts.Public {
		visibility = "--public"
	}
	args := []string{"repo", "create", result.Name, visibility, "--source", ".", "--remote", "origin", "--push"}
	if opts.Description != "" {
		args = append(args, "--description", opts.Description)
	}
	out, err := sys.OutputIn(projectPath, "gh", args...)
	if err != nil {
		return result, i18n.Errorf("gh repo create: %v %s", err, sys.Tail(out, 4))
	}
	result.URL = out
	if view, err := sys.OutputIn(projectPath, "gh", "repo", "view", "--json", "url", "--jq", ".url"); err == nil && view != "" {
		result.URL = view
	}
	result.Remote = remoteURL(projectPath)
	result.OK = true
	return result, nil
}
