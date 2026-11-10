package scaffold

import (
	"strings"

	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
	"romanctl/internal/sys"
	"romanctl/internal/tools"
	"romanctl/internal/ui"
)

const (
	setupDone    = "done"
	setupSkipped = "skipped"
	setupFailed  = "failed"
	setupOffline = "offline"
)

// missingTools são as ferramentas do template que ainda não rodam nesta máquina.
func missingTools(keys []string) []string {
	var out []string
	for _, key := range keys {
		t := tools.Find(key)
		if t == nil {
			out = append(out, key)
			continue
		}
		if _, ok := t.Detected(); !ok {
			out = append(out, key)
		}
	}
	return out
}

// ensureTools oferece instalar o que falta e devolve o que continuou faltando.
func ensureTools(keys []string) []string {
	missing := missingTools(keys)
	if len(missing) == 0 {
		return nil
	}
	list := strings.Join(missing, " ")
	if !ui.Interactive() || ui.JSONOutput {
		ui.Warnf("O projeto precisa de %s. Instale com: romanctl tools install %s", list, list)
		return missing
	}
	ok, err := ui.Confirm(i18n.Sprintf("O projeto precisa de %s. Instalar agora?", list))
	if err != nil || !ok {
		ui.Warnf("O projeto precisa de %s. Instale com: romanctl tools install %s", list, list)
		return missing
	}
	for _, key := range missing {
		if t := tools.Find(key); t != nil {
			tools.InstallOne(t, false)
		}
	}
	return missingTools(keys)
}

// runSetup baixa as dependências do projeto recém-criado. Uma falha aqui não desfaz o projeto.
func runSetup(t *Template, dir string, offline bool, missing []string) string {
	if len(t.Setup) == 0 {
		return ""
	}
	if offline {
		ui.Infof("Sem internet, as dependências ficam para depois: cd %s && %s", dir, strings.Join(t.Setup, " && "))
		return setupOffline
	}
	if len(missing) > 0 {
		ui.Infof("As dependências ficam para depois: cd %s && %s", dir, strings.Join(t.Setup, " && "))
		return setupSkipped
	}
	for _, command := range t.Setup {
		var err error
		ui.Spin(i18n.Sprintf("Rodando %s...", command), func(func(string)) { err = sys.Quiet(dir, nil, "sh", "-c", command) })
		if err != nil {
			ui.Warnf("%s falhou; o projeto foi criado, rode de novo dentro da pasta:\n%v", command, err)
			return setupFailed
		}
	}
	return setupDone
}

// Complete preenche, a partir do template, o que um romanctl.yaml mais antigo não traz.
func Complete(m *manifest.Project) {
	t := FindByKey(m.Stack, m.Template)
	if t == nil {
		return
	}
	if len(m.Requires) == 0 {
		m.Requires = t.Requires
	}
	if m.Publish.Kind != "" {
		return
	}
	complete := buildManifest(t, m.Name, m.Services)
	m.Publish = complete.Publish
	if m.Env == nil {
		m.Env = complete.Env
	}
}
