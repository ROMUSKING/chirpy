package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/romusking/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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

	chirpDB, err := cfg.db.CreateChirp(
		r.Context(), database.CreateChirpParams{
			Body:   params.Body,
			UserID: params.UserID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp, database error.", err)
		return
	}
	chirp := dBToChirpJSON(chirpDB)
	restpondWithJSON(w, 201, chirp)

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

func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {

	chirpsInDB, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirp, database error.", err)
		return
	}
	chirps := make([]Chirp, len(chirpsInDB))
	for i, chirpDB := range chirpsInDB {
		chirps[i] = dBToChirpJSON(chirpDB)

	}
	restpondWithJSON(w, 200, chirps)
}

func dBToChirpJSON(chirpDB database.Chirp) Chirp {
	chirp := Chirp{
		ID:        chirpDB.ID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
		Body:      chirpDB.Body,
		UserID:    chirpDB.UserID,
	}
	return chirp
}
