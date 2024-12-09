package main

import (
	"net/http"
	"time"

	"github.com/romusking/chirpy/internal/auth"
)

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

	respondWithJSON(w, http.StatusOK, params)

}

func (cfg *apiConfig) revokeRefreshToken(w http.ResponseWriter, r *http.Request) {

	refreshToken, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid token", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't revoke token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
