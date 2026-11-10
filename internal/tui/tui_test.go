package tui

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"romanctl/internal/sysinfo"
	"romanctl/internal/tools"
)

type fakeBackend struct {
	commands  [][]string
	upstreams []string
}

func (f *fakeBackend) Header() string { return "linux/arm64" }
func (f *fakeBackend) Services() []Row {
	return []Row{{Key: "postgres", Status: "parado", Columns: []string{"postgres", "PostgreSQL", "parado", "pg://"}},
		{Key: "redis", Status: "rodando", Columns: []string{"redis", "Redis", "rodando", "redis://"}},
		{Key: "kafka", Status: "nao-instalado", Columns: []string{"kafka", "Kafka", "não instalado", ""}},
		{Key: "docker", Status: "rodando", Group: engineGroup, Columns: []string{"docker", "Docker", "rodando", "unix:///var/run/docker.sock"}}}
}
func (f *fakeBackend) Projects() []Row {
	return []Row{{Key: "loja", Status: "parado", Columns: []string{"loja", "dotnet", "parado", "http://x", "01/01"}, Detail: [][2]string{{"Pasta", "/p/loja"}}},
		{Key: "site", Status: "rodando", Columns: []string{"site", "next", "rodando", "http://y", "01/01"}}}
}
func (f *fakeBackend) Tools() []Row {
	return []Row{{Key: "kubectl", Status: "instalado", Group: "k8s", Columns: []string{"kubectl", "k8s", "instalado", "v1.0.0", "", "CLI"}},
		{Key: "helm", Status: "disponivel", Group: "k8s", Columns: []string{"helm", "k8s", "disponível", "", "", "Pacotes"}},
		{Key: "psql", Status: "disponivel", Group: "db", Columns: []string{"psql", "db", "disponível", "", "", "Cliente"}}}
}
func (f *fakeBackend) System() sysinfo.Snapshot {
	return sysinfo.Snapshot{Memory: sysinfo.Memory{Total: 1 << 33, Available: 1 << 32}, Disk: sysinfo.Disk{Size: 1 << 34, Free: 1 << 33},
		Battery: &sysinfo.Battery{Percent: 80, Power: 3.5, HoursLeft: 2.5}, Top: []sysinfo.Proc{{PID: 7, Name: "claude", CPU: 12, Memory: 1 << 28}}}
}
func (f *fakeBackend) Upstream(key string) tools.Upstream {
	f.upstreams = append(f.upstreams, key)
	return tools.Upstream{Latest: "2.0.0", Source: "GitHub releases", URL: "https://github.com/x/" + key, Readme: "# " + key + "\n\nGuia de uso do " + key + "."}
}
func (f *fakeBackend) Latest(key string) tools.Upstream {
	return tools.Upstream{Latest: "2.0.0", Source: "GitHub releases"}
}
func (f *fakeBackend) Versions(key string) ([]tools.Version, error) {
	if key == "psql" {
		return nil, errors.New("o apt instala só a versão atual do repositório")
	}
	return []tools.Version{{ID: "v2.0.0", Label: "2.0.0"}, {ID: "v1.9.0", Label: "1.9.0"}, {ID: "v1.0.0", Label: "1.0.0"}}, nil
}
func (f *fakeBackend) Logs(view, key string) string { return "linha 1\nlinha 2\n" + key }
func (f *fakeBackend) Command(args ...string) *exec.Cmd {
	f.commands = append(f.commands, args)
	return exec.Command("true")
}

func setup(t *testing.T) (*Model, *fakeBackend) {
	t.Helper()
	f := &fakeBackend{}
	m := New(f)
	m.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	load(m, viewProjects)
	return m, f
}

// load executa o carregamento de forma síncrona.
func load(m *Model, view string) {
	if msg := m.load(view)(); msg != nil {
		m.Update(msg)
	}
}

func key(m *Model, keys ...string) tea.Cmd {
	var last tea.Cmd
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "right":
			msg = tea.KeyMsg{Type: tea.KeyRight}
		case "left":
			msg = tea.KeyMsg{Type: tea.KeyLeft}
		case "backspace":
			msg = tea.KeyMsg{Type: tea.KeyBackspace}
		case "ctrl+d":
			msg = tea.KeyMsg{Type: tea.KeyCtrlD}
		case "ctrl+c":
			msg = tea.KeyMsg{Type: tea.KeyCtrlC}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		_, last = m.Update(msg)
	}
	return last
}

