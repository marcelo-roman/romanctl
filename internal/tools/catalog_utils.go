package tools

import (
	"romanctl/internal/i18n"
	"romanctl/internal/platform"
)

func catalogUtils() []*Tool {
	return []*Tool{
		pkgTool("jq", "jq", GroupShell, i18n.T("JSON processor"), "jq", "jq"),
		releaseTool("yq", "yq", GroupShell, i18n.T("YAML processor"),
			Release{Repo: "mikefarah/yq", Asset: "yq_linux_{arch}", Bins: []string{"yq"}}),
		pkgTool("ripgrep", "ripgrep", GroupShell, i18n.T("Fast file search"), "rg", "ripgrep").aka("rg"),
		pkgTool("fzf", "fzf", GroupShell, i18n.T("Interactive terminal search"), "fzf", "fzf"),
		pkgTool("fd", "fd", GroupShell, i18n.T("Alternative to find"), "fdfind", "fd-find").aka("fd-find").on(platform.Pacman, "fd").on(platform.Zypper, "fd").on(platform.Apk, "fd").brew("fd"),
		pkgTool("bat", "bat", GroupShell, i18n.T("cat with syntax highlighting"), "batcat", "bat"),
		releaseTool("eza", "eza", GroupShell, i18n.T("Alternative to ls"),
			Release{Repo: "eza-community/eza", Asset: "eza_{uarch}-unknown-linux-gnu.tar.gz", Bins: []string{"eza"}}),
		releaseTool("zoxide", "zoxide", GroupShell, i18n.T("cd that learns your paths"),
			Release{Repo: "ajeetdsouza/zoxide", Asset: "zoxide-{ver}-{uarch}-unknown-linux-musl.tar.gz", Bins: []string{"zoxide"}}),
		releaseTool("starship", "Starship", GroupShell, i18n.T("Terminal prompt"),
			Release{Repo: "starship/starship", Asset: "starship-{uarch}-unknown-linux-musl.tar.gz", Sums: "starship-{uarch}-unknown-linux-musl.tar.gz.sha256", Bins: []string{"starship"}}),
		pkgTool("direnv", "direnv", GroupShell, i18n.T("Per-folder environment variables"), "direnv", "direnv"),
		pkgTool("tmux", "tmux", GroupShell, i18n.T("Terminal multiplexer"), "tmux", "tmux").ver("tmux", "-V"),
		{Key: "neovim", Name: "Neovim", Group: GroupShell, Description: i18n.T("Terminal editor"), Aliases: []string{"nvim"},
			Version: []string{"nvim", "--version"}, Install: installNeovim, Brew: []string{"neovim"}, RemovePaths: []string{opt("nvim"), "nvim"}},
		pkgTool("vim", "Vim", GroupShell, i18n.T("Terminal editor"), "vim", "vim"),
		pkgTool("tree", "tree", GroupShell, i18n.T("Shows folders as a tree"), "tree", "tree"),
		pkgTool("zip", "zip", GroupShell, i18n.T("zip and unzip"), "zip", "zip", "unzip").aka("unzip"),

		{Key: "ollama", Name: "Ollama", Group: GroupAI, Description: i18n.T("Local AI models; large package"),
			Version: []string{"ollama", "--version"}, Install: installOllama, Brew: []string{"ollama"}, RemovePaths: []string{opt("ollama"), "ollama"}},
	}
}
