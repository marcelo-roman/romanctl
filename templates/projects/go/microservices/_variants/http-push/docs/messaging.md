# Mensageria

Sem broker: o relay do `orders` entrega cada evento com um `POST` em `/events` do `billing`, no endereço de
`BILLING_URL` (padrão `http://localhost:8092`). Um barramento dentro de um processo não atravessaria de um serviço
para o outro.

O outbox e a inbox continuam valendo: o evento só sai do outbox quando o `billing` responde 2xx, e enquanto o
`billing` estiver fora do ar o relay tenta de novo com espera crescente. O `billing` responde 204 quando processou,
503 numa falha passageira e 422 para um evento que nunca vai ser aceito.

Para trocar por RabbitMQ, gere o projeto com `--with rabbitmq`.
