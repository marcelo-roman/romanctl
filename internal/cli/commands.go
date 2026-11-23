package cli

import (
	"fmt"
	"slices"
	"strings"

	"romanctl/internal/i18n"
	"romanctl/internal/ui"
)

// fullScreen lists the commands that take over the screen: on return, the list shows right away, without a pause.
var fullScreen = []string{"ui", "menu"}

// commandOptions builds the list from the same table as the help, in the sections of romanctl help.
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

// commandMenu is what romanctl does without arguments: the filterable command list, where Enter
// runs the chosen one, asking for whatever is missing.
func commandMenu() error {
	cursor := ""
	for {
		fmt.Print("\033[H\033[2J")
		picked, err := ui.Pick(ui.PickConfig{Title: i18n.T("Commands"), Initial: cursor, NoLeft: true, Options: commandOptions(),
			Subtitle: "romanctl " + version + "  " + i18n.T("type to filter, Enter runs, Esc exits; romanctl ui opens the panel")})
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
