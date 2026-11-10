package webui

import (
	"fmt"
	"regexp"
	"strings"

	"romanctl/internal/i18n"
)

// ActionRequest é o corpo de POST /api/jobs. Só as ações abaixo existem, e cada campo é conferido
// antes de virar argumento: nada chega à CLI como opção (começando por -) nem passa por shell.
type ActionRequest struct {
	Action  string   `json:"action"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Stack   string   `json:"stack"`
	Arch    []string `json:"arch"`
	With    []string `json:"with"`
	Git     *bool    `json:"git"`
	Kind    string   `json:"kind"`
	Entity  string   `json:"entity"`
	In      string   `json:"in"`
	Plural  string   `json:"plural"`
	Key     string   `json:"key"`
	Version string   `json:"version"`
}

var (
	reProject = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,62}$`)
	reKey     = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,62}$`)
	reEntity  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,62}$`)
	reVersion = regexp.MustCompile(`^[vV]?[0-9][0-9A-Za-z._+-]{0,40}$`)
)

type action struct {
	args  func(r ActionRequest) ([]string, error)
	title func(r ActionRequest) string
}

var actions = map[string]action{
	"project.create": {args: createArgs, title: func(r ActionRequest) string { return i18n.Sprintf("Criar o projeto %s", r.Name) }},
	"project.up":     {args: projectArgs("up", "-d"), title: func(r ActionRequest) string { return i18n.Sprintf("Subir %s", r.Name) }},
	"project.down":   {args: projectArgs("down"), title: func(r ActionRequest) string { return i18n.Sprintf("Derrubar %s", r.Name) }},
	"project.remove": {args: projectArgs("rm", "--yes"), title: func(r ActionRequest) string { return i18n.Sprintf("Remover o projeto %s", r.Name) }},
	"project.generate": {args: generateArgs, title: func(r ActionRequest) string {
		return i18n.Sprintf("Gerar %s %s em %s", r.Kind, r.Entity, r.Name)
	}},
	"service.install": {args: serviceArgs("install"), title: func(r ActionRequest) string { return i18n.Sprintf("Instalar %s", r.Key) }},
	"service.start":   {args: serviceArgs("start"), title: func(r ActionRequest) string { return i18n.Sprintf("Ligar %s", r.Key) }},
	"service.stop":    {args: serviceArgs("stop", "--yes"), title: func(r ActionRequest) string { return i18n.Sprintf("Desligar %s", r.Key) }},
	"service.restart": {args: serviceArgs("restart", "--yes"), title: func(r ActionRequest) string { return i18n.Sprintf("Reiniciar %s", r.Key) }},
	"tool.install":    {args: toolInstallArgs, title: func(r ActionRequest) string { return i18n.Sprintf("Instalar %s", r.Key) }},
	"tool.uninstall": {args: func(r ActionRequest) ([]string, error) {
		if err := check(reKey, "key", r.Key); err != nil {
			return nil, err
		}
		return []string{"tools", "uninstall", r.Key, "--yes"}, nil
	}, title: func(r ActionRequest) string { return i18n.Sprintf("Remover %s", r.Key) }},
	"templates.update": {args: func(ActionRequest) ([]string, error) { return []string{"templates", "update"}, nil },
		title: func(ActionRequest) string { return i18n.T("Atualizar os templates") }},
}

func check(re *regexp.Regexp, field, value string) error {
	if !re.MatchString(value) {
		return fmt.Errorf("%s %q", i18n.T("valor inválido em"), field+"="+value)
	}
	return nil
}

func checkAll(re *regexp.Regexp, field string, values []string) error {
	for _, v := range values {
		if err := check(re, field, v); err != nil {
			return err
		}
	}
	return nil
}

func projectArgs(verb string, extra ...string) func(ActionRequest) ([]string, error) {
	return func(r ActionRequest) ([]string, error) {
		if err := check(reProject, "name", r.Name); err != nil {
			return nil, err
		}
		return append([]string{verb, r.Name}, extra...), nil
	}
}

func serviceArgs(verb string, extra ...string) func(ActionRequest) ([]string, error) {
	return func(r ActionRequest) ([]string, error) {
		if err := check(reKey, "key", r.Key); err != nil {
			return nil, err
		}
		return append([]string{verb, r.Key}, extra...), nil
	}
}

func createArgs(r ActionRequest) ([]string, error) {
	if err := check(reProject, "name", r.Name); err != nil {
		return nil, err
	}
	var args []string
	switch {
	case r.Type != "":
		if err := check(reKey, "type", r.Type); err != nil {
			return nil, err
		}
		args = []string{"new", r.Type, r.Name}
	case r.Stack != "":
		if err := check(reKey, "stack", r.Stack); err != nil {
			return nil, err
		}
		if len(r.Arch) == 0 {
			return nil, fmt.Errorf("%s", i18n.T("informe a arquitetura do template"))
		}
		if err := checkAll(reKey, "arch", r.Arch); err != nil {
			return nil, err
		}
		args = []string{"new", r.Name, "--stack", r.Stack, "--arch", strings.Join(r.Arch, ",")}
		if len(r.With) > 0 {
			if err := checkAll(reKey, "with", r.With); err != nil {
				return nil, err
			}
			args = append(args, "--with", strings.Join(r.With, ","))
		}
	default:
		return nil, fmt.Errorf("%s", i18n.T("informe o tipo ou o template do projeto"))
	}
	if r.Git != nil && !*r.Git {
		args = append(args, "--sem-git")
	}
	return args, nil
}

func generateArgs(r ActionRequest) ([]string, error) {
	if err := check(reProject, "name", r.Name); err != nil {
		return nil, err
	}
	if err := check(reKey, "kind", r.Kind); err != nil {
		return nil, err
	}
	if err := check(reEntity, "entity", r.Entity); err != nil {
		return nil, err
	}
	args := []string{"add", r.Kind, r.Entity, "--project", r.Name}
	if r.In != "" {
		if err := check(reEntity, "in", r.In); err != nil {
			return nil, err
		}
		args = append(args, "--in", r.In)
	}
	if r.Plural != "" {
		if err := check(reEntity, "plural", r.Plural); err != nil {
			return nil, err
		}
		args = append(args, "--plural", r.Plural)
	}
	return args, nil
}

func toolInstallArgs(r ActionRequest) ([]string, error) {
	if err := check(reKey, "key", r.Key); err != nil {
		return nil, err
	}
	args := []string{"tools", "install", r.Key}
	if r.Version != "" {
		if err := check(reVersion, "version", r.Version); err != nil {
			return nil, err
		}
		args = append(args, "--version", r.Version)
	}
	return args, nil
}
