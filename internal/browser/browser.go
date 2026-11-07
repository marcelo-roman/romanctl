// Package browser abre URLs no navegador da máquina.
package browser

import (
	"os"
	"romanctl/internal/i18n"

	"romanctl/internal/platform"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
)

const termuxPrefix = "/data/data/com.termux/files/usr"

// opener escolhe o comando de cada ambiente; vazio significa nenhum disponível.
func opener(p platform.Info, have func(string) bool) []string {
	switch {
	case p.OS == "darwin":
		return []string{"open"}
	case have(termuxPrefix + "/bin/termux-open-url"):
		return []string{termuxPrefix + "/bin/termux-open-url"}
	case p.WSL && have("wslview"):
		return []string{"wslview"}
	case p.WSL && have("cmd.exe"):
		return []string{"cmd.exe", "/c", "start"}
	case have("xdg-open"):
		return []string{"xdg-open"}
	}
	return nil
}

// OpenURL abre a URL ou, sem navegador disponível, só informa o endereço.
func OpenURL(u string) error {
	command := opener(platform.Current, func(bin string) bool { return sys.Have(bin) || sys.Exists(bin) })
	if command == nil {
		ui.Warnf("Não achei como abrir o navegador; abra manualmente: %s", u)
		return nil
	}
	var env []string
	if command[0] == termuxPrefix+"/bin/termux-open-url" {
		env = []string{"PATH=" + termuxPrefix + "/bin:" + os.Getenv("PATH"), "PREFIX=" + termuxPrefix, "LD_LIBRARY_PATH="}
	}
	if out, err := sys.OutputEnv(env, command[0], append(command[1:], u)...); err != nil {
		return i18n.Errorf("não consegui abrir o navegador (%v) %s — abra manualmente: %s", err, out, u)
	}
	ui.Okf("Abrindo %s", ui.SBlue.Render(u))
	return nil
}
