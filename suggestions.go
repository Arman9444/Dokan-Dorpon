package main

import (
	"net/http"
)

type ExpiryAlert struct {
	BatchID    int    `json:"batch_id"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	Remaining  int    `json:"remaining"`
	ExpiryDate string `json:"expiry_date"`
	// negative means already expired
	DaysLeft   int    `json:"days_left"` 
	Expired    bool   `json:"expired"`
}

type DiscountSuggestion struct {
	ProductID int    `json:"product_id"`
	Name      string `json:"name"`
	Current   int    `json:"current"`
	Suggested int    `json:"suggested"`
	Reason    string `json:"reason"`
}

func getAlerts(w http.ResponseWriter,r *http.Request){
	products,err:=loadProducts()
	if(err!=nil){
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}

	lowStock:=[]Product{}
	for _,p:=range products{
		if(p.Stock<=p.MinStock){
			lowStock=append(lowStock,p)
		}
	}
	expiry,err:=loadExpiryAlerts()
	if(err!=nil){
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w,map[string]any{"low_stock":lowStock,"expiry":expiry,})
}