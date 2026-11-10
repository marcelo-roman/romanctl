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

Pronto. Depois, rode `romanctl` para ver a lista de comandos, ou `romanctl ui` para abrir o painel.

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

**1. Veja os comandos** e rode o que quiser pela lista:

```bash
romanctl         # lista de comandos: digite para filtrar, Enter roda e pergunta o que faltar
romanctl ui      # painel de tela cheia
```

No painel, `Tab` troca de tela, `/` filtra, `Enter` mostra o que dá para fazer com o item, e `?` mostra todas as teclas.

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
| `romanctl` | Lista os comandos: digite para filtrar, Enter roda e pergunta o que faltar |
| `ui` | Abre o painel |
| `tools list`, `tools info`, `tools install`, `tools uninstall`, `tools categories` | Ferramentas |
| `new`, `add`, `catalog`, `types` | Criar projetos e gerar código |
| `up`, `down`, `run`, `logs` | Rodar projetos com o ambiente |
| `code`, `open`, `list`, `rm` | Abrir, listar e remover projetos |
| `publish` | Publicar no GitHub |
| `status`, `start`, `stop`, `restart`, `install`, `restore` | Serviços |
| `system`, `platform`, `doctor` | CPU, memória, bateria e diagnóstico |
| `offline` | Preparar para trabalhar sem internet |
| `claude` | Skills para o Claude Code operar a máquina pelo romanctl |
| `config lang`, `self-update`, `menu` | Idioma, atualização e menu em lista |

Os projetos ficam em `~/projetos`; para usar outra pasta, defina `ROMANCTL_PROJETOS`. `--json` funciona nos comandos
de consulta, para scripts. Sem terminal, nada é perguntado. Os códigos de saída são
0 (sucesso), 1 (falha ao executar) e 2 (uso incorreto).

# Referência

## Painel

`romanctl ui` abre um painel de tela cheia, nos moldes do k9s, com telas de projetos, serviços, ferramentas
e sistema, que se atualizam sozinhas.

No alto ficam a máquina, a versão, quantos projetos e serviços estão no ar, CPU e memória; ao lado, as teclas que valem
agora. A lista vem numa moldura com o nome da tela, o filtro e a contagem no título, cada linha pintada pelo estado, e
embaixo a trilha de onde você está. Em terminal pequeno, o cabeçalho vira uma linha e as teclas descem para o rodapé.

| Tecla | O que faz |
|---|---|
| `Tab`, `1` a `4` | Troca de tela |
| `:` | Vai direto: `:tools`, `:svc`, `:prj`, `:sys`, ou uma categoria como `:db` e `:k8s` |
| `/` | Filtra a lista |
| `Enter` | Mostra as ações do item; em Ferramentas, as versões disponíveis |
| `d`, `l` | Detalhes e logs do item; em Ferramentas, os detalhes trazem o README, com `j`/`k` para rolar |
| `<` `>` | Troca de categoria, em Ferramentas |
| `?` | Mostra todas as teclas da tela atual |
| `Esc`, `q` | Volta e sai |

O cabeçalho mostra só as teclas que valem para o item selecionado. Ações longas rodam em segundo plano, com o andamento na
própria linha, e remover sempre pede confirmação.

Em Ferramentas, quando o cursor para numa linha, o painel consulta a versão que a instalação traria e mostra na coluna
VERSÃO, como `1.29 → 1.31`. Instalar e reinstalar pedem confirmação, com a versão e a origem. `romanctl menu` abre a versão em lista, mais simples.

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
romanctl tools info k9s             # instalada, disponível, origem, página e README
romanctl tools versions helm        # versões instaláveis, da mais nova para a mais antiga
romanctl tools install helm --version 3.18.4
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

A lista de ferramentas mostra sempre a última versão de cada uma, guardada por 6 horas para não ir à rede toda vez. No
painel, `Enter` numa ferramenta abre as versões disponíveis e instala a escolhida; em `romanctl tools list`, a opção é
"Escolher outra versão". As versões vêm da mesma origem do instalador (releases do GitHub, HashiCorp, npm, PyPI, `apt-cache
madison`, `dnf --showduplicates`, e as listas oficiais do Node, Go e Zig), sem pré-releases. Com o Homebrew e com os
instaladores próprios de aws, gcloud, dart, swift, rust e docker-cli, só existe a última.

