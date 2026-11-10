---
name: romanctl-maintenance
description: Mantém o próprio romanctl — atualizar para a versão nova (self-update), trocar o idioma, preparar a máquina para trabalhar sem internet (offline prep/test) e instalar ou atualizar estas skills do Claude. Use quando o usuário pedir para atualizar o romanctl, mudar o idioma, trabalhar offline, ou quando as skills estiverem desatualizadas.
---

# Manutenção do romanctl

Siga as regras gerais da skill `romanctl`.

## Versão e atualização

```bash
romanctl version
romanctl self-update --check --json        # current, latest, updateAvailable
romanctl self-update --yes --json          # atualiza; confere o SHA-256 e testa antes de trocar
romanctl self-update --version v0.1.0 --yes --json   # volta para uma versão específica
```

Depois de atualizar, rode `romanctl claude install` para as skills acompanharem a versão nova.

## Estas skills

```bash
romanctl claude status --json      # instalada, desatualizada, ausente ou alheia, por skill
romanctl claude install --json     # instala ou atualiza
romanctl claude uninstall --json   # remove só as que o romanctl instalou
```

O romanctl só mexe em pastas que ele mesmo criou (têm o arquivo `.romanctl`). Uma pasta com o
mesmo nome criada à mão aparece como `alheia` e só é sobrescrita com `--force`; pergunte ao
usuário antes.

## Idioma

```bash
romanctl config lang pt    # ou en, ou auto para seguir o sistema
```

Para uma única chamada: `--lang en`. As mensagens mudam; as chaves do JSON, não.

## Sem internet

```bash
romanctl offline status
romanctl offline prep [--force]   # baixa templates, caches e extensões; precisa de internet
romanctl templates update --json  # baixa de novo os templates de projeto e os geradores
romanctl offline test             # cria e compila cada tipo de projeto com a rede bloqueada
```

`offline prep` e `offline test` demoram; avise e rode em segundo plano. Com a preparação
feita, `romanctl new <tipo> <nome> --offline` cria projetos sem rede.
