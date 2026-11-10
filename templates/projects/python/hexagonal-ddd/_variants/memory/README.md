# SampleApp

API Python com arquitetura hexagonal e DDD, com FastAPI e uv, gerada pelo `romanctl`.

```bash
romanctl run sample-app
uv run pytest
uv run ruff check
```

Endereço: http://localhost:8000. A porta muda com a variável `PORT`, e a documentação interativa fica em `/docs`.

## Camadas

As dependências apontam sempre para dentro. O teste em `tests/architecture` garante isso.

| Pasta | Conteúdo | Depende de |
|---|---|---|
| `src/sample_app/domain` | Agregados, objetos de valor, eventos e regras de negócio | só a biblioteca padrão |
| `src/sample_app/application` | Casos de uso e as portas que eles precisam (`OrderRepository`, `EventPublisher`) | domain |
| `src/sample_app/adapters` | Implementações das portas: HTTP, persistência e eventos | application, domain |
| `src/sample_app/bootstrap.py` | Monta os adaptadores e a aplicação FastAPI | todos |

## Endpoints

| Método | Caminho | O que faz |
|---|---|---|
| `POST` | `/orders` | Cria um pedido: `{"customerName": "Ana"}` |
| `POST` | `/orders/{id}/items` | Adiciona um item: `{"productName": "Caneta", "quantity": 2, "unitPrice": 3.5}` |
| `POST` | `/orders/{id}/place` | Realiza o pedido |
| `GET` | `/orders/{id}` | Um pedido |
| `GET` | `/orders?limit=50` | Os pedidos mais recentes |
| `GET` | `/health` | Saúde da API |

Regras de negócio violadas respondem 422, pedidos que não existem, 404, e corpos inválidos, 400, sempre como
`application/problem+json`.

## Persistência

Em memória: os dados somem quando a API para. Para usar PostgreSQL, gere o projeto com `--with postgres`.
