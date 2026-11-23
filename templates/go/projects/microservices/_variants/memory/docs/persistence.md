# Persistence

In memory: each service's data is gone when it stops. The `orders` outbox lives in the same structure as the orders,
written under the same lock, and the `billing` inbox keeps the ids of events already processed next to the invoices.

To use PostgreSQL, generate the project with `--with postgres`.
