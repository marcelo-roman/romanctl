package cli

import (
	"fmt"
	"strings"

	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/ui"
)

// text guarda os dois idiomas juntos: cada ajuda é lida de uma vez, sem chave para procurar.
type text struct{ pt, en string }

func (t text) String() string {
	if i18n.Lang() == i18n.EN {
		return t.en
	}
	return t.pt
}

type helpOption struct {
	flag string
	desc text
}

// command é a ajuda de um comando; o caminho tem uma ou duas palavras.
type command struct {
	path     string
	aliases  []string
	usage    string
	summary  text
	details  text
	options  []helpOption
	examples []string
}

var (
	optJSON   = helpOption{"--json", text{"saída em JSON, sem menus", "JSON output, without menus"}}
	optFilter = helpOption{"--filter <texto>", text{"só o que contém o texto, sem lista interativa", "only what contains the text, without an interactive list"}}
	optPlain  = helpOption{"--plain", text{"imprime tudo, sem lista interativa", "prints everything, without an interactive list"}}
)

var commands = []command{
	{path: "status", aliases: []string{"st"}, usage: "romanctl status [--all] [--json]",
		summary:  text{"Mostra os serviços instalados e o Docker", "Shows the installed services and Docker"},
		details:  text{"Um serviço só conta como instalado quando o container dele existe; ter o Docker ou a imagem não basta.", "A service only counts as installed when its container exists; having Docker or the image is not enough."},
		options:  []helpOption{{"--all", text{"inclui o que dá para instalar", "includes what can be installed"}}, optJSON},
		examples: []string{"romanctl status", "romanctl status --all"}},
	{path: "start", aliases: []string{"iniciar"}, usage: "romanctl start [serviço|all]",
		summary:  text{"Inicia serviços; sem argumento, abre uma lista", "Starts services; without an argument, opens a list"},
		examples: []string{"romanctl start", "romanctl start postgres", "romanctl start all"}},
	{path: "stop", aliases: []string{"parar"}, usage: "romanctl stop [serviço|all] [--yes]",
		details:  text{"Parar o docker pede confirmação, porque derruba todos os containers da máquina; all nunca inclui o docker.", "Stopping docker asks for confirmation, because it brings down every container on the machine; all never includes docker."},
		summary:  text{"Para serviços; sem argumento, abre uma lista", "Stops services; without an argument, opens a list"},
		examples: []string{"romanctl stop", "romanctl stop all"}},
	{path: "restart", aliases: []string{"reiniciar"}, usage: "romanctl restart [serviço|all]",
		summary:  text{"Reinicia serviços; sem argumento, abre uma lista", "Restarts services; without an argument, opens a list"},
		examples: []string{"romanctl restart postgres"}},
	{path: "install", aliases: []string{"instalar"}, usage: "romanctl install [serviço]",
		summary:  text{"Prepara um serviço: baixa a imagem, se faltar, e cria o container parado", "Prepares a service: downloads the image if missing and creates the stopped container"},
		details:  text{"Para instalar ferramentas, como kubectl, use: romanctl tools install", "To install tools such as kubectl, use: romanctl tools install"},
		examples: []string{"romanctl install kafka"}},
	{path: "logs", aliases: []string{"log"}, usage: "romanctl logs [serviço]",
		summary:  text{"Acompanha o log de um serviço; Ctrl+C sai", "Follows a service's log; Ctrl+C exits"},
		examples: []string{"romanctl logs", "romanctl logs postgres"}},
	{path: "restore", aliases: []string{"restaurar"}, usage: "romanctl restore [status|save|clear]",
		summary:  text{"Sobe de novo o que estava ligado antes de uma queda", "Starts again what was running before a crash"},
		details:  text{"Em um terminal, sem subcomando, abre um menu. Sem terminal, restaura direto.", "In a terminal, without a subcommand, it opens a menu. Without a terminal, it restores directly."},
		examples: []string{"romanctl restore", "romanctl restore status"}},
	{path: "restore status", usage: "romanctl restore status",
		summary: text{"Mostra o que está salvo e o que caiu", "Shows what is saved and what crashed"}},
	{path: "restore save", aliases: []string{"salvar"}, usage: "romanctl restore save",
		summary: text{"Salva como estado o que está rodando agora", "Saves what is running now as the state"}},
	{path: "restore clear", aliases: []string{"limpar"}, usage: "romanctl restore clear",
		summary: text{"Limpa o estado: o restore não sobe mais nada", "Clears the state: restore starts nothing anymore"}},

	{path: "new", aliases: []string{"novo", "create"}, usage: "romanctl new [tipo] [nome]\n     romanctl new <nome> --stack <stack> --arch <a,b> [--with <serviços>]",
		summary: text{"Cria um projeto, perguntando o que faltar", "Creates a project, asking for whatever is missing"},
		options: []helpOption{
			{"--stack <stack>", text{"linguagem do template, como dotnet", "template language, such as dotnet"}},
			{"--arch <a,b>", text{"arquiteturas, como clean, ddd, cqrs e hexagonal; romanctl catalog lista todas", "architectures, such as clean, ddd, cqrs and hexagonal; romanctl catalog lists them all"}},
			{"--with <serviços>", text{"serviços do projeto, como postgres", "project services, such as postgres"}},
			{"--sem-git", text{"não cria repositório git", "does not create a git repository"}},
			{"--offline", text{"usa só templates e cache, sem baixar dependências", "uses only templates and cache, without downloading dependencies"}},
			{"--dry-run", text{"mostra o que seria criado", "shows what would be created"}}, optJSON},
		examples: []string{"romanctl new", "romanctl new loja", "romanctl new express api", "romanctl new loja --stack dotnet --arch clean,ddd,cqrs --with postgres"}},
	{path: "catalog", aliases: []string{"catalogo"}, usage: "romanctl catalog [--json]",
		summary: text{"Lista templates, combinações válidas e geradores", "Lists templates, valid combinations and generators"},
		options: []helpOption{optJSON}},
	{path: "types", aliases: []string{"tipos"}, usage: "romanctl types [--filter <texto>]",
		summary: text{"Lista os tipos de projeto simples", "Lists the simple project types"}, options: []helpOption{optFilter}},
	{path: "list", aliases: []string{"ls", "projetos"}, usage: "romanctl list [--filter <texto>] [--plain] [--json]",
		summary: text{"Lista os projetos; no terminal, com filtro e ações em cada um", "Lists the projects; in a terminal, with a filter and actions on each one"},
		options: []helpOption{optFilter, optPlain, optJSON}, examples: []string{"romanctl list", "romanctl list --filter dotnet"}},
	{path: "up", aliases: []string{"subir"}, usage: "romanctl up [projeto] [-d] [--no-run] [--open]",
		summary: text{"Sobe o ambiente do projeto: serviços, banco, .env e o app", "Brings up the project's environment: services, database, .env and the app"},
		details: text{"O .env só ganha o que falta: o que você mudou continua como está.", "The .env only gets what is missing: what you changed stays as it is."},
		options: []helpOption{
			{"-d, --detach", text{"roda em segundo plano e espera o app responder", "runs in the background and waits for the app to respond"}},
			{"--no-run", text{"prepara o ambiente, sem rodar o app", "prepares the environment, without running the app"}},
			{"--open", text{"abre o navegador quando o app subir", "opens the browser when the app is up"}}, optJSON},
		examples: []string{"romanctl up", "romanctl up loja -d", "romanctl up --no-run"}},
	{path: "down", aliases: []string{"derrubar"}, usage: "romanctl down [projeto] [--all] [--keep-services]",
		summary: text{"Para o app e os serviços que só ele estava usando", "Stops the app and the services only it was using"},
		details: text{"Um serviço que já estava ligado antes do up, ou que outro projeto usa, continua ligado.", "A service that was already on before up, or that another project uses, stays on."},
		options: []helpOption{
			{"--all", text{"derruba todos os projetos no ar", "brings down all running projects"}},
			{"--keep-services", text{"para só o app", "stops only the app"}}, optJSON},
		examples: []string{"romanctl down", "romanctl down loja", "romanctl down --all"}},
	{path: "run", aliases: []string{"dev", "rodar"}, usage: "romanctl run [projeto] [--open]",
		summary:  text{"Roda um projeto; sem argumento, abre uma lista", "Runs a project; without an argument, opens a list"},
		options:  []helpOption{{"--open", text{"abre o navegador quando o app subir", "opens the browser when the app is up"}}},
		examples: []string{"romanctl run", "romanctl run loja --open"}},
	{path: "rm", aliases: []string{"remove", "remover", "delete"}, usage: "romanctl rm [projeto] [-y]",
		summary:  text{"Remove um projeto e todos os arquivos dele", "Removes a project and all its files"},
		options:  []helpOption{{"-y, --yes", text{"não pede confirmação", "does not ask for confirmation"}}},
		examples: []string{"romanctl rm", "romanctl rm loja"}},
	{path: "code", aliases: []string{"editar", "edit"}, usage: "romanctl code [projeto|pasta]",
		summary:  text{"Abre no code-server; sem alvo, usa o projeto da pasta atual ou mostra uma lista", "Opens in code-server; without a target, uses the current folder's project or shows a list"},
		examples: []string{"romanctl code", "romanctl code loja", "romanctl code ."}},
	{path: "open", aliases: []string{"abrir"}, usage: "romanctl open [code|projeto|url] [--app]",
		summary:  text{"Abre no navegador; sem alvo, abre uma lista", "Opens in the browser; without a target, opens a list"},
		options:  []helpOption{{"--app", text{"abre o app rodando, em vez do code-server", "opens the running app instead of code-server"}}},
		examples: []string{"romanctl open", "romanctl open loja --app", "romanctl open https://example.com"}},
	{path: "add", aliases: []string{"adicionar", "generate", "g"}, usage: "romanctl add [tipo] [nome] [--in <agregado>]",
		summary: text{"Gera código conforme o projeto, perguntando o que faltar", "Generates code according to the project, asking for whatever is missing"},
		details: text{"Os tipos dependem do projeto: veja com romanctl add --list.", "The types depend on the project: see with romanctl add --list."},
		options: []helpOption{
			{"--in <agregado>", text{"agregado de destino, em command, query, event e value-object", "target aggregate, for command, query, event and value-object"}},
			{"--plural <nome>", text{"plural do nome, quando a regra do inglês erra, como --plural People", "plural of the name, when the English rule gets it wrong, such as --plural People"}},
			{"--project <nome>", text{"projeto de destino; sem isto, usa a pasta atual", "target project; without this, uses the current folder"}},
			{"--list", text{"lista o que dá para gerar", "lists what can be generated"}},
			{"--force", text{"sobrescreve arquivos existentes", "overwrites existing files"}},
			{"--dry-run", text{"mostra o que seria criado", "shows what would be created"}}, optJSON},
		examples: []string{"romanctl add", "romanctl add resource Product", "romanctl add command ArchiveProduct --in Product", "romanctl add page sobre --project site"}},
	{path: "publish", aliases: []string{"publicar"}, usage: "romanctl publish [projeto] [--public] [--description <texto>]",
		summary: text{"Gera CI e Dockerfile, faz o commit e publica no GitHub", "Generates CI and a Dockerfile, commits and publishes to GitHub"},
		details: text{"O repositório é privado, salvo se você escolher outra coisa. Arquivos como .env cancelam a publicação.", "The repository is private unless you choose otherwise. Files such as .env cancel the publishing."},
		options: []helpOption{
			{"--public", text{"cria um repositório público", "creates a public repository"}},
			{"--description <texto>", text{"descrição do repositório", "repository description"}},
			{"--no-ci", text{"não gera CI nem Dockerfile", "does not generate CI or a Dockerfile"}},
			{"-y, --yes", text{"não pergunta nada", "asks nothing"}},
			{"--dry-run", text{"mostra o que faria", "shows what it would do"}}, optJSON},
		examples: []string{"romanctl publish", "romanctl publish loja --description \"Loja\"", "romanctl publish --dry-run"}},

	{path: "tools", aliases: []string{"ferramentas"}, usage: "romanctl tools [list|info|versions|install|uninstall]",
		summary: text{"Instala e remove ferramentas de desenvolvimento", "Installs and removes development tools"},
		details: text{"Em um terminal, sem subcomando, abre um menu. Sem terminal, lista.", "In a terminal, without a subcommand, it opens a menu. Without a terminal, it lists."}},
	{path: "tools list", aliases: []string{"ls"}, usage: "romanctl tools list [--category <nomes>] [--all] [--filter <texto>] [--plain] [--json]",
		summary: text{"Lista o que roda nesta máquina; no terminal, com filtro, categorias e ações", "Lists what runs on this machine; in a terminal, with a filter, categories and actions"},
		details: text{"Na lista, digite para filtrar e use ←/→ para trocar de categoria. Digitar o nome da categoria também funciona.", "In the list, type to filter and use ←/→ to change category. Typing the category name also works."},
		options: []helpOption{
			{"--category <nomes>", text{"só estas categorias, separadas por vírgula", "only these categories, separated by commas"}},
			{"--all", text{"inclui o que não roda aqui, com o motivo", "includes what does not run here, with the reason"}}, optFilter, optPlain, optJSON},
		examples: []string{"romanctl tools list", "romanctl tools list --category \"banco de dados\"", "romanctl tools list --category db,sdk --plain", "romanctl tools list --filter kube"}},
	{path: "tools info", aliases: []string{"show", "readme"}, usage: "romanctl tools info <nome> [--no-readme] [--json]",
		summary:  text{"Versão instalada, versão disponível, origem e README de uma ferramenta", "Installed version, available version, source and README of a tool"},
		details:  text{"A versão disponível vem da mesma origem que a instalação usa: GitHub, gerenciador de pacotes, Homebrew, npm, PyPI ou o instalador oficial.", "The available version comes from the same source the installation uses: GitHub, package manager, Homebrew, npm, PyPI or the official installer."},
		options:  []helpOption{{"--no-readme", text{"não baixa nem mostra o README", "does not download or show the README"}}, optJSON},
		examples: []string{"romanctl tools info k9s", "romanctl tools info terraform --no-readme"}},
	{path: "tools versions", aliases: []string{"versoes", "versões"}, usage: "romanctl tools versions <nome> [--json]",
		summary:  text{"Lista as versões que dá para instalar, da mais nova para a mais antiga", "Lists the versions that can be installed, newest first"},
		details:  text{"Vêm da mesma origem do instalador: releases do GitHub, HashiCorp, npm, PyPI, apt ou dnf. Com o Homebrew e alguns instaladores próprios, só a última.", "They come from the installer's own source: GitHub releases, HashiCorp, npm, PyPI, apt or dnf. With Homebrew and some custom installers, only the latest."},
		options:  []helpOption{optJSON},
		examples: []string{"romanctl tools versions helm", "romanctl tools install helm --version 3.18.4"}},
	{path: "tools categories", aliases: []string{"categorias"}, usage: "romanctl tools categories [--json]",
		summary: text{"Mostra as categorias e quantas ferramentas de cada uma rodam aqui", "Shows the categories and how many tools of each run here"}, options: []helpOption{optJSON}},
	{path: "tools install", aliases: []string{"instalar", "add"}, usage: "romanctl tools install [nomes|grupos|all]",
		summary: text{"Instala ferramentas; um grupo abre uma lista com a opção Todos", "Installs tools; a group opens a list with an All option"},
		options: []helpOption{
			{"-y, --yes", text{"instala o grupo inteiro, sem lista", "installs the whole group, without a list"}},
			{"--force", text{"reinstala", "reinstalls"}},
			{"--version <versão>", text{"instala esta versão, mesmo mais antiga; uma ferramenta por vez", "installs this version, even if older; one tool at a time"}},
			{"--dry-run", text{"mostra o que seria instalado", "shows what would be installed"}}, optJSON},
		examples: []string{"romanctl tools install", "romanctl tools install k8s", "romanctl tools install kubectl gh", "romanctl tools install sdk --yes", "romanctl tools install terraform --version 1.9.8"}},
	{path: "tools uninstall", aliases: []string{"remove", "rm", "remover", "desinstalar"}, usage: "romanctl tools uninstall [nomes|grupos]",
		summary: text{"Remove ferramentas, com lista e confirmação", "Removes tools, with a list and confirmation"},
		options: []helpOption{
			{"-y, --yes", text{"não mostra lista nem pede confirmação; recusa all", "shows no list and asks no confirmation; refuses all"}},
			{"--dry-run", text{"mostra o que seria removido", "shows what would be removed"}}, optJSON},
		examples: []string{"romanctl tools uninstall", "romanctl tools uninstall k8s", "romanctl tools uninstall jq --yes"}},

	{path: "system", aliases: []string{"sistema", "sys"}, usage: "romanctl system [status|battery|watch]",
		summary: text{"Mostra CPU, memória, disco e bateria", "Shows CPU, memory, disk and battery"},
		details: text{"Em um terminal, sem subcomando, abre um menu. Sem terminal, mostra o status.", "In a terminal, without a subcommand, it opens a menu. Without a terminal, it shows the status."}},
	{path: "system status", usage: "romanctl system status [--json]",
		summary: text{"CPU, memória, disco, bateria e processos que mais consomem", "CPU, memory, disk, battery and top processes"}, options: []helpOption{optJSON}},
	{path: "system battery", aliases: []string{"bateria"}, usage: "romanctl system battery [--json]",
		summary: text{"Nível, consumo em watts, temperatura e previsão de autonomia", "Level, power in watts, temperature and time left"}, options: []helpOption{optJSON}},
	{path: "system watch", aliases: []string{"medir"}, usage: "romanctl system watch [--seconds N] [--interval N]",
		summary: text{"Mede consumo e CPU juntos e estima o gasto dos seus processos", "Measures power and CPU together and estimates what your processes use"},
		options: []helpOption{
			{"--seconds N", text{"duração; sem isto, pergunta em um terminal", "duration; without this, asks in a terminal"}},
			{"--interval N", text{"segundos entre as medições; o padrão é 5", "seconds between measurements; the default is 5"}}, optJSON},
		examples: []string{"romanctl system watch", "romanctl system watch --seconds 600"}},
	{path: "platform", aliases: []string{"plataforma"}, usage: "romanctl platform [--json]",
		summary: text{"Mostra sistema, arquitetura e capacidades detectados", "Shows the detected system, architecture and capabilities"}, options: []helpOption{optJSON}},
	{path: "doctor", aliases: []string{"info"}, usage: "romanctl doctor",
		summary: text{"Mostra as versões das ferramentas principais", "Shows the versions of the main tools"}},

	{path: "offline", usage: "romanctl offline [status|prep|test]",
		summary: text{"Prepara a máquina para trabalhar sem internet", "Prepares the machine to work without internet"},
		details: text{"Em um terminal, sem subcomando, abre um menu. Sem terminal, mostra o status.", "In a terminal, without a subcommand, it opens a menu. Without a terminal, it shows the status."}},
	{path: "offline status", usage: "romanctl offline status",
		summary: text{"Mostra o estado da preparação", "Shows the preparation status"}},
	{path: "offline prep", aliases: []string{"preparar"}, usage: "romanctl offline prep [--force]",
		summary: text{"Baixa templates, caches e extensões; precisa de internet", "Downloads templates, caches and extensions; needs internet"},
		options: []helpOption{{"--force", text{"refaz os templates", "redoes the templates"}}}},
	{path: "offline test", aliases: []string{"testar"}, usage: "romanctl offline test",
		summary: text{"Cria e compila cada tipo de projeto com a rede bloqueada", "Creates and builds each project type with the network blocked"}},

	{path: "login", aliases: []string{"entrar"}, usage: "romanctl login [--issuer URL] [--api URL]",
		summary: text{"Entra no romanctl pelo Keycloak; sem login a CLI não roda", "Signs in to romanctl through Keycloak; without a login the CLI does not run"},
		details: text{"Abre o navegador no realm romanctl de auth.marceloroman.com.br (authorization code com PKCE, redirect em 127.0.0.1) e guarda a sessão em ~/.config/romanctl/session.json, só para você. A sessão é offline: vale 30 dias sem uso e renova a cada comando, também sem internet. Em pipelines, ROMANCTL_CLIENT_ID e ROMANCTL_CLIENT_SECRET (o client romanctl-ci) substituem o login. Sem login rodam só login, logout, whoami, help, version, self-update e templates.", "Opens the browser on the romanctl realm at auth.marceloroman.com.br (authorization code with PKCE, redirect on 127.0.0.1) and keeps the session in ~/.config/romanctl/session.json, readable only by you. The session is offline: it lasts 30 days without use and renews on every command, even without internet. In pipelines, ROMANCTL_CLIENT_ID and ROMANCTL_CLIENT_SECRET (the romanctl-ci client) replace the login. Without a login only login, logout, whoami, help, version, self-update and templates run."},
		options: []helpOption{
			{"--issuer <url>", text{"outro realm; também ROMANCTL_AUTH_ISSUER", "another realm; also ROMANCTL_AUTH_ISSUER"}},
			{"--api <url>", text{"outra API do IAM Manager; também ROMANCTL_IAM_API", "another IAM Manager API; also ROMANCTL_IAM_API"}}, optJSON},
		examples: []string{"romanctl login", "romanctl login --issuer http://localhost:18081/realms/romanctl --api http://localhost:18091"}},
	{path: "logout", aliases: []string{"sair"}, usage: "romanctl logout",
		summary: text{"Encerra a sessão no Keycloak e apaga a sessão local", "Ends the session in Keycloak and deletes the local session"}, options: []helpOption{optJSON}},
	{path: "whoami", usage: "romanctl whoami [--json]",
		summary: text{"Mostra quem está logado, os papéis e até quando a sessão vale", "Shows who is logged in, the roles and until when the session lasts"}, options: []helpOption{optJSON}},
	{path: "iam", aliases: []string{"keycloak"}, usage: "romanctl iam <comando> [--json]",
		summary:  text{"Administra o Keycloak pela API do IAM Manager, com o seu login", "Manages Keycloak through the IAM Manager API, with your login"},
		details:  text{"O token do login vai para iam-api.marceloroman.com.br, que confere as permissões e registra tudo na auditoria. O que você pode fazer vem dos papéis do realm romanctl: iam-admin (tudo), iam-user-admin (usuários e API keys), iam-theme-admin (temas) e iam-auditor (leitura), ou delegações por realm (romanctl iam operator roles). Usuários aceitam o nome ou o id.", "The login token goes to iam-api.marceloroman.com.br, which checks permissions and records everything in the audit log. What you can do comes from the roles of the romanctl realm: iam-admin (everything), iam-user-admin (users and API keys), iam-theme-admin (themes) and iam-auditor (read-only), or per-realm delegations (romanctl iam operator roles). Users accept the username or the id."},
		options:  []helpOption{optJSON},
		examples: []string{"romanctl iam me", "romanctl iam realms", "romanctl iam users acme --search ana", "romanctl iam user create acme ana --email ana@acme.com", "romanctl iam api GET /api/themes"}},
	{path: "iam me", usage: "romanctl iam me",
		summary: text{"Quem você é na API e o que pode fazer", "Who you are in the API and what you can do"}},
	{path: "iam realms", usage: "romanctl iam realms",
		summary: text{"Lista os realms que você pode ver", "Lists the realms you can see"}},
	{path: "iam realm", usage: "romanctl iam realm create|enable|disable|delete <nome> [--display-name X] [--yes]",
		summary:  text{"Cria, liga, desliga ou exclui um realm", "Creates, enables, disables or deletes a realm"},
		examples: []string{"romanctl iam realm create acme --display-name ACME", "romanctl iam realm delete acme --yes"}},
	{path: "iam users", usage: "romanctl iam users <realm> [--search X]",
		summary: text{"Lista usuários de um realm", "Lists the users of a realm"}},
	{path: "iam user", usage: "romanctl iam user show|create|enable|disable|roles|reset|logout|delete <realm> <usuário>",
		summary:  text{"Cria e administra um usuário", "Creates and manages a user"},
		details:  text{"create aceita --email, --first-name, --last-name, --roles a,b e --credential: link (padrão, link de uso único para definir a senha), temporary, email ou none. reset usa o mesmo --credential. roles define o estado completo dos papéis do realm.", "create accepts --email, --first-name, --last-name, --roles a,b and --credential: link (default, a one-time link to set the password), temporary, email or none. reset uses the same --credential. roles sets the complete state of the realm roles."},
		examples: []string{"romanctl iam user create acme ana --email ana@acme.com --roles billing", "romanctl iam user disable acme ana", "romanctl iam user reset acme ana --credential temporary"}},
	{path: "iam api-keys", usage: "romanctl iam api-keys <realm>",
		summary: text{"Lista as API keys (client_credentials) de um realm", "Lists the API keys (client_credentials) of a realm"}},
	{path: "iam api-key", usage: "romanctl iam api-key create <realm> <nome> [--description X] [--roles a,b] | rotate|delete <realm> <id>",
		summary: text{"Cria, rotaciona ou exclui uma API key; o secret só aparece na hora", "Creates, rotates or deletes an API key; the secret is shown only once"}},
	{path: "iam operators", usage: "romanctl iam operators [--realm romanctl]",
		summary: text{"Quem administra o IAM Manager e com quais papéis", "Who manages IAM Manager and with which roles"},
		details: text{"Sem --realm, os operadores do realm iam-manager (o painel); com --realm romanctl, as pessoas que usam o romanctl.", "Without --realm, the operators of the iam-manager realm (the panel); with --realm romanctl, the people who use romanctl."}},
	{path: "iam operator", usage: "romanctl iam operator roles <usuário> [--realm romanctl] [--global iam-admin,...] [--realms acme=admin,beta=viewer]",
		summary:  text{"Define os papéis globais e as delegações por realm de alguém (só iam-admin)", "Sets someone's global roles and per-realm delegations (iam-admin only)"},
		examples: []string{"romanctl iam operator roles ana --realm romanctl --realms acme=admin"}},
	{path: "iam events", usage: "romanctl iam events <realm> [--limit N]",
		summary: text{"Logins e alterações administrativas guardados pelo Keycloak", "Logins and admin changes stored by Keycloak"}},
	{path: "iam audit", usage: "romanctl iam audit [--limit N]",
		summary: text{"O que foi feito pelo IAM Manager, por quem", "What was done through IAM Manager, and by whom"}},
	{path: "iam api", usage: "romanctl iam api <GET|POST|PUT|DELETE> <caminho> [--data JSON]",
		summary:  text{"Chama qualquer rota da API com o seu token", "Calls any API route with your token"},
		details:  text{"A especificação está em /api/openapi.json e o guia em /api/guide.md.", "The specification is at /api/openapi.json and the guide at /api/guide.md."},
		examples: []string{"romanctl iam api GET /api/realms/acme/settings", `romanctl iam api PUT /api/realms/acme/events/config --data '{"eventsEnabled":true}'`}},
	{path: "web-ui", aliases: []string{"webui", "web"}, usage: "romanctl web-ui [--port N] [--open]",
		summary: text{"Abre a interface web do romanctl no navegador, em localhost", "Opens the romanctl web interface in the browser, on localhost"},
		details: text{"Serve só em 127.0.0.1 e exige um token que vai na página, então outros sites não chamam a API. Projetos, serviços, ferramentas e templates, com as ações rodando o próprio romanctl.", "Listens only on 127.0.0.1 and requires a token embedded in the page, so other sites cannot call the API. Projects, services, tools and templates, with actions that run romanctl itself."},
		options: []helpOption{
			{"--port N", text{"porta; o padrão é 7777", "port; the default is 7777"}},
			{"--open", text{"abre o navegador", "opens the browser"}}},
		examples: []string{"romanctl web-ui", "romanctl web-ui --port 8888 --open"}},
	{path: "templates", aliases: []string{"template"}, usage: "romanctl templates [status|update] [--ref <tag ou branch>]",
		summary:  text{"Mostra e atualiza os templates de projeto e os geradores", "Shows and updates the project templates and generators"},
		details:  text{"Os templates ficam fora do binário, em ~/.local/share/romanctl/templates, e vêm da pasta templates do repositório, na tag da versão instalada. ROMANCTL_TEMPLATES aponta para outra pasta.", "Templates live outside the binary, in ~/.local/share/romanctl/templates, and come from the templates folder of the repository, at the tag of the installed version. ROMANCTL_TEMPLATES points to another folder."},
		options:  []helpOption{{"--ref <ref>", text{"baixa de uma tag ou branch, como refs/heads/main", "downloads from a tag or branch, such as refs/heads/main"}}, optJSON},
		examples: []string{"romanctl templates", "romanctl templates update", "romanctl templates update --ref refs/heads/main"}},
	{path: "templates status", usage: "romanctl templates status",
		summary: text{"Mostra onde estão os templates e de que versão vieram", "Shows where the templates are and which version they came from"}, options: []helpOption{optJSON}},
	{path: "templates update", aliases: []string{"atualizar"}, usage: "romanctl templates update [--ref <ref>]",
		summary: text{"Baixa os templates da versão instalada", "Downloads the templates of the installed version"}, options: []helpOption{optJSON}},
	{path: "claude", aliases: []string{"skills"}, usage: "romanctl claude [status|install|uninstall]",
		summary:  text{"Instala as skills que ensinam o Claude Code a usar o romanctl", "Installs the skills that teach Claude Code to use romanctl"},
		details:  text{"As skills vão para ~/.claude/skills (ou $CLAUDE_CONFIG_DIR/skills). O romanctl só troca ou apaga as pastas que ele mesmo criou.", "The skills go to ~/.claude/skills (or $CLAUDE_CONFIG_DIR/skills). romanctl only replaces or deletes the folders it created itself."},
		examples: []string{"romanctl claude install", "romanctl claude status"}},
	{path: "claude status", usage: "romanctl claude status [--json]",
		summary: text{"Mostra cada skill: instalada, desatualizada, ausente ou alheia", "Shows each skill: installed, outdated, missing or foreign"}, options: []helpOption{optJSON}},
	{path: "claude install", aliases: []string{"instalar", "update", "atualizar"}, usage: "romanctl claude install [--force] [--dry-run] [--json]",
		summary: text{"Instala ou atualiza as skills", "Installs or updates the skills"},
		options: []helpOption{
			{"--force", text{"troca também as pastas criadas fora do romanctl", "also replaces folders created outside romanctl"}},
			{"--dry-run", text{"mostra o que faria", "shows what it would do"}}, optJSON}},
	{path: "claude uninstall", aliases: []string{"remove", "rm", "remover"}, usage: "romanctl claude uninstall [--dry-run] [--json]",
		summary: text{"Remove as skills que o romanctl instalou", "Removes the skills romanctl installed"},
		options: []helpOption{{"--dry-run", text{"mostra o que faria", "shows what it would do"}}, optJSON}},
	{path: "ui", aliases: []string{"painel", "tui"}, usage: "romanctl ui",
		summary: text{"Abre o painel de tela cheia, nos moldes do k9s", "Opens the full-screen panel, k9s style"},
		details: text{"Dentro dele: Tab troca de tela, / filtra, : vai direto, Enter mostra as ações, ? mostra a ajuda.", "Inside it: Tab changes screens, / filters, : jumps, Enter shows the actions, ? shows the help."}},
	{path: "menu", usage: "romanctl menu",
		summary: text{"Abre o menu em lista, mais simples que o painel", "Opens the list menu, simpler than the panel"}},
	{path: "config", usage: "romanctl config [lang [pt|en|auto]]",
		summary: text{"Mostra ou grava as configurações", "Shows or saves the settings"}},
	{path: "config lang", usage: "romanctl config lang [pt|en|auto]",
		summary:  text{"Escolhe o idioma; auto volta a seguir o sistema", "Chooses the language; auto follows the system again"},
		examples: []string{"romanctl config lang", "romanctl config lang en"}},
	{path: "help", aliases: []string{"ajuda"}, usage: "romanctl help [comando]",
		summary: text{"Mostra esta ajuda", "Shows this help"}, examples: []string{"romanctl help tools install"}},
	{path: "self-update", aliases: []string{"update", "atualizar"}, usage: "romanctl self-update [--check] [--version <vX.Y.Z>] [--yes]",
		summary: text{"Atualiza o romanctl para a última versão publicada", "Updates romanctl to the latest published version"},
		details: text{"Confere o SHA-256 e roda a versão nova antes de trocar; se ela não funcionar, a atual fica.", "Checks the SHA-256 and runs the new version before swapping; if it does not work, the current one stays."},
		options: []helpOption{
			{"--check", text{"só mostra se há versão nova", "only shows whether there is a new version"}},
			{"--version <vX.Y.Z>", text{"instala esta versão, mesmo que mais antiga", "installs this version, even if older"}},
			{"-y, --yes", text{"não pede confirmação", "does not ask for confirmation"}}, optJSON},
		examples: []string{"romanctl self-update", "romanctl self-update --check", "romanctl self-update --version v0.1.0"}},
	{path: "version", aliases: []string{"versao"}, usage: "romanctl version",
		summary: text{"Mostra a versão", "Shows the version"}},
}

