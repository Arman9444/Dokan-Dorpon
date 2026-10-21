package main

import (
	"database/sql"
	"log"
	"math"
	"math/rand"
	"time"

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

type SampleProduct struct {
	Name       string
	Category   string
	Unit       string
	BuyPrice   float64
	SellPrice  float64
	MinStock   int
	DailySale  float64
	Stock      int
	ExpiryDays int
}

var samples = []SampleProduct{
	{"Garam Masala", "Spice", "packet", 40, 55, 20, 4, 60, 180},
	{"Chili Powder", "Spice", "packet", 35, 50, 20, 5, 45, 180},
	{"Turmeric Powder", "Spice", "packet", 30, 45, 20, 4, 70, 180},
	{"Onion", "Vegetable", "kg", 60, 75, 30, 15, 80, 0},
	{"Potato", "Vegetable", "kg", 25, 32, 40, 20, 150, 0},
	{"Miniket Rice", "Rice", "kg", 62, 70, 50, 25, 300, 0},
	{"Soybean Oil", "Oil", "liter", 165, 180, 15, 8, 40, 365},
	{"Sugar", "Sugar", "kg", 120, 135, 15, 6, 20, 0},
	{"Semai", "Sweet", "packet", 45, 60, 10, 2, 80, 120},
	{"Khejur (Dates)", "Iftar", "kg", 300, 380, 5, 1, 12, 90},
	{"Chola (Chickpea)", "Iftar", "kg", 90, 110, 5, 2, 25, 180},
	{"Rooh Afza", "Iftar", "bottle", 170, 200, 3, 0.2, 18, 300},
	{"Milk", "Dairy", "liter", 80, 95, 20, 10, 15, 2},
	{"Ghee", "Dairy", "jar", 380, 450, 5, 1.5, 6, 200},
	{"Ice Cream Cup", "Ice Cream", "pcs", 25, 35, 20, 6, 500, 90},
	{"Cold Drinks", "Drinks", "bottle", 35, 45, 20, 8, 120, 150},
	{"Tea Leaves", "Tea", "packet", 110, 130, 10, 5, 40, 365},
	{"Bread", "Bakery", "pcs", 50, 60, 10, 8, 12, -1},
}

func insertSampleData() {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	if count > 0 {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	for _, s := range samples {
		err = insertSampleProduct(tx, s)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func insertSampleProduct(tx *sql.Tx, s SampleProduct) error {
	result, err := tx.Exec("INSERT INTO products (name, category, unit, buy_price, sell_price, stock, min_stock, discount) VALUES (?, ?, ?, ?, ?, ?, ?, 0)",
		s.Name, s.Category, s.Unit, s.BuyPrice, s.SellPrice, s.Stock, s.MinStock)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	expiry := ""
	if s.ExpiryDays != 0 {
		expiry = daysFromToday(s.ExpiryDays)
	}

	_, err = tx.Exec("INSERT INTO purchases (product_id, quantity, remaining, buy_price, expiry_date, purchase_date) VALUES (?, ?, ?, ?, ?, ?)",
		id, s.Stock, s.Stock, s.BuyPrice, expiry, daysFromToday(-5))
	if err != nil {
		return err
	}

	for i := 364; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i)
		multiplier, _ := getMultiplier(s.Category, day)
		random := 0.7 + rand.Float64()*0.6
		quantity := int(math.Round(s.DailySale * multiplier * random))
		if quantity == 0 {
			continue
		}

		_, err = tx.Exec("INSERT INTO sales (product_id, quantity, sell_price, buy_price, discount, sale_date) VALUES (?, ?, ?, ?, 0, ?)",
			id, quantity, s.SellPrice, s.BuyPrice, day.Format("2006-01-02"))
		if err != nil {
			return err
		}
	}
	return nil
}
