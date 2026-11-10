# Mensageria

RabbitMQ, com `github.com/rabbitmq/amqp091-go`, lido de `RABBITMQ_URL`. O `romanctl up sample-app` liga o RabbitMQ e
grava a variável no `.env`.

| Peça | Nome |
|---|---|
| Exchange | `sample-app.events`, do tipo `topic` e durável |
| Chave do evento | `orders.order-placed` |
| Fila do faturamento | `billing.order-placed`, durável |

O `orders` publica com confirmação do broker e com `mandatory`: enquanto nenhuma fila estiver ligada à chave, a
mensagem volta, o relay trata como falha e ela continua no outbox. Assim nada se perde se o `orders` subir antes do
`billing`.

O `billing` declara a fila, consome com confirmação manual e reconecta sozinho se a conexão cair. Um erro passageiro
devolve a mensagem para a fila; um evento que nunca vai ser aceito, como um JSON inválido, é descartado.

O teste em `internal/platform/messaging/rabbitmq` faz a ida e volta num RabbitMQ de verdade e é pulado sem
`RABBITMQ_URL`.
