package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ruskiiamov/shortener/internal/url"
)

const pgx = "pgx"

type dbKeeper struct {
	db *sql.DB
}

func newDBKeeper(dsn string) (*dbKeeper, error) {
	db, err := sql.Open(pgx, dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if tableDoesntExist(ctx, db) {
		if err := createTable(ctx, db); err != nil {
			return nil, err
		}
	}

	return &dbKeeper{db: db}, nil
}

func tableDoesntExist(ctx context.Context, db *sql.DB) bool {
	err := db.QueryRowContext(ctx, "SELECT id FROM urls LIMIT 1;").Err()

	return err != nil
}

func createTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`CREATE TABLE urls (
			id serial PRIMARY KEY, 
			url varchar, 
			"user" varchar,
			deleted boolean DEFAULT FALSE
		);`,
	)
	if err != nil {
		return fmt.Errorf("cannot create db table: %w", err)
	}

	_, err = db.ExecContext(ctx, "CREATE UNIQUE INDEX url_idx ON urls (url);")
	if err != nil {
		return fmt.Errorf("cannot create index for url: %w", err)
	}

	return nil
}

func (d *dbKeeper) Add(ctx context.Context, userID, original string) (int, error) {
	var id int

	err := d.db.QueryRowContext(
		ctx,
		`INSERT INTO urls (url, "user") VALUES ($1, $2) ON CONFLICT (url) DO NOTHING RETURNING id;`,
		original,
		userID,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		err = d.db.QueryRowContext(ctx, `SELECT id FROM urls WHERE url=$1;`, original).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("cannot find url: %w", err)
		}
		return 0, url.NewErrURLDuplicate(id, original)
	}

	if err != nil {
		return 0, fmt.Errorf("cannot add url: %w", err)
	}

	return id, nil
}

func (d *dbKeeper) AddBatch(ctx context.Context, userID string, originals []string) (map[string]int, error) {
	added := make(map[string]int)

	tx, err := d.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("transaction error: %w", err)
	}
	defer tx.Rollback()

	insStmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO urls (url, "user") VALUES ($1, $2) ON CONFLICT (url) DO NOTHING RETURNING id;`,
	)
	if err != nil {
		return nil, fmt.Errorf("statement error: %w", err)
	}
	defer insStmt.Close()

	selStmt, err := tx.PrepareContext(ctx, `SELECT id FROM urls WHERE url=$1;`)
	if err != nil {
		return nil, fmt.Errorf("statement error: %w", err)
	}
	defer selStmt.Close()

	var id int

	for _, original := range originals {
		err = insStmt.QueryRowContext(ctx, original, userID).Scan(&id)

		if errors.Is(err, sql.ErrNoRows) {
			err = selStmt.QueryRowContext(ctx, original).Scan(&id)
			if err != nil {
				return nil, fmt.Errorf("cannot find url: %w", err)
			}
		}

		if err != nil {
			return nil, fmt.Errorf("cannot add url: %w", err)
		}

		added[original] = id
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("transaction commit error: %w", err)
	}

	return added, nil
}

func (d *dbKeeper) Get(ctx context.Context, id int) (string, error) {
	var original string
	var deleted bool

	err := d.db.QueryRowContext(ctx, "SELECT url, deleted FROM urls WHERE id=$1;", id).Scan(&original, &deleted)
	if err != nil {
		return "", fmt.Errorf("cannot find url: %w", err)
	}

	if deleted {
		return "", new(url.ErrURLDeleted)
	}

	return original, nil
}

func (d *dbKeeper) GetAllByUser(ctx context.Context, userID string) (map[string]int, error) {
	urls := make(map[string]int)

	rows, err := d.db.QueryContext(ctx, `SELECT id, url FROM urls WHERE "user" = $1 AND deleted = false;`, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot find urls: %w", err)
	}

	var id int
	var original string

	for rows.Next() {
		err = rows.Scan(&id, &original)
		if err != nil {
			return nil, fmt.Errorf("cannot scan values: %w", err)
		}

		urls[original] = id
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	return urls, nil
}

func (d *dbKeeper) DeleteBatch(ctx context.Context, batch map[string][]int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}
	defer tx.Rollback()

	updStmt, err := tx.PrepareContext(ctx, `UPDATE urls SET deleted = TRUE WHERE "user" = $1 AND id = ANY($2::int[]);`)
	if err != nil {
		return fmt.Errorf("statement error: %w", err)
	}
	defer updStmt.Close()

	for userID, IDs := range batch {
		_, err = updStmt.ExecContext(ctx, userID, IDs)
		if err != nil {
			return fmt.Errorf("update error: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("transaction commit error: %w", err)
	}

	return nil
}

func (d *dbKeeper) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (d *dbKeeper) Close(ctx context.Context) error {
	closed := make(chan error)

	go func() {
		closed <- d.db.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			close(closed)
			return ctx.Err()
		case err := <-closed:
			close(closed)
			if err != nil {
				return fmt.Errorf("cannot close DB: %w", err)
			}
			return nil
		}
	}
}