// placeholders traduz os nomes de argumento em uso e opções, na ajuda em inglês.
var placeholders = strings.NewReplacer("usuário", "user", "caminho", "path", "serviços", "services", "serviço", "service", "nomes", "names", "nome", "name",
	"grupos", "groups", "tipo", "type", "projeto", "project", "pasta", "folder", "versão", "version", "agregado", "aggregate",
	"texto", "text", "comando", "command", "alvo", "target")

func localized(s string) string {
	if i18n.Lang() == i18n.EN {
		return placeholders.Replace(s)
	}
	return s
}

func findCommand(words []string) *command {
	if len(words) == 0 {
		return nil
	}
	var top *command
	for i := range commands {
		c := &commands[i]
		if !strings.Contains(c.path, " ") && matches(c, c.path, words[0]) {
			top = c
			break
		}
	}
	if top == nil {
		return nil
	}
	if len(words) > 1 {
		prefix := top.path + " "
		for i := range commands {
			c := &commands[i]
			if name, ok := strings.CutPrefix(c.path, prefix); ok && matches(c, name, words[1]) {
				return c
			}
		}
	}
	return top
}

func matches(c *command, name, word string) bool {
	if name == word {
		return true
	}
	for _, a := range c.aliases {
		if a == word {
			return true
		}
	}
	return false
}

