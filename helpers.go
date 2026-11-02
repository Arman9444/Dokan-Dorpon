package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"time"
)

// returns current date->"YYYY-MM-DD".
// stored all the date as text in SQLite to compare
func today() string {
	return todayTime().Format("2006-01-02")
}

// returns the date n days from now.
// Negative n gives a past date
func daysFromToday(n int) string {
	return todayTime().AddDate(0, 0, n).Format("2006-01-02")
}

// parsed both strings, subtract, and convert nanoseconds to days.
func daysBetween(a, b string) int {
	ta, _ := time.Parse("2006-01-02", a)
	tb, _ := time.Parse("2006-01-02", b)
	return int(tb.Sub(ta).Hours() / 24)
}

// rounds a float to 2 decimal places.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// sell=100, discount=10 -> 90.00
func priceAfterDiscount(sell float64, discount int) float64 {
	return round2(sell * float64(100-discount) / 100)
}

// writeJSON sends any Go value as a JSON response.
// w.Header().Set must be called before w.WriteHeader (which Encode triggers).
func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("Could not send JSON:", err)
	}
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