func screen(m *Model) string { return ansi.Strip(m.View()) }

func TestTrocaDeTela(t *testing.T) {
	m, _ := setup(t)
	if !strings.Contains(screen(m), "loja") {
		t.Fatal(screen(m))
	}
	key(m, "tab")
	if m.view != viewServices {
		t.Fatal(m.view)
	}
	load(m, viewServices)
	if !strings.Contains(screen(m), "postgres") {
		t.Fatal(screen(m))
	}
	key(m, "3")
	if m.view != viewTools {
		t.Fatal(m.view)
	}
	key(m, "1")
	if m.view != viewProjects {
		t.Fatal(m.view)
	}
}

func TestComandos(t *testing.T) {
	m, _ := setup(t)
	for text, want := range map[string]string{"svc": viewServices, "tools": viewTools, "sys": viewSystem, "prj": viewProjects, "ser": viewServices} {
		key(m, ":", text, "enter")
		if m.view != want {
			t.Errorf(":%s foi para %s", text, m.view)
		}
	}
	key(m, ":", "nao-existe", "enter")
	if !strings.Contains(screen(m), "nao-existe") {
		t.Fatal(screen(m))
	}
	key(m, ":", "s", "e", "tab")
	if m.command != "services" {
		t.Fatalf("completar: %q", m.command)
	}
	key(m, "esc")
	cmd := key(m, ":", "q", "enter")
	if !m.quit || cmd == nil {
		t.Fatal("não saiu")
	}
}

func TestComandoDeCategoria(t *testing.T) {
	m, _ := setup(t)
	key(m, ":", "db", "enter")
	if m.view != viewTools {
		t.Fatal(m.view)
	}
	load(m, viewTools)
	rows := m.visible()
	if len(rows) != 1 || rows[0].Key != "psql" {
		t.Fatalf("%+v", rows)
	}
}

func TestFiltro(t *testing.T) {
	m, _ := setup(t)
	key(m, "/", "si")
	if rows := m.visible(); len(rows) != 1 || rows[0].Key != "site" {
		t.Fatalf("%+v", rows)
	}
	key(m, "enter")
	if m.mode != modeTable || m.filter != "si" {
		t.Fatalf("%v %q", m.mode, m.filter)
	}
	key(m, "esc")
	if m.filter != "" || len(m.visible()) != 2 {
		t.Fatal(m.filter)
	}
	key(m, "/", "zzz")
	if !strings.Contains(screen(m), "Nada encontrado") {
		t.Fatal(screen(m))
	}
}

func TestTeclasQueChegamJuntas(t *testing.T) {
	m, _ := setup(t)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/si")})
	if m.mode != modeFilter || m.filter != "si" {
		t.Fatalf("modo %v, filtro %q", m.mode, m.filter)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":svc")})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != viewServices {
		t.Fatal(m.view)
	}
}

func TestLetrasDoFiltroNaoDisparamAcoes(t *testing.T) {
	m, f := setup(t)
	key(m, "/", "u", "r", "x", "q")
	if len(f.commands) != 0 || m.quit {
		t.Fatalf("%v %v", f.commands, m.quit)
	}
	key(m, "esc", ":", "u", "q")
	if m.quit || len(f.commands) != 0 {
		t.Fatal("o comando também não dispara teclas")
	}
}

func TestAcaoRodaEmSegundoPlanoComEstadoNaLinha(t *testing.T) {
	m, f := setup(t)
	cmd := key(m, "u")
	if m.running["projects/loja"] == "" {
		t.Fatal("sem andamento")
	}
	if !strings.Contains(screen(m), "subir") {
		t.Fatal(screen(m))
	}
	if cmd2 := key(m, "u"); cmd2 != nil {
		t.Fatal("repetir enquanto roda não deveria disparar de novo")
	}
	msg := cmd()
	if got := strings.Join(f.commands[0], " "); got != "up loja -d" {
		t.Fatal(got)
	}
	m.Update(msg)
	if m.running["projects/loja"] != "" || !strings.Contains(screen(m), "concluído") {
		t.Fatal(screen(m))
	}
}

