// This file stores logic for the apiConfig struct.
package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/Boopitty/Chirpy/internal/database"
)

// Used to store server data.
type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

// Increment a counter to keep track of how many times the site has been visited.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Returns the number of hits that are stored in the config.
func (cfg *apiConfig) writeHitsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		body := fmt.Sprintf(
			`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(body))
	}
}

// Reset the hits count.
func (cfg *apiConfig) resetHitsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		cfg.fileserverHits.Store(0)
	}
}
