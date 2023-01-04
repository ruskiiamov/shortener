package data

import (
	"context"
	"database/sql"
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

	err := db.QueryRowContext(ctx, "SELECT * FROM urls LIMIT 1;").Err()

	return err != nil
}

func createTable(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE urls (id serial PRIMARY KEY, url varchar, user_id uuid);")

	return err
}

func (d *dbKeeper) Add(u url.OriginalURL) (id string, err error) {
	if id, ok := d.getID(u); ok {
		return id, nil
	}

	err = d.db.QueryRow("INSERT INTO urls (url, user_id) VALUES ($1, $2) RETURNING id;", u.URL, u.UserID).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (d *dbKeeper) Get(id string) (*url.OriginalURL, error) {
	var u url.OriginalURL

	err := d.db.QueryRow("SELECT * FROM urls WHERE id=$1;", id).Scan(&u.ID, &u.URL, &u.UserID)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (d *dbKeeper) GetAllByUser(userID string) ([]url.OriginalURL, error) {
	urls := make([]url.OriginalURL, 0)

	rows, err := d.db.Query("SELECT * FROM urls WHERE user_id=$1;", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var u url.OriginalURL

		err = rows.Scan(&u.ID, &u.URL, &u.UserID)
		if err != nil {
			return nil, err
		}

		urls = append(urls, u)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return urls, nil
}

func (d *dbKeeper) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (d *dbKeeper) Close() {
	d.db.Close()
}

func (d *dbKeeper) getID(u url.OriginalURL) (string, bool) {
	var id string

	err := d.db.QueryRow("SELECT id FROM urls WHERE url=$1 AND user_id=$2;", u.URL, u.UserID).Scan(&id)
	if err != nil {
		return "", false
	}

	return id, true
}