func TestFalhaApareceNoRodape(t *testing.T) {
	m, _ := setup(t)
	m.Update(doneMsg{id: "projects/loja", label: "up", err: exec.ErrNotFound, output: "x\n✘ a porta 5080 está em uso\n"})
	if !strings.Contains(screen(m), "porta 5080") {
		t.Fatal(screen(m))
	}
}

func TestAcoesDependemDoEstado(t *testing.T) {
	m, _ := setup(t)
	load(m, viewServices)
	key(m, "2")
	load(m, viewServices)
	view := screen(m)
	if !hasHint(view, "s", "iniciar") || hasHint(view, "x", "parar") {
		t.Fatalf("parado:\n%s", view)
	}
	key(m, "down")
	view = screen(m)
	if !hasHint(view, "x", "parar") || hasHint(view, "s", "iniciar") {
		t.Fatalf("rodando:\n%s", view)
	}
	cmd := key(m, "s")
	if cmd != nil {
		t.Fatal("iniciar o que já roda deveria ser ignorado")
	}
}

func TestRemoverPedeConfirmacao(t *testing.T) {
	m, f := setup(t)
	if cmd := key(m, "ctrl+d"); cmd != nil {
		t.Fatal("não pode remover direto")
	}
	if m.mode != modeConfirm || !strings.Contains(screen(m), "loja") {
		t.Fatal(screen(m))
	}
	key(m, "n")
	if m.mode != modeTable || len(f.commands) != 0 {
		t.Fatal("n deveria cancelar")
	}
	key(m, "ctrl+d")
	key(m, "enter")
	if len(f.commands) != 0 {
		t.Fatal("Enter não confirma: o padrão é não")
	}
	key(m, "ctrl+d")
	cmd := key(m, "y")
	if cmd == nil {
		t.Fatal("y deveria confirmar")
	}
	cmd()
	if got := strings.Join(f.commands[0], " "); got != "rm loja --yes" {
		t.Fatal(got)
	}
}

func TestEnterMostraAcoes(t *testing.T) {
	m, f := setup(t)
	key(m, "enter")
	if m.mode != modeActions || !strings.Contains(screen(m), "publicar") {
		t.Fatal(screen(m))
	}
	key(m, "esc")
	if m.mode != modeTable {
		t.Fatal(m.mode)
	}
	key(m, "enter")
	cmd := key(m, "enter")
	if m.mode != modeTable || cmd == nil {
		t.Fatal("Enter deveria executar a primeira ação")
	}
	cmd()
	if got := strings.Join(f.commands[0], " "); got != "up loja -d" {
		t.Fatal(got)
	}
}

func TestCategorias(t *testing.T) {
	m, _ := setup(t)
	key(m, "3")
	load(m, viewTools)
	if len(m.visible()) != 3 {
		t.Fatal(m.visible())
	}
	key(m, "right")
	if rows := m.visible(); len(rows) != 2 || rows[0].Group != "k8s" {
		t.Fatalf("%+v", rows)
	}
	key(m, "/", "helm", "enter")
	if rows := m.visible(); len(rows) != 1 {
		t.Fatalf("%+v", rows)
	}
	key(m, "esc", "left")
	if m.category != len(m.categories()) {
		t.Fatal(m.category)
	}
}

func TestLogs(t *testing.T) {
	m, _ := setup(t)
	cmd := key(m, "l")
	if m.mode != modeLogs {
		t.Fatal(m.mode)
	}
	m.Update(cmd())
	if !strings.Contains(screen(m), "linha 2") {
		t.Fatal(screen(m))
	}
	key(m, "esc")
	if m.mode != modeTable {
		t.Fatal(m.mode)
	}
}

func TestDetalhesEAjuda(t *testing.T) {
	m, _ := setup(t)
	key(m, "d")
	if !strings.Contains(screen(m), "/p/loja") {
		t.Fatal(screen(m))
	}
	key(m, "esc", "?")
	if !strings.Contains(screen(m), "NAVEGAR") {
		t.Fatal(screen(m))
	}
	key(m, "esc")
	if m.mode != modeTable {
		t.Fatal(m.mode)
	}
}

