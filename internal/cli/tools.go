package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"romanctl/internal/i18n"
	"slices"
	"strings"

	"romanctl/internal/platform"
	"romanctl/internal/tools"
	"romanctl/internal/ui"
)

func toolsCommand(args []string, flags map[string]bool) error {
	sub := arg(args, 0)
	if sub == "" && ui.Interactive() {
		var err error
		sub, err = ui.Choose(i18n.T("Ferramentas"),
			ui.Opt(i18n.T("Listar o que roda nesta máquina"), "list"),
			ui.Opt(i18n.T("Instalar"), "install"),
			ui.Opt(i18n.T("Remover"), "uninstall"))
		if err != nil {
			return err
		}
	}
	// Com o subcomando vindo do menu, não há o que cortar de args.
	rest := args[min(1, len(args)):]
	switch sub {
	case "", "list", "ls":
		return toolsList(flags["all"], flags["plain"])
	case "categories", "categorias":
		return toolsCategories()
	case "install", "instalar", "add":
		return toolsInstall(rest, installOptions{force: flags["force"], yes: flags["y"] || flags["yes"], dryRun: flags["dry-run"]})
	case "uninstall", "remove", "rm", "remover", "desinstalar":
		return toolsUninstall(rest, installOptions{yes: flags["y"] || flags["yes"], dryRun: flags["dry-run"]})
	}
	return ui.Usagef("use: romanctl tools [list|categories] | install [nomes|grupos|all] [--force] [--yes] [--dry-run] | uninstall [nomes|grupos] [--yes] [--dry-run]")
}

// toolsList mostra só o que roda nesta máquina; --all inclui o resto, com o motivo.
func toolsList(showAll, plain bool) error {
	filter := flagVals["filter"]
	categories, err := parseCategories(flagVals["category"])
	if err != nil {
		return err
	}
	if !plain && filter == "" && ui.Interactive() {
		return toolsBrowse(showAll, categories)
	}
	statuses := loadStatuses()
	visible := []tools.Status{}
	hidden := 0
	for _, s := range statuses {
		// Instalada aparece sempre, mesmo que a máquina não a suporte mais.
		if s.Unavailable != "" && !s.Installed && !showAll {
			hidden++
			continue
		}
		if len(categories) > 0 && !slices.Contains(categories, s.Group) {
			continue
		}
		if filter != "" && !ui.Matches(s.Key+" "+s.Name+" "+s.Group+" "+tools.GroupTitle(s.Group)+" "+s.Description+" "+s.Version, filter) {
			continue
		}
		visible = append(visible, s)
	}
	if ui.JSONOutput {
		return ui.PrintJSON(visible)
	}
	i18n.Printf("%s %s\n\n", ui.STitle.Render(i18n.T("Máquina")), platform.Current.Describe())
	for _, g := range tools.Groups {
		var rows []tools.Status
		for _, s := range visible {
			if s.Group == g.Key {
				rows = append(rows, s)
			}
		}
		if len(rows) == 0 {
			continue
		}
		i18n.Printf("%s %s\n", ui.STitle.Render(g.Title), ui.SGray.Render("("+g.Key+")"))
		for _, s := range rows {
			state := ui.SGray.Render("○ " + s.Description)
			switch {
			case s.Installed:
				state = ui.SGreen.Render("● ") + s.Version
			case s.Unavailable != "":
				state = ui.SGray.Render("✕ " + s.Unavailable)
			}
			i18n.Printf("  %-15s %s\n", s.Key, state)
		}
		fmt.Println()
	}
	if hidden > 0 {
		i18n.Printf("%d ferramentas não rodam nesta máquina e ficaram de fora (veja com: romanctl tools list --all)\n", hidden)
	}
	fmt.Println(i18n.T("Instalar: romanctl tools install <nomes|grupos|all> [--force]"))
	fmt.Println(i18n.T("Remover:  romanctl tools uninstall <nomes|grupos> [--yes]"))
	return nil
}

// parseCategories lê --category, que aceita vários nomes separados por vírgula.
func parseCategories(value string) ([]string, error) {
	var out []string
	for _, name := range strings.Split(value, ",") {
		if name = strings.TrimSpace(name); name == "" {
			continue
		}
		key, ok := tools.FindGroup(name)
		if !ok {
			return nil, ui.Usagef("categoria desconhecida %q (veja: romanctl tools categories)", name)
		}
		if !slices.Contains(out, key) {
			out = append(out, key)
		}
	}
	return out, nil
}

