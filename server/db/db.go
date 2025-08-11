package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	connStr := "postgres://postgres:vandan@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Create tables
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		sender_id INTEGER REFERENCES users(id),
		recipient_id INTEGER REFERENCES users(id),
		content TEXT NOT NULL,
		created_at BIGINT NOT NULL
	);
`)
	if err != nil {
		return nil, err
	}

	return db, nil

}
