package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
)

const (
	GenericError string = "Something went wrong"

	BodyDecodeError string = "Failed to read request body"
	DatabaseError   string = "Something went wrong with the database"
	FileMoveError   string = "Something went wrong while moving files"
	FileDeleteError string = "Something went wrong while deleting files"
	CoverURLError   string = "Failed to fetch the cover from the url. Only png and jpg are currently supported"
	NotFoundError   string = "Not Found"

	MetadataFetchError    string = "Something went wrong querying source api"
	MetadataApiKeyMissing string = "Api key for source not set"
	MetadataSourceError   string = "Metadata source missing or invalid. Sources are Open Library, Google Books and Audible. Audible requires a specified region"

	AuthBadAuthorization string = "Invalid Authorization"
	AuthExpired          string = "Authorization Expired"
)

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {

	if err != nil {
		log.Println(err)
		_, file, line, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("Called from %s at line %d\n", file, line)
			fmt.Println()
		}
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJson(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

const SecureCookies bool = false

func addJWTCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})
}

func addRefreshCookie(w http.ResponseWriter, refreshToken string, expirationTime int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/auth/refresh", // Optional: Only send to the refresh endpoint
		HttpOnly: true,
		Secure:   SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   expirationTime,
	})

}
