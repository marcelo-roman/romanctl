// Package claude installs the skills that teach Claude Code to operate the machine through romanctl.
package claude

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed skills
var bundle embed.FS

// Marker sits in every directory romanctl created: it is what allows overwriting or deleting it.
const Marker = ".romanctl"

// States of a skill on the machine.
const (
	StateInstalled = "installed"
	StateOutdated  = "outdated"
	StateMissing   = "missing"
	StateForeign   = "foreign"
)

// Skill is a bundled skill and what exists of it on disk.
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	State       string `json:"state"`
}

// Result is what happened to a skill on install or removal.
type Result struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Action string `json:"action"`
}

// Dir is where Claude Code looks for the user's skills.
func Dir() string {
	if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return filepath.Join(d, "skills")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "skills")
}

func names() []string {
	entries, _ := fs.ReadDir(bundle, "skills")
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// files returns a skill's files, keyed by path relative to its directory.
func files(name string) map[string][]byte {
	out := map[string][]byte{}
	root := path.Join("skills", name)
	fs.WalkDir(bundle, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := bundle.ReadFile(p)
		if err == nil {
			out[strings.TrimPrefix(p, root+"/")] = b
		}
		return err
	})
	return out
}

func description(skill []byte) string {
	for _, line := range strings.Split(string(skill), "\n") {
		if v, ok := strings.CutPrefix(line, "description:"); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func state(dir, name string) string {
	target := filepath.Join(dir, name)
	if _, err := os.Stat(target); err != nil {
		return StateMissing
	}
	if _, err := os.Stat(filepath.Join(target, Marker)); err != nil {
		return StateForeign
	}
	for rel, want := range files(name) {
		got, err := os.ReadFile(filepath.Join(target, rel))
		if err != nil || !bytes.Equal(got, want) {
			return StateOutdated
		}
	}
	return StateInstalled
}

// Status compares the bundle with what is in dir.
func Status(dir string) []Skill {
	var out []Skill
	for _, name := range names() {
		out = append(out, Skill{Name: name, Description: description(files(name)["SKILL.md"]),
			Path: filepath.Join(dir, name), State: state(dir, name)})
	}
	return out
}

// Install writes the skills to dir. A directory romanctl did not create is only replaced with force.
func Install(dir, version string, force, dryRun bool) ([]Result, error) {
	var out []Result
	for _, s := range Status(dir) {
		r := Result{Name: s.Name, Path: s.Path}
		switch {
		case s.State == StateInstalled:
			r.Action = "unchanged"
		case s.State == StateForeign && !force:
			r.Action = "skipped"
		case s.State == StateMissing:
			r.Action = "installed"
		default:
			r.Action = "updated"
		}
		if !dryRun && (r.Action == "installed" || r.Action == "updated") {
			if err := write(s.Path, s.Name, version); err != nil {
				return out, err
			}
		}
		out = append(out, r)
	}
	return out, nil
}

// write replaces the whole directory, so no file from an older version of the skill is left behind.
func write(target, name, version string) error {
	tmp := target + ".tmp"
	os.RemoveAll(tmp)
	for rel, b := range files(name) {
		p := filepath.Join(tmp, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, Marker), []byte("romanctl "+version+"\n"), 0o644); err != nil {
		return err
	}
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}

// Uninstall deletes only the directories that carry romanctl's marker.
func Uninstall(dir string, dryRun bool) ([]Result, error) {
	var out []Result
	for _, s := range Status(dir) {
		r := Result{Name: s.Name, Path: s.Path}
		switch s.State {
		case StateMissing:
			r.Action = "absent"
		case StateForeign:
			r.Action = "kept"
		default:
			r.Action = "removed"
			if !dryRun {
				if err := os.RemoveAll(s.Path); err != nil {
					return out, err
				}
			}
		}
		out = append(out, r)
	}
	return out, nil
}
