package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	openDatabase()
	createTables()

	http.HandleFunc("GET /api/products", listProducts)
	http.HandleFunc("POST /api/products", addProduct)
	http.HandleFunc("POST /api/products/update", updateProduct)
	http.HandleFunc("POST /api/products/discount", setDiscount)

	http.HandleFunc("POST /api/sales", addSale)
	http.HandleFunc("GET /api/sales/recent", recentSales)
	http.HandleFunc("POST /api/purchases", addPurchase)
	http.HandleFunc("POST /api/purchases/remove-expired", removeExpired)

	// Serve everything inside the static/ folder.
	// http.FileServer returns a handler that reads files from the given directory.
	// A request for "/" automatically maps to "static/index.html".
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
