# Persistência

Em memória: os dados de cada serviço somem quando ele para. O outbox do `orders` fica na mesma estrutura dos pedidos,
gravado sob a mesma trava, e a inbox do `billing` guarda os ids de eventos já processados ao lado das faturas.

Para usar PostgreSQL, gere o projeto com `--with postgres`.
