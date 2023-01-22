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

func newDBKeeper(dsn string) (url.DataKeeper, error) {
	db, err := sql.Open(pgx, dsn)
	if err != nil {
		return nil, err
	}

	if tableDoesntExist(db) {
		if err := createTable(db); err != nil {
			return nil, err
		}
	}

	return &dbKeeper{db: db}, nil
}

func tableDoesntExist(db *sql.DB) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, "SELECT id FROM urls LIMIT 1;").Err()

	return err != nil
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(
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

	_, err = db.Exec("CREATE UNIQUE INDEX url_idx ON urls (url);")
	if err != nil {
		return fmt.Errorf("cannot create index for url: %w", err)
	}

	return nil
}

func (d *dbKeeper) Add(userID, original string) (int, error) {
	var id int

	err := d.db.QueryRow(
		`INSERT INTO urls (url, "user") VALUES ($1, $2) ON CONFLICT (url) DO NOTHING RETURNING id;`,
		original,
		userID,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		err = d.db.QueryRow(`SELECT id FROM urls WHERE url=$1;`, original).Scan(&id)
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

func (d *dbKeeper) AddBatch(userID string, originals []string) (map[string]int, error) {
	added := make(map[string]int)

	tx, err := d.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("transaction error: %w", err)
	}
	defer tx.Rollback()

	insStmt, err := tx.Prepare(
		`INSERT INTO urls (url, "user") VALUES ($1, $2) ON CONFLICT (url) DO NOTHING RETURNING id;`,
	)
	if err != nil {
		return nil, fmt.Errorf("statement error: %w", err)
	}
	defer insStmt.Close()

	selStmt, err := tx.Prepare(`SELECT id FROM urls WHERE url=$1;`)
	if err != nil {
		return nil, fmt.Errorf("statement error: %w", err)
	}
	defer selStmt.Close()

	var id int

	for _, original := range originals {
		err = insStmt.QueryRow(original, userID).Scan(&id)

		if errors.Is(err, sql.ErrNoRows) {
			err = selStmt.QueryRow(original).Scan(&id)
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

func (d *dbKeeper) Get(id int) (string, error) {
	var original string
	var deleted bool

	err := d.db.QueryRow("SELECT url, deleted FROM urls WHERE id=$1;", id).Scan(&original, &deleted)
	if err != nil {
		return "", fmt.Errorf("cannot find url: %w", err)
	}

	if deleted {
		return "", new(url.ErrURLDeleted)
	}

	return original, nil
}

func (d *dbKeeper) GetAllByUser(userID string) (map[string]int, error) {
	urls := make(map[string]int)

	rows, err := d.db.Query(`SELECT id, url FROM urls WHERE "user" = $1 AND deleted = false;`, userID)
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

func (d *dbKeeper) DeleteBatch(userID string, IDs []int) error {
	_, err := d.db.Exec(
		`UPDATE urls SET deleted = true WHERE "user" = $1 AND id = ANY ($2);`,
		userID,
		IDs,
	)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	return nil
}

func (d *dbKeeper) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (d *dbKeeper) Close() error {
	return d.db.Close()
}
