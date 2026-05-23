package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func respond(w http.ResponseWriter, code int, payload any) {
	// payload should be a struct in json format.
	data, err := json.Marshal(payload)
	if err != nil {
		errResp := fmt.Sprintf("Marshaling Error: %v", err)
		respondWithError(w, http.StatusInternalServerError, errResp)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

// This function will accept an http.ResponseWriter and relevant info to create a response
func respondWithJson(w http.ResponseWriter, code int, payload any) {
	// payload should be a struct in json format.
	data, err := json.Marshal(payload)
	if err != nil {
		errResp := fmt.Sprintf("Marshaling Error: %v", err)
		respondWithError(w, http.StatusInternalServerError, errResp)
		return
	}

	// Format response. The order of these functions is very importaint.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(data)
}

// Creates a response to deliver and log and internal errors
func respondWithError(w http.ResponseWriter, code int, body string) {
	// Log error and write a server error response.
	log.Printf("%s", body)
	w.WriteHeader(code)
	w.Write([]byte("Interal Server Error"))
}

// Middleware to log requests in the terminal
func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func cleanString(input string) string {
	words := strings.Split(input, " ")
	forbidden := []string{"kerfuffle", "sharbert", "fornax"}

	for i, word := range words {
		for _, forbid := range forbidden {
			if strings.Contains(strings.ToLower(word), forbid) {
				words[i] = "****"
				break
			}
		}
	}
	return strings.Join(words, " ")
}

func decodeStruct(r *http.Request, req any) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(req)
	if err != nil {
		return err
	}
	return nil
}
