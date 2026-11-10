# SampleApp

API Java com arquitetura hexagonal e DDD, com Spring Boot e Maven, gerada pelo `romanctl`.

```bash
romanctl run sample-app
mvn test
```

Endereço: http://localhost:8080. A porta muda com a variável `PORT`.

## Camadas

As dependências apontam sempre para dentro. O teste com ArchUnit em `src/test/java/com/example/sampleapp/architecture`
garante isso.

| Pacote | Conteúdo | Depende de |
|---|---|---|
| `domain` | Agregados, objetos de valor, eventos e regras de negócio | só o JDK |
| `application` | Casos de uso e as portas que eles precisam (`OrderRepository`, `EventPublisher`) | domain |
| `adapters` | Implementações das portas: HTTP, persistência e eventos | application, domain, Spring |
| `config` | Monta os casos de uso com os adaptadores | todos |

O domínio e a aplicação não conhecem o Spring. Os adaptadores são achados pelo component scan, e os casos de uso
viram beans nas classes de `config`.

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
