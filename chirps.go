package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID  `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Body      string     `json:"body"`
	UserID    *uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't create chirp, invalid message.", err)
		return
	}
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	params.Body = filterProfane(params.Body)

	userDB, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user, database error.", err)
		return
	}
	user := User{
		ID:        userDB.ID,
		CreatedAt: userDB.CreatedAt,
		UpdatedAt: userDB.UpdatedAt,
		Email:     userDB.Email,
	}
	restpondWithJSON(w, 201, user)

}

func (cfg *apiConfig) resetChirpDB(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Can't remove all chirps", nil)
	}
	err := cfg.db.DeleteAllChirps(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't delete chirps, database error.", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Chirps removed from database"))

}

func filterProfane(msg string) string {
	profanities := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	words := strings.Split(msg, " ")

	for i, word := range words {
		for _, profane := range profanities {
			if profane == strings.ToLower(word) {
				words[i] = "****"
			}
		}

	}
	return strings.Join(words, " ")

}
