package offline

import (
	"fmt"
	"os"
	"path/filepath"
	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/projects"
	"romanctl/internal/scaffold"
	"romanctl/internal/sys"
	"romanctl/internal/ui"
	"strings"
	"time"
)

var offlineNpm = []string{
	"express", "cors", "dotenv", "zod", "axios", "nodemon", "typescript",
	"react-router-dom", "zustand", "@tanstack/react-query", "lucide-react", "recharts",
	"@dnd-kit/core", "@dnd-kit/sortable", "tailwindcss", "@tailwindcss/vite",
	"pg", "mongoose", "mongodb", "redis", "jsonwebtoken", "bcryptjs", "geist",
}

var offlineGo = []string{
	"github.com/gin-gonic/gin", "github.com/go-chi/chi/v5", "github.com/jackc/pgx/v5",
	"go.mongodb.org/mongo-driver/v2/mongo", "github.com/redis/go-redis/v9", "github.com/joho/godotenv",
	"gorm.io/gorm", "gorm.io/driver/postgres", "github.com/spf13/cobra", "github.com/charmbracelet/huh",
}

var offlineNuget = []string{
	"Npgsql.EntityFrameworkCore.PostgreSQL", "Microsoft.EntityFrameworkCore.Design",
	"MongoDB.Driver", "StackExchange.Redis", "Swashbuckle.AspNetCore", "Serilog.AspNetCore",
}

var offlineExtensions = []string{
	"esbenp.prettier-vscode", "dbaeumer.vscode-eslint", "bradlc.vscode-tailwindcss",
	"golang.Go", "muhammad-sammy.csharp",
}

var prepStamp = filepath.Join(projects.OfflineRoot, "preparado-em")

func offlinePrep(force bool) error {
	if !sys.IsOnline() {
		return i18n.Errorf("the preparation needs internet")
	}
	os.MkdirAll(projects.TemplatesDir, 0o755)
	var failed []string
	step := func(title string, f func() error) {
		fmt.Println("\n" + ui.STitle.Render(title))
		if err := f(); err != nil {
			ui.Failf("%v", err)
			failed = append(failed, title)
			return
		}
		ui.Okf("Done")
	}
	step(i18n.T("Project templates"), func() error { return prepTemplates(force) })
	if os.Getenv("ROMANCTL_TEMPLATES") == "" {
		step(i18n.T("Architecture templates"), func() error { _, err := scaffold.Update(""); return err })
	}
	step(i18n.T("npm packages"), prepNpm)
	step(i18n.T("Go modules"), prepGo)
	step(i18n.T("NuGet packages"), prepNuget)
	step(i18n.T("code-server extensions"), prepExtensions)
	if len(failed) > 0 {
		return i18n.Errorf("failed steps: %s. Run again with internet", strings.Join(failed, ", "))
	}
	os.Remove(filepath.Join(projects.OfflineRoot, "tamanhos.json")) // the caches have just changed
	os.WriteFile(prepStamp, []byte(time.Now().Format("02/01/2006 15:04")), 0o644)
	fmt.Println()
	ui.Okf("All set. Check with: romanctl offline test")
	return nil
}

