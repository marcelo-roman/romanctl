package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEverySkillHasNameAndDescription(t *testing.T) {
	list := Status(t.TempDir())
	if len(list) < 2 {
		t.Fatalf("bundle with %d skills", len(list))
	}
	for _, s := range list {
		body := string(files(s.Name)["SKILL.md"])
		if !strings.Contains(body, "\nname: "+s.Name+"\n") {
			t.Errorf("%s: the frontmatter name does not match the directory", s.Name)
		}
		if s.Description == "" {
			t.Errorf("%s: no description", s.Name)
		}
	}
}

func TestInstallUpdateAndRemove(t *testing.T) {
	dir := t.TempDir()
	results, err := Install(dir, "1.0.0", false, false)
	if err != nil || results[0].Action != "installed" {
		t.Fatalf("%v %+v", err, results)
	}
	for _, s := range Status(dir) {
		if s.State != StateInstalled {
			t.Fatalf("%s: %s", s.Name, s.State)
		}
	}
	first := filepath.Join(dir, results[0].Name, "SKILL.md")
	os.WriteFile(first, []byte("edited"), 0o644)
	os.WriteFile(filepath.Join(dir, results[0].Name, "leftover.md"), []byte("x"), 0o644)
	if got := Status(dir)[0].State; got != StateOutdated {
		t.Fatal(got)
	}
	results, _ = Install(dir, "1.0.1", false, false)
	if results[0].Action != "updated" || results[1].Action != "unchanged" {
		t.Fatalf("%+v", results)
	}
	if _, err := os.Stat(filepath.Join(dir, results[0].Name, "sobra.md")); err == nil {
		t.Fatal("a file from the old version was left behind")
	}
	results, _ = Uninstall(dir, false)
	for _, r := range results {
		if r.Action != "removed" {
			t.Fatalf("%+v", r)
		}
	}
}

func TestLeavesForeignSkillAlone(t *testing.T) {
	dir := t.TempDir()
	name := names()[0]
	mine := filepath.Join(dir, name, "SKILL.md")
	os.MkdirAll(filepath.Dir(mine), 0o755)
	os.WriteFile(mine, []byte("mine"), 0o644)
	results, _ := Install(dir, "1.0.0", false, false)
	if results[0].Action != "skipped" {
		t.Fatalf("%+v", results[0])
	}
	Uninstall(dir, false)
	if b, _ := os.ReadFile(mine); string(b) != "mine" {
		t.Fatal("deleted or replaced a skill romanctl did not create")
	}
	results, _ = Install(dir, "1.0.0", true, false)
	if results[0].Action != "updated" {
		t.Fatalf("--force should replace: %+v", results[0])
	}
}