var labels = struct{ usage, options, examples, subcommands, aliases, unknown text }{
	text{"Uso:", "Usage:"}, text{"Opções:", "Options:"}, text{"Exemplos:", "Examples:"},
	text{"Subcomandos:", "Subcommands:"}, text{"Apelidos:", "Aliases:"},
	text{"Comando desconhecido; a lista completa vem a seguir.", "Unknown command; the full list follows."},
}

// printHelp mostra a ajuda geral ou a de um comando. As palavras que não são comandos, como
// um nome de serviço em "start postgres --help", não atrapalham.
func printHelp(words []string) {
	if len(words) == 0 {
		usage()
		return
	}
	c := findCommand(words)
	if c == nil {
		ui.Warnf("%s", labels.unknown.String())
		usage()
		return
	}
	fmt.Printf("%s %s\n\n", ui.STitle.Render("romanctl "+c.path), "— "+c.summary.String())
	if c.details.pt != "" {
		fmt.Printf("%s\n\n", c.details)
	}
	fmt.Printf("%s %s\n", labels.usage, localized(c.usage))
	if len(c.aliases) > 0 {
		fmt.Printf("%s %s\n", labels.aliases, strings.Join(c.aliases, ", "))
	}
	var subs []command
	for _, other := range commands {
		if strings.HasPrefix(other.path, c.path+" ") {
			subs = append(subs, other)
		}
	}
	if len(subs) > 0 {
		fmt.Printf("\n%s\n", labels.subcommands)
		for _, s := range subs {
			fmt.Printf("  %-12s %s\n", strings.TrimPrefix(s.path, c.path+" "), s.summary)
		}
	}
	if len(c.options) > 0 {
		fmt.Printf("\n%s\n", labels.options)
		for _, o := range c.options {
			fmt.Printf("  %-24s %s\n", localized(o.flag), o.desc)
		}
	}
	if len(c.examples) > 0 {
		fmt.Printf("\n%s\n", labels.examples)
		for _, e := range c.examples {
			fmt.Println("  " + e)
		}
	}
}

