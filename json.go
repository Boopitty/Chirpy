package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	// Annonymous struct for storing the request.
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

	// If length of request is too long, give an error
	if len(req.Body) > 140 {
		type errResp struct {
			Error string `json:"error"`
		}
		respondWithJson(w, http.StatusBadRequest, errResp{Error: "Chirp is too long"})
		return
	}

	// Encode request struct if all is well
	type validResp struct {
		Valid bool `json:"valid"`
	}
	respondWithJson(w, http.StatusOK, validResp{Valid: true})
}
