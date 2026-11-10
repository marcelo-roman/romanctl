package tools

import (
	"regexp"
	"romanctl/internal/i18n"

	"romanctl/internal/platform"
)

func dotnetTool(key, name, desc, major, apt, pacmanName, apk, brewCask string) *Tool {
	t := &Tool{Key: key, Name: name, Group: GroupSDK, Description: desc, Detect: dotnetHasSDK(major),
		PostInstall: writeDotnetWrapper, AfterRemove: removeDotnetWrapperIfUnused,
		Pkgs: map[string][]string{platform.Apt: {apt}, platform.Dnf: {apt}, platform.Pacman: {pacmanName}, platform.Apk: {apk}},
		Brew: []string{brewCask}, Cask: true}
	return t
}

var dockerAsset = regexp.MustCompile(`docker-(\d+\.\d+\.\d+)\.tgz`)

func catalogSDK() []*Tool {
	return []*Tool{
		{Key: "node", Name: "Node.js", Group: GroupSDK, Description: i18n.T("Versão LTS, com npm"), Aliases: []string{"nodejs", "npm"},
			Version: []string{"node", "--version"}, Install: installNode, Brew: []string{"node"},
			Warning:     i18n.T("apaga também os pacotes globais do npm (mongosh, pnpm etc.); projetos Node, React e Next param de rodar"),
			RemovePaths: []string{opt("node"), "node", "npm", "npx", "corepack", "mongosh"}},
		{Key: "go", Name: "Go", Group: GroupSDK, Description: i18n.T("Última versão estável"), Aliases: []string{"golang"},
			Version: []string{"go", "version"}, Install: installGo, Brew: []string{"go"}, Warning: i18n.T("sem o Go não dá para recompilar o romanctl"),
			RemovePaths: []string{goDir(), "go", "gofmt"}},
		dotnetTool("dotnet", ".NET 10", "SDK LTS", "10", "dotnet-sdk-10.0", "dotnet-sdk", "dotnet10-sdk", "dotnet-sdk").aka(".net", "net", "dotnet-10", "dotnet10").
			warn(i18n.T("projetos .NET 10 e o template clean-ddd-cqrs param de compilar")),
		dotnetTool("dotnet-8", ".NET 8", i18n.T("SDK LTS anterior, ao lado do .NET 10"), "8", "dotnet-sdk-8.0", "dotnet-8.0", "dotnet8-sdk", "dotnet@8").aka("dotnet8"),
		{Key: "python", Name: "Python", Group: GroupSDK, Description: i18n.T("Python 3, com pip e venv"), Aliases: []string{"python3", "pip"},
			Detect: detectPython, Brew: []string{"python"},
			Pkgs: map[string][]string{platform.Apt: {"python3", "python3-venv", "python3-pip"}, platform.Dnf: {"python3", "python3-pip"},
				platform.Pacman: {"python", "python-pip"}, platform.Zypper: {"python3", "python3-pip"}, platform.Apk: {"python3", "py3-pip"}},
			RemovePkgs: map[string][]string{platform.Apt: {"python3-venv", "python3-pip"}, platform.Dnf: {"python3-pip"},
				platform.Pacman: {"python-pip"}, platform.Zypper: {"python3-pip"}, platform.Apk: {"py3-pip"}},
			Warning: i18n.T("no Linux remove só o pip e o venv: o Python faz parte do sistema e fica")},
		pkgTool("java", "Java", GroupSDK, i18n.T("OpenJDK 21, sem interface gráfica"), "javac", "openjdk-21-jdk-headless").aka("jdk", "openjdk").
			on(platform.Dnf, "java-21-openjdk-devel").on(platform.Pacman, "jdk21-openjdk").on(platform.Zypper, "java-21-openjdk-devel").on(platform.Apk, "openjdk21").brew("openjdk@21").ver("javac", "-version"),
		{Key: "kotlin", Name: "Kotlin", Group: GroupSDK, Description: i18n.T("Compilador Kotlin; precisa do Java"), Version: []string{"kotlinc", "-version"},
			Install: installKotlin, Brew: []string{"kotlin"}, RemovePaths: []string{opt("kotlin"), "kotlin", "kotlinc"}},
		pkgTool("scala", "Scala", GroupSDK, i18n.T("Compilador Scala; instala o Java junto"), "scala", "scala").ver("scala", "-version"),
		pkgTool("clojure", "Clojure", GroupSDK, i18n.T("Linguagem Clojure; instala o Java junto"), "clojure", "clojure").ver("clojure", "--version").on(platform.Dnf).on(platform.Zypper).on(platform.Apk),
		pkgTool("php", "PHP", GroupSDK, i18n.T("PHP com Composer"), "php", "php-cli", "php-mbstring", "php-xml", "php-curl", "php-zip", "php-pgsql", "composer").
			on(platform.Dnf, "php-cli", "php-mbstring", "php-xml", "php-pgsql", "composer").on(platform.Pacman, "php", "php-pgsql", "composer").
			on(platform.Zypper, "php8", "php8-mbstring", "php8-curl", "php8-pgsql", "php-composer2").on(platform.Apk, "php", "php-mbstring", "php-xml", "php-curl", "php-pdo_pgsql", "composer").brew("php", "composer"),
		{Key: "rust", Name: "Rust", Group: GroupSDK, Description: i18n.T("Instalado com o rustup"), Aliases: []string{"cargo", "rustc", "rustup"},
			Version: []string{"rustc", "--version"}, Install: installRust, Brew: []string{"rust"}, RemovePaths: rustPaths()},
		pkgTool("ruby", "Ruby", GroupSDK, i18n.T("Ruby com gem e bundler"), "ruby", "ruby-full", "ruby-bundler").on(platform.Dnf, "ruby", "rubygem-bundler").on(platform.Pacman, "ruby").on(platform.Zypper, "ruby").on(platform.Apk, "ruby", "ruby-bundler").brew("ruby"),
		pkgTool("elixir", "Elixir", GroupSDK, i18n.T("Elixir e Erlang"), "elixir", "elixir").ver("elixir", "--version").aka("erlang"),
		pkgTool("lua", "Lua", GroupSDK, "Lua 5.4", "lua", "lua5.4").on(platform.Dnf, "lua").on(platform.Pacman, "lua").on(platform.Zypper, "lua54").brew("lua").ver("lua5.4", "-v"),
		pkgTool("perl", "Perl", GroupSDK, i18n.T("Perl completo"), "perl", "perl"),
		pkgTool("r", "R", GroupSDK, i18n.T("Linguagem R, sem interface gráfica"), "R", "r-base-core").on(platform.Dnf, "R-core").on(platform.Pacman, "r").on(platform.Zypper, "R-base").on(platform.Apk, "R").brew("r").aka("r-lang"),
		pkgTool("haskell", "Haskell", GroupSDK, i18n.T("Compilador GHC"), "ghc", "ghc").aka("ghc"),
		pkgTool("gcc", "C e C++", GroupSDK, i18n.T("gcc, g++, make e libc de desenvolvimento"), "gcc", "build-essential").aka("build-essential", "c", "cpp", "c++").
			on(platform.Dnf, "gcc", "gcc-c++", "make").on(platform.Pacman, "base-devel").on(platform.Zypper, "gcc", "gcc-c++", "make").on(platform.Apk, "build-base").brew("gcc"),
		releaseTool("deno", "Deno", GroupSDK, i18n.T("Runtime de JavaScript e TypeScript"),
			Release{Repo: "denoland/deno", Asset: "deno-{uarch}-unknown-linux-gnu.zip", Sums: "deno-{uarch}-unknown-linux-gnu.zip.sha256sum", Bins: []string{"deno"}}),
		releaseTool("bun", "Bun", GroupSDK, i18n.T("Runtime e gerenciador de pacotes de JavaScript"),
			Release{Repo: "oven-sh/bun", Asset: "bun-linux-{m}.zip", Sums: "SHASUMS256.txt", Bins: []string{"bun"}, Map: map[string]string{"amd64": "x64", "arm64": "aarch64"}}),
		{Key: "zig", Name: "Zig", Group: GroupSDK, Description: i18n.T("Última versão estável"), Version: []string{"zig", "version"},
			Install: installZig, Brew: []string{"zig"}, RemovePaths: []string{opt("zig"), "zig"}},
		{Key: "dart", Name: "Dart", Group: GroupSDK, Description: i18n.T("SDK do Dart, sem o Flutter"), Version: []string{"dart", "--version"},
			Install: installDart, Brew: []string{"dart-lang/dart/dart"}, RemovePaths: []string{opt("dart-sdk"), "dart"}},
		{Key: "swift", Name: "Swift", Group: GroupSDK, Description: i18n.T("Toolchain oficial para Ubuntu; no macOS, instale o Xcode"), Version: []string{"swift", "--version"},
			Install: installSwift, RemovePaths: []string{opt("swift"), "swift", "swiftc"}},
	}
}
