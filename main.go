package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Serve everything inside the static/ folder.
	// http.FileServer returns a handler that reads files from the given directory.
	// http.StripPrefix removes "/static/" from the URL before looking for the file,
	// so a request for "/" maps to "static/index.html".
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
