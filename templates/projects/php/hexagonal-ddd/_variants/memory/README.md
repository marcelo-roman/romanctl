# SampleApp

API PHP com arquitetura hexagonal e DDD, com Symfony, gerada pelo `romanctl`.

```bash
romanctl run sample-app
composer test
```

Endereço: http://localhost:8002, servido por `php -S 127.0.0.1:8002 -t public`. Para outra porta, troque o `run` no
`romanctl.yaml`.

## Camadas

As dependências apontam sempre para dentro. O teste em `tests/Architecture` garante isso.

| Pasta | Conteúdo | Depende de |
|---|---|---|
| `src/Domain` | Agregados, objetos de valor, eventos e regras de negócio | só o PHP |
| `src/Application` | Casos de uso e as portas que eles precisam (`OrderRepository`, `EventPublisher`, `Clock`) | Domain |
| `src/Adapters` | Implementações das portas: HTTP, persistência, eventos e relógio | Application, Domain |
| `config/services.yaml` | Liga cada porta à sua implementação | todos |

Os controllers ficam em `src/Adapters/Http` e são registrados pelos atributos `#[Route]`.

## Endpoints

| Método | Caminho | O que faz |
|---|---|---|
| `POST` | `/orders` | Cria um pedido: `{"customerName": "Ana"}` |
| `POST` | `/orders/{id}/items` | Adiciona um item: `{"productName": "Caneta", "quantity": 2, "unitPrice": 3.5}` |
| `POST` | `/orders/{id}/place` | Realiza o pedido |
| `GET` | `/orders/{id}` | Um pedido |
| `GET` | `/orders?limit=50` | Os pedidos mais recentes, até 200 |
| `GET` | `/health` | Saúde da API |

Regras de negócio violadas respondem 422, pedidos que não existem, 404, e corpos inválidos, 400, sempre como
`application/problem+json`.

## Persistência

Em memória. Como o PHP começa cada requisição do zero, os repositórios de memória guardam uma cópia em
`var/cache/<ambiente>/memory`, que some quando o cache é limpo. Nos testes, nada é gravado. Para usar PostgreSQL, gere
o projeto com `--with postgres`.
