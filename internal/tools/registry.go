// Package tools instala ferramentas de desenvolvimento em Linux e macOS.
package tools

import (
	"context"
	"os/exec"
	"romanctl/internal/i18n"
	"slices"
	"strings"
	"time"

	"romanctl/internal/platform"
	"romanctl/internal/ui"
)

const (
	GroupCloud  = "cloud"
	GroupK8s    = "k8s"
	GroupIaC    = "iac"
	GroupImages = "containers"
	GroupSec    = "security"
	GroupDB     = "db"
	GroupObs    = "observability"
	GroupNet    = "network"
	GroupGit    = "git"
	GroupBuild  = "build"
	GroupShell  = "shell"
	GroupSDK    = "sdk"
	GroupNode   = "node-tools"
	GroupPython = "python-tools"
	GroupAI     = "ai"
)

// Groups define a ordem e os títulos da listagem.
var Groups = []struct{ Key, Title string }{
	{GroupSDK, i18n.T("Linguagens de programação e SDKs")}, {GroupNode, i18n.T("Ferramentas de Node.js")}, {GroupPython, i18n.T("Ferramentas de Python")},
	{GroupBuild, i18n.T("Build e automação")}, {GroupGit, "Git e GitHub"}, {GroupCloud, i18n.T("Nuvem")}, {GroupK8s, "Kubernetes"},
	{GroupIaC, i18n.T("Infraestrutura como código")}, {GroupImages, i18n.T("Containers e imagens")}, {GroupSec, i18n.T("Segurança")},
	{GroupDB, i18n.T("Clientes de banco de dados")}, {GroupObs, i18n.T("Observabilidade e carga")}, {GroupNet, i18n.T("Rede")},
	{GroupShell, i18n.T("Terminal e produtividade")}, {GroupAI, i18n.T("IA")},
}

var groupAliases = map[string]string{
	"nuvem": GroupCloud, "kubernetes": GroupK8s, "sdks": GroupSDK, "linguagens": GroupSDK, "linguagem": GroupSDK,
	"languages": GroupSDK, "programacao": GroupSDK, "linguagens-de-programacao": GroupSDK, "programming-languages": GroupSDK,
	"infra": GroupIaC, "infraestrutura": GroupIaC, "docker-tools": GroupImages, "imagens": GroupImages, "seguranca": GroupSec,
	"banco": GroupDB, "bancos": GroupDB, "banco-de-dados": GroupDB, "bancos-de-dados": GroupDB, "database": GroupDB, "databases": GroupDB,
	"monitoramento": GroupObs, "observabilidade": GroupObs, "rede": GroupNet, "redes": GroupNet, "net": GroupNet, "terminal": GroupShell, "cli": GroupShell,
	"node": "", "ia": GroupAI,
}

// Tool é uma ferramenta instalável. Detect substitui a checagem por versão quando
// o binário é compartilhado, como o dotnet com vários SDKs.
type Tool struct {
	Key         string
	Name        string
	Group       string
	Description string
	Aliases     []string
	Version     []string
	Detect      func() (string, bool)
	Install     func() error        // instalador próprio, usado no Linux
	Portable    bool                // Install também vale no macOS (npm, venv etc.)
	Pkgs        map[string][]string // pacotes por gerenciador; "*" vale para todos
	RemovePkgs  map[string][]string // o que remover, se for diferente de Pkgs
	Brew        []string            // fórmulas do Homebrew no macOS
	Cask        bool                // Brew são casks
	PostInstall func() error
	ProotReason string   // por que não roda em proot
	Arches      []string // vazio significa todas
	Next        string   // o que fazer depois de instalar
	Warning     string   // o que mais deixa de funcionar ao remover
	RemoveNpm   string   // pacote global do npm a remover
	RemovePaths []string
	AfterRemove func() error
}

func (t *Tool) aka(aliases ...string) *Tool { t.Aliases = aliases; return t }
func (t *Tool) ver(args ...string) *Tool    { t.Version = args; return t }
func (t *Tool) warn(w string) *Tool         { t.Warning = w; return t }

