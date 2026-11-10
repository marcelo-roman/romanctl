package postgres

import (
	"context"
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	platform "sample-app/internal/platform/postgres"
	"sample-app/internal/services/billing/application/invoices"
	"sample-app/internal/services/billing/domain/invoice"
	"sample-app/internal/services/billing/domain/shared"
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

func (s *Store) Issue(ctx context.Context, eventID string, i *invoice.Invoice) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `INSERT INTO billing.inbox (event_id, processed_at) VALUES ($1, now()) ON CONFLICT DO NOTHING`, eventID)
		if err != nil || tag.RowsAffected() == 0 {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO billing.invoices (id, order_id, amount, currency, status, issued_at)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (order_id) DO NOTHING`,
			string(i.ID()), string(i.OrderID()), i.Amount(), i.Currency(), string(i.Status()), i.IssuedAt())
		return err
	})
}

func (s *Store) ByOrder(ctx context.Context, orderID shared.ID) (*invoice.Invoice, error) {
	list, err := s.query(ctx, `SELECT id, order_id, amount, currency, status, issued_at FROM billing.invoices WHERE order_id = $1`, string(orderID))
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, invoices.ErrNotFound
	}
	return list[0], nil
}

func (s *Store) List(ctx context.Context, limit int) ([]*invoice.Invoice, error) {
	return s.query(ctx, `SELECT id, order_id, amount, currency, status, issued_at FROM billing.invoices ORDER BY issued_at DESC LIMIT $1`, limit)
}

func (s *Store) query(ctx context.Context, sql string, args ...any) ([]*invoice.Invoice, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (*invoice.Invoice, error) {
		var id, orderID, currency, status string
		var amount int64
		var issuedAt time.Time
		if err := row.Scan(&id, &orderID, &amount, &currency, &status, &issuedAt); err != nil {
			return nil, err
		}
		return invoice.Restore(shared.ID(id), shared.ID(orderID), amount, currency, invoice.Status(status), issuedAt), nil
	})
}
