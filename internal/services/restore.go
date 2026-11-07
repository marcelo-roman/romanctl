package services

import (
	"fmt"
	"romanctl/internal/i18n"
	"romanctl/internal/state"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"sort"
	"strings"
)

// projetos que estavam em "run" e cujo processo morreu.
func crashedProjects(st state.RunState) []string {
	var names []string
	for name, pid := range st.Projects {
		if !sys.ProcessIs(pid, "romanctl") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func RestoreRun() error {
	st := state.ReadState()
	crashed := crashedProjects(st)
	if len(st.Services) == 0 && len(crashed) == 0 {
		ui.Infof("Nada para restaurar: nenhum serviço estava ligado.")
		return nil
	}
	var failed []string
	for _, key := range st.Services {
		s := FindService(key)
		if s == nil {
			ui.Warnf("Serviço desconhecido no estado: %s", key)
			continue
		}
		if err := s.Start(); err != nil {
			ui.Failf("%v", err)
			failed = append(failed, s.Name)
		}
	}
	if len(crashed) > 0 {
		fmt.Println()
		ui.Warnf("Estes projetos estavam rodando e precisam de um terminal próprio:")
		for _, name := range crashed {
			i18n.Printf("    romanctl run %s\n", name)
			state.MarkProject(name, false)
		}
	}
	if len(failed) > 0 {
		return i18n.Errorf("não subiram: %s", strings.Join(failed, ", "))
	}
	return nil
}

func restoreStatus() {
	st := state.ReadState()
	fmt.Println(ui.STitle.Render(i18n.T("Estado salvo")))
	if st.Updated != "" {
		i18n.Printf("  %-14s %s\n", i18n.T("Atualizado em"), st.Updated)
	}
	if len(st.Services) == 0 {
		i18n.Printf("  %-14s %s\n", i18n.T("Serviços"), ui.SGray.Render("nenhum"))
	}
	for _, key := range st.Services {
		s := FindService(key)
		if s == nil {
			continue
		}
		state := ui.SRed.Render(i18n.T("○ caiu"))
		if s.Running() {
			state = ui.SGreen.Render(i18n.T("● rodando"))
		}
		i18n.Printf("  %-14s %s\n", s.Name, state)
	}
	for _, name := range crashedProjects(st) {
		i18n.Printf("  %-14s %s\n", name, ui.SRed.Render(i18n.T("○ projeto caiu")))
	}
}

func RestoreCommand(sub string) error {
	switch sub {
	case "":
		return RestoreRun()
	case "status":
		restoreStatus()
		return nil
	case "save", "salvar":
		var keys []string
		for _, s := range All {
			if s.Installed() && s.Running() {
				keys = append(keys, s.Key)
			}
		}
		sort.Strings(keys)
		if err := state.UpdateState(func(st *state.RunState) { st.Services = keys }); err != nil {
			return err
		}
		ui.Okf("Estado salvo: %s", orNone(keys))
		return nil
	case "clear", "limpar":
		if err := state.UpdateState(func(st *state.RunState) { st.Services, st.Projects = nil, nil }); err != nil {
			return err
		}
		ui.Okf("Estado limpo: o restore não vai subir nada")
		return nil
	}
	return i18n.Errorf("use: romanctl restore [status|save|clear]")
}

func orNone(keys []string) string {
	if len(keys) == 0 {
		return i18n.T("nenhum serviço")
	}
	return strings.Join(keys, ", ")
}
