package main

import (
	"fmt"
	"net/http"
	"slices"
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