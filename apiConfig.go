// This file stores logic for the apiConfig struct.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/Boopitty/Chirpy/internal/auth"
	"github.com/Boopitty/Chirpy/internal/database"
	"github.com/google/uuid"
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

// Accepts a json request containing an email address,
// adds the user to the database,
// and returns a json response of the new user info.
func (cfg *apiConfig) createUserHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode response body into the new req struct
		req := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			errBody := fmt.Sprintf("Decoding Error: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		now := time.Now()
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			errBody := fmt.Sprintf("Error hashing password: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
		}

		// Create user in database
		user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
			ID:             uuid.New(),
			CreatedAt:      now,
			UpdatedAt:      now,
			Email:          req.Email,
			HashedPassword: hash,
		})
		if err != nil {
			errBody := fmt.Sprintf("Database Error: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Respond with user info from database
		respondWithJson(w, http.StatusCreated, user)
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

// Creates a chirp
func (cfg *apiConfig) createChirpHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode the request
		req := struct {
			Body string    `json:"body"`
			User uuid.UUID `json:"user_id"`
		}{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			errBody := fmt.Sprintf("Decoding Error: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// If length of request is too long, give an error
		if len(req.Body) > 140 {
			type errResp struct {
				Error string `json:"error"`
			}
			respondWithJson(w, http.StatusBadRequest, errResp{Error: "Chirp is too long"})
			return
		}

		// Create the new chirp using the request
		chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Body:      req.Body,
			UserID:    req.User,
		})
		if err != nil {
			errBody := fmt.Sprintf("Problem creating chirp: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
		}

		// Respond with chirp if all is well
		respondWithJson(w, http.StatusCreated, chirp)
	}
}

// Gets all chirps in the database
func (cfg *apiConfig) getChirpsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		chirps, err := cfg.dbQueries.GetChirps(r.Context())
		if err != nil {
			errBody := fmt.Sprintf("Problem getting chirps: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Return chirps as JSON
		respond(w, http.StatusOK, chirps)
	}
}

// Get a single chirp from the database
func (cfg *apiConfig) getChirpHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the chirp id from the request URL
		idstr := r.PathValue("chirpID")
		id, err := uuid.Parse(idstr)
		if err != nil {
			errBody := fmt.Sprintf("Error parsing uuid: %v", err)
			respondWithError(w, http.StatusBadRequest, errBody)
			return
		}

		// Get the chrip from the db
		chirp, err := cfg.dbQueries.GetChirp(r.Context(), id)
		if err != nil {
			errBody := fmt.Sprintf("Problem getting chirp: %v", err)
			respondWithError(w, http.StatusNotFound, errBody)
			return
		}
		respond(w, http.StatusOK, chirp)
	}
}

// Login handler
func (cfg *apiConfig) loginHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode the request
		req := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}
		err := decodeStruct(r, &req)
		if err != nil {
			errBody := fmt.Sprintf("Decoding Error: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Find the user in the db with the Email
		user, err := cfg.dbQueries.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			errBody := fmt.Sprintf("Error getting user: %v", err)
			respondWithError(w, http.StatusBadRequest, errBody)
			return
		}

		// Check if passwords match
		valid, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
		if err != nil {
			errBody := fmt.Sprintf("Error checking password: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Behave according to validity
		if valid == false {
			respond(w, http.StatusUnauthorized, "Incorrect email or password")
		} else {
			respondWithJson(w, http.StatusOK, user)
		}

	}
}
