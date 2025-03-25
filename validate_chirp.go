package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type validResponse struct {
		// Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := chirp{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	// check if chirp contains profanity
	body, profane := checkProfanity(params.Body)
	if profane {
		respondWithJSON(w, http.StatusOK, validResponse{
			CleanedBody: body,
		})
		return
	}

	respondWithJSON(w, http.StatusOK, validResponse{
		CleanedBody: body,
	})

}

func checkProfanity(body string) (string, bool) {
	// words to filter: kerfuffle, sharbert, fornax
	profanity := []string{"kerfuffle", "sharbert", "fornax"}
	bodyLower := strings.ToLower(body)
	replaced := body
	hasProfanity := false

	for _, word := range profanity {
		wordLower := strings.ToLower(word)
		if strings.Contains(bodyLower, wordLower) {
			index := strings.Index(bodyLower, wordLower)
			actualWord := body[index : index+len(word)]
			replaced = strings.ReplaceAll(replaced, actualWord, "****")
			hasProfanity = true
		}
	}

	return replaced, hasProfanity

}
