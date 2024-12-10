package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/romusking/chirpy/internal/auth"
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

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no auth token in request", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token in request", err)
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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
			UserID: userID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp, database error.", err)
		return
	}
	chirp := chirpDBToChirpJSON(chirpDB)
	respondWithJSON(w, 201, chirp)

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

func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author_id")
	sort := r.URL.Query().Get("sort")

	var chirpsInDB []database.Chirp
	var err error
	if author == "" {
		chirpsInDB, err = cfg.db.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps, database error.", err)
			return
		}
	} else {
		authorID, err := uuid.Parse(author)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Couldn't get chirps, autor unknown.", err)
			return
		}
		chirpsInDB, err = cfg.db.GetAllChirpsByAuthor(r.Context(), authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps, database error.", err)
			return
		}
	}
	chirps := make([]Chirp, len(chirpsInDB))
	for i, chirpDB := range chirpsInDB {
		chirps[i] = chirpDBToChirpJSON(chirpDB)

	}
	if sort == "desc" {
		slices.Reverse(chirps)
	}
	respondWithJSON(w, 200, chirps)

}

func chirpDBToChirpJSON(chirpDB database.Chirp) Chirp {
	chirp := Chirp{
		ID:        chirpDB.ID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
		Body:      chirpDB.Body,
		UserID:    chirpDB.UserID,
	}
	return chirp
}

func (cfg *apiConfig) getOneChirp(w http.ResponseWriter, r *http.Request) {

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get chirp, wrong UUID.", err)
		return
	}

	chirpInDB, err := cfg.db.GetOneChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp doesn't exist.", err)
		return
	}

	chirp := chirpDBToChirpJSON(chirpInDB)

	respondWithJSON(w, 200, chirp)
}

func (cfg *apiConfig) deleteAChirp(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no auth token in request", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token in request", err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't delete chirp, wrong UUID.", err)
		return
	}

	chirpInDB, err := cfg.db.GetOneChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp doesn't exist.", err)
		return
	}

	if userID != chirpInDB.UserID {
		respondWithError(w, http.StatusForbidden, "Not yours, can't delete", err)
		return
	}

	err = cfg.db.DeleteAChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't delete chir.", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
