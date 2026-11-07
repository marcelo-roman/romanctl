package ui

import (
	"errors"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
	return []Option{Opt("kubectl CLI do Kubernetes", "kubectl"), Opt("helm Pacotes", "helm"), Opt(SGreen.Render("Segurança")+" trivy", "trivy"), Opt("kubens Troca", "kubens")}
}

func TestMatches(t *testing.T) {
	cases := []struct {
		text, filter string
		want         bool
	}{
		{"kubectl CLI", "kube", true}, {"kubectl CLI", "KUBE", true}, {"kubectl CLI", "cli kube", true},
		{"kubectl CLI", "kube helm", false}, {"Segurança", "seguranca", true}, {"Seguranca", "segurança", true},
		{SGreen.Render("verde") + " x", "verde", true}, {"qualquer", "", true}, {"qualquer", "   ", true},
	}
	for _, c := range cases {
		if got := matches(c.text, c.filter); got != c.want {
			t.Errorf("%q / %q: %v", c.text, c.filter, got)
		}
	}
}

func TestDigitarFiltraEEnterEscolhe(t *testing.T) {
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

func TestEscLimpaOFiltroAntesDeVoltar(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, "kube", tea.KeyEsc)
	if m.finished || m.filter != "" || len(m.visible) != 4 {
		t.Fatalf("primeiro Esc: %+v", m)
	}
	press(m, tea.KeyEsc)
	if !m.finished || !errors.Is(m.err, errBack) {
		t.Fatalf("segundo Esc: %+v", m)
	}
}

func TestSetaEsquerda(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, "k", tea.KeyLeft)
	if m.finished {
		t.Fatal("com filtro, a seta esquerda não deveria sair")
	}
	press(m, tea.KeyBackspace, tea.KeyLeft)
	if !errors.Is(m.err, errBack) {
		t.Fatalf("%v", m.err)
	}
	m = newPickModel(PickConfig{Options: options(), NoLeft: true})
	press(m, tea.KeyLeft)
	if m.finished {
		t.Fatal("NoLeft deveria ignorar a seta")
	}
}

func TestCtrlCCancela(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, tea.KeyCtrlC)
	if !errors.Is(m.err, errCancel) {
		t.Fatalf("%v", m.err)
	}
}

func TestTodosMarcaSoOQueOFiltroMostra(t *testing.T) {
	m := newPickModel(PickConfig{Options: options(), Multi: true, All: true})
	press(m, "kube", tea.KeySpace, tea.KeyEnter)
	if !slices.Equal(m.result, []string{"kubectl", "kubens"}) {
		t.Fatalf("%v", m.result)
	}
	m = newPickModel(PickConfig{Options: options(), Multi: true, All: true})
	press(m, tea.KeySpace, tea.KeySpace)
	press(m, tea.KeyEnter)
	if m.finished || !strings.Contains(m.warn, "Espaço") {
		t.Fatalf("sem marcação não deveria confirmar: %+v", m)
	}
	press(m, tea.KeyCtrlA, tea.KeyEnter)
	if len(m.result) != 4 {
		t.Fatalf("%v", m.result)
	}
}

func TestSemMarcacaoPermitida(t *testing.T) {
	m := newPickModel(PickConfig{Options: options(), Multi: true, AllowNone: true})
	press(m, tea.KeyEnter)
	if !m.finished || m.err != nil || len(m.result) != 0 {
		t.Fatalf("%+v", m)
	}
}

func TestMarcacaoSobreviveAoFiltro(t *testing.T) {
	m := newPickModel(PickConfig{Options: options(), Multi: true})
	press(m, "helm", tea.KeySpace, tea.KeyCtrlU, "trivy", tea.KeySpace, tea.KeyEnter)
	if !slices.Equal(m.result, []string{"helm", "trivy"}) {
		t.Fatalf("%v", m.result)
	}
}

