package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// O estado guarda o que o usuário quer ligado. Ele só muda por start/stop, então
// sobrevive quando o Android mata o Termux ou o tablet reinicia.
type runState struct {
	Services []string       `json:"services"`
	Projects map[string]int `json:"projects"` // projeto em "run" → PID do romanctl
	Updated  string         `json:"updated"`
}

var statePath = filepath.Join(dataDir, "estado.json")

func readState() runState {
	var st runState
	if b, err := os.ReadFile(statePath); err == nil {
		if err := json.Unmarshal(b, &st); err != nil {
			warnf("Estado ilegível em %s: %v", statePath, err)
		}
	}
	return st
}

// updateState altera o estado com trava, porque várias sessões podem usar o romanctl.
func updateState(f func(*runState)) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	lock, err := os.OpenFile(statePath+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	st := readState()
	f(&st)
	st.Updated = time.Now().Format("02/01/2006 15:04")
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := statePath + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, statePath)
}

func markService(key string, on bool) {
	err := updateState(func(st *runState) {
		st.Services = slices.DeleteFunc(st.Services, func(k string) bool { return k == key })
		if on {
			st.Services = append(st.Services, key)
			sort.Strings(st.Services)
		}
	})
	if err != nil {
		warnf("Não consegui salvar o estado: %v", err)
	}
}

func markProject(name string, on bool) {
	err := updateState(func(st *runState) {
		if st.Projects == nil {
			st.Projects = map[string]int{}
		}
		delete(st.Projects, name)
		if on {
			st.Projects[name] = os.Getpid()
		}
	})
	if err != nil {
		warnf("Não consegui salvar o estado: %v", err)
	}
}

func wantsRunning(key string) bool { return slices.Contains(readState().Services, key) }

// processIs confere se o PID está vivo e é o programa esperado. O PID sozinho não
// basta: depois de uma queda, o Android pode reusá-lo em outro processo.
func processIs(pid int, name string) bool {
	if pid <= 0 || syscall.Kill(pid, 0) != nil {
		return false
	}
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	return err == nil && strings.Contains(string(b), name)
}

// projetos que estavam em "run" e cujo processo morreu.
func crashedProjects(st runState) []string {
	var names []string
	for name, pid := range st.Projects {
		if !processIs(pid, "romanctl") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func restoreRun() error {
	st := readState()
	crashed := crashedProjects(st)
	if len(st.Services) == 0 && len(crashed) == 0 {
		infof("Nada para restaurar: nenhum serviço estava ligado.")
		return nil
	}
	var failed []string
	for _, key := range st.Services {
		s := findService(key)
		if s == nil {
			warnf("Serviço desconhecido no estado: %s", key)
			continue
		}
		if err := s.Start(); err != nil {
			failf("%v", err)
			failed = append(failed, s.Name)
		}
	}
	if len(crashed) > 0 {
		fmt.Println()
		warnf("Estes projetos estavam rodando e precisam de um terminal próprio:")
		for _, name := range crashed {
			fmt.Printf("    romanctl run %s\n", name)
			markProject(name, false)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("não subiram: %s", strings.Join(failed, ", "))
	}
	return nil
}

func restoreStatus() {
	st := readState()
	fmt.Println(sTitle.Render("Estado salvo"))
	if st.Updated != "" {
		fmt.Printf("  %-14s %s\n", "Atualizado em", st.Updated)
	}
	if len(st.Services) == 0 {
		fmt.Printf("  %-14s %s\n", "Serviços", sGray.Render("nenhum"))
	}
	for _, key := range st.Services {
		s := findService(key)
		if s == nil {
			continue
		}
		state := sRed.Render("○ caiu")
		if s.Running() {
			state = sGreen.Render("● rodando")
		}
		fmt.Printf("  %-14s %s\n", s.Name, state)
	}
	for _, name := range crashedProjects(st) {
		fmt.Printf("  %-14s %s\n", name, sRed.Render("○ projeto caiu"))
	}
}

func restoreCommand(sub string) error {
	switch sub {
	case "":
		return restoreRun()
	case "status":
		restoreStatus()
		return nil
	case "save", "salvar":
		var keys []string
		for _, s := range services {
			if s.Installed() && s.Running() {
				keys = append(keys, s.Key)
			}
		}
		sort.Strings(keys)
		if err := updateState(func(st *runState) { st.Services = keys }); err != nil {
			return err
		}
		okf("Estado salvo: %s", orNone(keys))
		return nil
	case "clear", "limpar":
		if err := updateState(func(st *runState) { st.Services, st.Projects = nil, nil }); err != nil {
			return err
		}
		okf("Estado limpo: o restore não vai subir nada")
		return nil
	}
	return fmt.Errorf("use: romanctl restore [status|save|clear]")
}

func orNone(keys []string) string {
	if len(keys) == 0 {
		return "nenhum serviço"
	}
	return strings.Join(keys, ", ")
}
