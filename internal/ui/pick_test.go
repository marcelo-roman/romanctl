package ui

import (
	"errors"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func press(m *pickModel, keys ...any) {
	for _, k := range keys {
		switch v := k.(type) {
		case string:
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(v)})
		case tea.KeyType:
			m.Update(tea.KeyMsg{Type: v})
		}
	}
}

func options() []Option {
	return []Option{Opt("kubectl Kubernetes CLI", "kubectl"), Opt("helm Packages", "helm"), Opt(SGreen.Render("Security")+" trivy", "trivy"), Opt("kubens Switch", "kubens")}
}

func TestMatches(t *testing.T) {
	cases := []struct {
		text, filter string
		want         bool
	}{
		{"kubectl CLI", "kube", true}, {"kubectl CLI", "KUBE", true}, {"kubectl CLI", "cli kube", true},
		{"kubectl CLI", "kube helm", false}, {"Cafe", "caf\u00e9", true}, {"Caf\u00e9", "cafe", true},
		{SGreen.Render("green") + " x", "green", true}, {"anything", "", true}, {"anything", "   ", true},
	}
	for _, c := range cases {
		if got := matches(c.text, c.filter); got != c.want {
			t.Errorf("%q / %q: %v", c.text, c.filter, got)
		}
	}
}

func TestTypingFiltersAndEnterChooses(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, "h", "e")
	if got := m.visibleOptions(); !slices.Equal(got, []string{"helm"}) {
		t.Fatalf("%v", got)
	}
	press(m, tea.KeyEnter)
	if !m.finished || m.err != nil || !slices.Equal(m.result, []string{"helm"}) {
		t.Fatalf("%+v", m)
	}
}

func TestEscClearsFilterBeforeGoingBack(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, "kube", tea.KeyEsc)
	if m.finished || m.filter != "" || len(m.visible) != 4 {
		t.Fatalf("first Esc: %+v", m)
	}
	press(m, tea.KeyEsc)
	if !m.finished || !errors.Is(m.err, errBack) {
		t.Fatalf("second Esc: %+v", m)
	}
}

func TestLeftArrow(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, "k", tea.KeyLeft)
	if m.finished {
		t.Fatal("with a filter, the left arrow should not exit")
	}
	press(m, tea.KeyBackspace, tea.KeyLeft)
	if !errors.Is(m.err, errBack) {
		t.Fatalf("%v", m.err)
	}
	m = newPickModel(PickConfig{Options: options(), NoLeft: true})
	press(m, tea.KeyLeft)
	if m.finished {
		t.Fatal("NoLeft should ignore the arrow")
	}
}

func TestCtrlCCancels(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, tea.KeyCtrlC)
	if !errors.Is(m.err, errCancel) {
		t.Fatalf("%v", m.err)
	}
}

func TestAllMarksOnlyWhatTheFilterShows(t *testing.T) {
	m := newPickModel(PickConfig{Options: options(), Multi: true, All: true})
	press(m, "kube", tea.KeySpace, tea.KeyEnter)
	if !slices.Equal(m.result, []string{"kubectl", "kubens"}) {
		t.Fatalf("%v", m.result)
	}
	m = newPickModel(PickConfig{Options: options(), Multi: true, All: true})
	press(m, tea.KeySpace, tea.KeySpace)
	press(m, tea.KeyEnter)
	if m.finished || !strings.Contains(m.warn, "Space") {
		t.Fatalf("should not confirm with nothing marked: %+v", m)
	}
	press(m, tea.KeyCtrlA, tea.KeyEnter)
	if len(m.result) != 4 {
		t.Fatalf("%v", m.result)
	}
}

func TestNoneMarkedAllowed(t *testing.T) {
	m := newPickModel(PickConfig{Options: options(), Multi: true, AllowNone: true})
	press(m, tea.KeyEnter)
	if !m.finished || m.err != nil || len(m.result) != 0 {
		t.Fatalf("%+v", m)
	}
}

func TestMarksSurviveTheFilter(t *testing.T) {
	m := newPickModel(PickConfig{Options: options(), Multi: true})
	press(m, "helm", tea.KeySpace, tea.KeyCtrlU, "trivy", tea.KeySpace, tea.KeyEnter)
	if !slices.Equal(m.result, []string{"helm", "trivy"}) {
		t.Fatalf("%v", m.result)
	}
}

func TestSpaceGoesIntoSingleChoiceFilter(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, "kube", tea.KeySpace, "cli")
	if got := m.visibleOptions(); !slices.Equal(got, []string{"kubectl"}) {
		t.Fatalf("%q %v", m.filter, got)
	}
}

func TestEnterWithNoResultDoesNotExit(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, "zzz", tea.KeyEnter)
	if m.finished {
		t.Fatal("should not finish")
	}
	if !strings.Contains(m.View(), "Nothing found") {
		t.Fatal(m.View())
	}
}