// toolCategories monta as abas, só com as categorias que têm o que mostrar.
func toolCategories(options []ui.Option) []ui.Category {
	var out []ui.Category
	for _, g := range tools.Groups {
		if slices.ContainsFunc(options, func(o ui.Option) bool { return o.Category == g.Key }) {
			out = append(out, ui.Category{Key: g.Key, Label: g.Title})
		}
	}
	if len(out) < 2 {
		// Numa lista de uma categoria só, buscar pelo nome dela casaria com todas as linhas.
		for i := range options {
			options[i].Search = ""
		}
		return nil
	}
	return out
}

func toolOption(label string, s tools.Status) ui.Option {
	return ui.Option{Label: label, Value: s.Key, Category: s.Group, Search: tools.GroupTitle(s.Group)}
}

// toolsCategories mostra as categorias e quantas ferramentas de cada uma rodam aqui.
func toolsCategories() error {
	type row struct {
		Key       string `json:"key"`
		Title     string `json:"title"`
		Available int    `json:"available"`
		Installed int    `json:"installed"`
	}
	rows := []row{}
	statuses := loadStatuses()
	for _, g := range tools.Groups {
		r := row{Key: g.Key, Title: g.Title}
		for _, s := range statuses {
			if s.Group != g.Key {
				continue
			}
			if s.Unavailable == "" || s.Installed {
				r.Available++
			}
			if s.Installed {
				r.Installed++
			}
		}
		rows = append(rows, r)
	}
	if ui.JSONOutput {
		return ui.PrintJSON(rows)
	}
	for _, r := range rows {
		i18n.Printf("  %-14s %-36s disponíveis: %d, instaladas: %d\n", r.Key, r.Title, r.Available, r.Installed)
	}
	return nil
}

func toolsInstall(names []string, opts installOptions) error {
	if len(names) == 0 {
		names = []string{"all"}
	}
	var err error
	// Em um terminal, cada grupo vira uma lista para escolher. Sem terminal, ou com
	// --yes, o grupo inteiro é instalado, porque ninguém responderia a um menu.
	if ui.Interactive() && !opts.yes {
		if names, err = expandGroups(names, opts.force); err != nil {
			return err
		}
		if len(names) == 0 {
			ui.Infof("Nada selecionado.")
			return nil
		}
	}
	selected, err := tools.Resolve(names)
	if err != nil {
		return err
	}
	if opts.dryRun {
		return printInstallPlan(selected)
	}
	results := []tools.Result{}
	var failed []string
	for _, t := range selected {
		r := tools.InstallOne(t, opts.force)
		results = append(results, r)
		if r.Action == "failed" || r.Action == "unavailable" {
			failed = append(failed, t.Key)
		}
	}
	if ui.JSONOutput {
		ui.PrintJSON(map[string]any{"ok": len(failed) == 0, "results": results})
	}
	if len(failed) > 0 {
		if ui.JSONOutput {
			return ui.ErrReported
		}
		return i18n.Errorf("não instaladas: %s", strings.Join(failed, ", "))
	}
	warnIfNotInPath()
	return nil
}

// Sem root, os binários vão para uma pasta que pode não estar no PATH do usuário.
func warnIfNotInPath() {
	bin := platform.Current.BinDir
	paths := filepath.SplitList(os.Getenv("ROMANCTL_ORIGINAL_PATH"))
	if slices.Contains(paths, bin) {
		return
	}
	ui.Warnf("Adicione ao PATH do seu shell: export PATH=\"%s:$PATH\"", bin)
}

type installOptions struct{ force, yes, dryRun bool }

// expandGroups troca cada grupo pelo que foi marcado na lista; nomes passam direto.
func expandGroups(names []string, force bool) ([]string, error) {
	var out []string
	for _, name := range names {
		keys := tools.GroupKeys(name)
		if keys == nil {
			out = append(out, name)
			continue
		}
		picked, err := pickFromGroup(name, keys, force)
		if err != nil {
			return nil, err
		}
		out = append(out, picked...)
	}
	return out, nil
}

func pickFromGroup(name string, keys []string, force bool) ([]string, error) {
	return pickTools(keys, i18n.T("Instalar quais de %s?"), i18n.T("Instalar quais ferramentas?"), name, func(s tools.Status) (string, bool) {
		return s.Description, s.Unavailable == "" && (!s.Installed || force)
	}, i18n.T("Tudo em %s já está instalado."))
}

