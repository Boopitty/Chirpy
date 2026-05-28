package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/Boopitty/Chirpy/internal/auth"
	"github.com/Boopitty/Chirpy/internal/database"
	"github.com/google/uuid"
)

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
