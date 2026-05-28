// This file stores logic for the apiConfig struct.
package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Boopitty/Chirpy/internal/database"
)

// Used to store server data.
type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	secret         string
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

// Delete all users from the database
// Only accessableto admins
func (cfg *apiConfig) resetHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the request is not from an admin, give a forbidden response
		platform := os.Getenv("PLATFORM")
		if platform != "dev" {
			errBody := struct {
				Body string `json:"body"`
			}{
				Body: "You are not an admin",
			}
			respondWithJson(w, http.StatusForbidden, errBody)
			return
		}

		// Reset the database
		err := cfg.dbQueries.Reset(r.Context())
		if err != nil {
			errBody := fmt.Sprintf("Problem executing 'reset' command: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
		}

		resp := struct {
			Body string `json:"body"`
		}{
			Body: "Database has been reset.",
		}
		respondWithJson(w, http.StatusOK, resp)
	}
}