// on troca os pacotes de um gerenciador; sem nomes, a ferramenta não existe nele.
func (t *Tool) on(manager string, pkgs ...string) *Tool {
	if t.Pkgs == nil {
		t.Pkgs = map[string][]string{}
	}
	t.Pkgs[manager] = pkgs
	return t
}

func (t *Tool) brew(formulas ...string) *Tool { t.Brew = formulas; return t }
func (t *Tool) cask(casks ...string) *Tool    { t.Brew, t.Cask = casks, true; return t }
func (t *Tool) noBrew() *Tool                 { t.Brew = nil; return t }
func (t *Tool) next(text string) *Tool        { t.Next = text; return t }
func (t *Tool) only(arches ...string) *Tool   { t.Arches = arches; return t }

// Unavailable explica por que a ferramenta não pode ser instalada nesta máquina.
func (t *Tool) Unavailable() string {
	// O sistema vem primeiro: "só existe para macOS" explica mais que a falta de uma capacidade.
	if why := t.unsupportedSystem(); why != "" {
		return why
	}
	p := platform.Current
	for _, need := range requirements[t.Key] {
		// glibc só importa onde o Linux instala o binário oficial.
		if need == platform.NeedGlibc && p.OS == "darwin" {
			continue
		}
		if why := p.Missing(need); why != "" {
			return why
		}
	}
	return ""
}

func (t *Tool) unsupportedSystem() string {
	p := platform.Current
	if len(t.Arches) > 0 && !slices.Contains(t.Arches, p.Arch) {
		return i18n.Sprintf("não há versão para %s", p.Arch)
	}
	if p.OS == "darwin" {
		if t.Portable {
			return ""
		}
		if len(t.Brew) == 0 {
			return i18n.T("só existe para Linux")
		}
		if p.Manager != platform.Brew {
			return i18n.T("precisa do Homebrew: https://brew.sh")
		}
		return ""
	}
	if t.Install != nil {
		return ""
	}
	if len(t.Pkgs) == 0 {
		return i18n.T("só existe para macOS")
	}
	if len(t.pkgsFor(p.Manager)) == 0 {
		if p.Manager == "" {
			return i18n.T("nenhum gerenciador de pacotes reconhecido nesta máquina")
		}
		return i18n.Sprintf("não há pacote para %s", p.Manager)
	}
	return ""
}

func (t *Tool) pkgsFor(manager string) []string {
	if pkgs, ok := t.Pkgs[manager]; ok {
		return pkgs
	}
	return t.Pkgs["*"]
}

func (t *Tool) removePkgsFor(manager string) []string {
	if t.RemovePkgs != nil {
		if pkgs, ok := t.RemovePkgs[manager]; ok {
			return pkgs
		}
		return t.RemovePkgs["*"]
	}
	return t.pkgsFor(manager)
}

// runInstall escolhe o caminho de instalação certo para esta máquina.
func (t *Tool) runInstall() error {
	p := platform.Current
	var err error
	switch {
	case p.OS == "darwin" && !t.Portable:
		err = brewInstall(t.Brew, t.Cask)
	case t.Install != nil:
		err = t.Install()
	default:
		err = pkgInstall(t.pkgsFor(p.Manager))
	}
	if err == nil && t.PostInstall != nil {
		err = t.PostInstall()
	}
	return err
}

// All é o catálogo completo, definido nos arquivos catalog_*.go.
var All = slices.Concat(catalogSDK(), catalogDev(), catalogInfra(), catalogUtils())

// Find busca uma ferramenta pelo nome ou por um apelido.
func Find(name string) *Tool {
	name = strings.ToLower(name)
	for _, t := range All {
		if t.Key == name || slices.Contains(t.Aliases, name) {
			return t
		}
	}
	return nil
}

func isGroup(name string) (string, bool) {
	// Aceita "Banco de dados", "banco-de-dados" e "BANCO".
	name = strings.Join(strings.Fields(strings.ReplaceAll(ui.Fold(name), "-", " ")), "-")
	if alias, ok := groupAliases[name]; ok && alias != "" {
		return alias, true
	}
	for _, g := range Groups {
		title := strings.Join(strings.Fields(ui.Fold(g.Title)), "-")
		if g.Key == name || title == name {
			return g.Key, true
		}
	}
	return "", false
}

