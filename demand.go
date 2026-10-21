package main

import (
	"fmt"
	"math"
	"net/http"
	"slices"
	"sort"
	"time"
)

type Festival struct {
	Name    string             `json:"name"`
	Start   string             `json:"start"` // "YYYY-MM-DD"
	End     string             `json:"end"`
	Effects map[string]float64 `json:"effects"` // category -> multiplier
}

type Season struct {
	Name      string             `json:"name"`
	Months    []int              `json:"months"`
	Effects   map[string]float64 `json:"effects"`
	LowSupply []string           `json:"low_supply"` // categories hard to get in this season
}

type BuySuggestion struct {
	ProductID int      `json:"product_id"`
	Name      string   `json:"name"`
	Unit      string   `json:"unit"`
	Stock     int      `json:"stock"`
	AvgDaily  float64  `json:"avg_daily"`
	CoverDays int      `json:"cover_days"`
	Expected  float64  `json:"expected"`
	Buy       int      `json:"buy"`
	Reasons   []string `json:"reasons"`
}

// Each effect is written once and reused for 2026 and 2027.
var ramadanEffects = map[string]float64{"Iftar": 4, "Oil": 1.5, "Sugar": 1.5}
var eidFitrEffects = map[string]float64{"Sweet": 4, "Sugar": 2, "Dairy": 2, "Spice": 1.5}
var boishakhEffects = map[string]float64{"Sweet": 1.5, "Dairy": 1.5}
var qurbaniEffects = map[string]float64{"Spice": 3, "Vegetable": 1.5, "Oil": 1.5}
var pujaEffects = map[string]float64{"Dairy": 2, "Sugar": 2, "Oil": 1.5}

// Eid and Ramadan depend on the moon, so these dates must be updated every year.
// The 2027 Eid/Ramadan dates are expected; fix them after the moon sighting.
// Order: name, start, end, effects
var festivals = []Festival{
	{"Ramadan 2026", "2026-02-19", "2026-03-20", ramadanEffects},
	{"Eid-ul-Fitr 2026", "2026-03-14", "2026-03-23", eidFitrEffects},
	{"Pohela Boishakh 2026", "2026-04-10", "2026-04-14", boishakhEffects},
	{"Qurbani Eid 2026", "2026-05-17", "2026-05-30", qurbaniEffects},
	{"Durga Puja 2026", "2026-10-14", "2026-10-21", pujaEffects},
	{"Ramadan 2027", "2027-02-09", "2027-03-10", ramadanEffects},
	{"Eid-ul-Fitr 2027", "2027-03-03", "2027-03-13", eidFitrEffects},
	{"Pohela Boishakh 2027", "2027-04-10", "2027-04-14", boishakhEffects},
	{"Qurbani Eid 2027", "2027-05-07", "2027-05-20", qurbaniEffects},
}

// Order: name, months, effects, low supply
var seasons = []Season{
	{"Summer", []int{3, 4, 5, 6}, map[string]float64{"Ice Cream": 2, "Drinks": 2}, []string{}},
	{"Monsoon", []int{7, 8, 9}, map[string]float64{"Ice Cream": 1.3, "Drinks": 1.3}, []string{"Vegetable"}},
	{"Winter", []int{12, 1, 2}, map[string]float64{"Ice Cream": 0.5, "Drinks": 0.6, "Tea": 1.5}, []string{}},
}

// getMultiplier returns how much the demand of a category changes on one day, and why.
// It starts at 1 (normal) and multiplies every festival and season that matches.
func getMultiplier(category string, day time.Time) (float64, []string) {
	multiplier := 1.0
	reasons := []string{}
	date := day.Format("2006-01-02")

	for _, f := range festivals {
		effect, ok := f.Effects[category]
		if ok && date >= f.Start && date <= f.End {
			multiplier = multiplier * effect
			reasons = append(reasons, fmt.Sprintf("%s x%v", f.Name, effect))
		}
	}

	month := int(day.Month())
	for _, s := range seasons {
		effect, ok := s.Effects[category]
		if ok && slices.Contains(s.Months, month) {
			multiplier = multiplier * effect
			reasons = append(reasons, fmt.Sprintf("%s x%v", s.Name, effect))
		}
	}
	return multiplier, reasons
}

