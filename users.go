package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/romusking/chirpy/internal/auth"
	"github.com/romusking/chirpy/internal/database"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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
		Email    string `json:"email"`
		Password string `json:"password"`
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
	const hour int = 3600 * 1000000000
	const exp int = hour * 24 * 60

	token, err := auth.MakeJWT(
		userDB.ID,
		cfg.secret,
		time.Duration(hour))

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't create token.", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't create token.", err)
		return
	}

	_, err = cfg.db.CreateRefToken(r.Context(),
		database.CreateRefTokenParams{
			Token:     refreshToken,
			UserID:    userDB.ID,
			ExpiresAt: time.Now().Add(time.Duration(exp))})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't create token.", err)
		return
	}

	user := User{
		ID:           userDB.ID,
		CreatedAt:    userDB.CreatedAt,
		UpdatedAt:    userDB.UpdatedAt,
		Email:        userDB.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}
	restpondWithJSON(w, http.StatusOK, user)

}

func (cfg *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid token", err)
		return
	}

	tokenDB, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)

	if err != nil || tokenDB.RevokedAt.Valid || time.Now().After(tokenDB.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "invalid token", err)
		return
	}

	const hour int = 3600 * 1000000000

	token, err := auth.MakeJWT(
		tokenDB.UserID,
		cfg.secret,
		time.Duration(hour))

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't create token.", err)
		return
	}

	params := parameters{Token: token}

	restpondWithJSON(w, http.StatusOK, params)

}

func (cfg *apiConfig) revokeRefreshToken(w http.ResponseWriter, r *http.Request) {

	refreshToken, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid token", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't revoke token", err)
		return
	}

	restpondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {

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
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't update user, invalid email or password.", err)
		return
	}

	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Cant use password.", err)
		return
	}

	userDB, err := cfg.db.UpdateUserDetails(
		r.Context(),
		database.UpdateUserDetailsParams{
			ID:             userID,
			Email:          params.Email,
			HashedPassword: hashed})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user, database error.", err)
		return
	}
	user := User{
		ID:        userDB.ID,
		CreatedAt: userDB.CreatedAt,
		UpdatedAt: userDB.UpdatedAt,
		Email:     userDB.Email,
	}
	restpondWithJSON(w, http.StatusOK, user)

}
