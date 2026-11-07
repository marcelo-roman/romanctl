# romanctl

CLI para quem desenvolve: instala ferramentas, cria projetos e traz utilitários de desenvolvimento.
Funciona em Linux e macOS, em `amd64` e `arm64`, e também em WSL e em Termux com `proot-distro`.

| Sistema | Instala por |
|---|---|
| Debian, Ubuntu | `apt` e binários oficiais |
| Fedora, RHEL | `dnf` e binários oficiais |
| Arch | `pacman` e binários oficiais |
| openSUSE | `zypper` e binários oficiais |
| Alpine | `apk` e binários oficiais; binários que exigem glibc não rodam |
| macOS | Homebrew |

Sem root, os binários vão para `~/.local/bin`, e só o gerenciador de pacotes usa `sudo`. Com root, vão
para `/usr/local/bin` e `/opt`. `romanctl platform` mostra o que foi detectado.

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

## Instalação

```bash
go build -ldflags="-s -w" -o ~/.local/bin/romanctl ./cmd/romanctl
```

`~/.local/bin` precisa estar no `PATH`.

## Uso

`romanctl` sem argumentos abre o menu interativo.
Esc volta um passo, ← volta em listas e Ctrl+C cancela.

```
romanctl status
romanctl start|stop|restart <code|postgres|mongo|redis|all>
romanctl restore [status|save|clear]  # sobe de novo o que estava ligado
romanctl logs <serviço>
romanctl install <serviço>

romanctl new [tipo] [nome]      # node, express, react, react-ts, next,
                                # dotnet-webapi, dotnet-console, dotnet-mvc,
                                # dotnet-blazor, go
romanctl new <nome> --stack <s> --arch <a,b> [--with <serviços>] [--dry-run]
romanctl catalog                # templates e combinações válidas
romanctl list
romanctl run <nome> [--open]
romanctl rm <nome> [-y]

romanctl code [projeto|pasta]   # abre no code-server
romanctl open [alvo] [--app]    # abre no navegador: code, projeto ou URL
romanctl offline [status|prep|test]
romanctl tools [list]
romanctl tools install [nomes|cloud|sdk|cli|all] [--force]
romanctl system [status|battery|watch]
romanctl doctor
```

Os projetos ficam em `~/projetos` (mude com `ROMANCTL_PROJETOS`).
Dados e logs dos serviços ficam em `~/.local/share/romanctl`.

## Templates de arquitetura

```bash
romanctl new pedidos --stack dotnet --arch clean,ddd,cqrs --with postgres
```

Cada template é uma pasta com um `template.yaml`: os embutidos ficam em `templates/`, e uma pasta em
`~/.config/romanctl/templates/<stack>/<id>/` tem prioridade sobre eles. Todo arquivo é processado
com `text/template`, e `__Name__` nos caminhos vira o nome do projeto em PascalCase. Cada projeto
gerado ganha um `romanctl.yaml`, que o `run` e o `open` usam.

## Automação

`--json` funciona em `status`, `list`, `catalog`, `new`, `add`, `publish`, `tools`, `system`, `platform` e `config`, e as mensagens de progresso vão
para o stderr. Códigos de saída: 0 sucesso, 1 falha ao
executar, 2 uso incorreto. Sem terminal, os comandos nunca abrem menu: informe tudo por argumentos.

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

## Publicar

```bash
romanctl publish                                   # repositório privado
romanctl publish loja --public --description "Loja"
romanctl publish --dry-run
```

Gera CI e Dockerfile conforme o projeto, sem sobrescrever os que existirem, faz o commit, cria o repositório com
o `gh` e envia. Se o repositório já existir, só envia as alterações. Arquivos como `.env` e chaves privadas
cancelam a publicação, assim como segredos encontrados pelo `gitleaks`, se ele estiver instalado.

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
| `minio` | `http://127.0.0.1:9000`; painel em `http://127.0.0.1:9001` |

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
