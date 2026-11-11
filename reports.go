package main

import (
	"net/http"
	"sort"
	"time"
)

type ReportItem struct {
	ProductID int     `json:"product_id"`
	Name      string  `json:"name"`
	Unit      string  `json:"unit"`
	Stock     int     `json:"stock"`
	Sold      int     `json:"sold_qty"`
	Income    float64 `json:"income"`
	Cost      float64 `json:"cost"`
	Profit    float64 `json:"profit"`
}

func findType(period string) string {
	_, err := time.Parse("2006-01-02", period)
	if err == nil {
		return "daily"
	}
	_, err = time.Parse("2006-01", period)
	if err == nil {
		return "monthly"
	}
	_, err = time.Parse("2006", period)
	if err == nil {
		return "yearly"
	}
	return ""
}

func loadReport(period string) ([]ReportItem, error) {
	rows, err := db.Query(`
		SELECT id, name, unit, stock,
		       (SELECT COALESCE(SUM(quantity), 0) FROM sales
		          WHERE sales.product_id = products.id AND sale_date LIKE ? || '%'),
		       (SELECT COALESCE(SUM(quantity * sell_price), 0) FROM sales
		          WHERE sales.product_id = products.id AND sale_date LIKE ? || '%'),
		       (SELECT COALESCE(SUM(quantity * buy_price), 0) FROM sales
		          WHERE sales.product_id = products.id AND sale_date LIKE ? || '%')
		FROM products`, period, period, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []ReportItem{}
	for rows.Next() {
		var p ReportItem
		err := rows.Scan(&p.ProductID, &p.Name, &p.Unit, &p.Stock, &p.Sold, &p.Income, &p.Cost)
		if err != nil {
			return nil, err
		}
		p.Income = round2(p.Income)
		p.Cost = round2(p.Cost)
		p.Profit = round2(p.Income - p.Cost)
		list = append(list, p)
	}
	return list, rows.Err()
}

func getReport(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	reportType := findType(period)
	if reportType == "" {
		http.Error(w, "Period must look like 2026, 2026-10 or 2026-10-15", http.StatusBadRequest)
		return
	}

	list, err := loadReport(period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Profit > list[j].Profit
	})

	totalSold := 0
	totalIncome := 0.0
	totalCost := 0.0
	for _, p := range list {
		totalSold = totalSold + p.Sold
		totalIncome = totalIncome + p.Income
		totalCost = totalCost + p.Cost
	}

	best := ""
	if len(list) > 0 && list[0].Profit > 0 {
		best = list[0].Name
	}

	writeJSON(w, map[string]any{
		"period":       period,
		"kind":         reportType,
		"rows":         list,
		"total_qty":    totalSold,
		"total_income": round2(totalIncome),
		"total_cost":   round2(totalCost),
		"total_profit": round2(totalIncome - totalCost),
		"best_product": best,
	})
}