// FindGroup traduz um nome de categoria na chave dela.
func FindGroup(name string) (string, bool) { return isGroup(name) }

// GroupTitle devolve o nome de uma categoria.
func GroupTitle(key string) string {
	for _, g := range Groups {
		if g.Key == key {
			return g.Title
		}
	}
	return key
}

// GroupKeys lista as ferramentas de um grupo ou de "all"; nil se name não for grupo.
// Um nome de ferramenta vale mais que um grupo: "node" é a ferramenta.
func GroupKeys(name string) []string {
	var keys []string
	if strings.EqualFold(name, "all") || strings.EqualFold(name, "tudo") {
		for _, t := range All {
			keys = append(keys, t.Key)
		}
		return keys
	}
	if Find(name) != nil {
		return nil
	}
	group, ok := isGroup(name)
	if !ok {
		return nil
	}
	for _, t := range All {
		if t.Group == group {
			keys = append(keys, t.Key)
		}
	}
	return keys
}

// Resolve traduz nomes, grupos e "all" em ferramentas, sem repetir nenhuma. Grupos trazem
// só o que roda nesta máquina; um nome pedido diretamente passa, e a instalação explica a recusa.
func Resolve(names []string) ([]*Tool, error) {
	var out []*Tool
	add := func(t *Tool) {
		if !slices.Contains(out, t) {
			out = append(out, t)
		}
	}
	for _, name := range names {
		if strings.EqualFold(name, "all") || strings.EqualFold(name, "tudo") {
			for _, t := range All {
				if t.Unavailable() == "" {
					add(t)
				}
			}
			continue
		}
		// Um nome de ferramenta vale mais que um grupo: "node" é a ferramenta.
		if t := Find(name); t != nil {
			add(t)
			continue
		}
		if group, ok := isGroup(name); ok {
			for _, t := range All {
				if t.Group == group && t.Unavailable() == "" {
					add(t)
				}
			}
			continue
		}
		return nil, ui.Usagef("ferramenta ou grupo desconhecido %q (veja: romanctl tools list)", name)
	}
	return out, nil
}

// Detected informa a versão instalada. A ferramenta conta como instalada só se rodar.
func (t *Tool) Detected() (string, bool) {
	if t.Detect != nil {
		return t.Detect()
	}
	if _, err := exec.LookPath(t.Version[0]); err != nil {
		return "", false
	}
	out, err := versionOutput(t.Version...)
	if err != nil {
		return "", false
	}
	return firstLine(out), true
}

// versionOutput roda um comando de versão com prazo. WaitDelay cobre o programa que deixa um
// processo filho com a saída aberta: sem ele, a espera passaria do prazo.
func versionOutput(command ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.WaitDelay = 2 * time.Second
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "WARNING") {
			return strings.Join(strings.Fields(line), " ")
		}
	}
	return ""
}

