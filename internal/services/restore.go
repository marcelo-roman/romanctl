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

// crashedProjects returns the projects that were in "run" and whose process died.
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
	var downed []string
	for name, e := range st.Ups {
		if e.PID != 0 && !e.Alive() {
			downed = append(downed, name)
		}
	}
	sort.Strings(downed)
	if len(st.Services) == 0 && len(crashed) == 0 && len(downed) == 0 {
		ui.Infof("Nothing to restore: no service was running.")
		return nil
	}
	var failed []string
	for _, key := range st.Services {
		s := FindService(key)
		if s == nil {
			ui.Warnf("Unknown service in the state: %s", key)
			continue
		}
		if err := s.Start(); err != nil {
			ui.Failf("%v", err)
			failed = append(failed, s.Name)
		}
	}
	if len(crashed) > 0 {
		fmt.Println()
		ui.Warnf("These projects were running and need their own terminal:")
		for _, name := range crashed {
			i18n.Printf("    romanctl run %s\n", name)
			state.MarkProject(name, false)
		}
	}
	if len(downed) > 0 {
		fmt.Println()
		ui.Warnf("These projects were up; to bring them up again:")
		for _, name := range downed {
			fmt.Printf("    romanctl up -d %s\n", name)
		}
	}
	if len(failed) > 0 {
		return i18n.Errorf("did not start: %s", strings.Join(failed, ", "))
	}
	return nil
}

func restoreStatus() {
	st := state.ReadState()
	fmt.Println(ui.STitle.Render(i18n.T("Saved state")))
	if st.Updated != "" {
		i18n.Printf("  %-14s %s\n", i18n.T("Updated on"), st.Updated)
	}
	if len(st.Services) == 0 {
		i18n.Printf("  %-14s %s\n", i18n.T("Services"), ui.SGray.Render(i18n.T("none")))
	}
	for _, key := range st.Services {
		s := FindService(key)
		if s == nil {
			continue
		}
		state := ui.SRed.Render(i18n.T("crashed"))
		if s.Running() {
			state = ui.SGreen.Render(i18n.T("running"))
		}
		i18n.Printf("  %-14s %s\n", s.Name, state)
	}
	for _, name := range crashedProjects(st) {
		i18n.Printf("  %-14s %s\n", name, ui.SRed.Render(i18n.T("project crashed")))
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
		ui.Okf("State saved: %s", orNone(keys))
		return nil
	case "clear", "limpar":
		if err := state.UpdateState(func(st *state.RunState) { st.Services, st.Projects = nil, nil }); err != nil {
			return err
		}
		ui.Okf("State cleared: restore will not start anything")
		return nil
	}
	return i18n.Errorf("usage: romanctl restore [status|save|clear]")
}

func orNone(keys []string) string {
	if len(keys) == 0 {
		return i18n.T("no services")
	}
	return strings.Join(keys, ", ")
}