`tools info` e a instalação mostram a versão que vai chegar, consultada na mesma origem que o instalador usa: releases do
GitHub, `apt-cache policy` e equivalentes, Homebrew, npm, PyPI ou o instalador oficial (dl.k8s.io, nodejs.org, go.dev,
HashiCorp). O README vem do GitHub, ou da descrição no PyPI, e é renderizado no terminal; fora de um terminal, sai o
markdown cru.

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
romanctl new estoque --stack go --arch hexagonal,ddd --with postgres
```

| Stack | Arquitetura | Base | Serviços |
|---|---|---|---|
| `dotnet` | `clean`, `ddd`, `cqrs` | Minimal API, EF Core, despachante próprio | `postgres` |
| `dotnet` | `clean`, `ddd`, `cqrs`, `event-sourcing` | Marten no PostgreSQL, projeções, concorrência otimista | `postgres` |
| `dotnet` | `modular-monolith`, `ddd` | Módulos Catalog e Ordering, contratos públicos, eventos de integração, um esquema por módulo | `postgres` |
| `dotnet` | `vertical-slice` | Minimal API, uma pasta por funcionalidade, EF Core | `postgres` |
| `go` | `hexagonal`, `ddd` | `net/http` da biblioteca padrão, pgx | `postgres` |
| `go` | `microservices`, `event-driven` | Pedidos e faturamento com outbox, inbox e RabbitMQ (ou HTTP sem ele) | `postgres`, `rabbitmq` |
| `kotlin` | `hexagonal`, `ddd` | Ktor, Gradle com wrapper, HikariCP, Konsist | `postgres` |
| `node` | `clean`, `ddd`, `cqrs` | NestJS com @nestjs/cqrs, TypeScript, pg, Vitest | `postgres` |
| `node` | `bff` | API de pedidos e BFF em Fastify, web em React; três processos | `postgres` |
| `node` | `layered` | Fastify, rotas, serviços e repositórios, pg | `postgres` |
| `node` | `trpc` | Monorepo com Next.js, tRPC, zod e pg | `postgres` |
| `java` | `hexagonal`, `ddd` | Spring Boot, Maven, JdbcClient, ArchUnit | `postgres` |
| `python` | `layered` | Django, Django REST Framework, uv | `postgres` |
| `react` | `feature-sliced` | Front da API de pedidos: Vite, React Router, TanStack Query, lint de camadas FSD | — |
| `rust` | `hexagonal`, `ddd` | axum, tokio, sqlx | `postgres` |
| `php` | `hexagonal`, `ddd` | Symfony 7.4, Doctrine DBAL, PHPUnit | `postgres` |
| `python` | `hexagonal`, `ddd` | FastAPI, uv, psycopg | `postgres` |

`romanctl catalog` lista os templates e as arquiteturas conhecidas: `bff`, `clean`, `cqrs`, `ddd`, `event-driven`,
`event-sourcing`, `feature-sliced`, `hexagonal`, `layered`, `microservices`, `modular-monolith`, `trpc` e
`vertical-slice`. Apelidos como `ports-and-adapters`, `vsa`, `modulith` e `fsd` também valem.

Depois de gerar, o `new` confere as ferramentas de que o projeto precisa e, em um terminal, oferece instalá-las. Com
tudo instalado, ele roda o preparo do template, como `dotnet restore` ou `go mod tidy`. Com `--offline`, o preparo
fica para depois. Uma falha no preparo não apaga o projeto.

### Onde ficam

Os templates não fazem parte do binário. Eles estão na pasta [`templates/`](templates) deste repositório, uma
subpasta por linguagem, e são baixados para `~/.local/share/romanctl/templates`, na tag da versão instalada:

```
templates/
  go/
    projects/hexagonal-ddd/          um projeto que compila e testa na própria pasta
    projects/microservices/
    generators/hexagonal-ddd/resource/
    generators/simple/handler/       geradores dos projetos Go simples
  node/
    projects/clean-ddd-cqrs/  projects/layered/  projects/trpc/  projects/bff/
    generators/express/  generators/next/  ...
  react/  dotnet/  java/  kotlin/  php/  python/  rust/
