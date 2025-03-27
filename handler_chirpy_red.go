package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/iamjoona/chippy/internal/auth"
)

func (cfg *apiConfig) userUpgradeHandler(w http.ResponseWriter, r *http.Request) {

	// read auth header and compare to polkaApiKey
	reqApiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	if reqApiKey != cfg.polkaApiKey {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	var upgradeRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	err = json.NewDecoder(r.Body).Decode(&upgradeRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if upgradeRequest.Event != "user.upgraded" {
		http.Error(w, "Invalid event type", http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(upgradeRequest.Data.UserID)
	if err != nil {
		http.Error(w, "User can't be found", http.StatusNotFound)
		return
	}

	_, err = cfg.db.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		http.Error(w, "Error upgrading user", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}
