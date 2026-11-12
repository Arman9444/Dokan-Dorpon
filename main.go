package main

import (
	"fmt"
	"log"
	"net/http"
	"flag" //giving extra option to terminal (like go run . -today=2026-05-14)
	

)

func main() {
	//

	demoDate:=flag.String("today","","demo mode: pretent today is this date, like 2026-05-14")
	flag.Parse()//reading flag from terminal
	fileName:="grocery.db"
	if(*demoDate!=""){
		fileName=startDemo(*demoDate)
	}
	openDatabase(fileName)
	createTables()
	insertDefaultFestivals()
	insertSampleData()

	http.HandleFunc("GET /api/products", listProducts)
	http.HandleFunc("POST /api/products", addProduct)
	http.HandleFunc("POST /api/products/update", updateProduct)
	http.HandleFunc("POST /api/products/discount", setDiscount)

	http.HandleFunc("POST /api/sales", addSale)
	http.HandleFunc("GET /api/sales/recent", recentSales)
	http.HandleFunc("POST /api/purchases", addPurchase)
	http.HandleFunc("POST /api/purchases/remove-expired", removeExpired)


	http.HandleFunc("GET /api/festivals", listFestivals)
	http.HandleFunc("POST /api/festivals", addFestival)
	http.HandleFunc("POST /api/festivals/update", updateFestival)
	http.HandleFunc("POST /api/festivals/delete", deleteFestival)


	http.HandleFunc("GET /api/occasions", getOccasions)
	http.HandleFunc("GET /api/suggestions/buy", buySuggestions)

	// Serve everything inside the static/ folder.
	// http.FileServer returns a handler that reads files from the given directory.
	// A request for "/" automatically maps to "static/index.html".
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
