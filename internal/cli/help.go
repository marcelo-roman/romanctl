package cli

import (
	"fmt"
	"regexp"
	"strings"

	"romanctl/internal/config"
	"romanctl/internal/i18n"
	"romanctl/internal/ui"
)

// text keeps both languages together: each help entry is read at once, with no key to look up.
type text struct{ en, pt string }

func (t text) String() string {
	if i18n.Lang() == i18n.PT {
		return t.pt
	}
	return t.en
}

type helpOption struct {
	flag string
	desc text
}

// command is the help of a command; the path has one or two words.
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
	optJSON   = helpOption{"--json", text{"JSON output, without menus", "saída em JSON, sem menus"}}
	optFilter = helpOption{"--filter <text>", text{"only what contains the text, without an interactive list", "só o que contém o texto, sem lista interativa"}}
	optPlain  = helpOption{"--plain", text{"prints everything, without an interactive list", "imprime tudo, sem lista interativa"}}
)

var commands = []command{
	{path: "status", aliases: []string{"st"}, usage: "romanctl status [--all] [--json]",
		summary:  text{"Shows the installed services and Docker", "Mostra os serviços instalados e o Docker"},
		details:  text{"A service only counts as installed when its container exists; having Docker or the image is not enough.", "Um serviço só conta como instalado quando o container dele existe; ter o Docker ou a imagem não basta."},
		options:  []helpOption{{"--all", text{"includes what can be installed", "inclui o que dá para instalar"}}, optJSON},
		examples: []string{"romanctl status", "romanctl status --all"}},
	{path: "start", aliases: []string{"iniciar"}, usage: "romanctl start [service|all]",
		summary:  text{"Starts services; without an argument, opens a list", "Inicia serviços; sem argumento, abre uma lista"},
		examples: []string{"romanctl start", "romanctl start postgres", "romanctl start all"}},
	{path: "stop", aliases: []string{"parar"}, usage: "romanctl stop [service|all] [--yes]",
		details:  text{"Stopping docker asks for confirmation, because it brings down every container on the machine; all never includes docker.", "Parar o docker pede confirmação, porque derruba todos os containers da máquina; all nunca inclui o docker."},
		summary:  text{"Stops services; without an argument, opens a list", "Para serviços; sem argumento, abre uma lista"},
		examples: []string{"romanctl stop", "romanctl stop all"}},
	{path: "restart", aliases: []string{"reiniciar"}, usage: "romanctl restart [service|all]",
		summary:  text{"Restarts services; without an argument, opens a list", "Reinicia serviços; sem argumento, abre uma lista"},
		examples: []string{"romanctl restart postgres"}},
	{path: "install", aliases: []string{"instalar"}, usage: "romanctl install [service]",
		summary:  text{"Prepares a service: downloads the image if missing and creates the stopped container", "Prepara um serviço: baixa a imagem, se faltar, e cria o container parado"},
		details:  text{"To install tools such as kubectl, use: romanctl tools install", "Para instalar ferramentas, como kubectl, use: romanctl tools install"},
		examples: []string{"romanctl install kafka"}},
	{path: "logs", aliases: []string{"log"}, usage: "romanctl logs [service]",
		summary:  text{"Follows a service's log; Ctrl+C exits", "Acompanha o log de um serviço; Ctrl+C sai"},
		examples: []string{"romanctl logs", "romanctl logs postgres"}},
	{path: "restore", aliases: []string{"restaurar"}, usage: "romanctl restore [status|save|clear]",
		summary:  text{"Starts again what was running before a crash", "Sobe de novo o que estava ligado antes de uma queda"},
		details:  text{"In a terminal, without a subcommand, it opens a menu. Without a terminal, it restores directly.", "Em um terminal, sem subcomando, abre um menu. Sem terminal, restaura direto."},
		examples: []string{"romanctl restore", "romanctl restore status"}},
	{path: "restore status", usage: "romanctl restore status",
		summary: text{"Shows what is saved and what crashed", "Mostra o que está salvo e o que caiu"}},
	{path: "restore save", aliases: []string{"salvar"}, usage: "romanctl restore save",
		summary: text{"Saves what is running now as the state", "Salva como estado o que está rodando agora"}},
	{path: "restore clear", aliases: []string{"limpar"}, usage: "romanctl restore clear",
		summary: text{"Clears the state: restore starts nothing anymore", "Limpa o estado: o restore não sobe mais nada"}},

	{path: "new", aliases: []string{"novo", "create"}, usage: "romanctl new [type] [name]\n     romanctl new <name> --stack <stack> --arch <a,b> [--with <services>]",
		summary: text{"Creates a project, asking for whatever is missing", "Cria um projeto, perguntando o que faltar"},
		options: []helpOption{
			{"--stack <stack>", text{"template language, such as dotnet", "linguagem do template, como dotnet"}},
			{"--arch <a,b>", text{"architectures, such as clean, ddd, cqrs and hexagonal; romanctl catalog lists them all", "arquiteturas, como clean, ddd, cqrs e hexagonal; romanctl catalog lista todas"}},
			{"--with <services>", text{"project services, such as postgres", "serviços do projeto, como postgres"}},
			{"--no-git", text{"does not create a git repository", "não cria repositório git"}},
			{"--offline", text{"uses only templates and cache, without downloading dependencies", "usa só templates e cache, sem baixar dependências"}},
			{"--dry-run", text{"shows what would be created", "mostra o que seria criado"}}, optJSON},
		examples: []string{"romanctl new", "romanctl new shop", "romanctl new express api", "romanctl new shop --stack dotnet --arch clean,ddd,cqrs --with postgres"}},
	{path: "catalog", aliases: []string{"catalogo"}, usage: "romanctl catalog [--json]",
		summary: text{"Lists templates, valid combinations and generators", "Lista templates, combinações válidas e geradores"},
		options: []helpOption{optJSON}},
	{path: "types", aliases: []string{"tipos"}, usage: "romanctl types [--filter <text>]",
		summary: text{"Lists the simple project types", "Lista os tipos de projeto simples"}, options: []helpOption{optFilter}},
	{path: "list", aliases: []string{"ls", "projetos"}, usage: "romanctl list [--filter <text>] [--plain] [--json]",
		summary: text{"Lists the projects; in a terminal, with a filter and actions on each one", "Lista os projetos; no terminal, com filtro e ações em cada um"},
		options: []helpOption{optFilter, optPlain, optJSON}, examples: []string{"romanctl list", "romanctl list --filter dotnet"}},
	{path: "up", aliases: []string{"subir"}, usage: "romanctl up [project] [-d] [--no-run] [--open]",
		summary: text{"Brings up the project's environment: services, database, .env and the app", "Sobe o ambiente do projeto: serviços, banco, .env e o app"},
		details: text{"The .env only gets what is missing: what you changed stays as it is.", "O .env só ganha o que falta: o que você mudou continua como está."},
		options: []helpOption{
			{"-d, --detach", text{"runs in the background and waits for the app to respond", "roda em segundo plano e espera o app responder"}},
			{"--no-run", text{"prepares the environment, without running the app", "prepara o ambiente, sem rodar o app"}},
			{"--open", text{"opens the browser when the app is up", "abre o navegador quando o app subir"}}, optJSON},
		examples: []string{"romanctl up", "romanctl up shop -d", "romanctl up --no-run"}},
	{path: "down", aliases: []string{"derrubar"}, usage: "romanctl down [project] [--all] [--keep-services]",
		summary: text{"Stops the app and the services only it was using", "Para o app e os serviços que só ele estava usando"},
		details: text{"A service that was already on before up, or that another project uses, stays on.", "Um serviço que já estava ligado antes do up, ou que outro projeto usa, continua ligado."},
		options: []helpOption{
			{"--all", text{"brings down all running projects", "derruba todos os projetos no ar"}},
			{"--keep-services", text{"stops only the app", "para só o app"}}, optJSON},
		examples: []string{"romanctl down", "romanctl down shop", "romanctl down --all"}},
	{path: "run", aliases: []string{"dev", "rodar"}, usage: "romanctl run [project] [--open]",
		summary:  text{"Runs a project; without an argument, opens a list", "Roda um projeto; sem argumento, abre uma lista"},
		options:  []helpOption{{"--open", text{"opens the browser when the app is up", "abre o navegador quando o app subir"}}},
		examples: []string{"romanctl run", "romanctl run shop --open"}},
	{path: "rm", aliases: []string{"remove", "remover", "delete"}, usage: "romanctl rm [project] [-y]",
		summary:  text{"Removes a project and all its files", "Remove um projeto e todos os arquivos dele"},
		options:  []helpOption{{"-y, --yes", text{"does not ask for confirmation", "não pede confirmação"}}},
		examples: []string{"romanctl rm", "romanctl rm shop"}},
	{path: "code", aliases: []string{"editar", "edit"}, usage: "romanctl code [project|folder]",
		summary:  text{"Opens in code-server; without a target, uses the current folder's project or shows a list", "Abre no code-server; sem alvo, usa o projeto da pasta atual ou mostra uma lista"},
		examples: []string{"romanctl code", "romanctl code shop", "romanctl code ."}},
	{path: "open", aliases: []string{"abrir"}, usage: "romanctl open [code|project|url] [--app]",
		summary:  text{"Opens in the browser; without a target, opens a list", "Abre no navegador; sem alvo, abre uma lista"},
		options:  []helpOption{{"--app", text{"opens the running app instead of code-server", "abre o app rodando, em vez do code-server"}}},
		examples: []string{"romanctl open", "romanctl open shop --app", "romanctl open https://example.com"}},
	{path: "add", aliases: []string{"adicionar", "generate", "g"}, usage: "romanctl add [type] [name] [--in <aggregate>]",
		summary: text{"Generates code according to the project, asking for whatever is missing", "Gera código conforme o projeto, perguntando o que faltar"},
		details: text{"The types depend on the project: see with romanctl add --list.", "Os tipos dependem do projeto: veja com romanctl add --list."},
		options: []helpOption{
			{"--in <aggregate>", text{"target aggregate, for command, query, event and value-object", "agregado de destino, em command, query, event e value-object"}},
			{"--plural <name>", text{"plural of the name, when the English rule gets it wrong, such as --plural People", "plural do nome, quando a regra do inglês erra, como --plural People"}},
			{"--project <name>", text{"target project; without this, uses the current folder", "projeto de destino; sem isto, usa a pasta atual"}},
			{"--list", text{"lists what can be generated", "lista o que dá para gerar"}},
			{"--force", text{"overwrites existing files", "sobrescreve arquivos existentes"}},
			{"--dry-run", text{"shows what would be created", "mostra o que seria criado"}}, optJSON},
		examples: []string{"romanctl add", "romanctl add resource Product", "romanctl add command ArchiveProduct --in Product", "romanctl add page about --project site"}},
	{path: "publish", aliases: []string{"publicar"}, usage: "romanctl publish [project] [--public] [--description <text>]",
		summary: text{"Generates CI and a Dockerfile, commits and publishes to GitHub", "Gera CI e Dockerfile, faz o commit e publica no GitHub"},
		details: text{"The repository is private unless you choose otherwise. Files such as .env cancel the publishing.", "O repositório é privado, salvo se você escolher outra coisa. Arquivos como .env cancelam a publicação."},
		options: []helpOption{
			{"--public", text{"creates a public repository", "cria um repositório público"}},
			{"--description <text>", text{"repository description", "descrição do repositório"}},
			{"--no-ci", text{"does not generate CI or a Dockerfile", "não gera CI nem Dockerfile"}},
			{"-y, --yes", text{"asks nothing", "não pergunta nada"}},
			{"--dry-run", text{"shows what it would do", "mostra o que faria"}}, optJSON},
		examples: []string{"romanctl publish", "romanctl publish shop --description \"Shop\"", "romanctl publish --dry-run"}},

	{path: "tools", aliases: []string{"ferramentas"}, usage: "romanctl tools [list|info|versions|install|uninstall]",
		summary: text{"Installs and removes development tools", "Instala e remove ferramentas de desenvolvimento"},
		details: text{"In a terminal, without a subcommand, it opens a menu. Without a terminal, it lists.", "Em um terminal, sem subcomando, abre um menu. Sem terminal, lista."}},
	{path: "tools list", aliases: []string{"ls"}, usage: "romanctl tools list [--category <names>] [--all] [--filter <text>] [--plain] [--json]",
		summary: text{"Lists what runs on this machine; in a terminal, with a filter, categories and actions", "Lista o que roda nesta máquina; no terminal, com filtro, categorias e ações"},
		details: text{"In the list, type to filter and use ←/→ to change category. Typing the category name also works.", "Na lista, digite para filtrar e use ←/→ para trocar de categoria. Digitar o nome da categoria também funciona."},
		options: []helpOption{
			{"--category <names>", text{"only these categories, separated by commas", "só estas categorias, separadas por vírgula"}},
			{"--all", text{"includes what does not run here, with the reason", "inclui o que não roda aqui, com o motivo"}}, optFilter, optPlain, optJSON},
		examples: []string{"romanctl tools list", "romanctl tools list --category databases", "romanctl tools list --category db,sdk --plain", "romanctl tools list --filter kube"}},
	{path: "tools info", aliases: []string{"show", "readme"}, usage: "romanctl tools info <name> [--no-readme] [--json]",
		summary:  text{"Installed version, available version, source and README of a tool", "Versão instalada, versão disponível, origem e README de uma ferramenta"},
		details:  text{"The available version comes from the same source the installation uses: GitHub, package manager, Homebrew, npm, PyPI or the official installer.", "A versão disponível vem da mesma origem que a instalação usa: GitHub, gerenciador de pacotes, Homebrew, npm, PyPI ou o instalador oficial."},
		options:  []helpOption{{"--no-readme", text{"does not download or show the README", "não baixa nem mostra o README"}}, optJSON},
		examples: []string{"romanctl tools info k9s", "romanctl tools info terraform --no-readme"}},
	{path: "tools versions", aliases: []string{"versoes"}, usage: "romanctl tools versions <name> [--json]",
		summary:  text{"Lists the versions that can be installed, newest first", "Lista as versões que dá para instalar, da mais nova para a mais antiga"},
		details:  text{"They come from the installer's own source: GitHub releases, HashiCorp, npm, PyPI, apt or dnf. With Homebrew and some custom installers, only the latest.", "Vêm da mesma origem do instalador: releases do GitHub, HashiCorp, npm, PyPI, apt ou dnf. Com o Homebrew e alguns instaladores próprios, só a última."},
		options:  []helpOption{optJSON},
		examples: []string{"romanctl tools versions helm", "romanctl tools install helm --version 3.18.4"}},
	{path: "tools categories", aliases: []string{"categorias"}, usage: "romanctl tools categories [--json]",
		summary: text{"Shows the categories and how many tools of each run here", "Mostra as categorias e quantas ferramentas de cada uma rodam aqui"}, options: []helpOption{optJSON}},
	{path: "tools install", aliases: []string{"instalar", "add"}, usage: "romanctl tools install [names|groups|all]",
		summary: text{"Installs tools; a group opens a list with an All option", "Instala ferramentas; um grupo abre uma lista com a opção Todos"},
		options: []helpOption{
			{"-y, --yes", text{"installs the whole group, without a list", "instala o grupo inteiro, sem lista"}},
			{"--force", text{"reinstalls", "reinstala"}},
			{"--version <version>", text{"installs this version, even if older; one tool at a time", "instala esta versão, mesmo mais antiga; uma ferramenta por vez"}},
			{"--dry-run", text{"shows what would be installed", "mostra o que seria instalado"}}, optJSON},
		examples: []string{"romanctl tools install", "romanctl tools install k8s", "romanctl tools install kubectl gh", "romanctl tools install sdk --yes", "romanctl tools install terraform --version 1.9.8"}},
	{path: "tools uninstall", aliases: []string{"remove", "rm", "remover", "desinstalar"}, usage: "romanctl tools uninstall [names|groups]",
		summary: text{"Removes tools, with a list and confirmation", "Remove ferramentas, com lista e confirmação"},
		options: []helpOption{
			{"-y, --yes", text{"shows no list and asks no confirmation; refuses all", "não mostra lista nem pede confirmação; recusa all"}},
			{"--dry-run", text{"shows what would be removed", "mostra o que seria removido"}}, optJSON},
		examples: []string{"romanctl tools uninstall", "romanctl tools uninstall k8s", "romanctl tools uninstall jq --yes"}},

	{path: "system", aliases: []string{"sistema", "sys"}, usage: "romanctl system [status|battery|watch]",
		summary: text{"Shows CPU, memory, disk and battery", "Mostra CPU, memória, disco e bateria"},
		details: text{"In a terminal, without a subcommand, it opens a menu. Without a terminal, it shows the status.", "Em um terminal, sem subcomando, abre um menu. Sem terminal, mostra o status."}},
	{path: "system status", usage: "romanctl system status [--json]",
		summary: text{"CPU, memory, disk, battery and top processes", "CPU, memória, disco, bateria e processos que mais consomem"}, options: []helpOption{optJSON}},
	{path: "system battery", aliases: []string{"bateria"}, usage: "romanctl system battery [--json]",
		summary: text{"Level, power in watts, temperature and time left", "Nível, consumo em watts, temperatura e previsão de autonomia"}, options: []helpOption{optJSON}},
	{path: "system watch", aliases: []string{"medir"}, usage: "romanctl system watch [--seconds N] [--interval N]",
		summary: text{"Measures power and CPU together and estimates what your processes use", "Mede consumo e CPU juntos e estima o gasto dos seus processos"},
		options: []helpOption{
			{"--seconds N", text{"duration; without this, asks in a terminal", "duração; sem isto, pergunta em um terminal"}},
			{"--interval N", text{"seconds between measurements; the default is 5", "segundos entre as medições; o padrão é 5"}}, optJSON},
		examples: []string{"romanctl system watch", "romanctl system watch --seconds 600"}},
	{path: "platform", aliases: []string{"plataforma"}, usage: "romanctl platform [--json]",
		summary: text{"Shows the detected system, architecture and capabilities", "Mostra sistema, arquitetura e capacidades detectados"}, options: []helpOption{optJSON}},
	{path: "doctor", aliases: []string{"info"}, usage: "romanctl doctor",
		summary: text{"Shows the versions of the main tools", "Mostra as versões das ferramentas principais"}},

	{path: "offline", usage: "romanctl offline [status|prep|test]",
		summary: text{"Prepares the machine to work without internet", "Prepara a máquina para trabalhar sem internet"},
		details: text{"In a terminal, without a subcommand, it opens a menu. Without a terminal, it shows the status.", "Em um terminal, sem subcomando, abre um menu. Sem terminal, mostra o status."}},
	{path: "offline status", usage: "romanctl offline status",
		summary: text{"Shows the preparation status", "Mostra o estado da preparação"}},
	{path: "offline prep", aliases: []string{"preparar"}, usage: "romanctl offline prep [--force]",
		summary: text{"Downloads templates, caches and extensions; needs internet", "Baixa templates, caches e extensões; precisa de internet"},
		options: []helpOption{{"--force", text{"redoes the templates", "refaz os templates"}}}},
	{path: "offline test", aliases: []string{"testar"}, usage: "romanctl offline test",
		summary: text{"Creates and builds each project type with the network blocked", "Cria e compila cada tipo de projeto com a rede bloqueada"}},

	{path: "login", aliases: []string{"entrar"}, usage: "romanctl login [--issuer URL] [--api URL]",
		summary: text{"Signs in to romanctl through Keycloak; without a login the CLI does not run", "Entra no romanctl pelo Keycloak; sem login a CLI não roda"},
		details: text{"Opens the browser on the private realm at auth.marceloroman.com.br (authorization code with PKCE, redirect on 127.0.0.1), which only accepts Google with a criarch Workspace account registered beforehand in IAM Manager, and keeps the session in ~/.config/romanctl/session.json, readable only by you. The session is offline: it lasts 30 days without use and renews on every command, even without internet. In pipelines, ROMANCTL_CLIENT_ID and ROMANCTL_CLIENT_SECRET (an IAM Manager API client) replace the login. Without a login only login, logout, whoami, help, version, self-update and templates run.", "Abre o navegador no realm private de auth.marceloroman.com.br (authorization code com PKCE, redirect em 127.0.0.1), que só aceita o Google com conta do Workspace da criarch cadastrada antes no IAM Manager, e guarda a sessão em ~/.config/romanctl/session.json, só para você. A sessão é offline: vale 30 dias sem uso e renova a cada comando, também sem internet. Em pipelines, ROMANCTL_CLIENT_ID e ROMANCTL_CLIENT_SECRET (um cliente da API do IAM Manager) substituem o login. Sem login rodam só login, logout, whoami, help, version, self-update e templates."},
		options: []helpOption{
			{"--issuer <url>", text{"another realm; also ROMANCTL_AUTH_ISSUER", "outro realm; também ROMANCTL_AUTH_ISSUER"}},
			{"--api <url>", text{"another IAM Manager API; also ROMANCTL_IAM_API", "outra API do IAM Manager; também ROMANCTL_IAM_API"}},
			{"--idp <alias>", text{"provider to sign in with; the default is google, and - lets the realm choose (also ROMANCTL_AUTH_IDP)", "provedor onde entrar; o padrão é google, e - deixa o realm escolher (também ROMANCTL_AUTH_IDP)"}}, optJSON},
		examples: []string{"romanctl login", "romanctl login --issuer http://localhost:18081/realms/private --api http://localhost:18091"}},
	{path: "logout", aliases: []string{"sair"}, usage: "romanctl logout",
		summary: text{"Ends the session in Keycloak and deletes the local session", "Encerra a sessão no Keycloak e apaga a sessão local"}, options: []helpOption{optJSON}},
	{path: "whoami", usage: "romanctl whoami [--json]",
		summary: text{"Shows who is logged in, the roles and until when the session lasts", "Mostra quem está logado, os papéis e até quando a sessão vale"}, options: []helpOption{optJSON}},
	{path: "iam", aliases: []string{"keycloak"}, usage: "romanctl iam <command> [--json]",
		summary:  text{"Manages Keycloak through the IAM Manager API, with an API client", "Administra o Keycloak pela API do IAM Manager, com um cliente da API"},
		details:  text{"Calls use an API client (client_credentials on the private realm), created in IAM Manager → Operators → API clients. The credential comes from IAM_CLIENT_ID and IAM_CLIENT_SECRET (with optional IAM_API_URL and IAM_TOKEN_URL) or from ~/.config/iam-manager/credentials.env, written by romanctl iam credentials set. The token lasts 5 minutes and renews by itself. The API checks the client's permissions (global roles iam-admin, iam-user-admin, iam-app-admin, iam-theme-admin, iam-auditor or the public:admin and public:viewer delegations) and records everything in the audit log under its client_id. The realms are private (operators) and public (application users). Users accept the username or the id.", "As chamadas usam um cliente da API (client_credentials no realm private), criado em IAM Manager → Operadores → Clientes da API. A credencial vem de IAM_CLIENT_ID e IAM_CLIENT_SECRET (com IAM_API_URL e IAM_TOKEN_URL opcionais) ou de ~/.config/iam-manager/credentials.env, gravado por romanctl iam credentials set. O token vale 5 minutos e é renovado sozinho. A API confere as permissões do cliente (papéis globais iam-admin, iam-user-admin, iam-app-admin, iam-theme-admin, iam-auditor ou delegações public:admin e public:viewer) e registra tudo na auditoria com o client_id dele. Os realms são private (operadores) e public (usuários das aplicações). Usuários aceitam o nome ou o id."},
		options:  []helpOption{optJSON},
		examples: []string{"romanctl iam credentials set --client-id iam-api-ci < secret.txt", "romanctl iam me", "romanctl iam realms", "romanctl iam users public --search ana", "romanctl iam user create public ana --email ana@acme.com", "romanctl iam api GET /api/themes"}},
	{path: "iam credentials", usage: "romanctl iam credentials [show|set|clear] [--client-id ID] [--api URL] [--token-url URL]",
		summary:  text{"Shows, saves or deletes the API client romanctl uses", "Mostra, grava ou apaga o cliente da API que o romanctl usa"},
		details:  text{"set reads the client_secret from standard input (or asks without echo in a terminal), checks it with GET /api/me and saves it to ~/.config/iam-manager/credentials.env with 0600. show never prints the secret. IAM_CLIENT_ID and IAM_CLIENT_SECRET in the environment override the file.", "set lê o client_secret da entrada padrão (ou pede sem eco no terminal), confere com GET /api/me e grava em ~/.config/iam-manager/credentials.env com 0600. show nunca mostra o secret. IAM_CLIENT_ID e IAM_CLIENT_SECRET no ambiente valem por cima do arquivo."},
		examples: []string{"romanctl iam credentials", "romanctl iam credentials set --client-id iam-api-romanctl", "romanctl iam credentials set --client-id iam-api-romanctl < ~/secret.txt"}},
	{path: "iam me", usage: "romanctl iam me",
		summary: text{"Who the client is in the API and what it can do", "Quem é o cliente na API e o que ele pode fazer"}},
	{path: "iam realms", usage: "romanctl iam realms",
		summary: text{"Lists the realms you can see", "Lista os realms que você pode ver"}},
	{path: "iam realm", usage: "romanctl iam realm <realm>",
		summary:  text{"Shows a realm's settings, which come from Terraform", "Mostra as configurações de um realm, que vêm do Terraform"},
		examples: []string{"romanctl iam realm public"}},
	{path: "iam users", usage: "romanctl iam users <realm> [--search X]",
		summary: text{"Lists the users of a realm", "Lista usuários de um realm"}},
	{path: "iam user", usage: "romanctl iam user show|create|enable|disable|roles|access|reset|logout|delete <realm> <user>",
		summary:  text{"Creates and manages a user", "Cria e administra um usuário"},
		details:  text{"create accepts --email, --first-name, --last-name, --roles a,b and --credential: link (default, a one-time link to set the password), temporary, email or none. reset uses the same --credential. roles sets the complete state of the realm roles. access shows the user's roles in each application; with the application id and --roles, it sets that application's roles, and an empty --roles revokes the access.", "create aceita --email, --first-name, --last-name, --roles a,b e --credential: link (padrão, link de uso único para definir a senha), temporary, email ou none. reset usa o mesmo --credential. roles define o estado completo dos papéis do realm. access mostra as roles do usuário em cada aplicação; com o id da aplicação e --roles, define as dela, e --roles vazio revoga o acesso."},
		examples: []string{"romanctl iam user create public ana --email ana@acme.com", "romanctl iam user access public ana notes --roles reader", "romanctl iam user disable public ana", "romanctl iam user reset public ana --credential temporary"}},
	{path: "iam apps", usage: "romanctl iam apps <realm>",
		summary: text{"Lists the realm's applications, with roles, clients and theme", "Lista as aplicações do realm, com roles, clients e tema"}},
	{path: "iam app", usage: "romanctl iam app create <realm> <id> --roles a,b [--name X] | role <realm> <app> <role> | theme <realm> <app> [theme]",
		summary:  text{"Registers an application, adds a role to it or changes its login theme", "Cadastra uma aplicação, acrescenta uma role ou troca o tema de login dela"},
		details:  text{"The application does not sign in: it provides the audience and the roles APIs check. Web apps linked to it are what sign in. theme without a name goes back to the realm theme. Those declared in Terraform are read-only.", "A aplicação não faz login: dá a audiência e as roles que as APIs conferem. Quem faz login são as web apps ligadas a ela. theme sem nome volta ao tema do realm. As declaradas no Terraform são só leitura."},
		examples: []string{"romanctl iam app create public notes --name Notes --roles reader,editor", "romanctl iam app role public notes admin", "romanctl iam app theme public notes acme-v3"}},
	{path: "iam api-keys", usage: "romanctl iam api-keys <realm>",
		summary: text{"Lists the API keys (client_credentials) of a realm", "Lista as API keys (client_credentials) de um realm"}},
	{path: "iam api-key", usage: "romanctl iam api-key create <realm> <name> [--description X] [--roles a,b] | rotate|delete <realm> <id>",
		summary: text{"Creates, rotates or deletes an API key; the secret is shown only once", "Cria, rotaciona ou exclui uma API key; o secret só aparece na hora"}},
	{path: "iam operators", usage: "romanctl iam operators",
		summary: text{"Who uses IAM Manager and romanctl, and with which roles", "Quem usa o IAM Manager e o romanctl, com quais papéis"},
		details: text{"Operators are the users of the private realm.", "Os operadores são os usuários do realm private."}},
	{path: "iam operator", usage: "romanctl iam operator add <e-mail> [--first-name X] [--last-name Y] [--global a,b] [--realms public=admin] | roles <operator> [--global a,b] [--realms public=viewer]",
		summary:  text{"Registers an operator or sets their roles (iam-admin only)", "Cadastra um operador ou define os papéis dele (só iam-admin)"},
		details:  text{"The private realm only lets in people registered beforehand: add takes the criarch Google Workspace email, and the account links on first login. roles accepts the username, the email or the id; whatever is not on the line stays as it is, and empty --global and --realms remove everything.", "O realm private só deixa entrar quem foi cadastrado antes: add recebe o e-mail do Google Workspace da criarch, e a conta se vincula no primeiro login. roles aceita o nome de usuário, o e-mail ou o id; o que não vier na linha fica como está, e --global e --realms vazios tiram tudo."},
		examples: []string{"romanctl iam operator add ana@criarch.com.br --first-name Ana --global iam-auditor", "romanctl iam operator roles ana@criarch.com.br --realms public=admin"}},
	{path: "iam api-clients", aliases: []string{"clients"}, usage: "romanctl iam api-clients",
		summary: text{"Lists the IAM Manager API clients (iam-admin only)", "Lista os clientes da API do IAM Manager (só iam-admin)"}},
	{path: "iam api-client", aliases: []string{"client"}, usage: "romanctl iam api-client create <name> [--description X] [--global a,b] [--realms public=admin] | roles|enable|disable|rotate|delete <client>",
		summary:  text{"Creates and manages an API client; the secret is shown only once", "Cria e administra um cliente da API; o secret só aparece na hora"},
		details:  text{"The client accepts the id, the client_id or the name. roles, enable and disable keep whatever is not on the line: empty --global and --realms remove everything. Rotating the client romanctl uses saves the new secret to credentials.env.", "O cliente aceita o id, o client_id ou o nome. roles, enable e disable mantêm o que não vier na linha: --global e --realms vazios tiram tudo. Rotacionar o cliente que o romanctl usa grava o secret novo no credentials.env."},
		examples: []string{"romanctl iam api-client create pipeline --realms public=admin", "romanctl iam api-client roles pipeline --global iam-auditor", "romanctl iam api-client rotate pipeline"}},
	{path: "iam events", usage: "romanctl iam events <realm> [--limit N]",
		summary: text{"Logins and admin changes stored by Keycloak", "Logins e alterações administrativas guardados pelo Keycloak"}},
	{path: "iam audit", usage: "romanctl iam audit [--limit N]",
		summary: text{"What was done through IAM Manager, and by whom", "O que foi feito pelo IAM Manager, por quem"}},
	{path: "iam api", usage: "romanctl iam api <GET|POST|PUT|DELETE> <path> [--data JSON]",
		summary:  text{"Calls any API route with the client's token", "Chama qualquer rota da API com o token do cliente"},
		details:  text{"The specification is at /api/openapi.json and the guide at /api/guide.md.", "A especificação está em /api/openapi.json e o guia em /api/guide.md."},
		examples: []string{"romanctl iam api GET /api/realms/public/applications", `romanctl iam api PUT /api/realms/public/events/config --data '{"eventsEnabled":true}'`}},
	{path: "web-ui", aliases: []string{"webui", "web"}, usage: "romanctl web-ui [--port N] [--open]",
		summary: text{"Opens the romanctl web interface in the browser, on localhost", "Abre a interface web do romanctl no navegador, em localhost"},
		details: text{"Listens only on 127.0.0.1 and requires a token embedded in the page, so other sites cannot call the API. Projects, services, tools and templates, with actions that run romanctl itself.", "Serve só em 127.0.0.1 e exige um token que vai na página, então outros sites não chamam a API. Projetos, serviços, ferramentas e templates, com as ações rodando o próprio romanctl."},
		options: []helpOption{
			{"--port N", text{"port; the default is 7777", "porta; o padrão é 7777"}},
			{"--open", text{"opens the browser", "abre o navegador"}}},
		examples: []string{"romanctl web-ui", "romanctl web-ui --port 8888 --open"}},
	{path: "templates", aliases: []string{"template"}, usage: "romanctl templates [status|update] [--ref <tag or branch>]",
		summary:  text{"Shows and updates the project templates and generators", "Mostra e atualiza os templates de projeto e os geradores"},
		details:  text{"Templates live outside the binary, in ~/.local/share/romanctl/templates, and come from the templates folder of the repository, at the tag of the installed version. ROMANCTL_TEMPLATES points to another folder.", "Os templates ficam fora do binário, em ~/.local/share/romanctl/templates, e vêm da pasta templates do repositório, na tag da versão instalada. ROMANCTL_TEMPLATES aponta para outra pasta."},
		options:  []helpOption{{"--ref <ref>", text{"downloads from a tag or branch, such as refs/heads/main", "baixa de uma tag ou branch, como refs/heads/main"}}, optJSON},
		examples: []string{"romanctl templates", "romanctl templates update", "romanctl templates update --ref refs/heads/main"}},
	{path: "templates status", usage: "romanctl templates status",
		summary: text{"Shows where the templates are and which version they came from", "Mostra onde estão os templates e de que versão vieram"}, options: []helpOption{optJSON}},
	{path: "templates update", aliases: []string{"atualizar"}, usage: "romanctl templates update [--ref <ref>]",
		summary: text{"Downloads the templates of the installed version", "Baixa os templates da versão instalada"}, options: []helpOption{optJSON}},
	{path: "claude", aliases: []string{"skills"}, usage: "romanctl claude [status|install|uninstall]",
		summary:  text{"Installs the skills that teach Claude Code to use romanctl", "Instala as skills que ensinam o Claude Code a usar o romanctl"},
		details:  text{"The skills go to ~/.claude/skills (or $CLAUDE_CONFIG_DIR/skills). romanctl only replaces or deletes the folders it created itself.", "As skills vão para ~/.claude/skills (ou $CLAUDE_CONFIG_DIR/skills). O romanctl só troca ou apaga as pastas que ele mesmo criou."},
		examples: []string{"romanctl claude install", "romanctl claude status"}},
	{path: "claude status", usage: "romanctl claude status [--json]",
		summary: text{"Shows each skill: installed, outdated, missing or foreign", "Mostra cada skill: instalada, desatualizada, ausente ou alheia"}, options: []helpOption{optJSON}},
	{path: "claude install", aliases: []string{"instalar", "update", "atualizar"}, usage: "romanctl claude install [--force] [--dry-run] [--json]",
		summary: text{"Installs or updates the skills", "Instala ou atualiza as skills"},
		options: []helpOption{
			{"--force", text{"also replaces folders created outside romanctl", "troca também as pastas criadas fora do romanctl"}},
			{"--dry-run", text{"shows what it would do", "mostra o que faria"}}, optJSON}},
	{path: "claude uninstall", aliases: []string{"remove", "rm", "remover"}, usage: "romanctl claude uninstall [--dry-run] [--json]",
		summary: text{"Removes the skills romanctl installed", "Remove as skills que o romanctl instalou"},
		options: []helpOption{{"--dry-run", text{"shows what it would do", "mostra o que faria"}}, optJSON}},
	{path: "ui", aliases: []string{"painel", "tui"}, usage: "romanctl ui",
		summary: text{"Opens the full-screen panel, k9s style", "Abre o painel de tela cheia, nos moldes do k9s"},
		details: text{"Inside it: Tab changes screens, / filters, : jumps, Enter shows the actions, ? shows the help.", "Dentro dele: Tab troca de tela, / filtra, : vai direto, Enter mostra as ações, ? mostra a ajuda."}},
	{path: "menu", usage: "romanctl menu",
		summary: text{"Opens the list menu, simpler than the panel", "Abre o menu em lista, mais simples que o painel"}},
	{path: "config", usage: "romanctl config [lang [pt|en|auto]]",
		summary: text{"Shows or saves the settings", "Mostra ou grava as configurações"}},
	{path: "config lang", usage: "romanctl config lang [pt|en|auto]",
		summary:  text{"Chooses the language; auto follows the system again", "Escolhe o idioma; auto volta a seguir o sistema"},
		examples: []string{"romanctl config lang", "romanctl config lang en"}},
	{path: "help", aliases: []string{"ajuda"}, usage: "romanctl help [command]",
		summary: text{"Shows this help", "Mostra esta ajuda"}, examples: []string{"romanctl help tools install"}},
	{path: "self-update", aliases: []string{"update", "atualizar"}, usage: "romanctl self-update [--check] [--version <vX.Y.Z>] [--yes]",
		summary: text{"Updates romanctl to the latest published version", "Atualiza o romanctl para a última versão publicada"},
		details: text{"Checks the SHA-256 and runs the new version before swapping; if it does not work, the current one stays.", "Confere o SHA-256 e roda a versão nova antes de trocar; se ela não funcionar, a atual fica."},
		options: []helpOption{
			{"--check", text{"only shows whether there is a new version", "só mostra se há versão nova"}},
			{"--version <vX.Y.Z>", text{"installs this version, even if older", "instala esta versão, mesmo que mais antiga"}},
			{"-y, --yes", text{"does not ask for confirmation", "não pede confirmação"}}, optJSON},
		examples: []string{"romanctl self-update", "romanctl self-update --check", "romanctl self-update --version v0.1.0"}},
	{path: "version", aliases: []string{"versao"}, usage: "romanctl version",
		summary: text{"Shows the version", "Mostra a versão"}},
}

