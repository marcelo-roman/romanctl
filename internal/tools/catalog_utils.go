package tools

import (
	"romanctl/internal/i18n"
	"romanctl/internal/platform"
)

func catalogUtils() []*Tool {
	return []*Tool{
		pkgTool("jq", "jq", GroupShell, i18n.T("Processador de JSON"), "jq", "jq"),
		releaseTool("yq", "yq", GroupShell, i18n.T("Processador de YAML"),
			Release{Repo: "mikefarah/yq", Asset: "yq_linux_{arch}", Bins: []string{"yq"}}),
		pkgTool("ripgrep", "ripgrep", GroupShell, i18n.T("Busca rápida em arquivos"), "rg", "ripgrep").aka("rg"),
		pkgTool("fzf", "fzf", GroupShell, i18n.T("Busca interativa no terminal"), "fzf", "fzf"),
		pkgTool("fd", "fd", GroupShell, i18n.T("Alternativa ao find"), "fdfind", "fd-find").aka("fd-find").on(platform.Pacman, "fd").on(platform.Zypper, "fd").on(platform.Apk, "fd").brew("fd"),
		pkgTool("bat", "bat", GroupShell, i18n.T("cat com destaque de sintaxe"), "batcat", "bat"),
		releaseTool("eza", "eza", GroupShell, i18n.T("Alternativa ao ls"),
			Release{Repo: "eza-community/eza", Asset: "eza_{uarch}-unknown-linux-gnu.tar.gz", Bins: []string{"eza"}}),
		releaseTool("zoxide", "zoxide", GroupShell, i18n.T("cd que aprende seus caminhos"),
			Release{Repo: "ajeetdsouza/zoxide", Asset: "zoxide-{ver}-{uarch}-unknown-linux-musl.tar.gz", Bins: []string{"zoxide"}}),
		releaseTool("starship", "Starship", GroupShell, i18n.T("Prompt do terminal"),
			Release{Repo: "starship/starship", Asset: "starship-{uarch}-unknown-linux-musl.tar.gz", Sums: "starship-{uarch}-unknown-linux-musl.tar.gz.sha256", Bins: []string{"starship"}}),
		pkgTool("direnv", "direnv", GroupShell, i18n.T("Variáveis de ambiente por pasta"), "direnv", "direnv"),
		pkgTool("tmux", "tmux", GroupShell, i18n.T("Multiplexador de terminal"), "tmux", "tmux").ver("tmux", "-V"),
		{Key: "neovim", Name: "Neovim", Group: GroupShell, Description: i18n.T("Editor de terminal"), Aliases: []string{"nvim"},
			Version: []string{"nvim", "--version"}, Install: installNeovim, Brew: []string{"neovim"}, RemovePaths: []string{opt("nvim"), "nvim"}},
		pkgTool("vim", "Vim", GroupShell, i18n.T("Editor de terminal"), "vim", "vim"),
		pkgTool("tree", "tree", GroupShell, i18n.T("Mostra pastas em árvore"), "tree", "tree"),
		pkgTool("zip", "zip", GroupShell, i18n.T("zip e unzip"), "zip", "zip", "unzip").aka("unzip"),

		{Key: "ollama", Name: "Ollama", Group: GroupAI, Description: i18n.T("Modelos de IA locais; pacote grande"),
			Version: []string{"ollama", "--version"}, Install: installOllama, Brew: []string{"ollama"}, RemovePaths: []string{opt("ollama"), "ollama"}},
	}
}