func prepTemplates(force bool) error {
	for _, key := range projects.OfflineTemplates {
		dst := filepath.Join(projects.TemplatesDir, key)
		if projects.HasTemplate(key) && !force {
			ui.Infof("%s already exists (use --force to redo it)", key)
			continue
		}
		ui.Infof("Generating %s...", key)
		tmp, err := os.MkdirTemp(projects.OfflineRoot, "generating-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmp)
		path := filepath.Join(tmp, key)
		if err := projects.FindType(key).Create(key, path); err != nil {
			return i18n.Errorf("template %s: %w", key, err)
		}
		os.RemoveAll(filepath.Join(path, ".next"))
		os.RemoveAll(filepath.Join(path, ".git"))
		os.RemoveAll(dst)
		if err := os.Rename(path, dst); err != nil {
			return err
		}
	}
	return nil
}

func prepNpm() error {
	tmp, err := os.MkdirTemp("", "romanctl-npm-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if out, err := sys.OutputIn(tmp, "npm", "init", "-y"); err != nil {
		return i18n.Errorf("npm init: %v %s", err, out)
	}
	ui.Infof("Downloading %d packages...", len(offlineNpm))
	args := append([]string{"install", "--legacy-peer-deps", "--ignore-scripts", "--no-audit", "--no-fund"}, offlineNpm...)
	return sys.Run(tmp, nil, "npm", args...)
}

func prepGo() error {
	// Tries the network and, if that fails for any reason, uses the cache itself as the proxy.
	proxy := "https://proxy.golang.org|file://" + goModCache() + "/cache/download|direct"
	if out, err := sys.Output("go", "env", "-w", "GOPROXY="+proxy); err != nil {
		return i18n.Errorf("go env: %v %s", err, out)
	}
	ui.Infof("GOPROXY configured to fall back to the local cache")
	tmp, err := os.MkdirTemp("", "romanctl-go-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if out, err := sys.OutputIn(tmp, "go", "mod", "init", "cache"); err != nil {
		return i18n.Errorf("go mod init: %v %s", err, out)
	}
	ui.Infof("Downloading %d modules...", len(offlineGo))
	if out, err := sys.OutputIn(tmp, "go", append([]string{"get"}, offlineGo...)...); err != nil {
		return i18n.Errorf("go get: %v %s", err, out)
	}
	if out, err := sys.OutputIn(tmp, "go", "mod", "download", "all"); err != nil {
		return i18n.Errorf("go mod download: %v %s", err, out)
	}
	ui.Infof("Installing gopls (used by the Go extension; this may take a while)...")
	if out, err := sys.OutputIn(tmp, "go", "install", "golang.org/x/tools/gopls@latest"); err != nil {
		return i18n.Errorf("gopls: %v %s", err, out)
	}
	return nil
}

func goModCache() string {
	out, err := sys.Output("go", "env", "GOMODCACHE")
	if err != nil || out == "" {
		return filepath.Join(config.Home, "go", "pkg", "mod")
	}
	return out
}

func prepNuget() error {
	tmp, err := os.MkdirTemp("", "romanctl-nuget-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	env := []string{"DOTNET_NOLOGO=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1"}
	for _, tpl := range []string{"webapi", "console", "mvc", "blazor"} {
		ui.Infof("Restoring template %s...", tpl)
		if err := sys.Quiet(tmp, env, "dotnet", "new", tpl, "-o", filepath.Join(tmp, tpl)); err != nil {
			return i18n.Errorf("dotnet new %s: %w", tpl, err)
		}
	}
	for _, t := range scaffold.LoadTemplates() {
		if t.Stack != "dotnet" {
			continue
		}
		for i, services := range [][]string{nil, t.Services} {
			ui.Infof("Restoring template %s (variant %d)...", t.Key(), i+1)
			dest := filepath.Join(tmp, i18n.Sprintf("arq-%s-%d", t.ID, i))
			if _, err := scaffold.WriteScaffold(t, "architecture", dest, services); err != nil {
				return i18n.Errorf("template %s: %w", t.Key(), err)
			}
			if err := sys.Quiet(dest, env, "dotnet", "restore"); err != nil {
				return i18n.Errorf("restore of %s: %w", t.Key(), err)
			}
		}
	}
	ui.Infof("Downloading %d packages...", len(offlineNuget))
	for _, pkg := range offlineNuget {
		if err := sys.Quiet(filepath.Join(tmp, "webapi"), env, "dotnet", "add", "package", pkg); err != nil {
			return i18n.Errorf("%s: %w", pkg, err)
		}
	}
	return nil
}

func prepExtensions() error {
	if !sys.Have("code-server") {
		return i18n.Errorf("code-server is not installed")
	}
	var failed []string
	for _, ext := range offlineExtensions {
		ui.Infof("Installing %s...", ext)
		if err := sys.Quiet("", nil, "code-server", "--install-extension", ext); err != nil {
			ui.Warnf("%s: %v", ext, err)
			failed = append(failed, ext)
		}
	}
	if len(failed) > 0 {
		return i18n.Errorf("not installed: %s", strings.Join(failed, ", "))
	}
	return nil
}
