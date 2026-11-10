# SampleApp

API Kotlin com arquitetura hexagonal e DDD, com Ktor e Gradle, gerada pelo `romanctl`.

```bash
romanctl run sample-app
./gradlew test
```

Endereço: http://localhost:8081. A porta muda com a variável `PORT`. O `gradlew` baixa o Gradle na primeira vez; basta
um JDK 21.

## Camadas

As dependências apontam sempre para dentro. O teste com Konsist em `src/test/kotlin/com/example/sampleapp/architecture`
garante isso.

| Pacote | Conteúdo | Depende de |
|---|---|---|
| `domain` | Agregados, objetos de valor, eventos e regras de negócio | só a biblioteca padrão |
| `application` | Casos de uso e as portas que eles precisam (`OrderRepository`, `EventPublisher`) | domain |
| `adapters` | Implementações das portas: HTTP com Ktor, persistência e eventos | application, domain |
| `Application.kt` | Monta os casos de uso com os adaptadores e sobe o servidor Netty | todos |

O domínio e a aplicação não conhecem o Ktor nem o JDBC.

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

PostgreSQL com JDBC e HikariCP, lido de `DATABASE_URL` (`postgres://usuario:senha@host:5432/banco`). O
`romanctl up sample-app` cria o banco `sample_app` e grava essa variável no `.env`. As tabelas de
`src/main/resources/schema.sql` são criadas na inicialização com `CREATE TABLE IF NOT EXISTS`, que não evolui o
esquema: adote migrações, como Flyway, antes de ter dados que importam.

O teste do repositório só roda com `DATABASE_URL` definida; sem ela, fica marcado como ignorado. Os testes HTTP usam
os repositórios em memória.
