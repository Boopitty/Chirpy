package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Boopitty/Chirpy/internal/auth"
	"github.com/Boopitty/Chirpy/internal/database"
	"github.com/google/uuid"
)

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
			IsChirpyRed  bool      `json:"is_chirpy_red"`
		}{
			Id:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        accessToken,
			RefreshToken: refreshToken,
			IsChirpyRed:  user.IsChirpyRed,
		}

		// Behave according to validity
		if valid == false {
			respond(w, http.StatusUnauthorized, "Incorrect email or password")
		} else {
			respondWithJson(w, http.StatusOK, resp)
		}

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

func (cfg *apiConfig) polkaWebhookHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode the request
		req := struct {
			Event string `json:"event"`
			Data  struct {
				UserID string `json:"user_id"`
			} `json:"data"`
		}{}
		err := decodeStruct(r, &req)
		if err != nil {
			errBody := fmt.Sprintf("Decoding Error: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// If the event is anything other than "user.upgraded",
		// return a 204 status code with no body.
		if req.Event != "user.upgraded" {
			respond(w, http.StatusNoContent, nil)
			return
		}

		// Upgrade the user with the given user ID in the database.
		err = cfg.dbQueries.MakeUserRed(r.Context(), uuid.MustParse(req.Data.UserID))
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				respond(w, http.StatusNotFound, "User not found")
				return
			}
			errBody := fmt.Sprintf("Error upgrading user in database: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		respond(w, http.StatusNoContent, nil)
	}
}
