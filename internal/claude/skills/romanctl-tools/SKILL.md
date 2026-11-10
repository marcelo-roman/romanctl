---
name: romanctl-tools
description: Instala, remove, atualiza e consulta ferramentas de desenvolvimento pelo romanctl (kubectl, helm, terraform, node, go, dotnet, gh, docker CLI, clientes de banco, linters, CLIs de nuvem, cerca de 170). Use quando o usuário pedir para instalar um programa, quando um comando falhar por "command not found", ou para saber versão instalada, versão disponível, origem e README de uma ferramenta.
---

# Ferramentas pelo romanctl

Siga as regras gerais da skill `romanctl` (sem terminal, `--json`, `--yes`).

## Descobrir

```bash
romanctl tools list --plain --json                 # o que roda nesta máquina, com estado e versão
romanctl tools list --all --filter <texto> --json  # inclui o que não roda aqui, com o motivo
romanctl tools list --category k8s,db --json       # por categoria
romanctl tools categories --json                   # categorias e contagens
romanctl tools info <nome> --json                  # instalada, disponível, origem, página
romanctl tools info <nome>                         # o mesmo, com o README em markdown
romanctl tools versions <nome> --json              # versões instaláveis, da mais nova para a mais antiga
```

Categorias: `sdk`, `node-tools`, `python-tools`, `build`, `git`, `cloud`, `k8s`, `iac`,
`containers`, `security`, `db`, `observability`, `network`, `shell`, `ai`. Apelidos em
português também valem (`banco`, `nuvem`, `linguagens`...).

`tools list --json` traz `latest` quando a última versão foi consultada nas últimas 6 horas.
`tools info` responde `upToDate` e `upstream.latest`: é a versão que a instalação traria
agora, consultada na mesma origem que o instalador usa (GitHub releases, apt/dnf/pacman,
Homebrew, npm, PyPI ou o instalador oficial do projeto).

## Instalar

```bash
romanctl tools install <nome> [<nome>...] --yes --json
romanctl tools install <grupo> --yes --json     # o grupo inteiro
romanctl tools install <nome> --force --yes     # reinstala: é assim que se atualiza
romanctl tools install <nome> --version <versão> --yes --json   # uma versão específica, mesmo mais antiga
romanctl tools install <nomes> --dry-run        # o plano, sem instalar
```

### Versão específica

Quando o usuário pedir uma versão ("o terraform 1.9", "o node 20"), consulte
`tools versions <nome> --json` e escolha o `id` (ou o `label`) que bate; com versão parcial,
prefira a mais nova daquela linha e diga qual escolheu. `--version` vale para uma ferramenta
por vez, e trocar a versão de algo instalado já é reinstalar, sem `--force`.

Nem toda ferramenta permite escolher: com o Homebrew no macOS e com alguns instaladores
próprios (aws, gcloud, dart, swift, rust, docker-cli), `versions` responde que só existe a
última. Diga isso ao usuário em vez de procurar outro caminho.

- Sem `--yes`, um grupo abriria uma lista de escolha; sem terminal, `--yes` é obrigatório
  para grupos.
- O resultado JSON traz `results[]` com `action` (`installed`, `skipped`, `failed`,
  `unavailable`) e `version`. `skipped` quer dizer que já estava instalada.
- Depois de instalar, se o romanctl avisar para pôr uma pasta no PATH, repasse o aviso ao
  usuário; não edite o `.zshrc`/`.bashrc` sem pedir.
- Instalação baixa da internet e pode pedir `sudo` (pacotes do sistema). Se falhar por
  permissão, diga ao usuário e sugira `! romanctl tools install <nome>` para ele rodar.

## Remover

```bash
romanctl tools uninstall <nome> --yes --json
romanctl tools uninstall <nomes> --dry-run
```

`--yes` recusa `all` de propósito. Algumas ferramentas avisam o que deixa de funcionar
(`tools info` mostra em "Ao remover"); confirme com o usuário antes de remover SDKs.

## Quando a ferramenta não está no catálogo

`romanctl tools list --all --filter <nome> --json` vazio quer dizer que o romanctl não a
conhece. Diga isso ao usuário e ofereça instalar por fora desta vez ou acrescentá-la ao
catálogo do romanctl (`internal/tools/catalog_*.go` no repositório do romanctl).
