package main

import (
	"encoding/json"
	"math"
	"net/http"
	"time"
)

// today returns the current date as "YYYY-MM-DD".
// We store all dates as text in SQLite so we can compare them as strings.
func today() string {
	return time.Now().Format("2006-01-02")
}

// daysFromToday returns the date n days from now as "YYYY-MM-DD".
// Negative n gives a past date (e.g. -1 = yesterday).
func daysFromToday(n int) string {
	return time.Now().AddDate(0, 0, n).Format("2006-01-02")
}

// daysBetween returns how many days are between date a and date b.
// We parse both strings, subtract, and convert nanoseconds to days.
func daysBetween(a, b string) int {
	ta, _ := time.Parse("2006-01-02", a)
	tb, _ := time.Parse("2006-01-02", b)
	return int(tb.Sub(ta).Hours() / 24)
}

// round2 rounds a float to 2 decimal places.
// math.Round rounds to the nearest integer, so we multiply by 100 first.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// priceAfterDiscount returns the sell price after applying a discount percent.
// e.g. sell=100, discount=10 -> 90.00
func priceAfterDiscount(sell float64, discount int) float64 {
	return round2(sell * float64(100-discount) / 100)
}

// writeJSON sends any Go value as a JSON response.
// w.Header().Set must be called before w.WriteHeader (which Encode triggers).
func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// readJSON reads the request body and fills the target struct.
// Returns false and sends a 400 error if the body is not valid JSON.
func readJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	err := json.NewDecoder(r.Body).Decode(target)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}
