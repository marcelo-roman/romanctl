# SampleApp

API Node enxuta em camadas, com Fastify e TypeScript, gerada pelo `romanctl`.

```bash
romanctl run sample-app
npm test
```

Endereço: http://localhost:3001. A porta muda com a variável `PORT`.

| Script | O que faz |
|---|---|
| `npm start` | Compila para `dist/` e sobe a API, lendo o `.env` se existir |
| `npm test` | Confere os tipos e roda os testes com Vitest |
| `npm run build` | Só compila |

## Camadas

Uma arquitetura em camadas clássica, sem agregados nem CQRS: cada requisição desce de rotas para serviços e de
serviços para repositórios. O teste em `test/layering.test.ts` garante a direção.

| Pasta | Conteúdo | Depende de |
|---|---|---|
| `src/models` | Os tipos dos dados, como `Task` | nada |
| `src/repositories` | As interfaces dos repositórios e a implementação em `memory/` | models |
| `src/services` | As regras de negócio | models e as interfaces dos repositórios, só com `import type` |
| `src/routes` | Rotas do Fastify, schemas JSON e a tradução de erros para problem JSON | services, models |
| `src/app.ts` | `buildApp`, que monta as rotas sobre os repositórios recebidos; os testes o usam com os de memória | todas |
| `src/server.ts` | Escolhe os repositórios e sobe o servidor | todas |

As rotas só validam o formato com os schemas do Fastify e chamam o serviço. Toda regra fica no serviço.

## Endpoints

| Método | Caminho | O que faz |
|---|---|---|
| `POST` | `/tasks` | Cria uma tarefa: `{"title": "Comprar pão", "dueDate": "2026-12-31"}` |
| `GET` | `/tasks?done=false&limit=50` | As tarefas mais recentes; `limit` vale 50 se omitido e no máximo 200 |
| `GET` | `/tasks/{id}` | Uma tarefa |
| `PATCH` | `/tasks/{id}` | Altera `title`, `done` ou `dueDate` (`null` tira a data) |
| `DELETE` | `/tasks/{id}` | Remove a tarefa |
| `GET` | `/health` | Saúde da API |

## Regras

- O título é obrigatório, perde os espaços das pontas e tem no máximo 200 caracteres.
- Na criação, `dueDate` não pode estar no passado; a comparação usa a data de hoje em UTC.
- Concluir uma tarefa que já está concluída responde 409: o cliente fica sabendo que a mudança não aconteceu.

Os erros saem como `application/problem+json`, no formato `{"status", "title", "detail"}`: corpo ou parâmetro
inválido, 400; tarefa que não existe, 404; tarefa já concluída, 409; regra de negócio violada, 422.

## Persistência

Em memória: os dados somem quando a API para. Para usar PostgreSQL, gere o projeto com `--with postgres`.