// GET /api/occasions
func getOccasions(w http.ResponseWriter, r *http.Request) {
	now := today()
	limit := daysFromToday(60)
	month := int(time.Now().Month())

	active := []Festival{}
	upcoming := []Festival{}
	for _, f := range festivals {
		if now >= f.Start && now <= f.End {
			active = append(active, f)
		} else if f.Start > now && f.Start <= limit {
			upcoming = append(upcoming, f)
		}
	}

	activeSeasons := []Season{}
	for _, s := range seasons {
		if slices.Contains(s.Months, month) {
			activeSeasons = append(activeSeasons, s)
		}
	}

	writeJSON(w, map[string]any{
		"today":     now,
		"festivals": active,
		"seasons":   activeSeasons,
		"upcoming":  upcoming,
	})
}

// isLowSupply returns true if a season of that month has this category in its low supply list.
// Example: Vegetable in July to September (Monsoon).
func isLowSupply(category string, day time.Time) bool {
	month := int(day.Month())
	for _, s := range seasons {
		if slices.Contains(s.Months, month) && slices.Contains(s.LowSupply, category) {
			return true
		}
	}
	return false
}

// soldLast30Days returns how much of one product was sold in the last 30 full days.
// Today is not counted, because today is not finished yet and would make the average look low.
// COALESCE turns NULL (no sales at all) into 0.
func soldLast30Days(productID int) (int, error) {
	var sold int
	err := db.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM sales WHERE product_id = ? AND sale_date >= ? AND sale_date < ?",
		productID, daysFromToday(-30), today()).Scan(&sold)
	return sold, err
}

// makeBuySuggestion works out how much of one product to buy. Only math, no database.
func makeBuySuggestion(p Product, soldLast30 int) BuySuggestion {
	now := time.Now()
	avgDaily := float64(soldLast30) / 30

	// Buy enough for 7 days, or 14 days if the market has low supply of this category.
	coverDays := 7
	lowSupply := isLowSupply(p.Category, now)
	if lowSupply {
		coverDays = 14
	}

	// We add each day separately, so a festival that starts in a few days is counted too.
	expected := 0.0
	reasons := []string{}
	for i := 0; i < coverDays; i++ {
		day := now.AddDate(0, 0, i)
		multiplier, dayReasons := getMultiplier(p.Category, day)
		expected = expected + avgDaily*multiplier
		// The same festival hits many days, but we show its reason only once.
		for _, reason := range dayReasons {
			if !slices.Contains(reasons, reason) {
				reasons = append(reasons, reason)
			}
		}
	}
	if lowSupply {
		reasons = append(reasons, "Low market supply")
	}

	// Round first, so the number we show and the number we use are the same.
	expected = round2(expected)
	buy := int(math.Ceil(expected)) - p.Stock
	// After buying, the stock should not be below min stock.
	if p.Stock+buy < p.MinStock {
		buy = p.MinStock - p.Stock
	}
	// We already have enough.
	if buy < 0 {
		buy = 0
	}

	return BuySuggestion{
		ProductID: p.ID,
		Name:      p.Name,
		Unit:      p.Unit,
		Stock:     p.Stock,
		AvgDaily:  round2(avgDaily),
		CoverDays: coverDays,
		Expected:  expected,
		Buy:       buy,
		Reasons:   reasons,
	}
}

// GET /api/suggestions/buy
func buySuggestions(w http.ResponseWriter, r *http.Request) {
	// loadProducts closes its rows before it returns. We have only 1 connection,
	// so this lets us run soldLast30Days inside the loop.
	products, err := loadProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	suggestions := []BuySuggestion{}
	for _, p := range products {
		sold, err := soldLast30Days(p.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		suggestions = append(suggestions, makeBuySuggestion(p, sold))
	}

	// Biggest buy first.
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Buy > suggestions[j].Buy
	})
	writeJSON(w, suggestions)
}
