package services

import (
	"fmt"
	"os"
	"os/signal"
	"romanctl/internal/i18n"
	"romanctl/internal/state"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"sort"
	"strings"
	"sync"
)

// StatusJSON lists what is installed; with all, also what can be installed.
func StatusJSON(all bool) []map[string]any {
	out := []map[string]any{}
	for _, s := range All {
		installed := s.Installed()
		if !installed && !all {
			continue
		}
		out = append(out, map[string]any{"key": s.Key, "name": s.Name, "installed": installed, "engine": s.Engine,
			"running": installed && s.Running(), "wanted": state.WantsRunning(s.Key), "url": s.URL(), "unsupported": s.Unsupported})
	}
	return out
}

// Available returns the services that run here and are not installed yet.
func Available() []*Service {
	var out []*Service
	for _, s := range All {
		if s.Unsupported == "" && !s.Engine && !s.Installed() {
			out = append(out, s)
		}
	}
	return out
}

// PrintStatus shows only what is installed: a database that was never pulled is not a service of
// this machine. What can be installed comes in one line, or in the list with all.
func PrintStatus(all bool) {
	labels := make([]string, len(All))
	installed := make([]bool, len(All))
	ui.Spin(i18n.T("Checking services..."), func(func(string)) {
		var wg sync.WaitGroup
		for i, s := range All {
			if s.Unsupported != "" {
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				installed[i] = s.Installed()
				labels[i] = statusLabel(s)
			}()
		}
		wg.Wait()
	})
	fmt.Println(ui.STitle.Render(i18n.T("Services")))
	var available []string
	shown := 0
	for i, s := range All {
		switch {
		case s.Unsupported != "":
		case installed[i] || all:
			fmt.Printf("  %-12s %s\n", s.Name, labels[i])
			shown++
		case !s.Engine:
			available = append(available, s.Key)
		}
	}
	if shown == 0 {
		fmt.Println("  " + ui.SGray.Render(i18n.T("none installed")))
	}
	if len(available) > 0 {
		fmt.Println("  " + ui.SGray.Render(i18n.Sprintf("To install: %s  (romanctl install <service>)", strings.Join(available, ", "))))
	}
	printUps()
}

// printUps shows the projects brought up with romanctl up.
func printUps() {
	ups := state.ReadState().Ups
	if len(ups) == 0 {
		return
	}
	names := make([]string, 0, len(ups))
	for name := range ups {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Println("\n" + ui.STitle.Render(i18n.T("Projects up")))
	for _, name := range names {
		e := ups[name]
		label := ui.SGray.Render(i18n.T("services only"))
		switch {
		case e.Alive():
			label = ui.SGreen.Render(i18n.T("running")) + "  " + e.URL
		case e.PID != 0:
			label = ui.SRed.Render(i18n.T("crashed")) + ui.SGray.Render("  (romanctl up -d "+name+")")
		}
		fmt.Printf("  %-12s %s\n", name, label)
	}
}

func statusLabel(s *Service) string {
	switch {
	case !s.Installed() && s.Engine:
		return ui.SGray.Render(i18n.T("not installed")) + ui.SGray.Render("  (romanctl tools install docker)")
	case !s.Installed():
		return ui.SGray.Render(i18n.T("not installed")) + ui.SGray.Render("  (romanctl install "+s.Key+")")
	case s.Running():
		return ui.SGreen.Render(i18n.T("running")) + "  " + s.URL()
	case state.WantsRunning(s.Key):
		return ui.SRed.Render(i18n.T("crashed")) + ui.SGray.Render("  (romanctl restore)")
	}
	return ui.SRed.Render(i18n.T("stopped"))
}

func TailLog(s *Service) error {
	if s.LogCmd != nil {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		defer signal.Stop(sig)
		command := s.LogCmd()
		sys.Run("", nil, command[0], command[1:]...)
		return nil
	}
	if !sys.Exists(s.LogFile) {
		return i18n.Errorf("no log in %s yet", s.LogFile)
	}
	ui.Infof("%s  (Ctrl+C to exit)", s.LogFile)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	sys.Run("", nil, "tail", "-n", "50", "-f", s.LogFile)
	return nil
}

// TailFile follows a log file; Ctrl+C exits.
func TailFile(path string) error {
	if !sys.Exists(path) {
		return i18n.Errorf("no log in %s yet", path)
	}
	ui.Infof("%s  (Ctrl+C to exit)", path)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	sys.Run("", nil, "tail", "-n", "50", "-f", path)
	return nil
}
