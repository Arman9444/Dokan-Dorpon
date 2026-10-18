package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// db is the global database connection, shared by all handlers.
// We use = not := because db is declared at package level, not inside a function.
var db *sql.DB

func openDatabase() {
	var err error
	// sql.Open does not actually connect; it just prepares the connection.
	// The real connection happens on the first query.
	db, err = sql.Open("sqlite", "grocery.db")
	if err != nil {
		log.Fatal(err)
	}

	// SQLite is a single file, so only one connection at a time is safe.
	db.SetMaxOpenConns(1)

	// Turn on foreign key checks. SQLite ignores foreign keys by default.
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal(err)
	}
}

func createTables() {
	// We run all three CREATE TABLE statements in one Exec call.
	// IF NOT EXISTS means this is safe to call every time the server starts.
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			name      TEXT    NOT NULL,
			category  TEXT    NOT NULL,
			unit      TEXT    NOT NULL,
			buy_price REAL    NOT NULL,
			sell_price REAL   NOT NULL,
			stock     INTEGER NOT NULL DEFAULT 0,
			min_stock INTEGER NOT NULL DEFAULT 0,
			discount  INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS purchases (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id    INTEGER NOT NULL REFERENCES products(id),
			quantity      INTEGER NOT NULL,
			remaining     INTEGER NOT NULL,
			buy_price     REAL    NOT NULL,
			expiry_date   TEXT    NOT NULL DEFAULT '',
			purchase_date TEXT    NOT NULL
		);

		CREATE TABLE IF NOT EXISTS sales (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id  INTEGER NOT NULL REFERENCES products(id),
			quantity    INTEGER NOT NULL,
			sell_price  REAL    NOT NULL,
			buy_price   REAL    NOT NULL,
			discount    INTEGER NOT NULL DEFAULT 0,
			sale_date   TEXT    NOT NULL
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
}
