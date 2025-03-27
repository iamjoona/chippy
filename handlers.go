package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/iamjoona/chippy/internal/auth"
	"github.com/iamjoona/chippy/internal/database"
)

func (cfg *apiConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>
	`, cfg.fileserverHits.Load())))
}

/*
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
*/

func (cfg *apiConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		http.Error(w, "Not allowed", http.StatusForbidden)
		return
	}

	cfg.fileserverHits.Store(0)

	_, err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Error resetting database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := createUserRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(params.Email) == 0 {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if len(params.Password) == 0 {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	// hash password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ID:             uuid.New(),
	})

	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	apiUser := User{
		ID:          dbUser.ID,
		Email:       dbUser.Email,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		IsChirpyRed: false,
	}

	respondWithJSON(w, http.StatusCreated, apiUser)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := createChirpRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(params.Body) == 0 {
		http.Error(w, "Message body is required", http.StatusBadRequest)
		return
	}

	// validate chirp
	cleanedBody, profane := checkProfanity(params.Body)
	if profane {
		params.Body = cleanedBody
	}

	// check JWT
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	dbChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:      params.Body,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ID:        uuid.New(),
		UserID:    userID,
	})

	if err != nil {
		http.Error(w, "Error creating chirp", http.StatusInternalServerError)
		return
	}

	apiChirp := Chirp{
		ID:        dbChirp.ID,
		Body:      dbChirp.Body,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, apiChirp)
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")

	// If we have an author ID, get their chirps
	if authorId != "" {
		parsedAuthorId, err := uuid.Parse(authorId)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
			return
		}

		authorChirps, err := cfg.db.GetChirpsByUserID(r.Context(), parsedAuthorId)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error fetching chirps", err)
			return
		}

		formattedChirps := make([]Chirp, len(authorChirps))
		for i, chirp := range authorChirps {
			formattedChirps[i] = Chirp{
				ID:        chirp.ID,
				Body:      chirp.Body,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				UserID:    chirp.UserID,
			}
		}

		respondWithJSON(w, http.StatusOK, formattedChirps)
		return
	}

	// If no author ID, get all chirps
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching chirps", err)
		return
	}

	formattedChirps := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		formattedChirps[i] = Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserID:    chirp.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, formattedChirps)
}

func (cfg *apiConfig) getSingleChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if len(chirpID) == 0 {
		http.Error(w, "Chirp ID is required", http.StatusNotFound)
		return
	}

	// parse chirpID to uuid.UUID
	parsedChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		http.Error(w, "Invalid Chirp ID format", http.StatusBadRequest)
		return
	}

	// get chirp from db
	chirp, err := cfg.db.GetChirpByID(r.Context(), parsedChirpID)
	if err != nil {
		http.Error(w, "Error fetching chirp", http.StatusNotFound)
		return
	}

	formattedChirp := Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, formattedChirp)
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Add debug logging
	log.Printf("Debug - Auth header: %s", r.Header.Get("Authorization"))

	// Get and validate token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Debug - Token error: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Get user ID from token
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		log.Printf("Debug - JWT validation error: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Decode request body
	decoder := json.NewDecoder(r.Body)
	params := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate inputs
	if len(params.Email) == 0 || len(params.Password) == 0 {
		respondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	// Update user
	dbUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user", err)
		return
	}

	// Return updated user
	respondWithJSON(w, http.StatusOK, User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	})
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	// Get and validate token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Debug - Token error: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Get user ID from token
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		log.Printf("Debug - JWT validation error: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// get chirp ID from path
	chirpID := r.PathValue("chirpID")
	if len(chirpID) == 0 {
		http.Error(w, "Chirp ID is required", http.StatusNotFound)
		return
	}

	parsedChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		http.Error(w, "Invalid Chirp ID format", http.StatusForbidden)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), parsedChirpID)
	if err != nil {
		http.Error(w, "Error fetching chirp", http.StatusInternalServerError)
		return
	}

	if chirp.UserID != userID {
		http.Error(w, "User does not own chirp", http.StatusForbidden)
		return
	}

	// delete chirp
	err = cfg.db.DeleteChirpByID(r.Context(), parsedChirpID)
	if err != nil {
		http.Error(w, "Error deleting chirp", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
