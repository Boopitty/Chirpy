package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Boopitty/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Could not load godotenv %v", err)
	}

	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("could not open db: %v", err)
	}

	secretKey := os.Getenv("SECRET")
	if secretKey == "" {
		log.Fatal("SECRET environment variable is not set")
	}

	var cfg apiConfig
	cfg.dbQueries = database.New(db)
	cfg.secret = secretKey

	// Create a server object with a mutex
	mux := http.NewServeMux() //Create a server mutex
	port := "8080"
	server := http.Server{
		Addr:    ":" + port, // this means the site will be run on a local server
		Handler: mux,
	}

	// serve the index.html file on the site on the home page
	var path http.Dir = "app"        // Root directory of the server
	handler := http.FileServer(path) // Returns an http.Handler. It is an interface.
	mux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(middlewareLog(handler))))

	// handler for the healthz file
	// This is a health check for the server to check if it's ready to recieve requests.
	h := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatalf("Write has failed: %s", err)
		}
	}

	// Handlers for multiple functions
	mux.HandleFunc("GET /api/healthz", h)
	mux.HandleFunc("GET /admin/metrics", cfg.writeHitsHandler())
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler())
	mux.HandleFunc("POST /api/chirps", cfg.createChirpHandler())
	mux.HandleFunc("POST /api/users", cfg.createUserHandler())
	mux.HandleFunc("GET /api/chirps", cfg.getChirpsHandler())
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirpHandler())
	mux.HandleFunc("POST /api/login", cfg.loginHandler())

	// Run ListenAndServe to run the site.
	// The code is blocked from this point until the server is closed or craches.
	log.Printf("Serving files from %s on port: %s\n", string(path), port)
	log.Fatal(server.ListenAndServe())
}
