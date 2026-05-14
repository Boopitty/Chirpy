package main

import (
	"log"
	"net/http"
)

func main() {
	// Create a server object with a mutex
	mux := http.NewServeMux()
	port := "8080"
	server := http.Server{
		Addr:    ":" + port, // this means the site will be run on a local server
		Handler: mux,
	}

	// serve the index.html file on the site on the home page
	var path http.Dir = "."
	filepathRoot := http.FileServer(path)
	mux.Handle("/", filepathRoot)

	// Run ListenAndServe to run the site.
	// The code is blocked from this point until the server is closed or craches.
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
