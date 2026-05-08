package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Ethanol2/book-organizer/internal/auth"
	"github.com/Ethanol2/book-organizer/internal/database"
)

func (cfg *apiConfig) handlerGetUsersCount(w http.ResponseWriter, r *http.Request) {

	err := cfg.db.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}
	defer cfg.db.Rollback()

	count, err := cfg.db.CountUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	err = cfg.db.Commit()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	respondWithJson(w, http.StatusOK, struct {
		Count int `json:"count"`
	}{Count: count})
}

func (cfg *apiConfig) handlerRegister(w http.ResponseWriter, r *http.Request) {

	err := cfg.db.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}
	defer cfg.db.Rollback()

	count, err := cfg.db.CountUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	// If there aren't any users, assume the first user is the admin
	if count > 0 {
		_, err := authorize(true, r, cfg.tokenSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
			return
		}
	} else {
		log.Println("REGISTERING FIRST USER WITHOUT AUTH. Reset the backend (-r) immediately if unexpected")
	}

	var params database.UserParams
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
		return
	}

	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, GenericError, err)
		return
	}

	user, err := cfg.db.AddUser(params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	err = cfg.db.Commit()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	user.Password = ""
	respondWithJson(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	var login database.UserParams

	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
	}

	user, err := cfg.db.GetUserWithUsername(login.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Username not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if ok, err := auth.CheckPassword(login.Password, user.Password); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Authentication error", err)
		return
	} else if !ok {
		respondWithError(w, http.StatusUnauthorized, "Incorrect password", nil)
		return
	}

	token, err := auth.MakeJWT(user.Id, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create token", err)
		return
	}

	refresh := auth.MakeRefreshToken()
	refreshLifetime := time.Now().UTC().Add(time.Hour * 24 * 60)
	_, err = cfg.db.AddRefreshToken(
		user.Id,
		refresh,
		refreshLifetime,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	addJWTCookie(w, token)
	addRefreshCookie(w, refresh, int(time.Until(refreshLifetime).Seconds()))

	user.Password = ""
	respondWithJson(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refresh, err := auth.GetBearerToken(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
		return
	}

	info, err := cfg.db.GetRefreshToken(refresh)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if time.Now().UTC().After(info.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, AuthExpired, nil)
		return
	}

	if info.RevokedAt != nil {
		respondWithError(w, http.StatusUnauthorized, AuthExpired, nil)
		return
	}

	user, err := cfg.db.GetUser(info.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	token, err := auth.MakeJWT(user.Id, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, GenericError, err)
		return
	}

	addJWTCookie(w, token)
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerUpdatePassword(w http.ResponseWriter, r *http.Request) {

	userId, err := authorize(true, r, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
		return
	}

	var params database.UserParams
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, GenericError, err)
		return
	}

	user, err := cfg.db.UpdatePassword(userId, hash)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}
	user.Password = ""

	respondWithJson(w, http.StatusOK, user)
}
