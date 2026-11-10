package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"romanctl/internal/config"
)

// A última versão muda pouco e consultá-la custa uma ida à rede por ferramenta: guardada por
// algumas horas, a lista abre já com a coluna preenchida.
const latestTTL = 6 * time.Hour

type latestEntry struct {
	Upstream
	At time.Time `json:"at"`
}

var (
	latestMu     sync.Mutex
	latestLoaded bool
	latestSeen   = map[string]latestEntry{}
)

func latestPath() string { return filepath.Join(config.DataDir, "ultimas-versoes.json") }

func loadLatest() {
	if latestLoaded {
		return
	}
	latestLoaded = true
	if b, err := os.ReadFile(latestPath()); err == nil {
		json.Unmarshal(b, &latestSeen)
	}
}

func saveLatest() {
	b, err := json.Marshal(latestSeen)
	if err != nil || os.MkdirAll(filepath.Dir(latestPath()), 0o755) != nil {
		return
	}
	tmp := latestPath() + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		os.Rename(tmp, latestPath())
	}
}

// CachedLatest devolve a última versão guardada, se ainda vale.
func CachedLatest(key string) (Upstream, bool) {
	latestMu.Lock()
	defer latestMu.Unlock()
	loadLatest()
	e, ok := latestSeen[key]
	if !ok || time.Since(e.At) > latestTTL {
		return Upstream{}, false
	}
	return e.Upstream, true
}

// Latest consulta a última versão, usando a guardada enquanto valer. Sem o README.
func (t *Tool) Latest() Upstream {
	if u, ok := CachedLatest(t.Key); ok {
		return u
	}
	u := t.Upstream(false)
	if u.Latest == "" {
		return u
	}
	latestMu.Lock()
	defer latestMu.Unlock()
	latestSeen[t.Key] = latestEntry{Upstream: u, At: time.Now()}
	saveLatest()
	return u
}

// RefreshLatest consulta, em paralelo, a última versão das ferramentas que não estão guardadas.
func RefreshLatest(keys []string, progress func(done, total int)) map[string]Upstream {
	out := map[string]Upstream{}
	var missing []*Tool
	for _, k := range keys {
		if u, ok := CachedLatest(k); ok {
			out[k] = u
		} else if t := Find(k); t != nil {
			missing = append(missing, t)
		}
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	slots := make(chan struct{}, 16)
	done := 0
	for _, t := range missing {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slots <- struct{}{}
			u := t.Latest()
			<-slots
			mu.Lock()
			defer mu.Unlock()
			out[t.Key] = u
			done++
			if progress != nil {
				progress(done, len(missing))
			}
		}()
	}
	wg.Wait()
	return out
}
