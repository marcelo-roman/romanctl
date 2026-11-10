package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sample-app/internal/services/orders/application/sampleitems"
	"sample-app/internal/services/orders/domain/sampleitem"
	"sample-app/internal/services/orders/domain/shared"
)

type SampleItemRepository struct {
	pool *pgxpool.Pool
}

func NewSampleItemRepository(pool *pgxpool.Pool) *SampleItemRepository {
	return &SampleItemRepository{pool: pool}
}

func (r *SampleItemRepository) Save(ctx context.Context, e *sampleitem.SampleItem) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO orders.sample_items (id, name, created_at) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`, string(e.ID()), e.Name(), e.CreatedAt())
	return err
}

func (r *SampleItemRepository) Get(ctx context.Context, id shared.ID) (*sampleitem.SampleItem, error) {
	list, err := r.query(ctx, `SELECT id, name, created_at FROM orders.sample_items WHERE id = $1`, string(id))
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, sampleitems.ErrNotFound
	}
	return list[0], nil
}

func (r *SampleItemRepository) List(ctx context.Context, limit int) ([]*sampleitem.SampleItem, error) {
	return r.query(ctx, `SELECT id, name, created_at FROM orders.sample_items ORDER BY created_at DESC LIMIT $1`, limit)
}

func (r *SampleItemRepository) Delete(ctx context.Context, id shared.ID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM orders.sample_items WHERE id = $1`, string(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return sampleitems.ErrNotFound
	}
	return nil
}

func (r *SampleItemRepository) query(ctx context.Context, sql string, args ...any) ([]*sampleitem.SampleItem, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (*sampleitem.SampleItem, error) {
		var id, name string
		var createdAt time.Time
		if err := row.Scan(&id, &name, &createdAt); err != nil {
			return nil, err
		}
		return sampleitem.Restore(shared.ID(id), name, createdAt), nil
	})
}
