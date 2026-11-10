---
name: romanctl-services
description: Liga, desliga, reinicia, instala e acompanha os serviços locais do romanctl — o daemon do Docker, os bancos e filas em container (PostgreSQL, MongoDB, Redis, MySQL, MariaDB, SQL Server, RabbitMQ, Kafka) e o code-server. Use quando o usuário precisar de um banco ou fila local, de uma connection string, quiser saber o que está rodando, ou ligar e desligar o Docker.
---

# Serviços pelo romanctl

Siga as regras gerais da skill `romanctl`. Use estes comandos em vez de `docker run` ou
`docker compose` para bancos e filas de desenvolvimento.

Chaves: `docker`, `code`, `postgres`, `mongo`, `redis`, `mysql`, `mariadb`, `sqlserver`,
`rabbitmq`, `kafka`.

## Estado e conexão

```bash
romanctl status --json         # só o que está instalado nesta máquina
romanctl status --all --json   # inclui o que dá para instalar
```

Cada item traz `installed`, `running`, `url` (a connection string ou endereço, com usuário e
senha de desenvolvimento), `wanted` (o romanctl quer que esteja ligado), `engine` (é o
daemon do Docker) e `unsupported`. Use o `url` para configurar o projeto.

Um serviço em container só conta como instalado quando o container `romanctl-<serviço>`
existe. Ter o Docker, ou a imagem baixada por outro projeto, não basta: se o serviço não
aparece em `status --json`, instale antes de usar.

## Ligar e desligar

```bash
romanctl install <serviço>     # baixa a imagem, se faltar, e cria o container parado; uma vez
romanctl start <serviço>
romanctl stop <serviço>
romanctl restart <serviço>
romanctl start all             # todos os instalados
```

Sem argumento, esses comandos abririam uma lista: sempre informe o serviço ou `all`. `all`
nunca inclui o Docker.
Num projeto, prefira `romanctl up <projeto>` (skill `romanctl-dev`), que liga só o que o
projeto usa e sabe desligar depois.

`stop` de um serviço que um projeto no ar usa derruba aquele projeto por baixo; confira
`romanctl list --json` (`up.services`) antes e avise o usuário.

## O Docker

O daemon aparece como o serviço `docker`. `running: false` nele explica por que nenhum
container sobe; diga isso ao usuário antes de tentar outra coisa.

```bash
romanctl start docker              # systemd ou service, com sudo
romanctl stop docker --yes         # derruba TODOS os containers da máquina
```

- Ligar e desligar pede `sudo`. Sem terminal, falha com "precisa da senha do sudo": peça ao
  usuário para rodar `! romanctl start docker`.
- **Parar o Docker derruba todos os containers da máquina**, inclusive os de outros projetos
  que não são do romanctl. Só com pedido explícito, depois de avisar; o romanctl recusa sem
  `--yes` fora de um terminal.
- Com o Docker Desktop, o daemon é ligado e desligado pelo Docker Desktop.

## Logs

`romanctl logs <serviço>` segue o log e não termina:

```bash
timeout 15 romanctl logs postgres
```

## Depois de uma queda ou reboot

```bash
romanctl restore status     # o que estava salvo como ligado e o que caiu
romanctl restore            # sobe de novo; sem terminal, restaura direto
romanctl restore save       # grava o que está rodando agora como o estado
romanctl restore clear      # esquece o estado
```