// placeholders translates argument names in usage lines and options for the Portuguese help.
var placeholders = map[string]string{"user": "usuário", "path": "caminho", "services": "serviços", "service": "serviço",
	"names": "nomes", "name": "nome", "groups": "grupos", "type": "tipo", "project": "projeto", "folder": "pasta",
	"version": "versão", "aggregate": "agregado", "text": "texto", "command": "comando", "target": "alvo",
	"theme": "tema", "client": "cliente", "operator": "operador", "or": "ou", "arguments": "argumentos"}

var (
	placeholderSpan = regexp.MustCompile(`[\[<][^\]>]*[\]>]`)
	placeholderWord = regexp.MustCompile(`(^|[^-\w])([a-z]+)`)
)

// localized translates only the placeholders inside [...] and <...>, never command or flag names.
func localized(s string) string {
	if i18n.Lang() != i18n.PT {
		return s
	}
	return placeholderSpan.ReplaceAllStringFunc(s, func(span string) string {
		return placeholderWord.ReplaceAllStringFunc(span, func(w string) string {
			m := placeholderWord.FindStringSubmatch(w)
			if pt, ok := placeholders[m[2]]; ok {
				return m[1] + pt
			}
			return w
		})
	})
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
	text{"Usage:", "Uso:"}, text{"Options:", "Opções:"}, text{"Examples:", "Exemplos:"},
	text{"Subcommands:", "Subcomandos:"}, text{"Aliases:", "Apelidos:"},
	text{"Unknown command; the full list follows.", "Comando desconhecido; a lista completa vem a seguir."},
}

