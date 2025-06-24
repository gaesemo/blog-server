package transaction

import (
	"context"
	"fmt"

	"github.com/gaesemo/tech-blog-server/gen/db/postgres"
	"github.com/jackc/pgx/v5"
)

// HACK: pgx, sqlc Queries에 매우 의존하고 있다.. 일단 ㄱ
type Transaction[R any] struct {
	db      *pgx.Conn
	queries *postgres.Queries
	opt     pgx.TxOptions
}

func New[R any](db *pgx.Conn, opt pgx.TxOptions, queries *postgres.Queries) *Transaction[R] {
	return &Transaction[R]{
		db:      db,
		opt:     opt,
		queries: queries,
	}
}

func (t *Transaction[R]) Exec(ctx context.Context, f func(c context.Context, q *postgres.Queries) (*R, error)) (*R, error) {
	tx, err := t.db.BeginTx(ctx, t.opt)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %v", err)
	}
	q := t.queries.WithTx(tx)
	result, err := f(ctx, q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return nil, fmt.Errorf("%v, rollback failed: %v", err, rbErr)
		}
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit failed: %v", err)
	}
	return result, nil
}