func TestInitialAndWrappingCursor(t *testing.T) {
	m := newPickModel(PickConfig{Options: options(), Initial: "trivy"})
	if m.cfg.Options[m.visible[m.cursor]].Value != "trivy" {
		t.Fatal(m.cursor)
	}
	press(m, tea.KeyDown, tea.KeyDown)
	if m.cursor != 0 {
		t.Fatal(m.cursor)
	}
	press(m, tea.KeyUp)
	if m.cursor != 3 {
		t.Fatal(m.cursor)
	}
}

func TestScrollKeepsCursorVisible(t *testing.T) {
	var many []Option
	for i := 0; i < 100; i++ {
		many = append(many, Opt("item", string(rune('a'+i%26))+string(rune('a'+i/26))))
	}
	m := newPickModel(PickConfig{Options: many})
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	for i := 0; i < 50; i++ {
		press(m, tea.KeyDown)
	}
	if m.cursor < m.offset || m.cursor >= m.offset+m.pageSize() {
		t.Fatalf("cursor %d outside %d+%d", m.cursor, m.offset, m.pageSize())
	}
	if lines := strings.Count(m.View(), "\n"); lines > 12 {
		t.Fatalf("%d lines do not fit in 12:\n%s", lines, m.View())
	}
}

func TestSectionsFollowTheFilter(t *testing.T) {
	m := newPickModel(PickConfig{Options: []Option{
		{Label: "new", Value: "new", Section: "Projects"},
		{Label: "add", Value: "add", Section: "Projects"},
		{Label: "status", Value: "status", Section: "Services"},
	}})
	view := m.View()
	if strings.Count(view, "Projects") != 1 || !strings.Contains(view, "Services") {
		t.Fatal(view)
	}
	press(m, "status")
	if view := m.View(); strings.Contains(view, "Projects") || !strings.Contains(view, "Services") {
		t.Fatal(view)
	}
}

func TestListHasNoIcons(t *testing.T) {
	for _, multi := range []bool{false, true} {
		m := newPickModel(PickConfig{Title: "T", Options: options(), Multi: multi, All: multi,
			Categories: []Category{{"x", "X"}}})
		for _, r := range ansi.Strip(m.View()) {
			if r > 0x2000 {
				t.Errorf("symbol %q in the list", r)
			}
		}
	}
}

func TestSpinWithoutTerminalJustRuns(t *testing.T) {
	ran := false
	Spin("x", func(update func(string)) { update("y"); ran = true })
	if !ran {
		t.Fatal("did not run")
	}
}

func categorized() PickConfig {
	return PickConfig{
		Categories: []Category{{"db", "Databases"}, {"sdk", "Languages"}},
		Options: []Option{
			{Label: "psql", Value: "psql", Category: "db", Search: "Databases"},
			{Label: "pgcli", Value: "pgcli", Category: "db", Search: "Databases"},
			{Label: "go", Value: "go", Category: "sdk", Search: "Languages"},
		},
	}
}

func TestArrowsSwitchCategory(t *testing.T) {
	m := newPickModel(categorized())
	if len(m.visibleOptions()) != 3 {
		t.Fatal(m.visibleOptions())
	}
	press(m, tea.KeyRight)
	if got := m.visibleOptions(); !slices.Equal(got, []string{"psql", "pgcli"}) || !strings.Contains(m.View(), "Databases") {
		t.Fatalf("%v", got)
	}
	press(m, tea.KeyRight)
	if got := m.visibleOptions(); !slices.Equal(got, []string{"go"}) {
		t.Fatalf("%v", got)
	}
	press(m, tea.KeyRight)
	if m.category != 0 || len(m.visibleOptions()) != 3 {
		t.Fatal("should wrap back to All")
	}
	press(m, tea.KeyLeft)
	if got := m.visibleOptions(); !slices.Equal(got, []string{"go"}) || m.finished {
		t.Fatalf("the left arrow should cycle, not exit: %v", got)
	}
}

func TestCategoryCombinesWithFilter(t *testing.T) {
	m := newPickModel(categorized())
	press(m, tea.KeyRight, "pg")
	if got := m.visibleOptions(); !slices.Equal(got, []string{"pgcli"}) {
		t.Fatalf("%v", got)
	}
	press(m, tea.KeyRight)
	if got := m.visibleOptions(); len(got) != 0 {
		t.Fatalf("the filter should still apply: %v", got)
	}
}

func TestTypingTheCategoryName(t *testing.T) {
	m := newPickModel(categorized())
	press(m, "databases")
	if got := m.visibleOptions(); !slices.Equal(got, []string{"psql", "pgcli"}) {
		t.Fatalf("%v", got)
	}
	if strings.Contains(m.View(), "Languages") {
		t.Fatal("the search text should not appear on the row")
	}
}

func TestAllRespectsTheCategory(t *testing.T) {
	cfg := categorized()
	cfg.Multi, cfg.All = true, true
	m := newPickModel(cfg)
	press(m, tea.KeyRight, tea.KeySpace, tea.KeyEnter)
	if !slices.Equal(m.result, []string{"psql", "pgcli"}) {
		t.Fatalf("%v", m.result)
	}
}
