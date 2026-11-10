package cli

import (
	"fmt"
	"slices"
	"strings"

	"romanctl/internal/i18n"
	"romanctl/internal/ui"
)

// fullScreen são os comandos que tomam a tela: na volta, a lista aparece direto, sem pausa.
var fullScreen = []string{"ui", "menu"}

// commandOptions monta a lista a partir da mesma tabela da ajuda, nas seções de romanctl help.
func commandOptions() []ui.Option {
	var out []ui.Option
	for _, g := range usageGroups {
		for _, top := range g.paths {
			for _, c := range commands {
				if c.path != top && !strings.HasPrefix(c.path, top+" ") {
					continue
				}
				label := fmt.Sprintf("%-18s %s", c.path, ui.SGray.Render(c.summary.String()))
				out = append(out, ui.Option{Label: label, Value: c.path, Section: g.title.String(),
					Search: strings.Join(c.aliases, " ") + " " + localized(c.usage)})
			}
		}
	}
	return out
}

// commandMenu é o que romanctl faz sem argumentos: a lista de comandos, com filtro, e Enter
// roda o escolhido, perguntando o que faltar.
func commandMenu() error {
	cursor := ""
	for {
		fmt.Print("\033[H\033[2J")
		picked, err := ui.Pick(ui.PickConfig{Title: i18n.T("Comandos"), Initial: cursor, NoLeft: true, Options: commandOptions(),
			Subtitle: "romanctl " + version + "  " + i18n.T("digite para filtrar, Enter roda, Esc sai; romanctl ui abre o painel")})
		if err != nil {
			if ui.IsAbort(err) {
				fmt.Print("\033[H\033[2J")
				return nil
			}
			return err
		}
		cursor = picked[0]
		words := strings.Fields(cursor)
		fmt.Printf("\n%s\n\n", ui.STitle.Render("romanctl "+cursor))
		err = dispatch(words, map[string]bool{})
		if err != nil && !ui.IsAbort(err) {
			ui.Failf("%v", err)
		}
		if !slices.Contains(fullScreen, words[0]) || err != nil {
			pause()
		}
	}
}