// printHelp shows the general help or a command's help. Words that are not commands, such as
// a service name in "start postgres --help", do not get in the way.
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
	if c.details.en != "" {
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

var generalHelp = text{`
Usage: romanctl [command] [arguments] [--lang pt|en]
       romanctl            opens the command list: type to filter, Enter runs
       romanctl ui         opens the full-screen panel
       romanctl <command> --help   shows the help of any command or subcommand

In a terminal, whatever is missing is asked for, and in lists you just type to filter.
Without a terminal or with --json, nothing is asked.
`, `
Uso: romanctl [comando] [argumentos] [--lang pt|en]
     romanctl              abre a lista de comandos: digite para filtrar, Enter roda
     romanctl ui           abre o painel de tela cheia
     romanctl <comando> --help   mostra a ajuda de qualquer comando ou subcomando

Em um terminal, o que faltar é perguntado, e nas listas basta digitar para filtrar.
Sem terminal ou com --json, nada é perguntado.
`}

var usageGroups = []struct {
	title text
	paths []string
}{
	{text{"Tools", "Ferramentas"}, []string{"tools"}},
	{text{"Projects", "Projetos"}, []string{"new", "add", "up", "down", "run", "publish", "list", "code", "open", "rm", "catalog", "types", "templates"}},
	{text{"Services", "Serviços"}, []string{"status", "start", "stop", "restart", "restore", "logs", "install"}},
	{text{"System", "Sistema"}, []string{"system", "platform", "doctor", "offline"}},
	{text{"Access", "Acesso"}, []string{"login", "logout", "whoami", "iam"}},
	{text{"Other", "Outros"}, []string{"web-ui", "claude", "ui", "menu", "config", "self-update", "help", "version"}},
}

func usage() {
	fmt.Println(ui.STitle.Render("romanctl") + " " + ui.SGray.Render("v"+version) + i18n.T(" — installs tools, creates projects and brings developer utilities"))
	fmt.Print(generalHelp)
	for _, g := range usageGroups {
		fmt.Printf("\n%s\n", ui.STitle.Render(g.title.String()))
		for _, p := range g.paths {
			if c := findCommand([]string{p}); c != nil {
				fmt.Printf("  %-12s %s\n", c.path, c.summary)
			}
		}
	}
	fmt.Printf("\n%s %s\n", text{"Projects live in", "Projetos ficam em"}, config.ProjDir)
}