// Status é o estado de uma ferramenta, em tela ou em --json.
type Status struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Group       string `json:"group"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
	Version     string `json:"version"`
	Unavailable string `json:"unavailable,omitempty"`
}

// Statuses consulta as ferramentas em paralelo. A versão vem do cache enquanto o binário não
// mudar, porque rodar 170 programas a cada listagem é o que deixava a tela parada.
func Statuses(progress func(done, total int)) []Status {
	cache := loadVersionCache()
	type finished struct {
		index     int
		version   string
		installed bool
	}
	results := make(chan finished, len(All))
	slots := make(chan struct{}, 10)
	for i, t := range All {
		go func() {
			slots <- struct{}{}
			defer func() { <-slots }()
			version, installed := t.cachedDetect(cache)
			results <- finished{i, version, installed}
		}()
	}
	out := make([]Status, len(All))
	seen := make([]bool, len(All))
	deadline := time.After(statusBudget)
	done := 0
collect:
	for done < len(All) {
		select {
		case r := <-results:
			t := All[r.index]
			out[r.index] = Status{t.Key, t.Name, t.Group, i18n.T(t.Description), r.installed, r.version, t.Unavailable()}
			seen[r.index] = true
			done++
			if progress != nil {
				progress(done, len(All))
			}
		case <-deadline:
			break collect
		}
	}
	// Uma ferramenta que não responde não pode segurar a lista inteira.
	for i, t := range All {
		if !seen[i] {
			installed := false
			if len(t.Version) > 0 {
				_, err := exec.LookPath(t.Version[0])
				installed = err == nil
			}
			out[i] = Status{t.Key, t.Name, t.Group, i18n.T(t.Description), installed, "?", t.Unavailable()}
		}
	}
	cache.save()
	return out
}

// statusBudget é o máximo que uma listagem espera.
var statusBudget = 12 * time.Second

// Result é o resultado da instalação de uma ferramenta.
type Result struct {
	Key     string `json:"key"`
	Action  string `json:"action"` // installed, skipped ou failed
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

// InstallOne instala e confirma que a ferramenta realmente roda depois.
func InstallOne(t *Tool, force bool) Result {
	if version, ok := t.Detected(); ok && !force {
		ui.Okf("%s já está instalado: %s", t.Name, version)
		return Result{Key: t.Key, Action: "skipped", Version: version}
	}
	if why := t.Unavailable(); why != "" {
		ui.Failf("%s: %s", t.Name, why)
		return Result{Key: t.Key, Action: "unavailable", Error: why}
	}
	var err0 error
	ui.Spin(i18n.Sprintf("Instalando %s...", t.Name), func(func(string)) { err0 = t.runInstall() })
	if err := err0; err != nil {
		ui.Failf("%s: %v", t.Name, err)
		return Result{Key: t.Key, Action: "failed", Error: err.Error()}
	}
	version, ok := t.Detected()
	if !ok {
		err := i18n.Sprintf("%s foi instalado, mas não rodou", t.Name)
		ui.Failf("%s", err)
		// Não deixa para trás algo que não funciona, às vezes com gigabytes.
		if rmErr := t.remove(); rmErr != nil {
			ui.Warnf("%v", rmErr)
		}
		return Result{Key: t.Key, Action: "failed", Error: err}
	}
	ui.Okf("%s instalado: %s", t.Name, version)
	if t.Next != "" {
		ui.Infof("Próximo passo: %s", t.Next)
	}
	return Result{Key: t.Key, Action: "installed", Version: version}
}

// isExternal diz se o binário que sobrou não é nosso: o do Termux e os que vêm com o macOS,
// como o jq em /usr/bin, que nem o Homebrew nem o romanctl podem tirar.
func isExternal(path, goos string) bool {
	if strings.HasPrefix(path, termuxPrefix) {
		return true
	}
	return goos == "darwin" && (strings.HasPrefix(path, "/usr/bin/") || strings.HasPrefix(path, "/bin/"))
}

// UninstallOne remove o que a declaração da ferramenta indica e confere o resultado.
func UninstallOne(t *Tool) Result {
	if _, ok := t.Detected(); !ok && !hasLeftovers(t) {
		ui.Infof("%s não está instalado", t.Name)
		return Result{Key: t.Key, Action: "skipped"}
	}
	var rmErr0 error
	ui.Spin(i18n.Sprintf("Removendo %s...", t.Name), func(func(string)) { rmErr0 = t.remove() })
	if err := rmErr0; err != nil {
		ui.Failf("%s: %v", t.Name, err)
		return Result{Key: t.Key, Action: "failed", Error: err.Error()}
	}
	if _, still := t.Detected(); still {
		where := t.Key
		if len(t.Version) > 0 {
			if p, err := exec.LookPath(t.Version[0]); err == nil {
				where = p
			}
		}
		if isExternal(where, platform.Current.OS) {
			ui.Warnf("%s continua disponível pelo sistema, fora do romanctl: %s", t.Name, where)
			return Result{Key: t.Key, Action: "external", Version: where}
		}
		err := i18n.Sprintf("%s continua instalado em %s", t.Name, where)
		ui.Failf("%s", err)
		return Result{Key: t.Key, Action: "failed", Error: err}
	}
	ui.Okf("%s removido", t.Name)
	return Result{Key: t.Key, Action: "removed"}
}
