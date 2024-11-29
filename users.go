package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/romusking/chirpy/internal/auth"
	"github.com/romusking/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't create user, invalid input.", err)
		return
	}
	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Cant use password.", err)
		return
	}

	userDB, err := cfg.db.CreateUser(
		r.Context(),
		database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashed})

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

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password.", err)
		return
	}
	userDB, err := cfg.db.GetUserPassword(r.Context(), params.Email)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password.", err)
		return
	}
	err = auth.CheckPasswordHash(params.Password, userDB.HashedPassword)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password.", err)
		return
	}
	const defaultExpiration int = 3600
	if params.ExpiresInSeconds > defaultExpiration || params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = defaultExpiration
	}
	fmt.Println(time.Duration(params.ExpiresInSeconds * 1000000000))
	token, err := auth.MakeJWT(
		userDB.ID,
		cfg.secret,
		time.Duration(params.ExpiresInSeconds*1000000000))

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't create token.", err)
		return
	}

	user := User{
		ID:        userDB.ID,
		CreatedAt: userDB.CreatedAt,
		UpdatedAt: userDB.UpdatedAt,
		Email:     userDB.Email,
		Token:     token,
	}
	restpondWithJSON(w, http.StatusOK, user)

}
