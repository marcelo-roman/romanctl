package tools

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"romanctl/internal/config"
)

type cachedVersion struct {
	Path    string    `json:"path"`
	ModTime time.Time `json:"modTime"`
	Version string    `json:"version"`
}

type versionCache struct {
	mu      sync.Mutex
	entries map[string]cachedVersion
	dirty   bool
}

func versionCachePath() string { return filepath.Join(config.DataDir, "versoes.json") }

func loadVersionCache() *versionCache {
	c := &versionCache{entries: map[string]cachedVersion{}}
	if b, err := os.ReadFile(versionCachePath()); err == nil {
		json.Unmarshal(b, &c.entries)
	}
	return c
}

func (c *versionCache) save() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.dirty {
		return
	}
	b, err := json.Marshal(c.entries)
	if err != nil || os.MkdirAll(filepath.Dir(versionCachePath()), 0o755) != nil {
		return
	}
	tmp := versionCachePath() + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		os.Rename(tmp, versionCachePath())
	}
}

// cachedDetect só roda o programa se o binário mudou desde a última consulta. Quem instala e
// remove continua usando Detected, que sempre executa de verdade.
func (t *Tool) cachedDetect(c *versionCache) (string, bool) {
	if t.Detect != nil {
		return t.Detected()
	}
	path, err := exec.LookPath(t.Version[0])
	if err != nil {
		return "", false
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	c.mu.Lock()
	entry, ok := c.entries[t.Key]
	c.mu.Unlock()
	if ok && entry.Path == path && entry.ModTime.Equal(info.ModTime()) {
		return entry.Version, true
	}
	version, installed := t.Detected()
	if installed {
		c.mu.Lock()
		c.entries[t.Key] = cachedVersion{path, info.ModTime(), version}
		c.dirty = true
		c.mu.Unlock()
	}
	return version, installed
}
