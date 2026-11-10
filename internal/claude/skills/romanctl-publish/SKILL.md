---
name: romanctl-publish
description: Publica um projeto no GitHub pelo romanctl — gera CI (GitHub Actions) e Dockerfile conforme o template, recusa arquivos sensíveis como .env, faz o commit e cria o repositório (privado por padrão) com o gh. Use quando o usuário pedir para publicar, subir para o GitHub ou criar o repositório de um projeto.
---

# Publicação pelo romanctl

Siga as regras gerais da skill `romanctl`. Publicar é uma ação externa: **só com pedido
explícito do usuário**, e mostre o plano antes.

## 1. Plano

```bash
romanctl publish <projeto> --dry-run --json
```

Mostra `name`, `public`, `remoteExists`, `remote` e os arquivos que seriam gerados
(`generated`). Repasse ao usuário: nome do repositório, se é público ou privado, e o que
entra de CI e Dockerfile.

## 2. Publicar

```bash
romanctl publish <projeto> --yes --json
romanctl publish <projeto> --yes --description "Loja de exemplo" --json
romanctl publish <projeto> --yes --public --json   # só se o usuário pedir público
romanctl publish <projeto> --yes --no-ci --json    # sem CI nem Dockerfile
```

- Sem terminal, `--yes` é obrigatório; ele também pula as perguntas de visibilidade, então o
  repositório sai **privado** a menos que você passe `--public`.
- O resultado traz `url` do repositório e `committed`.
- Se o romanctl recusar por causa de um arquivo sensível (`.env`, chaves), não force nem
  apague o arquivo: diga ao usuário qual é e sugira pô-lo no `.gitignore`.
- Precisa do `gh` autenticado. Se faltar: `romanctl tools install gh --yes` e peça ao
  usuário para rodar `! gh auth login`, que é interativo.
