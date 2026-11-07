package offline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
	"sync"
	"time"
)

func offlineStatus() {
	fmt.Println(ui.STitle.Render(i18n.T("Preparação offline")))
	if b, err := os.ReadFile(prepStamp); err == nil {
		i18n.Printf("  %-14s %s\n", i18n.T("Preparado em"), ui.SGreen.Render(string(b)))
	} else {
		i18n.Printf("  %-14s %s\n", i18n.T("Preparado em"), ui.SYellow.Render(i18n.T("nunca — rode: romanctl offline prep")))
	}
	conn := ui.SRed.Render(i18n.T("sem conexão"))
	if sys.IsOnline() {
		conn = ui.SGreen.Render(i18n.T("conectado"))
	}
	i18n.Printf("  %-14s %s\n", "Internet", conn)
	var tpls []string
	for _, k := range projects.OfflineTemplates {
		if projects.HasTemplate(k) {
			tpls = append(tpls, k)
		}
	}
	i18n.Printf("  %-14s %s\n", "Templates", strings.Join(tpls, ", "))
	caches := [][2]string{
		{i18n.T("Templates"), projects.TemplatesDir},
		{i18n.T("Cache npm"), filepath.Join(config.Home, ".npm")},
		{i18n.T("Cache Go"), goModCache()},
		{i18n.T("Cache NuGet"), filepath.Join(config.Home, ".nuget", "packages")},
	}
	for i, size := range cacheSizes(caches) {
		fmt.Printf("  %-14s %s\n", caches[i][0], size)
	}
	if out, err := sys.Output("code-server", "--list-extensions"); err == nil {
		i18n.Printf("  %-14s %d instaladas\n", i18n.T("Extensões"), len(strings.Fields(out)))
	}

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Sem internet")))
	fmt.Println(i18n.T("  romanctl new <tipo> <nome>             usa templates e cache sozinho"))
	fmt.Println(i18n.T("  npm install <pacote> --offline         instala do cache"))
	fmt.Println(i18n.T("  go get <módulo>                        usa o cache sozinho"))
	fmt.Println(i18n.T("  dotnet add package <pacote>            usa o cache se o pacote foi baixado"))

	fmt.Println("\n" + ui.STitle.Render(i18n.T("O que você precisa fazer manualmente")))
	fmt.Println(i18n.T("  • Abrir devdocs.io e ativar as documentações para uso offline"))
	fmt.Println(i18n.T("  • Testar de verdade: ligar o modo avião e rodar romanctl offline test"))
	fmt.Println(i18n.T("  • Sem internet, o Claude Code não funciona"))
}

// cacheSizes mede em paralelo e guarda por uma hora: percorrer esses caches arquivo por
// arquivo é o que fazia o comando parecer travado.
func cacheSizes(caches [][2]string) []string {
	cachePath := filepath.Join(projects.OfflineRoot, "tamanhos.json")
	type stored struct {
		At    time.Time        `json:"at"`
		Sizes map[string]int64 `json:"sizes"`
	}
	var saved stored
	if b, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(b, &saved)
	}
	out := make([]string, len(caches))
	if time.Since(saved.At) < time.Hour && saved.Sizes != nil {
		complete := true
		for i, c := range caches {
			size, ok := saved.Sizes[c[1]]
			if !ok {
				complete = false
				break
			}
			out[i] = sys.Human(size)
		}
		if complete {
			return out
		}
	}
	sizes := make([]int64, len(caches))
	ui.Spin(i18n.T("Calculando o tamanho dos caches..."), func(update func(string)) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		done := 0
		for i, c := range caches {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sizes[i] = sys.DirSize(c[1])
				mu.Lock()
				done++
				update(i18n.Sprintf("Calculando o tamanho dos caches %d/%d", done, len(caches)))
				mu.Unlock()
			}()
		}
		wg.Wait()
	})
	saved = stored{At: time.Now(), Sizes: map[string]int64{}}
	for i, c := range caches {
		saved.Sizes[c[1]] = sizes[i]
		out[i] = sys.Human(sizes[i])
	}
	if b, err := json.Marshal(saved); err == nil && os.MkdirAll(projects.OfflineRoot, 0o755) == nil {
		os.WriteFile(cachePath, b, 0o644)
	}
	return out
}
