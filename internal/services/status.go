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

// StatusJSON lista o que está instalado; com all, também o que dá para instalar.
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

// Available são os serviços que rodam aqui e ainda não foram instalados.
func Available() []*Service {
	var out []*Service
	for _, s := range All {
		if s.Unsupported == "" && !s.Engine && !s.Installed() {
			out = append(out, s)
		}
	}
	return out
}

// PrintStatus mostra só o que está instalado: um banco que nunca foi baixado não é um serviço
// desta máquina. O que dá para instalar vem numa linha, ou na lista com all.
func PrintStatus(all bool) {
	labels := make([]string, len(All))
	installed := make([]bool, len(All))
	ui.Spin(i18n.T("Verificando serviços..."), func(func(string)) {
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
	fmt.Println(ui.STitle.Render(i18n.T("Serviços")))
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
		fmt.Println("  " + ui.SGray.Render(i18n.T("nenhum instalado")))
	}
	if len(available) > 0 {
		fmt.Println("  " + ui.SGray.Render(i18n.Sprintf("Para instalar: %s  (romanctl install <serviço>)", strings.Join(available, ", "))))
	}
	printUps()
}

// printUps mostra os projetos levantados com romanctl up.
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
	fmt.Println("\n" + ui.STitle.Render(i18n.T("Projetos no ar")))
	for _, name := range names {
		e := ups[name]
		label := ui.SGray.Render(i18n.T("só os serviços"))
		switch {
		case e.Alive():
			label = ui.SGreen.Render(i18n.T("rodando")) + "  " + e.URL
		case e.PID != 0:
			label = ui.SRed.Render(i18n.T("caiu")) + ui.SGray.Render("  (romanctl up -d "+name+")")
		}
		fmt.Printf("  %-12s %s\n", name, label)
	}
}

func statusLabel(s *Service) string {
	switch {
	case !s.Installed() && s.Engine:
		return ui.SGray.Render(i18n.T("não instalado")) + ui.SGray.Render("  (romanctl tools install docker)")
	case !s.Installed():
		return ui.SGray.Render(i18n.T("não instalado")) + ui.SGray.Render("  (romanctl install "+s.Key+")")
	case s.Running():
		return ui.SGreen.Render(i18n.T("rodando")) + "  " + s.URL()
	case state.WantsRunning(s.Key):
		return ui.SRed.Render(i18n.T("caiu")) + ui.SGray.Render("  (romanctl restore)")
	}
	return ui.SRed.Render(i18n.T("parado"))
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
		return i18n.Errorf("ainda não há log em %s", s.LogFile)
	}
	ui.Infof("%s  (Ctrl+C para sair)", s.LogFile)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	sys.Run("", nil, "tail", "-n", "50", "-f", s.LogFile)
	return nil
}

// TailFile acompanha um arquivo de log; Ctrl+C sai.
func TailFile(path string) error {
	if !sys.Exists(path) {
		return i18n.Errorf("ainda não há log em %s", path)
	}
	ui.Infof("%s  (Ctrl+C para sair)", path)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	sys.Run("", nil, "tail", "-n", "50", "-f", path)
	return nil
}
