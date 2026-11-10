package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"sample-app/internal/application/orders"
	"sample-app/internal/domain/order"
	"sample-app/internal/domain/shared"
)

type OrderRepository struct {
	db *DB
}

func NewOrderRepository(db *DB) *OrderRepository { return &OrderRepository{db: db} }

func (r *OrderRepository) Save(ctx context.Context, o *order.Order) error {
	return pgx.BeginFunc(ctx, r.db.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO orders (id, customer_name, status, created_at) VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET customer_name = EXCLUDED.customer_name, status = EXCLUDED.status`,
			string(o.ID()), o.CustomerName(), string(o.Status()), o.CreatedAt()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, string(o.ID())); err != nil {
			return err
		}
		for i, item := range o.Items() {
			if _, err := tx.Exec(ctx, `INSERT INTO order_items (id, order_id, position, product_name, quantity, unit_price, currency)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`, string(item.ID()), string(o.ID()), i, item.ProductName(), item.Quantity(),
				item.UnitPrice().Cents(), item.UnitPrice().Currency()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *OrderRepository) Get(ctx context.Context, id shared.ID) (*order.Order, error) {
	list, err := r.load(ctx, `SELECT id, customer_name, status, created_at FROM orders WHERE id = $1`, string(id))
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, orders.ErrNotFound
	}
	return list[0], nil
}

func (r *OrderRepository) List(ctx context.Context, limit int) ([]*order.Order, error) {
	return r.load(ctx, `SELECT id, customer_name, status, created_at FROM orders ORDER BY created_at DESC LIMIT $1`, limit)
}

type orderRow struct {
	id           string
	customerName string
	status       string
	createdAt    time.Time
}

func (r *OrderRepository) load(ctx context.Context, query string, args ...any) ([]*order.Order, error) {
	rows, err := r.db.pool.Query(ctx, query, args...)
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
	items, err := r.items(ctx, ids)
	if err != nil {
		return nil, err
	}
	list := make([]*order.Order, 0, len(heads))
	for _, h := range heads {
		list = append(list, order.Restore(shared.ID(h.id), h.customerName, order.Status(h.status), h.createdAt, items[h.id]))
	}
	return list, nil
}

func (r *OrderRepository) items(ctx context.Context, orderIDs []string) (map[string][]order.Item, error) {
	out := map[string][]order.Item{}
	if len(orderIDs) == 0 {
		return out, nil
	}
	rows, err := r.db.pool.Query(ctx, `SELECT order_id, id, product_name, quantity, unit_price, currency
		FROM order_items WHERE order_id = ANY($1::uuid[]) ORDER BY position`, orderIDs)
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
			return nil, errors.Join(errors.New("invalid price stored for order "+orderID), err)
		}
		out[orderID] = append(out[orderID], order.RestoreItem(shared.ID(id), product, quantity, price))
	}
	return out, rows.Err()
}
