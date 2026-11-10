---
name: romanctl-dev
description: Sobe e derruba o ambiente de desenvolvimento de um projeto pelo romanctl (serviços em Docker, banco, .env e o app em segundo plano), e lê os logs do app. Use para rodar, testar no navegador, reiniciar ou parar um projeto, ou quando o usuário disser "sobe", "roda", "derruba", "para o projeto" ou "o que está no ar".
---

# Ciclo de desenvolvimento pelo romanctl

Siga as regras gerais da skill `romanctl`. O projeto pode ser passado pelo nome; sem nome,
o romanctl usa o projeto da pasta atual.

## Subir

```bash
romanctl up <projeto> -d --json        # serviços + banco + .env + app em segundo plano
romanctl up <projeto> --no-run --json  # só o ambiente (serviços, banco, .env), sem o app
```

- **Sempre `-d` ou `--no-run`.** Sem eles, o comando só termina quando o app para.
- `-d` espera o app responder. O JSON traz `url`, `pid`, `log`, `services`,
  `startedNow` (o que foi ligado agora), `envAdded` (variáveis que entraram no `.env`) e
  `alreadyUp`.
- O `.env` só ganha o que falta; o que o usuário mudou fica. Não sobrescreva o `.env` à mão
  com valores do romanctl.
- Porta ocupada é recusada citando quem a usa: repasse ao usuário em vez de matar processos.
- Sem terminal, `up` não instala serviço sozinho: se falhar com "não está instalado — rode:
  romanctl install <serviço>", rode esse `install` e repita o `up`.

## Conferir o que está no ar

```bash
romanctl list --json      # projetos; os que estão no ar têm up.alive, up.url e up.services
romanctl status --json    # serviços: installed, running, url
```

Para testar o app, use a `url` devolvida pelo `up` (por exemplo com `curl`).

## Logs

`romanctl logs <projeto>` **segue o log e não termina**. Prefira ler o arquivo: o caminho
vem no campo `log` do `up -d --json`.

```bash
tail -n 100 <log>
timeout 15 romanctl logs <projeto>     # se precisar seguir por um tempo
```

## Derrubar

```bash
romanctl down <projeto> --json                  # app + serviços que só ele usava
romanctl down <projeto> --keep-services --json  # só o app
romanctl down --all --json                      # tudo que o romanctl subiu
```

O `down` não desliga serviço que já estava ligado antes do `up` nem serviço que outro
projeto usa: o JSON separa `servicesStopped` de `servicesKept`.

## Reiniciar o app depois de mudar código

```bash
romanctl down <projeto> --keep-services --json && romanctl up <projeto> -d --json
```
