---
name: romanctl
description: Regras para operar esta máquina pelo romanctl, a CLI que centraliza instalação de ferramentas, criação e execução de projetos (web, .NET, Go, Node, Python, Java, Kotlin, PHP, Rust), serviços em Docker (Postgres, Redis, Kafka...), geração de código, publicação no GitHub e monitoramento da máquina. Use antes de instalar qualquer programa, criar ou rodar um projeto, subir um banco, ver CPU/memória/bateria ou publicar um repositório.
---

# romanctl

Nesta máquina, o romanctl é o caminho padrão para o ambiente de desenvolvimento. Antes de
recorrer a `apt`, `brew`, `npm -g`, `docker run`, `dotnet new`, `npx create-*` ou scripts
avulsos, veja se o romanctl faz aquilo. Ele registra o que instalou, o que subiu e o que
cada projeto usa, e é isso que permite derrubar e replicar o ambiente depois.

## Como chamar

O Claude roda sem terminal interativo. Então:

- **Passe tudo por argumento.** Sem terminal, o romanctl não pergunta nada; se faltar algo,
  ele falha com `usage: true` e diz o que falta.
- **Use `--json` para ler resultado.** Erros também vêm em JSON: `{"ok": false, "error": "...", "usage": bool}`.
  Código de saída 2 é erro de uso (corrija os argumentos); 1 é falha da operação.
- **Use `--yes` quando o comando pedir confirmação** e você já tiver o aval do usuário.
  `--dry-run` mostra o plano sem executar; use antes de algo destrutivo ou grande.
- **O romanctl só roda com login.** Se um comando falhar com "só funciona com login" ou "sessão
  expirou", não tente contornar: o login abre o navegador, então peça ao usuário para rodar
  `! romanctl login`. `romanctl whoami --json` diz se há sessão. Para administrar o Keycloak
  (realms, usuários, API keys) com esse login, veja `romanctl help iam`.
- **Na dúvida, `romanctl help <comando> [subcomando]`.** A ajuda é a fonte da verdade sobre
  opções; esta skill é um mapa.

### Nunca rode estes, porque prendem o processo ou abrem tela cheia

| Não rode                               | Em vez disso                                            |
| -------------------------------------- | ------------------------------------------------------- |
| `romanctl` sem argumentos, `ui`, `menu` | o subcomando específico                                 |
| `romanctl run <projeto>`               | `romanctl up <projeto> -d --json`                       |
| `romanctl up <projeto>` sem `-d`       | `romanctl up <projeto> -d --json`                       |
| `romanctl logs <alvo>` (segue o log)   | `timeout 15 romanctl logs <alvo>` ou leia o arquivo de log |
| `romanctl system watch` sem `--seconds` | `romanctl system watch --seconds 60 --json`            |
| `romanctl code`, `romanctl open`       | só quando o usuário pedir para abrir algo no navegador  |
| `romanctl web-ui`                      | só quando o usuário pedir a interface web; prende o terminal, então rode em segundo plano |

## Mapa das skills

| Para                                                      | Skill                  |
| --------------------------------------------------------- | ---------------------- |
| instalar, remover, atualizar ou consultar ferramentas      | `romanctl-tools`       |
| criar, listar ou remover projetos; templates              | `romanctl-projects`    |
| subir e derrubar o ambiente de um projeto; logs do app    | `romanctl-dev`         |
| gerar código dentro de um projeto (resource, page...)     | `romanctl-add`         |
| o Docker, bancos, filas e code-server                     | `romanctl-services`    |
| CPU, memória, disco, bateria, processos, plataforma        | `romanctl-system`      |
| publicar no GitHub com CI e Dockerfile                    | `romanctl-publish`     |
| atualizar o romanctl, idioma, modo offline, estas skills  | `romanctl-maintenance` |

## Quando o romanctl não cobre o pedido

Hoje ele não tem template de mobile (Expo, Flutter, React Native), de desktop, nem de outras
linguagens além de .NET, Go, Node, React, Python, Java, Kotlin, PHP e Rust. Se o pedido cair fora
do que `romanctl catalog` e `romanctl types` mostram:

1. Diga ao usuário que o romanctl ainda não faz aquilo.
2. Ofereça fazer manualmente desta vez **ou** acrescentar o suporte ao próprio romanctl.
3. Não instale nem crie nada por fora sem dizer, porque o objetivo é que tudo passe por ele.

Para ferramentas, o mesmo vale: `romanctl tools list --all --filter <nome> --json` antes de
instalar por outro meio.

## Onde as coisas ficam

- Projetos: na pasta mostrada no fim de `romanctl help` (padrão `~/projetos`).
- Binários das ferramentas: `binDir` em `romanctl platform --json`.
- Estado do que está no ar: o romanctl guarda; consulte com `romanctl status --json` e
  `romanctl list --json`, sem ler arquivos internos.