func TestSistema(t *testing.T) {
	m, _ := setup(t)
	key(m, "4")
	load(m, viewSystem)
	view := screen(m)
	for _, want := range []string{"Memória", "Bateria", "80%", "claude", "#"} {
		if !strings.Contains(view, want) {
			t.Errorf("faltou %q:\n%s", want, view)
		}
	}
}

func TestListaLongaEmTelaPequenaCabe(t *testing.T) {
	m, _ := setup(t)
	var many []Row
	for i := 0; i < 200; i++ {
		many = append(many, Row{Key: "item", Status: "parado", Columns: []string{"item", "t", "parado", "url muito longa aqui para estourar a largura da tela pequena do celular", "x"}})
	}
	m.rows[viewProjects] = many
	for _, size := range [][2]int{{80, 24}, {40, 12}, {200, 60}} {
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		for i := 0; i < 120; i++ {
			key(m, "down")
		}
		lines := strings.Split(screen(m), "\n")
		if len(lines) != size[1] {
			t.Errorf("%dx%d: %d linhas", size[0], size[1], len(lines))
		}
		for _, l := range lines {
			if w := len([]rune(l)); w > size[0] {
				t.Errorf("%dx%d: linha de %d: %q", size[0], size[1], w, l)
			}
		}
	}
}

func TestSemIcones(t *testing.T) {
	m, _ := setup(t)
	for _, v := range []string{"1", "2", "3", "4", "?"} {
		key(m, v)
		load(m, m.view)
		for _, r := range screen(m) {
			if r >= 0x2190 && r != 0x2190 && r != 0x2192 && !isBoxDrawing(r) {
				t.Errorf("tela %s tem o símbolo %q", v, r)
			}
		}
		key(m, "esc")
	}
}

func TestCtrlCSaiDeQualquerLugar(t *testing.T) {
	for _, keys := range [][]string{{"/"}, {":"}, {"?"}, {"d"}, {"l"}, {"enter"}, {"ctrl+d"}} {
		m, _ := setup(t)
		key(m, keys...)
		if cmd := key(m, "ctrl+c"); cmd == nil || !m.quit {
			t.Errorf("%v: não saiu", keys)
		}
	}
}

// A moldura não é ícone: o desenho de caixa existe em qualquer terminal.
func isBoxDrawing(r rune) bool { return r >= 0x2500 && r <= 0x257F }

func hasHint(view, key, label string) bool {
	return regexp.MustCompile(`<` + regexp.QuoteMeta(key) + `>\s+` + regexp.QuoteMeta(label) + `\b`).MatchString(view)
}

func toolsView(t *testing.T) (*Model, *fakeBackend) {
	m, f := setup(t)
	key(m, "3")
	load(m, viewTools)
	return m, f
}

func TestVersaoDisponivelNaLista(t *testing.T) {
	m, f := toolsView(t)
	cmd := m.fetchUpstream("kubectl")
	m.Update(cmd())
	if len(f.upstreams) != 1 || !regexp.MustCompile(`kubectl\s+k8s\s+instalado\s+v1\.0\.0\s+2\.0\.0`).MatchString(screen(m)) {
		t.Fatal(screen(m))
	}
	if m.fetchUpstream("kubectl") != nil {
		t.Fatal("consultou duas vezes a mesma ferramenta")
	}
}

func TestListaMostraSempreAUltima(t *testing.T) {
	m, _ := toolsView(t)
	if !strings.Contains(screen(m), "ÚLTIMA") || !strings.Contains(screen(m), "...") {
		t.Fatal(screen(m))
	}
	cmd := m.fetchLatest()
	if cmd != nil {
		t.Fatal("a lista já pediu as últimas ao carregar; não pede de novo")
	}
	for _, r := range m.rows[viewTools] {
		m.Update(latestMsg{r.Key, tools.Upstream{Latest: "2.0.0"}})
	}
	if strings.Contains(screen(m), "...") || strings.Count(screen(m), "2.0.0") < 3 {
		t.Fatal(screen(m))
	}
}

