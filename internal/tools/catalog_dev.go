package tools

import (
	"romanctl/internal/i18n"
	"romanctl/internal/platform"
)

func catalogDev() []*Tool {
	return []*Tool{
		npmTool("pnpm", "pnpm", GroupNode, i18n.T("Gerenciador de pacotes"), "pnpm", "pnpm", "pnpx"),
		npmTool("yarn", "Yarn", GroupNode, i18n.T("Gerenciador de pacotes"), "yarn", "yarn", "yarnpkg"),
		npmTool("typescript", "TypeScript", GroupNode, i18n.T("Compilador tsc"), "typescript", "tsc", "tsserver").aka("tsc"),
		npmTool("tsx", "tsx", GroupNode, i18n.T("Roda TypeScript direto"), "tsx", "tsx"),
		npmTool("nest", "Nest CLI", GroupNode, i18n.T("Projetos NestJS"), "@nestjs/cli", "nest").aka("nestjs"),
		npmTool("angular", "Angular CLI", GroupNode, i18n.T("Projetos Angular"), "@angular/cli", "ng").aka("ng"),
		npmTool("pm2", "PM2", GroupNode, i18n.T("Gerenciador de processos Node"), "pm2", "pm2"),
		npmTool("prisma", "Prisma", GroupNode, i18n.T("ORM e migrações"), "prisma", "prisma"),
		npmTool("vercel", "Vercel CLI", GroupNode, i18n.T("Deploy na Vercel"), "vercel", "vercel"),
		npmTool("netlify", "Netlify CLI", GroupNode, i18n.T("Deploy na Netlify"), "netlify-cli", "netlify"),
		npmTool("wrangler", "Wrangler", GroupNode, "Cloudflare Workers", "wrangler", "wrangler"),
		npmTool("firebase", "Firebase CLI", GroupNode, i18n.T("Projetos Firebase"), "firebase-tools", "firebase"),
		npmTool("aws-cdk", "AWS CDK", GroupNode, i18n.T("Infraestrutura AWS em código"), "aws-cdk", "cdk").aka("cdk"),
		npmTool("serverless", "Serverless", GroupNode, i18n.T("Framework Serverless"), "serverless", "serverless", "sls").aka("sls"),

		releaseTool("uv", "uv", GroupPython, i18n.T("Gerenciador rápido de projetos Python"),
			Release{Repo: "astral-sh/uv", Asset: "uv-{uarch}-unknown-linux-gnu.tar.gz", Sums: "uv-{uarch}-unknown-linux-gnu.tar.gz.sha256", Bins: []string{"uv", "uvx"}}),
		venvTool("poetry", "Poetry", GroupPython, i18n.T("Gerenciador de projetos Python"), "poetry", "poetry"),
		pkgTool("pipx", "pipx", GroupPython, i18n.T("Instala aplicativos Python isolados"), "pipx", "pipx"),
		venvTool("ruff", "Ruff", GroupPython, i18n.T("Linter e formatador"), "ruff", "ruff"),
		venvTool("pre-commit", "pre-commit", GroupPython, i18n.T("Ganchos de commit"), "pre-commit", "pre-commit"),

		{Key: "dotnet-ef", Name: "EF Core CLI", Group: GroupSDK, Description: i18n.T("Migrações do Entity Framework; precisa do .NET"),
			Aliases: []string{"ef"}, Version: []string{"dotnet-ef", "--version"}, Install: installDotnetEF,
			Portable: true, AfterRemove: removeDotnetEF, RemovePaths: []string{"dotnet-ef"}},

		pkgTool("make", "Make", GroupBuild, i18n.T("Automação de build"), "make", "make"),
		pkgTool("cmake", "CMake", GroupBuild, i18n.T("Gerador de build para C e C++"), "cmake", "cmake"),
		pkgTool("ninja", "Ninja", GroupBuild, i18n.T("Build rápido"), "ninja", "ninja-build").on(platform.Pacman, "ninja").on(platform.Zypper, "ninja").on(platform.Apk, "ninja").brew("ninja"),
		pkgTool("maven", "Maven", GroupBuild, i18n.T("Build para Java; instala o Java junto"), "mvn", "maven").aka("mvn"),
		pkgTool("gradle", "Gradle", GroupBuild, i18n.T("Build para Java e Kotlin"), "gradle", "gradle"),
		pkgTool("protoc", "Protocol Buffers", GroupBuild, i18n.T("Compilador protoc"), "protoc", "protobuf-compiler").aka("protobuf").on(platform.Dnf, "protobuf-compiler").on(platform.Pacman, "protobuf").on(platform.Zypper, "protobuf-devel").on(platform.Apk, "protoc").brew("protobuf"),
		releaseTool("just", "just", GroupBuild, i18n.T("Executor de comandos do projeto"),
			Release{Repo: "casey/just", Asset: "just-{ver}-{uarch}-unknown-linux-musl.tar.gz", Sums: "SHA256SUMS", Bins: []string{"just"}}),
		releaseTool("task", "Task", GroupBuild, i18n.T("Executor de tarefas em YAML"),
			Release{Repo: "go-task/task", Asset: "task_linux_{arch}.tar.gz", Sums: "task_checksums.txt", Bins: []string{"task"}}).aka("go-task").brew("go-task"),
		pkgTool("shellcheck", "ShellCheck", GroupBuild, i18n.T("Análise de scripts shell"), "shellcheck", "shellcheck"),
		releaseTool("shfmt", "shfmt", GroupBuild, i18n.T("Formatador de scripts shell"),
			Release{Repo: "mvdan/sh", Asset: "shfmt_{tag}_linux_{arch}", Bins: []string{"shfmt"}}),

		releaseTool("gh", "GitHub CLI", GroupGit, i18n.T("CLI do GitHub"),
			Release{Repo: "cli/cli", Asset: "gh_{ver}_linux_{arch}.tar.gz", Sums: "gh_{ver}_checksums.txt", Bins: []string{"gh"}}).aka("github").also(opt("gh")),
		pkgTool("git", "Git", GroupGit, i18n.T("Controle de versão"), "git", "git"),
		pkgTool("git-lfs", "Git LFS", GroupGit, i18n.T("Arquivos grandes no Git"), "git-lfs", "git-lfs").aka("lfs"),
		releaseTool("lazygit", "lazygit", GroupGit, i18n.T("Interface de terminal para Git"),
			Release{Repo: "jesseduffield/lazygit", Asset: "lazygit_{ver}_linux_{xarch}.tar.gz", Sums: "checksums.txt", Bins: []string{"lazygit"}}),
		pkgTool("tig", "tig", GroupGit, i18n.T("Navegador de histórico do Git"), "tig", "tig"),
		releaseTool("delta", "delta", GroupGit, i18n.T("Diff com destaque de sintaxe"),
			Release{Repo: "dandavison/delta", Asset: "delta-{ver}-{uarch}-unknown-linux-gnu.tar.gz", Bins: []string{"delta"}}).brew("git-delta"),
		releaseTool("act", "act", GroupGit, i18n.T("GitHub Actions local; precisa de um Docker remoto"),
			Release{Repo: "nektos/act", Asset: "act_Linux_{xarch}.tar.gz", Sums: "checksums.txt", Bins: []string{"act"}}),
		{Key: "code-server", Name: "code-server", Group: GroupShell, Description: i18n.T("VS Code no navegador"), Aliases: []string{"code"},
			Version: []string{"code-server", "--version"}, Install: installCodeServer, Brew: []string{"code-server"},
			RemovePkgs: map[string][]string{platform.Apt: {"code-server"}}, RemovePaths: []string{opt("code-server"), "code-server"},
			Warning: i18n.T("pare o serviço antes: romanctl stop code")},
	}
}

// also acrescenta caminhos à remoção de uma ferramenta.
func (t *Tool) also(paths ...string) *Tool {
	t.RemovePaths = append(t.RemovePaths, paths...)
	return t
}
