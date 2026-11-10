package postgres

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sample-app/internal/platform/messaging"
	platform "sample-app/internal/platform/postgres"
	"sample-app/internal/services/orders/application/orders"
	"sample-app/internal/services/orders/domain/order"
	"sample-app/internal/services/orders/domain/shared"
)

//go:embed schema.sql
var schema string

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return platform.Migrate(ctx, pool, schema)
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Save(ctx context.Context, o *order.Order, outbox []messaging.Envelope) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO orders.orders (id, customer_name, status, created_at) VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET customer_name = EXCLUDED.customer_name, status = EXCLUDED.status`,
			string(o.ID()), o.CustomerName(), string(o.Status()), o.CreatedAt()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM orders.order_items WHERE order_id = $1`, string(o.ID())); err != nil {
			return err
		}
		for i, item := range o.Items() {
			if _, err := tx.Exec(ctx, `INSERT INTO orders.order_items (id, order_id, position, product_name, quantity, unit_price, currency)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`, string(item.ID()), string(o.ID()), i, item.ProductName(), item.Quantity(),
				item.UnitPrice().Cents(), item.UnitPrice().Currency()); err != nil {
				return err
			}
		}
		for _, m := range outbox {
			if _, err := tx.Exec(ctx, `INSERT INTO orders.outbox (id, type, payload, occurred_at) VALUES ($1, $2, $3, $4)`,
				m.ID, m.Type, string(m.Payload), m.OccurredAt); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Get(ctx context.Context, id shared.ID) (*order.Order, error) {
	list, err := s.load(ctx, `SELECT id, customer_name, status, created_at FROM orders.orders WHERE id = $1`, string(id))
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, orders.ErrNotFound
	}
	return list[0], nil
}

func (s *Store) List(ctx context.Context, limit int) ([]*order.Order, error) {
	return s.load(ctx, `SELECT id, customer_name, status, created_at FROM orders.orders ORDER BY created_at DESC LIMIT $1`, limit)
}

func (s *Store) Pending(ctx context.Context, limit int) ([]messaging.Envelope, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, type, occurred_at, payload FROM orders.outbox
		WHERE sent_at IS NULL ORDER BY occurred_at, id LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (messaging.Envelope, error) {
		var e messaging.Envelope
		var payload []byte
		err := row.Scan(&e.ID, &e.Type, &e.OccurredAt, &payload)
		e.Payload = payload
		return e, err
	})
}

func (s *Store) MarkSent(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE orders.outbox SET sent_at = now() WHERE id = $1 AND sent_at IS NULL`, id)
	return err
}

type orderRow struct {
	id           string
	customerName string
	status       string
	createdAt    time.Time
}

func (s *Store) load(ctx context.Context, query string, args ...any) ([]*order.Order, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	heads, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (orderRow, error) {
		var h orderRow
		err := row.Scan(&h.id, &h.customerName, &h.status, &h.createdAt)
		return h, err
	})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(heads))
	for _, h := range heads {
		ids = append(ids, h.id)
	}
	items, err := s.items(ctx, ids)
	if err != nil {
		return nil, err
	}
	list := make([]*order.Order, 0, len(heads))
	for _, h := range heads {
		list = append(list, order.Restore(shared.ID(h.id), h.customerName, order.Status(h.status), h.createdAt, items[h.id]))
	}
	return list, nil
}

func (s *Store) items(ctx context.Context, orderIDs []string) (map[string][]order.Item, error) {
	out := map[string][]order.Item{}
	if len(orderIDs) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `SELECT order_id, id, product_name, quantity, unit_price, currency
		FROM orders.order_items WHERE order_id = ANY($1::uuid[]) ORDER BY position`, orderIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var orderID, id, product, currency string
		var quantity int
		var cents int64
		if err := rows.Scan(&orderID, &id, &product, &quantity, &cents, &currency); err != nil {
			return nil, err
		}
		price, err := order.NewMoney(cents, currency)
		if err != nil {
			return nil, errors.Join(errors.New("preço inválido gravado no pedido "+orderID), err)
		}
		out[orderID] = append(out[orderID], order.RestoreItem(shared.ID(id), product, quantity, price))
	}
	return out, rows.Err()
}
