package manifest

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"romanctl/internal/ui"
)

const FileName = "romanctl.yaml"

// Project é o romanctl.yaml gravado em cada projeto gerado.
type Project struct {
	Name     string   `yaml:"name" json:"name"`
	Stack    string   `yaml:"stack" json:"stack"`
	Template string   `yaml:"template" json:"template"`
	Arch     []string `yaml:"arch" json:"arch"`
	Services []string `yaml:"services" json:"services"`
	Run      string   `yaml:"run" json:"run"`
	Test     string   `yaml:"test" json:"test"`
	URL      string   `yaml:"url" json:"url"`
	Requires []string `yaml:"requires,omitempty" json:"requires,omitempty"`
	Env      Env      `yaml:"env,omitempty" json:"env,omitempty"`
	Publish  Publish  `yaml:"publish,omitempty" json:"publish,omitempty"`
}

// Env são as variáveis que o projeto lê de cada serviço, em text/template sobre services.Conn.
type Env map[string]map[string]string

// Publish diz qual conjunto de CI e Dockerfile o projeto usa, e com que valores.
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
		ui.Warnf("%s inválido em %s: %v", FileName, projectPath, err)
		return nil, false
	}
	return &m, true
}
