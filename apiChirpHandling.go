// This file stores logic for the apiConfig struct.
package main

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/Boopitty/Chirpy/internal/database"
	"github.com/google/uuid"
)

// Creates a chirp
func (cfg *apiConfig) createChirpHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode the request
		req := struct {
			Body string `json:"body"`
		}{}

		err := decodeStruct(r, &req)
		if err != nil {
			errBody := fmt.Sprintf("Decoding Error: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
			return
		}

		// Get the token from the header
		userID, err := cfg.validateAccessToken(w, r)
		if err != nil {
			errBody := fmt.Sprintf("Error getting bearer token: %v", err)
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
			UserID:    userID,
		})
		if err != nil {
			errBody := fmt.Sprintf("Problem creating chirp: %v", err)
			respondWithError(w, http.StatusInternalServerError, errBody)
		}

		// Respond with chirp if all is well
		respondWithJson(w, http.StatusCreated, chirp)
	}
}

// Gets all chirps in the database. If an "author_id" query parameter is provided, it filters chirps by author.
func (cfg *apiConfig) getChirpsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Accept an optional "author_id" query parameter to filter chirps by author
		authorIDStr := r.URL.Query().Get("author_id")
		var chirps []database.Chirp
		var err error

		if authorIDStr != "" {
			// Parse author ID as a UUID
			authorID, err := uuid.Parse(authorIDStr)
			if err != nil {
				errBody := fmt.Sprintf("Error parsing uuid: %v", err)
				respondWithError(w, http.StatusBadRequest, errBody)
				return
			}

			// Gets chirps by author from the db
			chirps, err = cfg.dbQueries.GetChirpsByAuthor(r.Context(), authorID)
			if err != nil {
				errBody := fmt.Sprintf("Problem getting chirps by author: %v", err)
				respondWithError(w, http.StatusInternalServerError, errBody)
				return
			}
			respond(w, http.StatusOK, chirps)
			return
		} else {
			// Get all chirps from the db
			chirps, err = cfg.dbQueries.GetChirps(r.Context())
			if err != nil {
				errBody := fmt.Sprintf("Problem getting chirps: %v", err)
				respondWithError(w, http.StatusInternalServerError, errBody)
				return
			}
		}

		// Accept an optional "sort" query parameter to sort chirps by creation time
		// The list is sorted in ascending order by default.
		sortOrder := r.URL.Query().Get("sort")
		if sortOrder == "desc" {
			sort.Slice(chirps, func(i, j int) bool {
				return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
			})
		}
		// Return chirps as JSON
		respond(w, http.StatusOK, chirps)
	}
}

// Get a single chirp from the database
func (cfg *apiConfig) getChirpHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the chirp id from the request URL
		idString := r.PathValue("chirpID")
		id, err := uuid.Parse(idString)
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

// Deletes a chirp from the database
func (cfg *apiConfig) deleteChirpHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate the access token and get the user id from it
		userID, err := cfg.validateAccessToken(w, r)
		if err != nil {
			// Error response is handled in the validateRefreshToken function
			return
		}

		// get the chirp id from the request URL
		idString := r.PathValue("chirpID")
		id, err := uuid.Parse(idString)
		if err != nil {
			errBody := fmt.Sprintf("Error parsing uuid: %v", err)
			respondWithError(w, http.StatusBadRequest, errBody)
			return
		}

		// Get the chirp from the db
		chirp, err := cfg.dbQueries.GetChirp(r.Context(), id)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				errBody := fmt.Sprintf("Chirp with id %s not found", idString)
				respondWithError(w, http.StatusNotFound, errBody)
				return
			}
			errBody := fmt.Sprintf("Problem getting chirp: %v", err)
			respondWithError(w, http.StatusNotFound, errBody)
			return
		}

		// Only the author of the chirp can delete it.
		if chirp.UserID != userID {
			errBody := "User is not the author of the chirp"
			respondWithError(w, http.StatusForbidden, errBody)
			return

		} else {
			// Delete the chirp from the db
			err = cfg.dbQueries.DeleteChirp(r.Context(), id)
			if err != nil {
				errBody := fmt.Sprintf("Problem deleting chirp: %v", err)
				respondWithError(w, http.StatusInternalServerError, errBody)
				return
			}
		}

		respond(w, http.StatusNoContent, nil)
	}
}
