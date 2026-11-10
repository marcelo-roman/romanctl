package webui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Job é uma ação da interface que roda o próprio romanctl com --json, como a CLI faria.
type Job struct {
	ID       string          `json:"id"`
	Action   string          `json:"action"`
	Title    string          `json:"title"`
	Args     []string        `json:"args"`
	Status   string          `json:"status"`
	Started  time.Time       `json:"started"`
	Finished *time.Time      `json:"finished,omitempty"`
	Lines    []string        `json:"lines"`
	Result   json.RawMessage `json:"result,omitempty"`
	Error    string          `json:"error,omitempty"`
}

const (
	jobRunning   = "running"
	jobSucceeded = "succeeded"
	jobFailed    = "failed"
	maxJobs      = 100
	maxJobLines  = 2000
)

type jobStore struct {
	mu   sync.Mutex
	jobs []*Job
	seq  int
	exe  string
	env  []string
}

func newJobStore(exe string) *jobStore {
	return &jobStore{exe: exe, env: append(os.Environ(), "NO_COLOR=1", "ROMANCTL_WEBUI=1")}
}

func (s *jobStore) start(action, title string, args []string) *Job {
	s.mu.Lock()
	s.seq++
	j := &Job{ID: fmt.Sprintf("%d", s.seq), Action: action, Title: title, Args: args, Status: jobRunning,
		Started: time.Now(), Lines: []string{}}
	s.jobs = append(s.jobs, j)
	if len(s.jobs) > maxJobs {
		s.jobs = s.jobs[len(s.jobs)-maxJobs:]
	}
	s.mu.Unlock()
	go s.run(j)
	return j
}

func (s *jobStore) run(j *Job) {
	cmd := exec.Command(s.exe, append(j.Args, "--json")...)
	cmd.Env = s.env
	cmd.Stdin = nil
	var stdout bytes.Buffer
	stderr, err := cmd.StderrPipe()
	if err != nil {
		s.finish(j, nil, err)
		return
	}
	cmd.Stdout = &stdout
	if err := cmd.Start(); err != nil {
		s.finish(j, nil, err)
		return
	}
	s.collect(j, stderr)
	err = cmd.Wait()
	s.finish(j, stdout.Bytes(), err)
}

func (s *jobStore) collect(j *Job, r io.Reader) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "" {
			continue
		}
		s.mu.Lock()
		j.Lines = append(j.Lines, line)
		if len(j.Lines) > maxJobLines {
			j.Lines = j.Lines[len(j.Lines)-maxJobLines:]
		}
		s.mu.Unlock()
	}
}

func (s *jobStore) finish(j *Job, stdout []byte, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	j.Finished = &now
	result := lastJSON(stdout)
	if result != nil {
		j.Result = result
	}
	var failure struct {
		OK    *bool  `json:"ok"`
		Error string `json:"error"`
	}
	if result != nil {
		_ = json.Unmarshal(result, &failure)
	}
	switch {
	case err != nil:
		j.Status = jobFailed
		j.Error = failure.Error
		if j.Error == "" {
			j.Error = err.Error()
		}
	case failure.OK != nil && !*failure.OK:
		j.Status = jobFailed
		j.Error = failure.Error
	default:
		j.Status = jobSucceeded
	}
}

// lastJSON acha o documento JSON da saída padrão; a CLI imprime um só no fim.
func lastJSON(b []byte) json.RawMessage {
	b = bytes.TrimSpace(b)
	if len(b) == 0 {
		return nil
	}
	if json.Valid(b) {
		return json.RawMessage(b)
	}
	for _, open := range []byte{'{', '['} {
		if i := bytes.LastIndexByte(b, open); i >= 0 && json.Valid(b[i:]) {
			return json.RawMessage(b[i:])
		}
	}
	return nil
}

func (s *jobStore) list() []Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Job, 0, len(s.jobs))
	for i := len(s.jobs) - 1; i >= 0; i-- {
		j := *s.jobs[i]
		j.Lines = append([]string{}, j.Lines...)
		out = append(out, j)
	}
	return out
}

func (s *jobStore) get(id string) (Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, j := range s.jobs {
		if j.ID == id {
			copy := *j
			copy.Lines = append([]string{}, j.Lines...)
			return copy, true
		}
	}
	return Job{}, false
}
