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

func loadExpiryAlerts()([]ExpiryAlert,error){
	rows,err:=db.Query(`
	SELECT name FROM products WHERE products.id=purchases.product_id),
	SELECT unit FROM products WHERE products.id=purchases.product_id),
	remaining,expiry_date FROM purchases WHERE 
	remaining>0 AND expiry_date!='' AND expiry_date<=?
	ORDER BY expiry_date`,daysFromToday(7))

	if(err!=nil){
		return nil,err
	}
	defer rows.Close()

	now:=today()
	alerts:=[]ExpiryAlert{}
	for rows.Next(){
		var a ExpiryAlert
		err:=rows.Scan(&a.BatchID,&a.Name,&a.Unit,&a.Remaining,&a.ExpiryDate)
		if(err!=nil){
			return nil,err
		}
		a.DaysLeft=daysBetween(now,a.ExpiryDate)
		alerts=append(alerts, a)
	}
	return alerts,rows.Err()

}