package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	openDatabase()
	createTables()

	// Serve everything inside the static/ folder.
	// http.FileServer returns a handler that reads files from the given directory.
	// A request for "/" automatically maps to "static/index.html".
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
