package tools

import (
	"romanctl/internal/i18n"
	"romanctl/internal/platform"
)

func catalogDev() []*Tool {
	return []*Tool{
		npmTool("pnpm", "pnpm", GroupNode, i18n.T("Package manager"), "pnpm", "pnpm", "pnpx"),
		npmTool("yarn", "Yarn", GroupNode, i18n.T("Package manager"), "yarn", "yarn", "yarnpkg"),
		npmTool("typescript", "TypeScript", GroupNode, i18n.T("tsc compiler"), "typescript", "tsc", "tsserver").aka("tsc"),
		npmTool("tsx", "tsx", GroupNode, i18n.T("Runs TypeScript directly"), "tsx", "tsx"),
		npmTool("nest", "Nest CLI", GroupNode, i18n.T("NestJS projects"), "@nestjs/cli", "nest").aka("nestjs"),
		npmTool("angular", "Angular CLI", GroupNode, i18n.T("Angular projects"), "@angular/cli", "ng").aka("ng"),
		npmTool("pm2", "PM2", GroupNode, i18n.T("Node process manager"), "pm2", "pm2"),
		npmTool("prisma", "Prisma", GroupNode, i18n.T("ORM and migrations"), "prisma", "prisma"),
		npmTool("vercel", "Vercel CLI", GroupNode, i18n.T("Deploy to Vercel"), "vercel", "vercel"),
		npmTool("netlify", "Netlify CLI", GroupNode, i18n.T("Deploy to Netlify"), "netlify-cli", "netlify"),
		npmTool("wrangler", "Wrangler", GroupNode, "Cloudflare Workers", "wrangler", "wrangler"),
		npmTool("firebase", "Firebase CLI", GroupNode, i18n.T("Firebase projects"), "firebase-tools", "firebase"),
		npmTool("aws-cdk", "AWS CDK", GroupNode, i18n.T("AWS infrastructure in code"), "aws-cdk", "cdk").aka("cdk"),
		npmTool("serverless", "Serverless", GroupNode, i18n.T("Serverless framework"), "serverless", "serverless", "sls").aka("sls"),

		releaseTool("uv", "uv", GroupPython, i18n.T("Fast Python project manager"),
			Release{Repo: "astral-sh/uv", Asset: "uv-{uarch}-unknown-linux-gnu.tar.gz", Sums: "uv-{uarch}-unknown-linux-gnu.tar.gz.sha256", Bins: []string{"uv", "uvx"}}),
		venvTool("poetry", "Poetry", GroupPython, i18n.T("Python project manager"), "poetry", "poetry"),
		pkgTool("pipx", "pipx", GroupPython, i18n.T("Installs isolated Python applications"), "pipx", "pipx"),
		venvTool("ruff", "Ruff", GroupPython, i18n.T("Linter and formatter"), "ruff", "ruff"),
		venvTool("pre-commit", "pre-commit", GroupPython, i18n.T("Commit hooks"), "pre-commit", "pre-commit"),

		{Key: "dotnet-ef", Name: "EF Core CLI", Group: GroupSDK, Description: i18n.T("Entity Framework migrations; needs .NET"),
			Aliases: []string{"ef"}, Version: []string{"dotnet-ef", "--version"}, Install: installDotnetEF,
			Portable: true, AfterRemove: removeDotnetEF, RemovePaths: []string{"dotnet-ef"}},

		pkgTool("make", "Make", GroupBuild, i18n.T("Build automation"), "make", "make"),
		pkgTool("cmake", "CMake", GroupBuild, i18n.T("Build generator for C and C++"), "cmake", "cmake"),
		pkgTool("ninja", "Ninja", GroupBuild, i18n.T("Fast builds"), "ninja", "ninja-build").on(platform.Pacman, "ninja").on(platform.Zypper, "ninja").on(platform.Apk, "ninja").brew("ninja"),
		pkgTool("maven", "Maven", GroupBuild, i18n.T("Build for Java; installs Java too"), "mvn", "maven").aka("mvn"),
		{Key: "gradle", Name: "Gradle", Group: GroupBuild, Description: i18n.T("Build for Java and Kotlin"),
			Version: []string{"gradle", "--version"}, VersionLine: "Gradle ", Install: installGradle, Brew: []string{"gradle"},
			RemovePaths: []string{opt("gradle"), "gradle"}},
		pkgTool("protoc", "Protocol Buffers", GroupBuild, i18n.T("protoc compiler"), "protoc", "protobuf-compiler").aka("protobuf").on(platform.Dnf, "protobuf-compiler").on(platform.Pacman, "protobuf").on(platform.Zypper, "protobuf-devel").on(platform.Apk, "protoc").brew("protobuf"),
		releaseTool("just", "just", GroupBuild, i18n.T("Project command runner"),
			Release{Repo: "casey/just", Asset: "just-{ver}-{uarch}-unknown-linux-musl.tar.gz", Sums: "SHA256SUMS", Bins: []string{"just"}}),
		releaseTool("task", "Task", GroupBuild, i18n.T("YAML task runner"),
			Release{Repo: "go-task/task", Asset: "task_linux_{arch}.tar.gz", Sums: "task_checksums.txt", Bins: []string{"task"}}).aka("go-task").brew("go-task"),
		pkgTool("shellcheck", "ShellCheck", GroupBuild, i18n.T("Shell script analysis"), "shellcheck", "shellcheck"),
		releaseTool("shfmt", "shfmt", GroupBuild, i18n.T("Shell script formatter"),
			Release{Repo: "mvdan/sh", Asset: "shfmt_{tag}_linux_{arch}", Bins: []string{"shfmt"}}),

		releaseTool("gh", "GitHub CLI", GroupGit, i18n.T("GitHub CLI"),
			Release{Repo: "cli/cli", Asset: "gh_{ver}_linux_{arch}.tar.gz", Sums: "gh_{ver}_checksums.txt", Bins: []string{"gh"}}).aka("github").also(opt("gh")),
		pkgTool("git", "Git", GroupGit, i18n.T("Version control"), "git", "git"),
		pkgTool("git-lfs", "Git LFS", GroupGit, i18n.T("Large files in Git"), "git-lfs", "git-lfs").aka("lfs"),
		releaseTool("lazygit", "lazygit", GroupGit, i18n.T("Terminal interface for Git"),
			Release{Repo: "jesseduffield/lazygit", Asset: "lazygit_{ver}_linux_{xarch}.tar.gz", Sums: "checksums.txt", Bins: []string{"lazygit"}}),
		pkgTool("tig", "tig", GroupGit, i18n.T("Git history browser"), "tig", "tig"),
		releaseTool("delta", "delta", GroupGit, i18n.T("Diff with syntax highlighting"),
			Release{Repo: "dandavison/delta", Asset: "delta-{ver}-{uarch}-unknown-linux-gnu.tar.gz", Bins: []string{"delta"}}).brew("git-delta"),
		releaseTool("act", "act", GroupGit, i18n.T("Local GitHub Actions; needs a remote Docker"),
			Release{Repo: "nektos/act", Asset: "act_Linux_{xarch}.tar.gz", Sums: "checksums.txt", Bins: []string{"act"}}),
		{Key: "code-server", Name: "code-server", Group: GroupShell, Description: i18n.T("VS Code in the browser"), Aliases: []string{"code"},
			Version: []string{"code-server", "--version"}, Install: installCodeServer, Brew: []string{"code-server"},
			RemovePkgs: map[string][]string{platform.Apt: {"code-server"}}, RemovePaths: []string{opt("code-server"), "code-server"},
			Warning: i18n.T("stop the service first: romanctl stop code")},
	}
}

// also adds paths to a tool's removal.
func (t *Tool) also(paths ...string) *Tool {
	t.RemovePaths = append(t.RemovePaths, paths...)
	return t
}
