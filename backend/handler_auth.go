package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Ethanol2/book-organizer/internal/auth"
	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) createRefreshToken(userId uuid.UUID) (string, int, error) {
	refresh := auth.MakeRefreshToken()
	// Lifetime is 60 days
	refreshLifetime := time.Now().UTC().Add(time.Hour * 24 * 60)
	_, err := cfg.db.AddRefreshToken(
		userId,
		refresh,
		refreshLifetime,
	)
	if err != nil {
		return "", -1, err
	}
	return refresh, int(refreshLifetime.Second()), nil
}

func (cfg *apiConfig) handlerGetAuthStatus(w http.ResponseWriter, r *http.Request) {

	var err error
	count := -1

	err = cfg.db.HandleTransaction(func(c *database.Client) error {
		count, err = c.CountUsers()
		return err
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	var user *database.User
	if id, err := authenticate(true, r, cfg.tokenSecret); err == nil {
		usr, _, err := cfg.db.GetUser(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
			return
		}
		user = &usr
	}

	respondWithJson(w, http.StatusOK, struct {
		User  *database.User `json:"user,omitempty"`
		Count int            `json:"user_count"`
	}{Count: count, User: user})
}

func (cfg *apiConfig) handlerRegister(w http.ResponseWriter, r *http.Request) {

	count, err := cfg.db.CountUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	// If there aren't any users, assume the first user is the admin
	if count > 0 {
		_, err := authenticate(true, r, cfg.tokenSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
			return
		}
	} else {
		log.Println("THE FIRST USER HAS BEEN REGISTERED. AUTHENTICATION NOW REQUIRED")
		log.Println(`If this was not intentional or wasn't caused by you, restart the backend using -u flag. 
		This will clear all users and remove the need for authentication. Consider using anthentication to 
		protect your information.`)
	}

	var params database.UserParams
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
		return
	}

	params.Username = strings.TrimSpace(params.Username)

	if params.Username == "" {
		respondWithError(w, http.StatusBadRequest, "Username can't be empty", errors.New("username can't be empty"))
		return
	}

	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, GenericError, err)
		return
	}

	var user database.User
	err = cfg.db.HandleTransaction(func(c *database.Client) error {
		user, _, err = cfg.db.AddUser(params)
		return err
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	respondWithJson(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	var login database.UserParams

	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
	}

	user, password, err := cfg.db.GetUserWithUsername(login.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Username not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if ok, err := auth.CheckPassword(login.Password, password); err != nil {
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

	refresh, refreshLifetime, err := cfg.createRefreshToken(user.Id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	addJWTCookie(w, token)
	addRefreshCookie(w, refresh, refreshLifetime)

	respondWithJson(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerLogout(w http.ResponseWriter, r *http.Request) {

	if refresh, err := r.Cookie("refresh_token"); err == nil {
		err = cfg.db.RevokeRefreshToken(refresh.Value)
		if err != nil {
			clearAuthCookies(w)
			respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
			return
		}
	}

	clearAuthCookies(w)
	w.WriteHeader(http.StatusOK)
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

	user, _, err := cfg.db.GetUser(info.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	err = cfg.db.DeleteRefreshToken(refresh)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, nil)
		return
	}

	refresh, refreshLifetime, err := cfg.createRefreshToken(user.Id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	token, err := auth.MakeJWT(user.Id, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, GenericError, err)
		return
	}

	cfg.authRequired = true

	addJWTCookie(w, token)
	addRefreshCookie(w, refresh, refreshLifetime)
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerDeleteUser(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	count, err := cfg.db.DeleteUser(id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, NotFoundError, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if count == 0 {
		log.Println("ALL USERS HAVE BEEN DELETED. THE APP WILL NO LONGER USE AUTHENTICATION")
		cfg.authRequired = false
	}

	clearAuthCookies(w)
	respondWithJson(w, http.StatusOK, struct {
		Count int `json:"user_count"`
	}{Count: count})
}

func (cfg *apiConfig) handlerUpdatePassword(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	var params struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
		return
	}

	user, password, err := cfg.db.GetUser(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if ok, err := auth.CheckPassword(params.OldPassword, password); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Authentication error", err)
		return
	} else if !ok {
		respondWithError(w, http.StatusBadRequest, "Old password doesn't match", nil)
		return
	}

	hash, err := auth.HashPassword(params.NewPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, GenericError, err)
		return
	}

	user, _, err = cfg.db.UpdatePassword(id, hash)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	respondWithJson(w, http.StatusOK, user)
}
