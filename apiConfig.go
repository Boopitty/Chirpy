// This file stores logic for the apiConfig struct.
package main

import (
	"database/sql"
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
			Body string `json:"body"`
		}{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			errBody := fmt.Sprintf("Decoding Error: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Get the token from the header
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			errBody := fmt.Sprintf("Error getting bearer token: %v", err)
			respondWithError(w, http.StatusUnauthorized, errBody)
			return
		}

		// Validate the token and get the user id from the token
		parsedId, err := auth.ValidateJWT(token, cfg.secret)
		if err != nil {
			errBody := fmt.Sprintf("Error validating JWT: %v", err)
			respondWithError(w, http.StatusUnauthorized, errBody)
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
			Body:      cleanString(req.Body),
			UserID:    parsedId,
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

		// If the password is correct, make a JWT for the user
		accessToken, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(time.Duration(3600)*time.Second))
		if err != nil {
			errBody := fmt.Sprintf("Error making JWT: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Make a refresh token for the user and add it to the database
		refreshToken := auth.MakeRefreshToken()
		if refreshToken == "" {
			errBody := "Error making refresh token: got empty string"
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}
		_, err = cfg.dbQueries.CreateToken(r.Context(), database.CreateTokenParams{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Duration(1440) * time.Hour),
		})
		if err != nil {
			errBody := fmt.Sprintf("Error creating refresh token in database: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Response Structure
		resp := struct {
			Id           uuid.UUID `json:"id"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			Email        string    `json:"email"`
			Token        string    `json:"token"`
			RefreshToken string    `json:"refresh_token"`
		}{
			Id:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        accessToken,
			RefreshToken: refreshToken,
		}

		// Behave according to validity
		if valid == false {
			respond(w, http.StatusUnauthorized, "Incorrect email or password")
		} else {
			respondWithJson(w, http.StatusOK, resp)
		}

	}
}

// Handler for the refresh endpoint.
// This will give a new access token if the refresh token in the header is valid.
func (cfg *apiConfig) refreshHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// This endpoint doesn't accept a request body,
		// but it does require a refresh token in the header
		dbToken, err := cfg.validateRefreshToken(w, r)
		if err != nil {
			return
		}

		// If the token is expired or revoked, give an error
		if dbToken.ExpiresAt.Before(time.Now()) {
			errBody := "Refresh token has expired"
			respondWithError(w, http.StatusUnauthorized, errBody)
			return
		}
		if dbToken.RevokedAt.Valid && dbToken.RevokedAt.Time.Before(time.Now()) {
			errBody := "Refresh token has been revoked"
			respondWithError(w, http.StatusUnauthorized, errBody)
			return
		}

		// If the refresh token is valid, make a new access token and update the database
		accessToken, err := auth.MakeJWT(dbToken.UserID, cfg.secret, time.Duration(time.Duration(3600)*time.Second))
		if err != nil {
			errBody := fmt.Sprintf("Error making JWT: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// If the token is valid, make a new access token for the user
		resp := struct {
			Token string `json:"token"`
		}{
			Token: accessToken,
		}
		respondWithJson(w, http.StatusOK, resp)
	}
}

// Handler for the revoke endpoint.
// This will revoke the refresh token in the header, making it invalid for future use.
func (cfg *apiConfig) revokeHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			errBody := fmt.Sprintf("Error getting bearer token: %v", err)
			respondWithError(w, http.StatusUnauthorized, errBody)
			return
		}

		_, err = cfg.dbQueries.RevokeToken(r.Context(), database.RevokeTokenParams{
			RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
			Token:     token,
		})
		if err != nil {
			errBody := fmt.Sprintf("Error revoking token in database: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for the update user endpoint.
// This will update the user's email and password in the database if the token is valid.
func (cfg *apiConfig) updateUserHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate Access Token
		userID, err := cfg.validateAccessToken(w, r)
		if err != nil {
			return
		}

		// Decode the request
		req := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}
		err = decodeStruct(r, &req)
		if err != nil {
			errBody := fmt.Sprintf("Decoding Error: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Hash the password from the request
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			errBody := fmt.Sprintf("Error hashing password: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Update the user in the database with the new email and password
		user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
			ID:             userID,
			UpdatedAt:      time.Now(),
			Email:          req.Email,
			HashedPassword: hash,
		})
		if err != nil {
			errBody := fmt.Sprintf("Error updating user in database: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Respond with the updated user info
		respondWithJson(w, http.StatusOK, user)
	}
}

// Returns a user ID if the refresh token is valid, and gives an error response if not.
func (cfg *apiConfig) validateAccessToken(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errBody := fmt.Sprintf("Error getting bearer token: %v", err)
		respondWithError(w, http.StatusUnauthorized, errBody)
		return uuid.Nil, err
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		errBody := fmt.Sprintf("Error validating JWT: %v", err)
		respondWithError(w, http.StatusUnauthorized, errBody)
		return uuid.Nil, err
	}

	return userId, nil
}

// Returns a refresh token if the token in the header's Access Token is valid, and gives an error response if not.
func (cfg *apiConfig) validateRefreshToken(w http.ResponseWriter, r *http.Request) (database.RefreshToken, error) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errBody := fmt.Sprintf("Error getting bearer token: %v", err)
		respondWithError(w, http.StatusInternalServerError, errBody)
		return database.RefreshToken{}, err
	}

	dbToken, err := cfg.dbQueries.GetToken(r.Context(), token)
	if err != nil {
		errBody := fmt.Sprintf("Error validating JWT: %v", err)
		respondWithError(w, http.StatusUnauthorized, errBody)
		return database.RefreshToken{}, err
	}

	return dbToken, nil
}
