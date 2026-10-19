package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type SaleRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type PurchaseRequest struct {
	ProductID  int     `json:"product_id"`
	Quantity   int     `json:"quantity"`
	BuyPrice   float64 `json:"buy_price"`
	ExpiryDate string  `json:"expiry_date"` // "" = never expires
}

type RemoveRequest struct {
	BatchID int `json:"batch_id"`
}

type Sale struct {
	ID          int     `json:"id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	SellPrice   float64 `json:"sell_price"`
	Discount    int     `json:"discount"`
	Total       float64 `json:"total"`
	SaleDate    string  `json:"sale_date"`
}

// Batch is one purchase row. Used only inside FEFO.
type Batch struct {
	ID        int
	Remaining int
}

// POST /api/sales
func addSale(w http.ResponseWriter, r *http.Request) {
	var req SaleRequest
	if !readJSON(w, r, &req) {
		return
	}
	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be more than 0", http.StatusBadRequest)
		return
	}

	// A sale changes 3 tables, so everything runs in one transaction.
	// If we return before Commit, the deferred Rollback undoes every change.
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// We read the product inside the transaction, so nobody can change the stock
	// between our check and our update.
	var stock, discount int
	var sellPrice, buyPrice float64
	err = tx.QueryRow("SELECT stock, sell_price, buy_price, discount FROM products WHERE id = ?", req.ProductID).
		Scan(&stock, &sellPrice, &buyPrice, &discount)
	if err == sql.ErrNoRows {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if req.Quantity > stock {
		http.Error(w, fmt.Sprintf("Not enough stock. Only %d left", stock), http.StatusBadRequest)
		return
	}

	// Save today's prices inside the sale, so old profit stays correct if prices change later.
	price := priceAfterDiscount(sellPrice, discount)
	_, err = tx.Exec("INSERT INTO sales (product_id, quantity, sell_price, buy_price, discount, sale_date) VALUES (?, ?, ?, ?, ?, ?)",
		req.ProductID, req.Quantity, price, buyPrice, discount, today())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", req.Quantity, req.ProductID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = takeFromBatches(tx, req.ProductID, req.Quantity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"message": "Sale saved", "total": round2(float64(req.Quantity) * price)})
}

// loadBatches returns the batches that still have items, in FEFO order:
// the one that expires first comes first, batches with no expiry date go last,
// and on the same expiry date the older batch (smaller id) comes first.
func loadBatches(tx *sql.Tx, productID int) ([]Batch, error) {
	rows, err := tx.Query("SELECT id, remaining FROM purchases WHERE product_id = ? AND remaining > 0 ORDER BY expiry_date = '', expiry_date, id", productID)
	if err != nil {
		return nil, err
	}
	// rows are closed when this function returns, so the caller can run UPDATEs after it.
	defer rows.Close()

	batches := []Batch{}
	for rows.Next() {
		var b Batch
		err := rows.Scan(&b.ID, &b.Remaining)
		if err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, rows.Err()
}

// takeFromBatches takes the sold quantity out of the batches, first-expire-first-out.
func takeFromBatches(tx *sql.Tx, productID int, quantity int) error {
	batches, err := loadBatches(tx, productID)
	if err != nil {
		return err
	}

	need := quantity
	for _, b := range batches {
		if need == 0 {
			break
		}
		// Take the whole batch, or only what we still need, whichever is smaller.
		take := min(need, b.Remaining)
		_, err := tx.Exec("UPDATE purchases SET remaining = remaining - ? WHERE id = ?", take, b.ID)
		if err != nil {
			return err
		}
		need = need - take
	}

	// Should never happen (stock = sum of remaining), but if the data is wrong
	// this error cancels the whole sale instead of saving bad data.
	if need > 0 {
		return errors.New("batches do not have enough items")
	}
	return nil
}

// GET /api/sales/recent
func recentSales(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id,
		       (SELECT name FROM products WHERE products.id = sales.product_id),
		       quantity, sell_price, discount, sale_date
		FROM sales
		ORDER BY sale_date DESC, id DESC
		LIMIT 15`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	sales := []Sale{}
	for rows.Next() {
		var s Sale
		err := rows.Scan(&s.ID, &s.ProductName, &s.Quantity, &s.SellPrice, &s.Discount, &s.SaleDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.Total = round2(float64(s.Quantity) * s.SellPrice)
		sales = append(sales, s)
	}
	if rows.Err() != nil {
		http.Error(w, rows.Err().Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, sales)
}

// POST /api/purchases (stock in)
func addPurchase(w http.ResponseWriter, r *http.Request) {
	var req PurchaseRequest
	if !readJSON(w, r, &req) {
		return
	}
	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be more than 0", http.StatusBadRequest)
		return
	}
	if req.BuyPrice <= 0 {
		http.Error(w, "Buy price must be more than 0", http.StatusBadRequest)
		return
	}
	if req.ExpiryDate != "" {
		_, err := time.Parse("2006-01-02", req.ExpiryDate)
		if err != nil {
			http.Error(w, "Expiry date must look like 2026-12-31", http.StatusBadRequest)
			return
		}
		if req.ExpiryDate < today() {
			http.Error(w, "Expiry date is already past", http.StatusBadRequest)
			return
		}
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// The newest buy price becomes the product's buy price.
	result, err := tx.Exec("UPDATE products SET stock = stock + ?, buy_price = ? WHERE id = ?", req.Quantity, req.BuyPrice, req.ProductID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	count, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// A new batch starts with remaining = quantity.
	_, err = tx.Exec("INSERT INTO purchases (product_id, quantity, remaining, buy_price, expiry_date, purchase_date) VALUES (?, ?, ?, ?, ?, ?)",
		req.ProductID, req.Quantity, req.Quantity, req.BuyPrice, req.ExpiryDate, today())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"message": "Stock added"})
}

// POST /api/purchases/remove-expired
func removeExpired(w http.ResponseWriter, r *http.Request) {
	var req RemoveRequest
	if !readJSON(w, r, &req) {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var productID, remaining int
	var expiryDate string
	err = tx.QueryRow("SELECT product_id, remaining, expiry_date FROM purchases WHERE id = ?", req.BatchID).
		Scan(&productID, &remaining, &expiryDate)
	if err == sql.ErrNoRows {
		http.Error(w, "Batch not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Something that expires today can still be sold today.
	if expiryDate == "" || expiryDate >= today() {
		http.Error(w, "This batch is not expired", http.StatusBadRequest)
		return
	}
	if remaining == 0 {
		http.Error(w, "This batch is already empty", http.StatusBadRequest)
		return
	}

	_, err = tx.Exec("UPDATE purchases SET remaining = 0 WHERE id = ?", req.BatchID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", remaining, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"message": "Expired batch removed", "removed": remaining})
}