func TestEspacoEntraNoFiltroDaEscolhaUnica(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, "kube", tea.KeySpace, "cli")
	if got := m.visibleOptions(); !slices.Equal(got, []string{"kubectl"}) {
		t.Fatalf("%q %v", m.filter, got)
	}
}

func TestSemResultadoEnterNaoSai(t *testing.T) {
	m := newPickModel(PickConfig{Options: options()})
	press(m, "zzz", tea.KeyEnter)
	if m.finished {
		t.Fatal("não deveria terminar")
	}
	if !strings.Contains(m.View(), "Nada encontrado") {
		t.Fatal(m.View())
	}
}

func TestInicialECursorCircular(t *testing.T) {
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

func TestRolagemMantemOCursorVisivel(t *testing.T) {
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
		t.Fatalf("cursor %d fora de %d+%d", m.cursor, m.offset, m.pageSize())
	}
	if lines := strings.Count(m.View(), "\n"); lines > 12 {
		t.Fatalf("%d linhas não cabem em 12:\n%s", lines, m.View())
	}
}

func TestSpinSemTerminalSoRoda(t *testing.T) {
	ran := false
	Spin("x", func(update func(string)) { update("y"); ran = true })
	if !ran {
		t.Fatal("não rodou")
	}
}

func categorized() PickConfig {
	return PickConfig{
		Categories: []Category{{"db", "Banco de dados"}, {"sdk", "Linguagens"}},
		Options: []Option{
			{Label: "psql", Value: "psql", Category: "db", Search: "Banco de dados"},
			{Label: "pgcli", Value: "pgcli", Category: "db", Search: "Banco de dados"},
			{Label: "go", Value: "go", Category: "sdk", Search: "Linguagens"},
		},
	}
}

func TestSetasTrocamCategoria(t *testing.T) {
	m := newPickModel(categorized())
	if len(m.visibleOptions()) != 3 {
		t.Fatal(m.visibleOptions())
	}
	press(m, tea.KeyRight)
	if got := m.visibleOptions(); !slices.Equal(got, []string{"psql", "pgcli"}) || !strings.Contains(m.View(), "Banco de dados") {
		t.Fatalf("%v", got)
	}
	press(m, tea.KeyRight)
	if got := m.visibleOptions(); !slices.Equal(got, []string{"go"}) {
		t.Fatalf("%v", got)
	}
	press(m, tea.KeyRight)
	if m.category != 0 || len(m.visibleOptions()) != 3 {
		t.Fatal("deveria voltar para Todas")
	}
	press(m, tea.KeyLeft)
	if got := m.visibleOptions(); !slices.Equal(got, []string{"go"}) || m.finished {
		t.Fatalf("a seta esquerda deveria girar, não sair: %v", got)
	}
}

func TestCategoriaCombinaComFiltro(t *testing.T) {
	m := newPickModel(categorized())
	press(m, tea.KeyRight, "pg")
	if got := m.visibleOptions(); !slices.Equal(got, []string{"pgcli"}) {
		t.Fatalf("%v", got)
	}
	press(m, tea.KeyRight)
	if got := m.visibleOptions(); len(got) != 0 {
		t.Fatalf("o filtro deveria continuar valendo: %v", got)
	}
}

func TestDigitarONomeDaCategoria(t *testing.T) {
	m := newPickModel(categorized())
	press(m, "banco de dados")
	if got := m.visibleOptions(); !slices.Equal(got, []string{"psql", "pgcli"}) {
		t.Fatalf("%v", got)
	}
	if strings.Contains(m.View(), "Linguagens") {
		t.Fatal("o texto de busca não deveria aparecer na linha")
	}
}

func TestTodosRespeitaACategoria(t *testing.T) {
	cfg := categorized()
	cfg.Multi, cfg.All = true, true
	m := newPickModel(cfg)
	press(m, tea.KeyRight, tea.KeySpace, tea.KeyEnter)
	if !slices.Equal(m.result, []string{"psql", "pgcli"}) {
		t.Fatalf("%v", m.result)
	}
}
