package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Tests written by copilot

const testSecret = "9jyB8cNA_TN1+),+nw(1bs^snIXL[M/seDZE5*nL88," // Use a secure key in real scenarios

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "mypassword123", false},
		{"empty password", "", false}, // Argon2id handles empty strings, but check behavior
		{"long password", string(make([]byte, 1000)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash for valid input")
			}
			// Additional check: hashes should be different for different passwords
			if tt.password != "" {
				hash2, _ := HashPassword(tt.password + "diff")
				if hash == hash2 {
					t.Error("HashPassword() produced identical hashes for different passwords")
				}
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	validPassword := "mypassword123"
	hash, _ := HashPassword(validPassword) // Pre-hash for testing

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
		wantErr  bool
	}{
		{"correct password", validPassword, hash, true, false},
		{"incorrect password", "wrongpassword", hash, false, false},
		{"empty password", "", hash, false, false},
		{"invalid hash", validPassword, "invalidhash", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckPassword(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	expiresIn := time.Hour

	tests := []struct {
		name        string
		userID      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
		wantErr     bool
	}{
		{"valid JWT", userID, testSecret, expiresIn, false},
		{"empty secret", userID, "", expiresIn, true},     // Should fail due to invalid secret
		{"zero expiration", userID, testSecret, 0, false}, // JWT allows zero, but check claims
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := MakeJWT(tt.userID, tt.tokenSecret, tt.expiresIn)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("MakeJWT() returned empty token")
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	expiresIn := time.Hour
	validToken, _ := MakeJWT(userID, testSecret, expiresIn)

	expiredToken, _ := MakeJWT(userID, testSecret, -time.Hour) // Already expired

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{"valid token", validToken, testSecret, userID, false},
		{"invalid secret", validToken, "wrongsecret", uuid.Nil, true},
		{"expired token", expiredToken, testSecret, uuid.Nil, true},
		{"empty token", "", testSecret, uuid.Nil, true},
		{"malformed token", "invalid.jwt.token", testSecret, uuid.Nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

// This one was written by me
func TestBearerToken(t *testing.T) {

	validToken := http.Header{}
	validToken.Add("Authorization", "Bearer TOKEN_STRING")

	malformedToken := http.Header{}
	malformedToken.Add("Authorization", "Not Valid")

	noToken := http.Header{}

	validCookie := http.Cookie{
		Name:  "access_token",
		Value: "TOKEN_STRING",
	}
	validRequest, err := http.NewRequest("POST", "/api/library", nil)
	if err != nil {
		t.Fatal(err)
	}
	validRequest.AddCookie(&validCookie)

	invalidCookie := http.Cookie{
		Name:  "accesstoken",
		Value: "TOKEN_STRING",
	}
	invalidRequest, err := http.NewRequest("POST", "/api/library", nil)
	if err != nil {
		t.Fatal(err)
	}
	invalidRequest.AddCookie(&invalidCookie)

	tests := []struct {
		name          string
		request       http.Request
		token         string
		errorExpected bool
	}{
		{"Valid Token in Header", http.Request{Header: validToken}, "TOKEN_STRING", false},
		{"Malformed Token in Header", http.Request{Header: malformedToken}, "", true},
		{"Missing Header", http.Request{Header: noToken}, "", true},
		{"Valid Cookie", *validRequest, "TOKEN_STRING", false},
		{"Invalid Cookie", *invalidRequest, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(&tt.request)
			if (tt.errorExpected && err == nil) || (!tt.errorExpected && err != nil) {
				t.Errorf("GetBearerToken() error = %v, wantError = %v", err, tt.errorExpected)
			} else if tt.token != token {
				t.Errorf("GetBearerToken() returned the wrong token. Expected = %s, Got = %s", tt.token, token)
			}
		})
	}
}
