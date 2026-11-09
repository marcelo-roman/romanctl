# romanctl

[![CI](https://github.com/marcelo-roman/romanctl/actions/workflows/ci.yml/badge.svg)](https://github.com/marcelo-roman/romanctl/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/marcelo-roman/romanctl)](https://github.com/marcelo-roman/romanctl/releases/latest)

CLI para quem desenvolve. Com um único comando, ela:

- **instala ferramentas**: mais de 170, como kubectl, gh, AWS e Azure CLI, Terraform, Docker, SDKs de várias
  linguagens e clientes de banco de dados;
- **cria projetos**, inclusive com arquitetura pronta (Clean Architecture, DDD e CQRS em .NET), e **gera código**
  dentro deles, como recursos completos, controllers, comandos e páginas;
- **sobe o ambiente** do projeto: PostgreSQL, MongoDB, Redis, MySQL, SQL Server, RabbitMQ e Kafka, o banco, o `.env`
  e o próprio app;
- **publica no GitHub**, com CI e Dockerfile;
- traz um **painel de terminal**, nos moldes do k9s, para fazer tudo isso sem decorar comandos.

Funciona em Linux e macOS, em `amd64` e `arm64`, e também no WSL e no Termux com `proot-distro`. Fala português e
inglês, conforme o idioma do sistema.

## Instalação

```bash
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | sh
```

Pronto. Depois, rode `romanctl` para abrir o painel.

### O que o instalador faz

1. Descobre o sistema (Linux ou macOS) e a arquitetura (`amd64` ou `arm64`).
2. Baixa o binário da [última release](https://github.com/marcelo-roman/romanctl/releases/latest).
3. Confere o SHA-256 com o arquivo `SHA256SUMS` da release. Se não bater, para sem instalar nada.
4. Instala em `~/.local/bin/romanctl`, ou em `/usr/local/bin/romanctl` quando roda como root.
5. Avisa se essa pasta não estiver no seu `PATH`.

Requisitos: `curl` ou `wget`, e `sha256sum` ou `shasum`. Já vêm instalados na maioria dos sistemas.

### Prefere ler o script antes?

```bash
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh -o install.sh
less install.sh
sh install.sh
```

### Escolher a versão ou a pasta

```bash
# uma versão específica
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | ROMANCTL_VERSION=v0.1.0 sh

# outra pasta
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | ROMANCTL_INSTALL_DIR="$HOME/bin" sh

# para todos os usuários da máquina
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | sudo sh
```

### Colocar no PATH

Se o instalador avisar que `~/.local/bin` não está no `PATH`, acrescente ao arquivo do seu shell e abra um terminal
novo:

| Shell | Comando |
|---|---|
| bash | `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc` |
| zsh | `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc` |
| fish | `fish_add_path ~/.local/bin` |

### Conferir

```bash
romanctl version      # romanctl 0.1.0
romanctl platform     # o que ele detectou na sua máquina
romanctl doctor       # versões das ferramentas principais
```

### Por sistema

**Linux.** Funciona em Debian, Ubuntu, Fedora, RHEL, Arch, openSUSE e Alpine. As ferramentas são instaladas pelo
gerenciador de pacotes da distribuição (`apt`, `dnf`, `pacman`, `zypper` ou `apk`) ou pelo binário oficial do
fornecedor. Sem root, só o gerenciador de pacotes usa `sudo`. Para os serviços, como PostgreSQL e Kafka, é preciso
Docker ou Podman: `romanctl tools install docker`.

**macOS.** As ferramentas são instaladas pelo [Homebrew](https://brew.sh), que precisa estar instalado. Para os
serviços, `romanctl tools install docker` instala o Colima com o Docker.

**WSL.** Funciona como o Linux da distribuição que você usa.

**Termux (Android).** Use um Ubuntu dentro do Termux e instale lá dentro:

```bash
pkg install proot-distro
proot-distro install ubuntu
proot-distro login ubuntu
apt update && apt install -y curl
curl -fsSL https://raw.githubusercontent.com/marcelo-roman/romanctl/main/install.sh | sh
```

No proot, containers não rodam. O `romanctl` detecta isso: PostgreSQL, MongoDB e Redis rodam como processos locais,
e o que precisa de container fica fora das listas. Detalhes em [Termux e proot](#termux-e-proot).

### A partir do código

Precisa do Go:

```bash
git clone https://github.com/marcelo-roman/romanctl.git
cd romanctl
go build -ldflags="-s -w" -o ~/.local/bin/romanctl ./cmd/romanctl
```

## Atualizar

```bash
romanctl self-update            # atualiza para a última versão
romanctl self-update --check    # só mostra se há versão nova
romanctl self-update --version v0.1.0
```

O `self-update` confere o SHA-256 e roda a versão nova antes de trocar a atual: se ela não funcionar na sua máquina,
nada muda. Se o `romanctl` estiver numa pasta do sistema, como `/usr/local/bin`, use `sudo`. Rodar o instalador de
novo também atualiza.

## Desinstalar

```bash
rm "$(command -v romanctl)"
```

Isso remove só o programa. O que ele criou fica até você apagar:

| Pasta | O que tem |
|---|---|
| `~/.local/share/romanctl` | dados dos serviços locais, logs, estado e caches |
| `~/.config/romanctl` | idioma escolhido, templates e geradores próprios |
| `~/projetos` | os seus projetos |

As ferramentas instaladas continuam no sistema. Para removê-las antes, use `romanctl tools uninstall`.
Os containers dos serviços se chamam `romanctl-<serviço>`, e os dados ficam em volumes de mesmo nome.

## Primeiros passos

**1. Abra o painel** e explore com o teclado:

```bash
romanctl
```

`Tab` troca de tela, `/` filtra, `Enter` mostra o que dá para fazer com o item, e `?` mostra todas as teclas.

**2. Instale as ferramentas** de que você precisa:

```bash
romanctl tools list                  # o que roda na sua máquina
romanctl tools install k8s           # escolha numa lista do grupo Kubernetes
romanctl tools install gh node go    # ou diga quais
```

**3. Crie um projeto** e gere código nele:

```bash
romanctl new loja --stack dotnet --arch clean,ddd,cqrs --with postgres
cd ~/projetos/loja
romanctl add resource Product        # CRUD completo, do domínio aos endpoints
```

**4. Suba tudo**: PostgreSQL, banco, `.env` e a API.

```bash
romanctl up -d
curl http://localhost:5080/products
romanctl down
```

**5. Publique** no GitHub, em repositório privado:

```bash
romanctl publish
```

Qualquer comando sem argumentos pergunta o que falta. Para ajuda, use `romanctl <comando> --help`.

## Comandos

| Comando | Para quê |
|---|---|
| `romanctl` | Abre o painel |
| `tools list`, `tools install`, `tools uninstall`, `tools categories` | Ferramentas |
| `new`, `add`, `catalog`, `types` | Criar projetos e gerar código |
| `up`, `down`, `run`, `logs` | Rodar projetos com o ambiente |
| `code`, `open`, `list`, `rm` | Abrir, listar e remover projetos |
| `publish` | Publicar no GitHub |
| `status`, `start`, `stop`, `restart`, `install`, `restore` | Serviços |
| `system`, `platform`, `doctor` | CPU, memória, bateria e diagnóstico |
| `offline` | Preparar para trabalhar sem internet |
| `config lang`, `self-update`, `menu` | Idioma, atualização e menu em lista |

Os projetos ficam em `~/projetos`; para usar outra pasta, defina `ROMANCTL_PROJETOS`. `--json` funciona nos comandos
de consulta, para scripts. Sem terminal, nada é perguntado. Os códigos de saída são
0 (sucesso), 1 (falha ao executar) e 2 (uso incorreto).

# Referência

## Painel

`romanctl` sem argumentos abre um painel de tela cheia, nos moldes do k9s, com telas de projetos, serviços, ferramentas
e sistema, que se atualizam sozinhas.

| Tecla | O que faz |
|---|---|
| `Tab`, `1` a `4` | Troca de tela |
| `:` | Vai direto: `:tools`, `:svc`, `:prj`, `:sys`, ou uma categoria como `:db` e `:k8s` |
| `/` | Filtra a lista |
| `Enter` | Mostra as ações do item |
| `d`, `l` | Detalhes e logs do item |
| `<` `>` | Troca de categoria, em Ferramentas |
| `?` | Mostra todas as teclas da tela atual |
| `Esc`, `q` | Volta e sai |

O rodapé mostra só as teclas que valem para o item selecionado. Ações longas rodam em segundo plano, com o andamento na
própria linha, e remover sempre pede confirmação. `romanctl menu` abre a versão em lista, mais simples.

## Idioma

Português e inglês. O idioma segue o sistema: `LC_ALL`, `LC_MESSAGES`, `LANGUAGE`, `LANG` e, no macOS, o idioma
da conta. Sem nenhuma pista, usa inglês. Para escolher:

```bash
romanctl config lang pt      # grava a escolha; auto volta a seguir o sistema
romanctl status --lang en    # só neste comando
ROMANCTL_LANG=en romanctl status
```

No Termux com `proot-distro`, o Android não informa o idioma ao Ubuntu, então vale gravar a escolha. Os projetos
gerados pelos templates continuam em português.

As mensagens em português são as chaves de `internal/i18n/en.go`, e um teste falha se alguma ficar sem tradução.

## Ferramentas

```bash
romanctl tools list                 # só o que roda nesta máquina
romanctl tools list --all           # inclui o resto, com o motivo
romanctl tools install kubectl gh aws az docker
romanctl tools install k8s          # lista o grupo, com a opção Todos
romanctl tools install k8s --yes    # instala o grupo inteiro
romanctl tools install k8s --dry-run
romanctl tools uninstall kubectl --yes
romanctl tools uninstall k8s        # lista o que está instalado no grupo
```

O catálogo fica em `internal/tools/catalog_*.go`, em grupos: `sdk`, `node-tools`, `python-tools`, `build`,
`git`, `cloud`, `k8s`, `iac`, `containers`, `security`, `db`, `observability`, `network`, `shell` e `ai`.
Downloads são conferidos com o SHA-256 do fornecedor quando ele publica um. Depois de instalar, a
ferramenta é executada para confirmar que roda.

Sem terminal, ou com `--json`, um grupo é instalado inteiro, porque não há quem responda ao menu.

A listagem, as listas de seleção e os grupos mostram só o que roda na máquina, conforme o sistema, a
arquitetura, o gerenciador de pacotes, a libc e as capacidades detectadas: containers (namespaces e cgroups,
e VM no macOS), virtualização (`/dev/kvm` ou `kern.hv_support`) e interface gráfica (`DISPLAY` ou
`WAYLAND_DISPLAY`). Se a detecção errar, force com `ROMANCTL_CONTAINERS`, `ROMANCTL_VM` e `ROMANCTL_GUI`,
valendo `0` ou `1`. O que cada ferramenta exige está em `internal/tools/requirements.go`.

`uninstall` funciona igual, mostrando só o que está instalado, e sempre pede confirmação antes de
remover. `--yes` dispensa a lista e a confirmação, e nesse caso `all` é recusado. Antes de remover pelo `apt`, ele simula a
operação e recusa se levar junto um pacote vital ou outro pacote instalado explicitamente.

## Listas e loading

Em um terminal, toda lista filtra ao digitar, sem diferenciar maiúsculas e acentos, e cada palavra precisa aparecer
na linha. O primeiro Esc limpa o filtro, e o segundo volta. Nas listas de múltipla escolha, Espaço marca e "Todos"
ou Ctrl+A marcam só o que o filtro está mostrando.

Nas listas de ferramentas, ←/→ trocam de categoria, e digitar o nome da categoria também funciona. Em comandos, use
`--category`, que aceita nomes como `db`, `"banco de dados"` e `linguagens`, separados por vírgula. `romanctl tools
categories` mostra todas.

`romanctl tools list` e `romanctl list` são listas navegáveis, com ações em cada item. Para imprimir, use `--plain`
ou `--filter <texto>`; com pipe, `--json` ou sem terminal, elas imprimem sozinhas.

O que demora mostra uma animação com progresso, só em terminal. As versões das ferramentas ficam em cache até o
binário mudar, e o tamanho dos caches do `offline` fica guardado por uma hora.

## Templates de arquitetura

```bash
romanctl new pedidos --stack dotnet --arch clean,ddd,cqrs --with postgres
```

Cada template é uma pasta com um `template.yaml`: os embutidos ficam em `templates/`, e uma pasta em
`~/.config/romanctl/templates/<stack>/<id>/` tem prioridade sobre eles. Todo arquivo é processado
com `text/template`, e `__Name__` nos caminhos vira o nome do projeto em PascalCase. Cada projeto
gerado ganha um `romanctl.yaml`, que o `run` e o `open` usam.

## Adicionar ao projeto

`romanctl add` gera código conforme o tipo do projeto, e `romanctl add --list` mostra as opções:

```bash
romanctl add resource Product                      # CRUD completo
romanctl add domain Customer
romanctl add controller Customer
romanctl add command ArchiveProduct --in Product
romanctl add query ListActiveProducts --in Product
romanctl add event ProductArchived --in Product
romanctl add value-object Sku --in Product
romanctl add resource Product --dry-run            # só mostra
```

| Projeto | Geradores |
|---|---|
| .NET Clean, DDD e CQRS | `resource`, `domain`, `application`, `repository`, `controller`, `command`, `query`, `event`, `value-object`, `test` |
| .NET simples | `controller`, `service`, `model` |
| Express | `route`, `middleware`, `service` |
| Next.js | `page`, `api`, `component` |
| React | `component`, `hook`, `page` |
| Go | `handler`, `service`, `model` |

Sem `--force`, nada é sobrescrito. O que precisa de registro em outro arquivo é inserido antes de um marcador,
como `// romanctl:endpoints`; sem o marcador, o comando mostra o que fazer à mão. Cada gerador é uma pasta em
`internal/scaffold/generators/<alvo>/<tipo>`, e uma pasta em `~/.config/romanctl/generators` tem prioridade.

## Subir o ambiente

```bash
romanctl up                  # serviços, banco, .env e o app, em primeiro plano
romanctl up loja -d          # em segundo plano; espera o app responder
romanctl up --no-run         # só prepara o ambiente
romanctl logs loja
romanctl down loja           # ou --all, ou --keep-services
```

O `up` lê os serviços do `romanctl.yaml`, cria o banco do projeto e acrescenta ao `.env` só o que falta, sem mudar
o que você editou. O `.env` entra no `.gitignore`, e as variáveis valem para o app, salvo as que já estiverem no
ambiente. Se a porta estiver ocupada, o comando para e diz por qual projeto, quando for o caso.

O `down` desliga só os serviços que o `up` ligou e que nenhum outro projeto no ar usa. O que já estava ligado antes
continua ligado. Projetos sem `romanctl.yaml` não têm serviços: o `up` só roda o app.

## Publicar

```bash
romanctl publish                                   # repositório privado
romanctl publish loja --public --description "Loja"
romanctl publish --dry-run
```

Gera CI e Dockerfile conforme o projeto, sem sobrescrever os que existirem, faz o commit, cria o repositório com
o `gh` e envia. Se o repositório já existir, só envia as alterações. Arquivos como `.env` e chaves privadas
cancelam a publicação, assim como segredos encontrados pelo `gitleaks`, se ele estiver instalado.

## Serviços

Fora do proot, os serviços rodam em Docker ou Podman, com portas só em `127.0.0.1` e dados em volumes
nomeados. O code-server roda como processo local.

| Serviço | Endereço |
|---|---|
| `code` | `http://127.0.0.1:8080` |
| `postgres` | `postgres://postgres:postgres@127.0.0.1:5432/postgres` |
| `mongo` | `mongodb://127.0.0.1:27017` |
| `redis` | `redis://127.0.0.1:6379` |
| `mysql` | `mysql://root:mysql@127.0.0.1:3306` |
| `mariadb` | `mysql://root:mariadb@127.0.0.1:3307` |
| `sqlserver` | `Server=127.0.0.1,1433`; só `amd64`; a imagem aceita a licença do SQL Server Developer |
| `rabbitmq` | `amqp://guest:guest@127.0.0.1:5672`; painel em `http://127.0.0.1:15672` |
| `kafka` | `127.0.0.1:9092` |

As senhas são fixas e servem só para desenvolvimento local.

## Sistema

```bash
romanctl system           # CPU, memória, disco, bateria e processos
romanctl system battery   # nível, watts, temperatura e previsão de autonomia
romanctl system watch     # mede por um tempo e estima o gasto dos seus processos
```

A autonomia vem da carga restante dividida pela corrente, e cada leitura entra em um histórico que dá a
queda real em %/h. No Linux a bateria vem de `/sys/class/power_supply`, no macOS de `pmset` e `ioreg` e
no Termux do `termux-battery-status`, que exige o app Termux:API.

Nenhum desses sistemas informa o consumo por processo a um programa comum. O `watch` calcula por regressão
quanto o consumo sobe com o uso de CPU e se recusa a dar um número quando os dados não sustentam.

## Depois de uma queda

Cada `start` registra o serviço em `~/.local/share/romanctl/estado.json` e cada `stop` tira.
Depois de uma queda ou de reiniciar a máquina, `romanctl restore` sobe de novo só o que
estava ligado. Projetos em `romanctl run` não voltam sozinhos, porque precisam de um terminal
próprio: o `restore` mostra o comando de cada um.

## Sem internet

`romanctl offline prep` baixa, com internet, templates de React e Next com `node_modules`,
caches de npm, Go e NuGet e extensões do code-server. Também configura o `GOPROXY` para cair no
cache local quando a rede falhar.

Sem internet, `romanctl new` usa os templates e os caches sozinho. Para adicionar pacotes:

```bash
npm install <pacote> --offline
go get <módulo>
dotnet add package <pacote>
```

`romanctl offline test` cria e compila cada tipo de projeto com a rede bloqueada por um proxy
morto. Esse teste não substitui o modo avião.

## Automação

`--json` funciona em `status`, `list`, `catalog`, `new`, `add`, `publish`, `tools`, `system`, `platform` e `config`, e as mensagens de progresso vão
para o stderr. Códigos de saída: 0 sucesso, 1 falha ao
executar, 2 uso incorreto. Sem terminal, os comandos nunca abrem menu: informe tudo por argumentos.

## Termux e proot

`romanctl` detecta `proot-distro` e muda de comportamento só nesse ambiente:

- **Containers não rodam**, porque o Android não libera namespaces e cgroups. `docker`, `podman`, `kind` e
  afins são recusados, com o motivo. `docker-cli` funciona com um Docker remoto.
- **Postgres, Mongo e Redis rodam como processo local.** Os demais serviços ficam indisponíveis.
- **PostgreSQL** precisa do shim `pgshim.c.txt`, carregado via `LD_PRELOAD`. Ele corrige o dono que o
  `stat()` reporta e emula a memória compartilhada SysV. É um contorno: faça backup do que for importante.
- **MongoDB** fica na série 7.0, porque as versões novas testadas falham ao alocar memória no kernel Android.
- **.NET** ganha um wrapper com limite de heap e o caminho do host para o `dotnet test`.
- **Os serviços param** quando a sessão do Ubuntu que os iniciou é encerrada (`--kill-on-exit`).

## Estrutura do código

| Pasta | Responsabilidade |
|---|---|
| `cmd/romanctl` | Ponto de entrada |
| `internal/cli` | Comandos e menus |
| `internal/tools` | Registro e instalação de ferramentas |
| `internal/services` | code-server, PostgreSQL, MongoDB, Redis e `restore` |
| `internal/projects` | Criar, listar, rodar e remover projetos simples |
| `internal/scaffold` | Templates de arquitetura, em `internal/scaffold/templates` |
| `internal/publish` | CI, Dockerfile e publicação no GitHub |
| `internal/offline` | Preparação e teste sem internet |
| `internal/manifest` | O `romanctl.yaml` de cada projeto |
| `internal/state` | O que deve estar ligado, para o `restore` |
| `internal/platform` | Detecta sistema, arquitetura, gerenciador de pacotes, root, WSL e proot |
| `internal/browser` | Abrir URLs em cada sistema |
| `internal/sys`, `internal/ui`, `internal/config` | Processos e rede, saída e prompts, caminhos |

```bash
go test ./...
```

O CI roda em Linux e macOS, nas duas arquiteturas, e em containers de Debian, Fedora, Arch, openSUSE e
Alpine. O `scripts/toolsweep.py` instala e remove todas as ferramentas; no CI, rode à mão com `sweep`.
Uma tag `v*` publica os binários e os checksums em uma release.
