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
	var path http.Dir = "app"
	handler := http.FileServer(path)
	mux.Handle("/app/", http.StripPrefix("/app", middlewareLog(handler)))

	// handler for the healthz file
	h := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatalf("Write has failed: %s", err)
		}
	}

	mux.HandleFunc("/healthz", h)

	// Run ListenAndServe to run the site.
	// The code is blocked from this point until the server is closed or craches.
	log.Printf("Serving files from %s on port: %s\n", handler, port)
	log.Fatal(server.ListenAndServe())
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
