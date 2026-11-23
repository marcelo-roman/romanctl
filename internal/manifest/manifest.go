package manifest

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"romanctl/internal/ui"
)

const FileName = "romanctl.yaml"

// Project is the romanctl.yaml written to every generated project.
type Project struct {
	Name      string    `yaml:"name" json:"name"`
	Stack     string    `yaml:"stack" json:"stack"`
	Template  string    `yaml:"template" json:"template"`
	Arch      []string  `yaml:"arch" json:"arch"`
	Services  []string  `yaml:"services" json:"services"`
	Run       string    `yaml:"run" json:"run"`
	Test      string    `yaml:"test" json:"test"`
	URL       string    `yaml:"url" json:"url"`
	Requires  []string  `yaml:"requires,omitempty" json:"requires,omitempty"`
	Env       Env       `yaml:"env,omitempty" json:"env,omitempty"`
	Publish   Publish   `yaml:"publish,omitempty" json:"publish,omitempty"`
	Processes []Process `yaml:"processes,omitempty" json:"processes,omitempty"`
}

// Process is one of the programs of a multi-process project, like a monorepo's API and frontend.
type Process struct {
	Name string            `yaml:"name" json:"name"`
	Dir  string            `yaml:"dir,omitempty" json:"dir,omitempty"`
	Run  string            `yaml:"run" json:"run"`
	URL  string            `yaml:"url,omitempty" json:"url,omitempty"`
	Env  map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

// URLs are the addresses that must respond for the project to be up.
func (p *Project) URLs() []string {
	var out []string
	if p.URL != "" {
		out = append(out, p.URL)
	}
	for _, proc := range p.Processes {
		if proc.URL != "" && proc.URL != p.URL {
			out = append(out, proc.URL)
		}
	}
	return out
}

// Env holds the variables the project reads from each service, as text/template over services.Conn.
type Env map[string]map[string]string

// Publish says which CI and Dockerfile set the project uses, and with which values.
type Publish struct {
	Kind string            `yaml:"kind,omitempty" json:"kind,omitempty"`
	Vars map[string]string `yaml:"vars,omitempty" json:"vars,omitempty"`
}

func Read(projectPath string) (*Project, bool) {
	b, err := os.ReadFile(filepath.Join(projectPath, FileName))
	if err != nil {
		return nil, false
	}
	var m Project
	if err := yaml.Unmarshal(b, &m); err != nil {
		ui.Warnf("invalid %s in %s: %v", FileName, projectPath, err)
		return nil, false
	}
	return &m, true
}
