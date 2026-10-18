package main

import (
	"net/http"
	"strings"
)

type Product struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Unit       string  `json:"unit"`
	BuyPrice   float64 `json:"buy_price"`
	SellPrice  float64 `json:"sell_price"`
	Stock      int     `json:"stock"`
	MinStock   int     `json:"min_stock"`
	Discount   int     `json:"discount"`
	FinalPrice float64 `json:"final_price"`
}

type DiscountRequest struct {
	ID       int `json:"id"`
	Discount int `json:"discount"`
}

func loadProducts() ([]Product, error) {
	rows, err := db.Query("SELECT id, name, category, unit, buy_price, sell_price, stock, min_stock, discount FROM products ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Unit, &p.BuyPrice, &p.SellPrice, &p.Stock, &p.MinStock, &p.Discount)
		if err != nil {
			return nil, err
		}
		p.FinalPrice = priceAfterDiscount(p.SellPrice, p.Discount)
		products = append(products, p)
	}
	return products, rows.Err()
}

func listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := loadProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, products)
}

func readProduct(w http.ResponseWriter, r *http.Request) (Product, bool) {
	var p Product
	if !readJSON(w, r, &p) {
		return p, false
	}

	p.Name = strings.TrimSpace(p.Name)
	p.Category = strings.TrimSpace(p.Category)
	p.Unit = strings.TrimSpace(p.Unit)

	if p.Name == "" || p.Category == "" || p.Unit == "" {
		http.Error(w, "Name, category and unit are required", http.StatusBadRequest)
		return p, false
	}
	if p.BuyPrice <= 0 || p.SellPrice <= 0 {
		http.Error(w, "Prices must be more than 0", http.StatusBadRequest)
		return p, false
	}
	if p.SellPrice < p.BuyPrice {
		http.Error(w, "Sell price can not be less than buy price", http.StatusBadRequest)
		return p, false
	}
	if p.MinStock < 0 {
		http.Error(w, "Min stock can not be negative", http.StatusBadRequest)
		return p, false
	}
	return p, true
}

func addProduct(w http.ResponseWriter, r *http.Request) {
	p, ok := readProduct(w, r)
	if !ok {
		return
	}

	result, err := db.Exec("INSERT INTO products (name, category, unit, buy_price, sell_price, stock, min_stock, discount) VALUES (?, ?, ?, ?, ?, 0, ?, 0)",
		p.Name, p.Category, p.Unit, p.BuyPrice, p.SellPrice, p.MinStock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p.ID = int(id)
	p.Stock = 0
	p.Discount = 0
	p.FinalPrice = p.SellPrice
	writeJSON(w, p)
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	p, ok := readProduct(w, r)
	if !ok {
		return
	}

	result, err := db.Exec("UPDATE products SET name = ?, category = ?, unit = ?, buy_price = ?, sell_price = ?, min_stock = ? WHERE id = ?",
		p.Name, p.Category, p.Unit, p.BuyPrice, p.SellPrice, p.MinStock, p.ID)
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
	writeJSON(w, map[string]string{"message": "Product updated"})
}

func setDiscount(w http.ResponseWriter, r *http.Request) {
	var req DiscountRequest
	if !readJSON(w, r, &req) {
		return
	}
	if req.Discount < 0 || req.Discount > 90 {
		http.Error(w, "Discount must be between 0 and 90", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE products SET discount = ? WHERE id = ?", req.Discount, req.ID)
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
	writeJSON(w, map[string]string{"message": "Discount saved"})
}
