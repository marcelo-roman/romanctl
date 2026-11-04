# romanctl

Utilitário de desenvolvimento para um tablet Android com Ubuntu em `proot-distro` (Termux, ARM64).
Sobe serviços sem Docker, cria e roda projetos e abre tudo no navegador.

## Instalação

```bash
go build -ldflags="-s -w" -o ~/.local/bin/romanctl .
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
romanctl list
romanctl run <nome> [--open]
romanctl rm <nome> [-y]

romanctl code [projeto|pasta]   # abre no code-server
romanctl open [alvo] [--app]    # abre no navegador: code, projeto ou URL
romanctl offline [status|prep|test]
romanctl doctor
```

Os projetos ficam em `~/projetos` (mude com `ROMANCTL_PROJETOS`).
Dados e logs dos serviços ficam em `~/.local/share/romanctl`.

## Serviços

| Serviço     | Endereço                                              |
|-------------|-------------------------------------------------------|
| code-server | `http://127.0.0.1:8080`                               |
| PostgreSQL  | `postgres://postgres:postgres@127.0.0.1:5432/postgres` |
| MongoDB     | `mongodb://127.0.0.1:27017`                           |
| Redis       | `redis://127.0.0.1:6379`                              |

## Depois de uma queda

Cada `start` registra o serviço em `~/.local/share/romanctl/estado.json` e cada `stop` tira.
Quando o Android mata o Termux ou o tablet reinicia, `romanctl restore` sobe de novo só o que
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

## Limitações do proot

- **Docker não roda**: o Android não libera namespaces e cgroups.
- **PostgreSQL** precisa do shim `pgshim.c.txt`, carregado via `LD_PRELOAD`. Ele corrige o dono
  que o `stat()` reporta e emula a memória compartilhada SysV. O banco usa `auth=trust` e escuta
  só em `127.0.0.1`. É um contorno para desenvolvimento: faça backup do que for importante.
- **MongoDB** fica na série 7.0, porque as versões mais novas testadas falham ao alocar memória
  no espaço de endereços do kernel Android.
- **Os serviços param** quando a sessão do Ubuntu que os iniciou é encerrada (`--kill-on-exit`
  do `proot-distro`).
