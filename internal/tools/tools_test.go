package tools

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/ui"
)

func keys(list []*Tool) string {
	var out []string
	for _, t := range list {
		out = append(out, t.Key)
	}
	return strings.Join(out, ",")
}

func TestResolveAceitaApelidosSemRepetir(t *testing.T) {
	got, err := Resolve([]string{"azure-cli", "az", "GOLANG"})
	if err != nil {
		t.Fatal(err)
	}
	if keys(got) != "az,go" {
		t.Fatalf("obtido %s", keys(got))
	}
}

func TestResolveExpandeGruposEAll(t *testing.T) {
	cloud, err := Resolve([]string{"cloud"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range cloud {
		if tool.Group != GroupCloud {
			t.Fatalf("%s não é do grupo cloud", tool.Key)
		}
	}
	all, err := Resolve([]string{"sdks", "all"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range all {
		if why := tool.Unavailable(); why != "" {
			t.Errorf("all trouxe %s, que não roda aqui: %s", tool.Key, why)
		}
	}
	if len(all) == 0 {
		t.Fatal("all veio vazio")
	}
}

func TestFerramentaDesconhecida(t *testing.T) {
	if _, err := Resolve([]string{"nao-existe"}); err == nil || !ui.IsUsage(err) {
		t.Fatalf("erro inesperado: %v", err)
	}
}

func TestRegistroNaoTemChavesRepetidas(t *testing.T) {
	seen := map[string]string{}
	for _, tool := range All {
		for _, name := range append([]string{tool.Key}, tool.Aliases...) {
			if other, dup := seen[name]; dup {
				t.Errorf("%q aponta para %s e %s", name, other, tool.Key)
			}
			seen[name] = tool.Key
		}
		if (tool.Install == nil && len(tool.Pkgs) == 0 && len(tool.Brew) == 0) || (tool.Detect == nil && len(tool.Version) == 0) {
			t.Errorf("%s está incompleta", tool.Key)
		}
	}
}

func TestFirstLineIgnoraAvisosEEspacos(t *testing.T) {
	got := firstLine("\nWARNING: algo\nVersion     v1.0  \noutra\n")
	if got != "Version v1.0" {
		t.Fatalf("obtido %q", got)
	}
}

func TestLinhaDaVersao(t *testing.T) {
	gradle := "openjdk version \"21\"\n\n------\nGradle 9.2.0\n------"
	if got := versionLine(gradle, "Gradle "); got != "Gradle 9.2.0" {
		t.Fatalf("%q", got)
	}
	if got := versionLine("go: downloading go1.27.3 (linux/amd64)\ngo version go1.27.3 linux/amd64", "go version"); got != "go version go1.27.3 linux/amd64" {
		t.Fatalf("%q", got)
	}
	if got := firstLine("info: kotlinc-jvm 2.4.21 (JRE 21)"); got != "kotlinc-jvm 2.4.21 (JRE 21)" {
		t.Fatalf("%q", got)
	}
}

func TestFirstLineTiraCores(t *testing.T) {
	if got := firstLine("\x1b[1mApache Maven 3.8.7\x1b[m\nMaven home: /usr"); got != "Apache Maven 3.8.7" {
		t.Fatalf("%q", got)
	}
}

func TestNomeDeFerramentaValeMaisQueGrupo(t *testing.T) {
	got, err := Resolve([]string{"node"})
	if err != nil || keys(got) != "node" {
		t.Fatalf("obtido %s, %v", keys(got), err)
	}
}

func TestTodoGrupoTemFerramentaETodaFerramentaTemGrupo(t *testing.T) {
	count := map[string]int{}
	for _, tool := range All {
		count[tool.Group]++
		if tool.Description == "" {
			t.Errorf("%s sem descrição", tool.Key)
		}
		if tool.RemoveNpm == "" && len(tool.Pkgs) == 0 && len(tool.RemovePkgs) == 0 && len(tool.RemovePaths) == 0 && tool.Brew == nil {
			t.Errorf("%s não sabe se remover", tool.Key)
		}
	}
	for _, g := range Groups {
		if count[g.Key] == 0 {
			t.Errorf("grupo %s vazio", g.Key)
		}
		delete(count, g.Key)
	}
	if len(count) > 0 {
		t.Errorf("grupos fora da listagem: %v", count)
	}
}

func TestPacotesVitaisSaoProtegidos(t *testing.T) {
	for _, pkg := range []string{"apt", "python3", "python3.12", "libc6", "coreutils"} {
		if !aptProtected.MatchString(pkg) {
			t.Errorf("%s deveria ser protegido", pkg)
		}
	}
	for _, pkg := range []string{"jq", "php-cli", "python3-pip", "aptitude"} {
		if aptProtected.MatchString(pkg) {
			t.Errorf("%s não deveria ser protegido", pkg)
		}
	}
}

func TestAptBlockedRecusaOQueFoiInstaladoExplicitamente(t *testing.T) {
	got := aptBlocked([]string{"perl", "git", "libfoo", "apt"}, []string{"perl"}, []string{"perl", "git"})
	if strings.Join(got, ",") != "git,apt (vital)" {
		t.Fatalf("obtido %v", got)
	}
	if got := aptBlocked([]string{"php-cli", "php8.3-cli"}, []string{"php-cli"}, []string{"php-cli"}); len(got) != 0 {
		t.Fatalf("dependência automática deveria sair: %v", got)
	}
}

func TestTodaFerramentaTemCaminhoNoMacOS(t *testing.T) {
	for _, tool := range All {
		// Runtimes de container e Kubernetes do Linux não têm versão para macOS.
		linuxOnly := []string{"swift", "buildah", "containerd", "nerdctl", "crictl", "k3s", "k0s"}
		if !slices.Contains(linuxOnly, tool.Key) && !tool.Portable && len(tool.Brew) == 0 {
			t.Errorf("%s não instala no macOS", tool.Key)
		}
	}
}

func TestAssetSemArquiteturaFixa(t *testing.T) {
	for _, tool := range All {
		for _, name := range tool.RemovePaths {
			if strings.Contains(name, "arm64") || strings.Contains(name, "aarch64") {
				t.Errorf("%s: arquitetura fixa em %s", tool.Key, name)
			}
		}
	}
	r := Release{Map: map[string]string{"amd64": "64bit", "arm64": "ARM64"}}
	got := assetVars(r, "v1.2", "1.2", platform.Info{Arch: "amd64"}).Replace("a-{ver}-{arch}-{uarch}-{xarch}-{x64}-{m}")
	if got != "a-1.2-amd64-x86_64-x86_64-x64-64bit" {
		t.Fatalf("obtido %s", got)
	}
	got = assetVars(r, "kustomize/v5", "5", platform.Info{Arch: "arm64"}).Replace("{shorttag}-{arch}-{uarch}-{xarch}-{x64}-{m}")
	if got != "v5-arm64-aarch64-arm64-arm64-ARM64" {
		t.Fatalf("obtido %s", got)
	}
}

func withPlatform(t *testing.T, p platform.Info) {
	t.Helper()
	saved := platform.Current
	platform.Current = p
	t.Cleanup(func() { platform.Current = saved })
}

func TestRequisitosDeCadaMaquina(t *testing.T) {
	desktop := platform.Info{OS: "linux", Arch: "amd64", Manager: platform.Dnf, Containers: true, Virtualization: true, GUI: true, Libc: "glibc"}
	server := desktop
	server.GUI, server.Virtualization = false, false
	proot := platform.Info{OS: "linux", Arch: "arm64", Manager: platform.Apt, Proot: true, Libc: "glibc"}
	alpine := platform.Info{OS: "linux", Arch: "amd64", Manager: platform.Apk, Containers: true, Libc: "musl"}
	mac := platform.Info{OS: "darwin", Arch: "arm64", Manager: platform.Brew, Containers: true, Virtualization: true, GUI: true}

	cases := []struct {
		machine platform.Info
		tool    string
		runs    bool
	}{
		{desktop, "docker", true}, {desktop, "rancher-desktop", true}, {desktop, "orbstack", false}, {desktop, "docker-desktop", false},
		{server, "docker", true}, {server, "rancher-desktop", false}, {server, "colima", false},
		{proot, "docker", false}, {proot, "podman", false}, {proot, "kind", false}, {proot, "docker-cli", true}, {proot, "kubectl", true}, {proot, "rancher", true},
		{alpine, "node", false}, {alpine, "aws", false}, {alpine, "kubectl", true}, {alpine, "docker", true},
		{mac, "orbstack", true}, {mac, "docker-desktop", true}, {mac, "containerd", false}, {mac, "node", true},
		{platform.Info{OS: "linux", Arch: "arm64", Manager: platform.Apt, Containers: true, GUI: true, Virtualization: true, Libc: "glibc"}, "rancher-desktop", false},
	}
	for _, c := range cases {
		withPlatform(t, c.machine)
		tool := Find(c.tool)
		if tool == nil {
			t.Fatalf("%s não existe", c.tool)
		}
		if why := tool.Unavailable(); (why == "") != c.runs {
			t.Errorf("%s em %+v: motivo %q, esperado rodar=%v", c.tool, c.machine, why, c.runs)
		}
	}
	withPlatform(t, platform.Info{OS: "linux", Arch: "arm64", Manager: platform.Apt, Proot: true, Libc: "glibc"})
	if why := Find("orbstack").Unavailable(); !strings.Contains(why, "macOS") {
		t.Errorf("OrbStack no Linux deveria falar do sistema: %q", why)
	}
	withPlatform(t, desktop)
	if got := strings.Join(Find("docker").pkgsFor(platform.Dnf), ","); got != "moby-engine,docker-compose" {
		t.Fatalf("pacotes errados: %s", got)
	}
}

func TestGrupoTrazSoOQueRoda(t *testing.T) {
	withPlatform(t, platform.Info{OS: "linux", Arch: "arm64", Manager: platform.Apt, Proot: true, Libc: "glibc"})
	got, err := Resolve([]string{"containers"})
	if err != nil {
		t.Fatal(err)
	}
	list := "," + keys(got) + ","
	if strings.Contains(list, ",docker,") || strings.Contains(list, ",podman,") || !strings.Contains(list, ",docker-cli,") {
		t.Fatalf("obtido %s", list)
	}
	// Pedido pelo nome, passa e é recusado na instalação, com o motivo.
	named, err := Resolve([]string{"docker"})
	if err != nil || keys(named) != "docker" {
		t.Fatalf("%s %v", keys(named), err)
	}
	// Sem o binário no PATH: a máquina que roda o teste pode ter Docker de verdade.
	t.Setenv("PATH", t.TempDir())
	if r := InstallOne(named[0], false); r.Action != "unavailable" || !strings.Contains(r.Error, "proot") {
		t.Fatalf("%+v", r)
	}
}

func TestTabelaDeRequisitosApontaParaFerramentasReais(t *testing.T) {
	for key := range requirements {
		if Find(key) == nil || Find(key).Key != key {
			t.Errorf("requisito de ferramenta inexistente: %s", key)
		}
	}
}

func TestDescricoesEAvisosTraduzidos(t *testing.T) {
	defer i18n.Set(i18n.Lang())
	i18n.Set(i18n.EN)
	for _, tool := range All {
		for _, text := range []string{tool.Description, tool.Warning} {
			// Os textos do catálogo são traduzidos quando o catálogo é montado, então aqui
			// só é preciso garantir que nenhum deles é uma frase em português sem entrada.
			if text != "" && i18n.T(text) == text && portuguese.MatchString(text) {
				t.Errorf("%s: sem tradução: %q", tool.Key, text)
			}
		}
	}
}

var portuguese = regexp.MustCompile(`[áàâãéêíóôõúç]| (de|do|da|para|com|sem|em|e) `)

func fakeBin(t *testing.T, dir, name, script string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestFerramentaLentaNaoSeguraALista(t *testing.T) {
	dir := t.TempDir()
	fakeBin(t, dir, "rapida", "echo rapida 1.0")
	fakeBin(t, dir, "lenta", "sleep 30")
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	t.Setenv("HOME", t.TempDir())

	savedAll, savedBudget := All, statusBudget
	defer func() { All, statusBudget = savedAll, savedBudget }()
	All = []*Tool{
		{Key: "rapida", Name: "Rapida", Version: []string{"rapida", "--version"}},
		{Key: "lenta", Name: "Lenta", Version: []string{"lenta", "--version"}},
		{Key: "ausente", Name: "Ausente", Version: []string{"nao-existe-mesmo", "--version"}},
	}
	statusBudget = 300 * time.Millisecond

	began := time.Now()
	var calls int
	got := Statuses(func(done, total int) { calls++ })
	if elapsed := time.Since(began); elapsed > 3*time.Second {
		t.Fatalf("demorou %s", elapsed)
	}
	byKey := map[string]Status{}
	for _, s := range got {
		byKey[s.Key] = s
	}
	if s := byKey["rapida"]; !s.Installed || s.Version != "rapida 1.0" {
		t.Errorf("%+v", s)
	}
	if s := byKey["lenta"]; !s.Installed || s.Version != "?" {
		t.Errorf("%+v", s)
	}
	if s := byKey["ausente"]; s.Installed {
		t.Errorf("%+v", s)
	}
	if calls != 2 {
		t.Errorf("progresso chamado %d vezes", calls)
	}
}

func TestCacheSoRepeteQuandoOBinarioMuda(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "chamadas")
	fakeBin(t, dir, "ferramenta", "echo x >> "+log+"; echo v1")
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	home := t.TempDir()
	cache := &versionCache{entries: map[string]cachedVersion{}}
	tool := &Tool{Key: "ferramenta", Version: []string{"ferramenta"}}
	_ = home

	for i := 0; i < 3; i++ {
		if v, ok := tool.cachedDetect(cache); !ok || v != "v1" {
			t.Fatalf("%q %v", v, ok)
		}
	}
	calls, _ := os.ReadFile(log)
	if strings.Count(string(calls), "x") != 1 {
		t.Fatalf("executou %d vezes", strings.Count(string(calls), "x"))
	}
	bin := filepath.Join(dir, "ferramenta")
	future := time.Now().Add(time.Hour)
	fakeBin(t, dir, "ferramenta", "echo v2")
	os.Chtimes(bin, future, future)
	if v, _ := tool.cachedDetect(cache); v != "v2" {
		t.Fatalf("depois de trocar o binário: %q", v)
	}
	os.Remove(bin)
	if _, ok := tool.cachedDetect(cache); ok {
		t.Fatal("binário removido não deveria contar como instalado")
	}
}

func TestNomesDeCategoria(t *testing.T) {
	for name, want := range map[string]string{
		"db": GroupDB, "banco de dados": GroupDB, "Banco-de-Dados": GroupDB, "BANCO": GroupDB, "Clientes de banco de dados": GroupDB,
		"linguagens": GroupSDK, "Linguagens de Programação": GroupSDK, "sdk": GroupSDK, "segurança": GroupSec, "k8s": GroupK8s,
	} {
		if got, ok := FindGroup(name); !ok || got != want {
			t.Errorf("%q: %q %v", name, got, ok)
		}
	}
	if _, ok := FindGroup("inventada"); ok {
		t.Error("categoria inventada")
	}
	if GroupTitle(GroupDB) == GroupDB {
		t.Error("sem título")
	}
}

func TestBinarioQueNaoENosso(t *testing.T) {
	ours := []string{"/home/a/.local/bin", "/home/a/.local/share/romanctl/tools"}
	cases := []struct {
		path, goos string
		managed    bool
		want       bool
	}{
		{"/usr/bin/jq", "darwin", false, true},
		{"/usr/bin/kubectl", "linux", false, true},            // o runner já tinha
		{"/home/a/.local/bin/kubectl", "linux", false, false}, // o nosso continua lá
		{"/usr/bin/jq", "linux", true, false},                 // apt removeu e sobrou: falha
		{termuxPrefix + "/bin/jq", "linux", true, true},
		{"/home/a/.local/share/romanctl/tools/go/bin/go", "linux", false, false},
	}
	for _, c := range cases {
		if got := isExternal(c.path, c.goos, c.managed, ours); got != c.want {
			t.Errorf("%s em %s (pacote=%v): %v", c.path, c.goos, c.managed, got)
		}
	}
}

func TestDownloadConfereChecksum(t *testing.T) {
	const body = "conteúdo"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/faltando" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(body))
	}))
	defer srv.Close()
	sum := sha256.Sum256([]byte(body))

	if _, err := download(srv.URL+"/a", t.TempDir(), hex.EncodeToString(sum[:])); err != nil {
		t.Fatalf("checksum certo falhou: %v", err)
	}
	if _, err := download(srv.URL+"/a", t.TempDir(), strings.Repeat("0", 64)); err == nil {
		t.Fatal("aceitou checksum errado")
	}
	if _, err := download(srv.URL+"/faltando", t.TempDir(), ""); err == nil {
		t.Fatal("aceitou HTTP 404")
	}
}

func TestSumFromFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("aaa  outro.tar.gz\nbbb *gh_linux_arm64.tar.gz\n"))
	}))
	defer srv.Close()
	got, err := sumFromFile(srv.URL, "gh_linux_arm64.tar.gz")
	if err != nil || got != "bbb" {
		t.Fatalf("obtido %q, %v", got, err)
	}
	if _, err := sumFromFile(srv.URL, "nao-tem"); err == nil {
		t.Fatal("deveria falhar sem o arquivo na lista")
	}
}

func TestExtractZipRecusaCaminhoParaForaDoDestino(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "mau.zip")
	f, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	entry, _ := w.Create("../fora.txt")
	entry.Write([]byte("x"))
	w.Close()
	f.Close()

	dest := filepath.Join(dir, "destino")
	if err := extractZip(file, dest); err == nil {
		t.Fatal("aceitou caminho inseguro")
	}
	if _, err := os.Stat(filepath.Join(dir, "fora.txt")); err == nil {
		t.Fatal("gravou fora do destino")
	}
}
