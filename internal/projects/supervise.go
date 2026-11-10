package projects

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"

	"romanctl/internal/i18n"
	"romanctl/internal/manifest"
)

// SuperviseArg é o subcomando interno que roda os processos de um projeto. Não aparece na ajuda.
const SuperviseArg = "__supervise"

const stopGrace = 5 * time.Second

type child struct {
	name string
	cmd  *exec.Cmd
}

type exit struct {
	name string
	code int
}

// SuperviseHere roda os processos do romanctl.yaml da pasta atual.
func SuperviseHere() int {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	m, ok := manifest.Read(dir)
	if !ok || len(m.Processes) == 0 {
		fmt.Fprintln(os.Stderr, i18n.T("o romanctl.yaml desta pasta não declara processos"))
		return 2
	}
	return Supervise(dir, m.Processes, os.Stdout)
}

// Supervise sobe cada processo no próprio grupo, com o nome no começo de cada linha. Quando um
// termina, ou quando chega Ctrl+C ou SIGTERM, para os outros e devolve o código do primeiro que
// terminou. Assim quem espera o projeto subir percebe a queda na hora.
func Supervise(dir string, processes []manifest.Process, out io.Writer) int {
	var mu sync.Mutex
	exits := make(chan exit, len(processes))
	var children []child
	var output sync.WaitGroup
	for _, p := range processes {
		c, err := startChild(dir, p, out, &mu, &output, exits)
		if err != nil {
			fmt.Fprintf(out, "[%s] %v\n", p.Name, err)
			stopAll(children)
			return 1
		}
		children = append(children, child{p.Name, c})
	}

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	code := 0
	remaining := len(children)
	select {
	case first := <-exits:
		remaining--
		code = first.code
		if code == 0 && remaining > 0 {
			code = 1
		}
		mu.Lock()
		fmt.Fprintf(out, i18n.T("[%s] terminou com código %d; parando os outros processos")+"\n", first.name, first.code)
		mu.Unlock()
	case <-signals:
	}
	stopAll(children)
	deadline := time.After(stopGrace)
	for remaining > 0 {
		select {
		case <-exits:
			remaining--
		case <-deadline:
			killAll(children)
			deadline = nil
		}
	}
	output.Wait()
	return code
}

func startChild(dir string, p manifest.Process, out io.Writer, mu *sync.Mutex, output *sync.WaitGroup, exits chan<- exit) (*exec.Cmd, error) {
	cmd := exec.Command("sh", "-c", p.Run)
	cmd.Dir = filepath.Join(dir, p.Dir)
	cmd.Env = append(os.Environ(), envList(p.Env)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd.Stdout, cmd.Stderr = writer, writer
	err = cmd.Start()
	writer.Close()
	if err != nil {
		reader.Close()
		return nil, err
	}
	output.Add(1)
	go func() {
		defer output.Done()
		defer reader.Close()
		prefix(reader, out, p.Name, mu)
	}()
	go func() {
		exits <- exit{p.Name, exitCode(cmd.Wait())}
	}()
	return cmd, nil
}

func prefix(r io.Reader, out io.Writer, name string, mu *sync.Mutex) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		mu.Lock()
		fmt.Fprintf(out, "[%s] %s\n", name, scanner.Text())
		mu.Unlock()
	}
	io.Copy(io.Discard, r)
}

func envList(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+env[k])
	}
	return out
}

func signalAll(children []child, sig syscall.Signal) {
	for _, c := range children {
		if c.cmd.Process != nil {
			syscall.Kill(-c.cmd.Process.Pid, sig)
		}
	}
}

func stopAll(children []child) { signalAll(children, syscall.SIGTERM) }
func killAll(children []child) { signalAll(children, syscall.SIGKILL) }

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if e, ok := err.(*exec.ExitError); ok && e.ExitCode() >= 0 {
		return e.ExitCode()
	}
	return 1
}
