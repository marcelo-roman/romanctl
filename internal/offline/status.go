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
	fmt.Println(ui.STitle.Render(i18n.T("Offline preparation")))
	if b, err := os.ReadFile(prepStamp); err == nil {
		i18n.Printf("  %-14s %s\n", i18n.T("Prepared on"), ui.SGreen.Render(string(b)))
	} else {
		i18n.Printf("  %-14s %s\n", i18n.T("Prepared on"), ui.SYellow.Render(i18n.T("never — run: romanctl offline prep")))
	}
	conn := ui.SRed.Render(i18n.T("no connection"))
	if sys.IsOnline() {
		conn = ui.SGreen.Render(i18n.T("connected"))
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
		{i18n.T("npm cache"), filepath.Join(config.Home, ".npm")},
		{i18n.T("Go cache"), goModCache()},
		{i18n.T("NuGet cache"), filepath.Join(config.Home, ".nuget", "packages")},
	}
	for i, size := range cacheSizes(caches) {
		fmt.Printf("  %-14s %s\n", caches[i][0], size)
	}
	if out, err := sys.Output("code-server", "--list-extensions"); err == nil {
		i18n.Printf("  %-14s %d installed\n", i18n.T("Extensions"), len(strings.Fields(out)))
	}

	fmt.Println("\n" + ui.STitle.Render(i18n.T("Without internet")))
	fmt.Println(i18n.T("  romanctl new <type> <name>             uses templates and cache on its own"))
	fmt.Println(i18n.T("  npm install <package> --offline        installs from the cache"))
	fmt.Println(i18n.T("  go get <module>                        uses the cache on its own"))
	fmt.Println(i18n.T("  dotnet add package <package>           uses the cache if the package was downloaded"))

	fmt.Println("\n" + ui.STitle.Render(i18n.T("What you need to do manually")))
	fmt.Println(i18n.T("  • Open devdocs.io and enable the documentation for offline use"))
	fmt.Println(i18n.T("  • Test for real: turn on airplane mode and run romanctl offline test"))
	fmt.Println(i18n.T("  • Claude Code does not work without internet"))
}

// cacheSizes measures in parallel and keeps the result for an hour: walking these caches file by
// file is what made the command look stuck.
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
	ui.Spin(i18n.T("Calculating cache sizes..."), func(update func(string)) {
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
				update(i18n.Sprintf("Calculating cache sizes %d/%d", done, len(caches)))
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
