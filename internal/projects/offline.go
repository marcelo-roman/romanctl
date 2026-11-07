package projects

import (
	"os"
	"path/filepath"
	"regexp"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"time"
)

// Tipos cujo gerador precisa de internet; sem rede, são copiados de um template.
var OfflineTemplates = []string{"react", "react-ts", "next"}

var OfflineRoot = filepath.Join(config.DataDir, "offline")

var TemplatesDir = filepath.Join(OfflineRoot, "templates")

// offlineMode é definido por "new": faz o npm usar só o cache.
var offlineMode bool

func npmInstall(dir string, pkgs ...string) error {
	args := append([]string{"install"}, pkgs...)
	if offlineMode {
		args = append(args, "--offline")
	}
	return sys.Run(dir, nil, "npm", args...)
}

func HasTemplate(key string) bool {
	return sys.Exists(filepath.Join(TemplatesDir, key, "package.json"))
}

func isTemplated(key string) bool {
	for _, k := range OfflineTemplates {
		if k == key {
			return true
		}
	}
	return false
}

// createFromTemplate copia um template pronto, com node_modules, e ajusta o nome.
func createFromTemplate(key, name, path string) error {
	if !HasTemplate(key) {
		return i18n.Errorf("sem internet e sem template de %q — rode com internet: romanctl offline prep", key)
	}
	ui.Infof("Usando o template offline de %s", key)
	if out, err := sys.Output("cp", "-a", filepath.Join(TemplatesDir, key), path); err != nil {
		return i18n.Errorf("cópia do template: %v %s", err, out)
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
