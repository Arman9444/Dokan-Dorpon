package main

import (
	"net/http"
	"fmt"
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

func nearestExpiry(productID int)(string,error){
	var date string
	err:=db.QueryRow("SELECT COLESCE(MIN(expiry_date),'') FROM purchases WHERE product_id=? AND remaining>0 AND expiry_date>=?",
		 productID,today()).Scan(&date)

	return date,err
}

func pickDiscount(p Product,soldLast30 int,expiryDate string)(int,string) {
	daysLeft:=999
	if(expiryDate!=""){
		daysLeft=daysBetween(today(),expiryDate)
	}
	if(daysLeft<=3){
		return 30,fmt.Sprintf("Expires in %d days",daysLeft)
	}
	if(soldLast30==0){
		return 20,"No sale in the last 30 days"
	}
	if(daysLeft<=7){
		return 15,fmt.Sprintf("Expires in %d days",daysLeft)
	}

	daysOfStock:=float64(p.Stock)/(float64(soldLast30)/30)
	if(daysOfStock>60){
		return 10,fmt.Sprintf("Stock will last about %.0f days",daysOfStock)
	}
	return 0,""
}

func discountSuggestions(w http.ResponseWriter,r *http.Request){
	products,err:=loadProducts()
	if(err!=nil){
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}

	suggestions:=[]DiscountSuggestions{}
	for _,p:=range products{
		if(p.Stock==0){
			continue
		}

		sold,err:=soldLast30Days(p.ID)
		if(err!=nil){
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}

		expiry,err:=nearestExpiry(p.ID)
		if(err!=nil){
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		percent,reason:=pickDiscount(p,sold,expiry)

		if(percent>p.Discount){
			suggestions=append(suggestions,DiscoundSuggestions{
				productID:p.ID,
				Name:p.Name,
				Current:p.Discount,
				Suggested:percent,
				Reason:reason,
			})
		}
	}
	writeJSON(w,suggestions)
}