const helpPT = `
Uso: romanctl [comando] [argumentos] [--lang pt|en]
     romanctl              abre a lista de comandos: digite para filtrar, Enter roda
     romanctl ui           abre o painel de tela cheia
     romanctl <comando> --help   mostra a ajuda de qualquer comando ou subcomando

Em um terminal, o que faltar é perguntado, e nas listas basta digitar para filtrar.
Sem terminal ou com --json, nada é perguntado.
`

const helpEN = `
Usage: romanctl [command] [arguments] [--lang pt|en]
       romanctl            opens the command list: type to filter, Enter runs
       romanctl ui         opens the full-screen panel
       romanctl <command> --help   shows the help of any command or subcommand

In a terminal, whatever is missing is asked for, and in lists you just type to filter.
Without a terminal or with --json, nothing is asked.
`

var usageGroups = []struct {
	title text
	paths []string
}{
	{text{"Ferramentas", "Tools"}, []string{"tools"}},
	{text{"Projetos", "Projects"}, []string{"new", "add", "up", "down", "run", "publish", "list", "code", "open", "rm", "catalog", "types", "templates"}},
	{text{"Serviços", "Services"}, []string{"status", "start", "stop", "restart", "restore", "logs", "install"}},
	{text{"Sistema", "System"}, []string{"system", "platform", "doctor", "offline"}},
	{text{"Acesso", "Access"}, []string{"login", "logout", "whoami", "iam"}},
	{text{"Outros", "Other"}, []string{"web-ui", "claude", "ui", "menu", "config", "self-update", "help", "version"}},
}

func init() { i18n.Add(map[string]string{helpPT: helpEN}) }

func usage() {
	fmt.Println(ui.STitle.Render("romanctl") + " " + ui.SGray.Render("v"+version) + i18n.T(" — instala ferramentas, cria projetos e traz utilitários de desenvolvimento"))
	i18n.Printf(helpPT)
	for _, g := range usageGroups {
		fmt.Printf("\n%s\n", ui.STitle.Render(g.title.String()))
		for _, p := range g.paths {
			if c := findCommand([]string{p}); c != nil {
				fmt.Printf("  %-12s %s\n", c.path, c.summary)
			}
		}
	}
	fmt.Printf("\n%s %s\n", text{"Projetos ficam em", "Projects live in"}, config.ProjDir)
}