func TestEnterEscolheOutraVersao(t *testing.T) {
	m, f := toolsView(t)
	cmd := key(m, "enter")
	if m.mode != modeVersions || cmd == nil {
		t.Fatal("Enter em Ferramentas abre as versões")
	}
	m.Update(cmd())
	view := screen(m)
	for _, want := range []string{"versões(kubectl)", "2.0.0", "última", "1.0.0", "instalada"} {
		if !strings.Contains(view, want) {
			t.Errorf("faltou %q:\n%s", want, view)
		}
	}
	key(m, "down", "enter")
	if m.mode != modeConfirm || !strings.Contains(screen(m), "Instalar kubectl 1.9.0?") {
		t.Fatal(screen(m))
	}
	run := key(m, "y")
	run()
	if got := strings.Join(f.commands[0], " "); got != "tools install kubectl --version v1.9.0 --yes" {
		t.Fatal(got)
	}
}

func TestFerramentaSemOutrasVersoes(t *testing.T) {
	m, _ := toolsView(t)
	key(m, "G")
	cmd := key(m, "enter")
	m.Update(cmd())
	if !strings.Contains(screen(m), "só a versão atual") {
		t.Fatal(screen(m))
	}
	if key(m, "enter"); m.mode != modeVersions {
		t.Fatal("sem versões, Enter não faz nada")
	}
	key(m, "esc")
	if m.mode != modeTable {
		t.Fatal(m.mode)
	}
}

func TestSoConsultaQuandoOCursorPara(t *testing.T) {
	m, f := toolsView(t)
	key(m, "down", "down")
	if len(f.upstreams) != 0 {
		t.Fatal("mover o cursor não pode ir à rede")
	}
	m.Update(settleMsg{"helm"})
	if len(f.upstreams) != 0 {
		t.Fatal("helm já não está selecionado")
	}
	cmd := func() tea.Cmd { _, c := m.Update(settleMsg{"psql"}); return c }()
	if cmd == nil {
		t.Fatal("psql está selecionado e deveria ser consultado")
	}
}

func TestInstalarMostraVersaoEOrigem(t *testing.T) {
	m, f := toolsView(t)
	key(m, "down")
	cmd := key(m, "i")
	if m.mode != modeConfirm || cmd == nil {
		t.Fatal("instalar pede confirmação e consulta a versão")
	}
	m.Update(cmd())
	view := screen(m)
	for _, want := range []string{"Instalar helm?", "2.0.0", "GitHub releases"} {
		if !strings.Contains(view, want) {
			t.Errorf("faltou %q:\n%s", want, view)
		}
	}
	run := key(m, "y")
	run()
	if got := strings.Join(f.commands[0], " "); got != "tools install helm --yes" {
		t.Fatal(got)
	}
}

func TestDetalhesDaFerramentaTemReadme(t *testing.T) {
	m, _ := toolsView(t)
	cmd := key(m, "d")
	m.Update(cmd())
	view := screen(m)
	for _, want := range []string{"README", "Guia de uso do kubectl", "github.com/x/kubectl", "atualização disponível"} {
		if !strings.Contains(view, want) {
			t.Errorf("faltou %q:\n%s", want, view)
		}
	}
}

func TestPararODockerAvisaEUsaOTerminal(t *testing.T) {
	m, f := setup(t)
	key(m, "2")
	load(m, viewServices)
	key(m, "G")
	if r, _ := m.selected(); r.Key != "docker" {
		t.Fatal(r.Key)
	}
	if !hasHint(screen(m), "x", "parar") {
		t.Fatal(screen(m))
	}
	if cmd := key(m, "x"); cmd != nil || m.mode != modeConfirm {
		t.Fatal("parar o Docker precisa de confirmação")
	}
	if !strings.Contains(screen(m), "todos os containers") {
		t.Fatal(screen(m))
	}
	cmd := key(m, "y")
	if cmd == nil || len(f.commands) != 1 || strings.Join(f.commands[0], " ") != "stop docker --yes" {
		t.Fatalf("deveria tomar o terminal com stop docker --yes: %v", f.commands)
	}
}
