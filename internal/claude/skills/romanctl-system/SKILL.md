---
name: romanctl-system
description: Monitora a máquina pelo romanctl — CPU, memória, swap, disco, bateria (nível, watts, autonomia), processos que mais consomem, consumo estimado em watts, sistema, arquitetura e capacidades detectadas (WSL, containers, virtualização). Use quando o usuário perguntar se a máquina está pesada, o que está consumindo, quanto de bateria resta, ou antes de instalar algo grande.
---

# Monitoramento pelo romanctl

Siga as regras gerais da skill `romanctl`. Leia sempre com `--json` e resuma para o
usuário em frases, não despeje o JSON.

## Retrato agora

```bash
romanctl system status --json
```

Traz `platform`, `cores`, `systemCpuPercent` (nulo quando a máquina não informa; aí use
`visibleProcessesCores`), `memory` (`totalBytes`, `availableBytes`, `swapTotalBytes`,
`swapFreeBytes`), `disk` (`path`, `sizeBytes`, `freeBytes`), `battery` (quando existe) e
`top` (processos com `pid`, `name`, `cpuPercentOfOneCore`, `rssBytes`).

## Bateria

```bash
romanctl system battery --json
```

Nível, consumo em watts, temperatura e previsão de autonomia, quando o sistema informa.

## Medir por um período

```bash
romanctl system watch --seconds 60 --interval 5 --json
```

**Sempre com `--seconds`**: sem ele, o comando pergunta a duração. Mede consumo e CPU juntos
e estima quantos watts os processos do usuário gastam (`fit.estimatedW`, `fit.meanPowerW`). Para medições longas (minutos), rode
em segundo plano e avise o usuário quanto tempo vai levar.

## A máquina

```bash
romanctl platform --json   # os, arch, wsl, proot, root, manager, binDir, containers, virtualization, gui
romanctl doctor            # versões das ferramentas principais
```

Use `platform` antes de sugerir algo que dependa do sistema: `containers: false` quer dizer
que os serviços em Docker não sobem; `manager` diz qual gerenciador de pacotes existe.

## Ao interpretar

- Memória: o que importa é `availableBytes`.
- `cpuPercentOfOneCore` passa de 100% em processos com várias threads; divida por `cores`
  para a fatia da máquina.
- Disco abaixo de ~10% livre atrapalha imagens Docker e builds; avise.
- Um processo do próprio Claude ou do code-server no topo é normal; diga o nome antes de
  sugerir encerrar qualquer coisa, e não encerre processos sem pedido.