// pickTools mostra as ferramentas aceitas por filter, com "Todos" no topo.
// pickTools mostra as ferramentas aceitas por filter, com "Todos" no topo.
func pickTools(keys []string, groupTitle, allTitle, name string, filter func(tools.Status) (string, bool), empty string) ([]string, error) {
	var opts []ui.Option
	for _, s := range loadStatuses() {
		detail, ok := filter(s)
		if !slices.Contains(keys, s.Key) || !ok {
			continue
		}
		opts = append(opts, toolOption(fmt.Sprintf("%-15s %s", s.Key, ui.SGray.Render(detail)), s))
	}
	if len(opts) == 0 {
		ui.Okf(empty, name)
		return nil, nil
	}
	title := i18n.Sprintf(groupTitle, name)
	if name == "all" {
		title = allTitle
	}
	return ui.Pick(ui.PickConfig{Title: title, Options: opts, Multi: true, All: true, Categories: toolCategories(opts)})
}

// loadStatuses mostra o progresso, porque consultar cada ferramenta pode demorar.
func loadStatuses() []tools.Status {
	var out []tools.Status
	ui.Spin(i18n.T("Verificando ferramentas..."), func(update func(string)) {
		out = tools.Statuses(func(done, total int) {
			update(i18n.Sprintf("Verificando ferramentas %d/%d", done, total))
		})
	})
	return out
}

func printInstallPlan(selected []*tools.Tool) error {
	var keys []string
	for _, t := range selected {
		keys = append(keys, t.Key)
	}
	if ui.JSONOutput {
		return ui.PrintJSON(map[string]any{"ok": true, "dryRun": true, "tools": keys})
	}
	ui.Infof("Simulação: nada foi instalado. Seriam instaladas: %d", len(keys))
	for _, t := range selected {
		line := "  " + t.Key
		if why := t.Unavailable(); why != "" {
			line += ui.SGray.Render(i18n.T("  (indisponível: ") + why + ")")
		}
		fmt.Println(line)
	}
	return nil
}

func toolsUninstall(names []string, opts installOptions) error {
	if len(names) == 0 {
		names = []string{"all"}
	}
	// Com lista e confirmação, "all" é seguro. Sem terminal, ou com --yes, seria apagar
	// os SDKs e o code-server sem ninguém ver a lista.
	interactive := ui.Interactive() && !opts.yes
	for _, name := range names {
		if !interactive && (strings.EqualFold(name, "all") || strings.EqualFold(name, "tudo")) {
			return ui.Usagef("sem a lista, uninstall não aceita %q, porque apagaria os SDKs e o code-server: informe nomes ou grupos", name)
		}
	}
	if interactive {
		var picked []string
		for _, name := range names {
			keys := tools.GroupKeys(name)
			if keys == nil {
				picked = append(picked, name)
				continue
			}
			chosen, err := pickTools(keys, i18n.T("Remover quais de %s?"), i18n.T("Remover quais ferramentas?"), name, func(s tools.Status) (string, bool) {
				return s.Version, s.Installed
			}, i18n.T("Nada instalado em %s."))
			if err != nil {
				return err
			}
			picked = append(picked, chosen...)
		}
		if len(picked) == 0 {
			ui.Infof("Nada selecionado.")
			return nil
		}
		names = picked
	}
	selected, err := tools.Resolve(names)
	if err != nil {
		return err
	}
	if opts.dryRun {
		return printRemovalPlan(selected)
	}
	if !opts.yes {
		var list []string
		for _, t := range selected {
			line := t.Name
			if t.Warning != "" {
				line += " — " + t.Warning
			}
			list = append(list, line)
		}
		ok, err := ui.Confirm(i18n.T("Remover?\n  ") + strings.Join(list, "\n  "))
		if err != nil {
			return err
		}
		if !ok {
			ui.Infof("Cancelado")
			return nil
		}
	}
	results := []tools.Result{}
	var failed []string
	for _, t := range selected {
		r := tools.UninstallOne(t)
		results = append(results, r)
		if r.Action == "failed" {
			failed = append(failed, t.Key)
		}
	}
	if ui.JSONOutput {
		ui.PrintJSON(map[string]any{"ok": len(failed) == 0, "results": results})
		if len(failed) > 0 {
			return ui.ErrReported
		}
		return nil
	}
	if len(failed) > 0 {
		return i18n.Errorf("não removidas: %s", strings.Join(failed, ", "))
	}
	return nil
}

func printRemovalPlan(selected []*tools.Tool) error {
	var keys []string
	for _, t := range selected {
		keys = append(keys, t.Key)
	}
	if ui.JSONOutput {
		return ui.PrintJSON(map[string]any{"ok": true, "dryRun": true, "tools": keys})
	}
	ui.Infof("Simulação: nada foi removido. Seriam removidas: %d", len(keys))
	for _, t := range selected {
		line := "  " + t.Key
		if t.Warning != "" {
			line += ui.SGray.Render("  (" + t.Warning + ")")
		}
		fmt.Println(line)
	}
	return nil
}