```

```bash
romanctl templates                 # onde estão e de que versão vieram
romanctl templates update          # baixa de novo, da tag da versão instalada
romanctl templates update --ref refs/heads/main
```

O instalador, o `self-update` e o `offline prep` baixam os templates sozinhos, e o primeiro `new` também, se ainda não
houver nenhum. Para desenvolver um template, aponte `ROMANCTL_TEMPLATES` para a pasta do repositório.

### Como um template é feito

Cada template é um projeto de verdade, que compila e passa nos testes dentro da própria pasta, com o nome de exemplo
`sample-app`. Gerar é copiar a pasta e trocar esse nome, em todas as grafias, pelo do projeto:

| No template | `minha-loja` vira |
|---|---|
| `SampleApp` | `MinhaLoja` |
| `sampleApp` | `minhaLoja` |
| `sample_app` | `minha_loja` |
| `sample-app` | `minha-loja` |
| `SAMPLE_APP` | `MINHA_LOJA` |
| `Sample App` | `Minha Loja` |
| `sample app` | `minha loja` |
| `sampleapp` | `minhaloja`, como em `com.example.sampleapp` |

A troca vale para o conteúdo e para os caminhos, como `src/SampleApp.Api`. A pasta do template é a variante com todos
os serviços. O que muda quando um serviço fica de fora vai em `_variants/`, que nem o Go nem o .NET compilam:

```yaml
id: hexagonal-ddd
stack: go
arch: [ddd, hexagonal]
description: API Go com arquitetura hexagonal e DDD, com HTTP da biblioteca padrão
services: [postgres]                  # o que --with aceita
requires: [go]                        # ferramentas do romanctl tools
setup: [go mod tidy]                  # roda na pasta do projeto, depois de gerar
run: go run ./cmd/api
test: go test ./...
url: http://localhost:8090
env:                                  # variáveis extras que o up grava no .env, por serviço
  postgres:
    PGHOST: "{{.Host}}"
publish:                              # CI e Dockerfile do romanctl publish
  kind: go
  vars: {Main: ./cmd/api}
conditional:                          # pastas que só existem com ou sem um serviço
  - path: internal/adapters/postgres
    when: postgres
overlays:                             # arquivos que substituem os da raiz quando o serviço fica de fora
  - from: _variants/memory
    when: "!postgres"
```

Um projeto com vários programas, como um monorepo com API e front, declara `processes` no lugar de `run`:

```yaml
processes:
  - name: api
    dir: apps/api
    run: npm run dev
    url: http://localhost:3000
  - name: web
    dir: apps/web
    run: npm run dev
    url: http://localhost:5173
