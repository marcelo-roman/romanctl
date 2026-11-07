package cli

import (
	"fmt"
	"os/exec"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/platform"
	"romanctl/internal/services"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
)

func doctor() {
	i18n.Printf("%s %s\n\n", ui.STitle.Render(i18n.T("Plataforma")), platform.Current.Describe())
	fmt.Println(ui.STitle.Render(i18n.T("Ferramentas")))
	tools := [][]string{
		{"node", "node", "--version"},
		{"npm", "npm", "--version"},
		{"go", "go", "version"},
		{".NET", "dotnet", "--version"},
		{"git", "git", "--version"},
		{"code-server", "code-server", "--version"},
		{"PostgreSQL", "/usr/lib/postgresql/16/bin/postgres", "--version"},
		{"MongoDB", services.MongoHome + "/bin/mongod", "--version"},
		{"mongosh", "mongosh", "--version"},
		{"Redis", "redis-server", "--version"},
	}
	for _, t := range tools {
		if b := services.PgBin(); t[0] == "PostgreSQL" && b != "" {
			t[1] = b + "/postgres"
		}
		v := ui.SGray.Render(i18n.T("não instalado"))
		if _, err := exec.LookPath(t[1]); err == nil {
			if out, err := sys.Output(t[1], t[2:]...); err == nil {
				v = strings.SplitN(out, "\n", 2)[0]
			}
		}
		i18n.Printf("  %-12s %s\n", t[0], v)
	}
	if out, err := sys.Output("sh", "-c", "free -h | awk '/Mem:/{print $7\" livres de \"$2}'"); err == nil {
		i18n.Printf("\n  %-12s %s\n", i18n.T("Memória"), out)
	}
	if out, err := sys.Output("sh", "-c", "df -h "+config.Home+" | awk 'NR==2{print $4\" livres de \"$2}'"); err == nil {
		i18n.Printf("  %-12s %s\n", i18n.T("Disco"), out)
	}
	i18n.Printf("  %-12s %s\n  %-12s %s\n", i18n.T("Projetos"), config.ProjDir, i18n.T("Dados"), config.DataDir)
}
