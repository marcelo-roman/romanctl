Output docs/demo.mp4

Set Shell bash
Set FontFamily "JetBrains Mono"
Set FontSize 15
Set Width 1440
Set Height 860
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 35ms
Set CursorBlink false

Hide
Type "export ROMANCTL_LANG=en ROMANCTL_PROJETOS=$HOME/romanctl-demo ROMANCTL_TEMPLATES=$PWD/templates PS1='\[\e[1;35m\]demo\[\e[0m\] \[\e[36m\]\W\[\e[0m\] $ ' && mkdir -p ~/romanctl-demo && cd ~/romanctl-demo && clear"
Enter
Sleep 1s
Show

# ------------------------------------------------------------ basics
Type "# romanctl: tools, projects, services and system in a single CLI"
Enter
Sleep 1s
Type "romanctl version"
Enter
Sleep 2s
Type "romanctl help"
Enter
Sleep 6s
Type "clear"
Enter

Type "# Without arguments: the command list. Type to filter, Enter runs it"
Enter
Type "romanctl"
Enter
Sleep 2s
Type@120ms "platform"
Sleep 1.5s
Enter
Sleep 4s
Enter
Sleep 1s
Escape
Sleep 500ms
Escape
Sleep 1s
Ctrl+C
Sleep 300ms
Type "clear"
Enter

# ------------------------------------------------------------ access
Type "# Access: login through Keycloak (browser + PKCE), whoami, logout and iam"
Enter
Type "romanctl login --help"
Enter
Sleep 4s
Type "romanctl whoami --help && romanctl logout --help"
Enter
Sleep 4s
Type "romanctl iam --help | head -30"
Enter
Sleep 5s
Type "clear"
Enter

# ------------------------------------------------------------ system
Type "# System: platform, doctor and machine status"
Enter
Type "romanctl platform"
Enter
Sleep 4s
Type "romanctl doctor"
Enter
Sleep 5s
Type "clear"
Enter
Type "romanctl system status"
Enter
Sleep 5s
Type "romanctl system battery"
Enter
Sleep 2s
Type "clear"
Enter

# ------------------------------------------------------------ tools
Type "# Tools: more than 170, filtered by what runs on this machine"
Enter
Type "romanctl tools categories"
Enter
Sleep 4s
Type "clear"
Enter
Type "romanctl tools list --category k8s --plain"
Enter
Sleep 4s
Type "romanctl tools info jq"
Enter
Sleep 4s
Type "clear"
Enter
Type "romanctl tools versions helm | head -8"
Enter
Sleep 4s
Type "romanctl tools install helm --version 3.18.4 --dry-run"
Enter
Sleep 4s
Type "romanctl tools uninstall jq --dry-run"
Enter
Sleep 3s
Type "clear"
Enter
Type "# The interactive list: type to filter, arrows to change category"
Enter
Type "romanctl tools list"
Enter
Sleep 2s
Right
Sleep 1s
Right
Sleep 1s
Left
Sleep 700ms
Left
Sleep 700ms
Type@120ms "terra"
Sleep 2s
Escape
Sleep 500ms
Escape
Sleep 1s
Ctrl+C
Sleep 300ms
Type "clear"
Enter

# ------------------------------------------------------------ templates
Type "# Templates and generators"
Enter
Type "romanctl templates"
Enter
Sleep 3s
Type "romanctl types"
Enter
Sleep 4s
Type "clear"
Enter
Type "romanctl catalog"
Enter
Sleep 6s
Type "clear"
Enter

# ------------------------------------------------------------ projects
Type "# Projects: create one from an architecture template"
Enter
Type "romanctl new shop --stack go --arch hexagonal,ddd --no-git"
Enter
Sleep 5s
Type "cd shop && romanctl add --list"
Enter
Sleep 3s
Type "romanctl add resource Product"
Enter
Sleep 4s
Type "clear"
Enter
Type "romanctl list --plain"
Enter
Sleep 3s
Type "# run: in the foreground, Ctrl+C stops it"
Enter
Type "romanctl run shop"
Enter
Sleep 6s
Ctrl+C
Sleep 1s
Type "clear"
Enter
Type "# up: services, database, .env and the app; -d runs it in the background"
Enter
Type "romanctl up -d"
Enter
Sleep 5s
Type `curl -s -X POST localhost:8090/products -H 'content-type: application/json' -d '{"name":"Pen"}'; echo`
Enter
Sleep 2s
Type "curl -s localhost:8090/products; echo"
Enter
Sleep 2s
Type "romanctl logs shop"
Enter
Sleep 4s
Ctrl+C
Sleep 1s
Type "clear"
Enter
Type "# code and open: code-server and the browser"
Enter
Type "romanctl code --help && romanctl open --help"
Enter
Sleep 5s
Type "clear"
Enter
Type "# publish: CI, Dockerfile, commit and a GitHub repository"
Enter
Type "romanctl publish --dry-run"
Enter
Sleep 4s
Type "romanctl down"
Enter
Sleep 3s
Type "clear"
Enter

# ------------------------------------------------------------ services
Type "# Services in Docker: PostgreSQL, MongoDB, Redis, MySQL, Kafka..."
Enter
Type "romanctl status --all"
Enter
Sleep 4s
Type "romanctl install postgres"
Enter
Sleep 6s
Type "romanctl start postgres"
Enter
Sleep 8s
Type "clear"
Enter
Type "romanctl status"
Enter
Sleep 3s
Type "romanctl logs postgres"
Enter
Sleep 4s
Ctrl+C
Sleep 1s
Type "clear"
Enter
Type "romanctl restart postgres"
Enter
Sleep 8s
Type "romanctl stop postgres"
Enter
Sleep 4s
Type "romanctl restore status"
Enter
Sleep 4s
Type "clear"
Enter

# ------------------------------------------------------------ maintenance
Type "# Offline, Claude Code skills, config and updates"
Enter
Type "romanctl offline status"
Enter
Sleep 5s
Type "clear"
Enter
Type "romanctl claude status"
Enter
Sleep 3s
Type "# config: the language picker (Esc cancels) and the current setting"
Enter
Type "romanctl config"
Enter
Sleep 3s
Escape
Sleep 1s
Ctrl+C
Sleep 300ms
Type "romanctl config --json"
Enter
Sleep 3s
Type "romanctl self-update --check"
Enter
Sleep 4s
Type "clear"
Enter
Type "# web-ui: the web interface on localhost"
Enter
Type "romanctl web-ui"
Enter
Sleep 4s
Ctrl+C
Sleep 1s
Type "clear"
Enter

# ------------------------------------------------------------ dashboards
Type "# ui: the full-screen dashboard, in the spirit of k9s"
Enter
Type "romanctl ui"
Enter
Sleep 4s
Tab
Sleep 3s
Tab
Sleep 3s
Tab
Sleep 3s
Type "?"
Sleep 3s
Escape
Sleep 1s
Type "q"
Sleep 2s
Type "clear"
Enter
Type "# menu: the simpler list version"
Enter
Type "romanctl menu"
Enter
Sleep 3s
Down
Sleep 700ms
Down
Sleep 700ms
Down
Sleep 2s
Escape
Sleep 1s
Escape
Sleep 1s
Ctrl+C
Sleep 300ms
Type "clear"
Enter

# ------------------------------------------------------------ cleanup
Type "# rm: removes a project and all its files"
Enter
Type "cd .. && romanctl rm shop -y"
Enter
Sleep 3s
Type "romanctl list --plain"
Enter
Sleep 4s