```

O `run` e o `up` sobem todos juntos, com o nome do processo no começo de cada linha do log, e o `up -d` espera cada
`url` responder. Se um processo termina, os outros são parados e o comando falha na hora, com o fim do log. O `down`
para todos. Cada processo pode ter variáveis próprias em `env:`, além das do `.env`, como `PORT` diferente para cada um.

Com `processes`, o `run` do topo é opcional.

`run`, `test`, `publish.vars` e os campos `dir` e `run` dos processos também passam pela troca de nomes. Os valores de `env` são montados no `up`, com o
endereço do serviço: `{{.Host}}`, `{{.Port}}`, `{{.User}}`, `{{.Password}}`, `{{.Database}}` e `{{.URL}}`. Além
deles, todo projeto recebe as variáveis padrão de cada serviço, como `DATABASE_URL`.

Uma pasta em `~/.config/romanctl/templates/<stack>/<id>/` tem prioridade sobre os templates do romanctl. Cada projeto
gerado ganha um `romanctl.yaml` com o que o `run`, o `up`, o `add` e o `publish` usam.

Os testes do romanctl geram cada template com todas as combinações de serviços e falham se sobrar um nome de exemplo.
O CI roda os testes de cada template na própria pasta e também os do projeto gerado, contra um PostgreSQL de verdade.

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
| React feature-sliced | `entity`, `feature`, `page` |
| .NET event sourcing | `resource`, `domain`, `application`, `controller`, `test`, `event` |
| .NET monólito modular | `module`, `resource` (com `--in <módulo>`) |
| .NET Vertical Slice | `resource` |
| Go hexagonal e DDD | `resource` |
| Go microsserviços | `resource` (no serviço orders) |
| Java hexagonal e DDD | `resource` |
| Kotlin hexagonal e DDD | `resource` |
| Node Clean, DDD e CQRS | `resource` |
| Node BFF | `resource` |
| Node em camadas | `resource` |
| Node tRPC | `resource` |
| PHP hexagonal e DDD | `resource` |
| Python hexagonal e DDD | `resource` |
| Python em camadas (Django) | `resource` |
| Rust hexagonal e DDD | `resource` |
| Go simples | `handler`, `service`, `model` |

Sem `--force`, nada é sobrescrito. O que precisa de registro em outro arquivo é inserido antes de um marcador,
como `// romanctl:endpoints`; sem o marcador, o comando mostra o que fazer à mão.

Cada gerador é uma pasta em `templates/<linguagem>/generators/<alvo>/<tipo>`, com um `generator.yaml` e os arquivos em
`files/`. O alvo é o id do template (`go/hexagonal-ddd`) ou o tipo de projeto simples (`node/express`, `node/next`,
`react/vite`, `go/simple`, `dotnet/simple`). Eles usam a entidade de exemplo `SampleItem` (e `SampleItems`, `sample-item`,
`sample_items`, `sample item`...), o agregado de `--in` `SampleParent` e o projeto `SampleApp`, trocados como nos templates. Um patch
com `when: postgres` só vale com o serviço, e `after:` roda comandos depois de gravar, como `gofmt -w`. Uma pasta em
`~/.config/romanctl/generators/<linguagem>/<alvo>/<tipo>` tem prioridade.

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

`romanctl status` e o painel mostram só o que está instalado: o Docker, quando existe, e os serviços cujo container
`romanctl-<serviço>` existe. Ter o Docker não instala nada, e nem uma imagem baixada por outro projeto: `romanctl install
<serviço>` baixa a imagem, se faltar, e cria o container parado. O resto aparece numa linha, ou com `romanctl status --all`.

O Docker aparece como serviço porque os demais dependem dele. Ligar e desligar o daemon usa o systemd (ou `service`) com
`sudo`; parar pede confirmação, porque derruba todos os containers da máquina, e `stop all` nunca o inclui. Com o Docker
Desktop, o daemon é ligado e desligado pelo próprio Docker Desktop.

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

`--json` funciona em `status`, `list`, `catalog`, `new`, `add`, `publish`, `tools`, `system`, `platform`, `claude` e `config`, e as mensagens de progresso vão
para o stderr. Códigos de saída: 0 sucesso, 1 falha ao
executar, 2 uso incorreto. Sem terminal, os comandos nunca abrem menu: informe tudo por argumentos.

## Claude Code

```bash
romanctl claude install      # grava as skills em ~/.claude/skills
romanctl claude status       # instalada, desatualizada, ausente ou alheia
romanctl claude uninstall
```

As skills vêm dentro do binário, em `internal/claude/skills`, e ensinam o Claude Code a fazer tudo pelo romanctl: uma
geral, com as regras de uso sem terminal, e uma por área (ferramentas, projetos, ambiente, geradores, serviços,
sistema, publicação e manutenção). O romanctl marca as pastas que cria com um arquivo `.romanctl` e só troca ou apaga
essas; uma skill com o mesmo nome feita à mão só é trocada com `--force`. Depois de um `self-update`, rode
`romanctl claude install` de novo.

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
| `internal/tools` | Registro e instalação de ferramentas, e a versão e o README de cada uma |
| `internal/tui` | O painel |
| `internal/claude` | As skills do Claude Code e a instalação delas |
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
