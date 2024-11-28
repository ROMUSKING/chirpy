package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't create user, invalid email.", err)
		return
	}
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

func (cfg *apiConfig) resetUserDB(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Can't remove all users", nil)
	}
	err := cfg.db.DeleteAllUsers(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't delete users, database error.", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users removed from database"))

}
