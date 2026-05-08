package database

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserParams
}

type UserParams struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

type RefreshTokenInfo struct {
	Token     string
	UserId    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// #region Users

// Requires an active db transaction
func (c Client) CountUsers() (int, error) {

	if c.tx == nil {
		return 0, errors.New("count users requires an active transaction")
	}

	var count int
	err := c.tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (c Client) AddUser(params UserParams) (User, error) {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return User{}, err
		}
		defer c.Rollback()
	}

	_, err := c.db.Exec(
		"INSERT INTO users (id, username, password_hash, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		uuid.New(),
		params.Username,
		params.Password,
	)
	if err != nil {
		return User{}, err
	}

	if indyTx {
		err = c.Commit()
		if err != nil {
			return User{}, err
		}
	}

	return c.GetUserWithUsername(params.Username)
}

func (c Client) UpdatePassword(id uuid.UUID, password string) (User, error) {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return User{}, err
		}
		defer c.Rollback()
	}

	_, err := c.tx.Exec("UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", password, id)
	if err != nil {
		return User{}, err
	}

	if indyTx {
		err = c.Commit()
		if err != nil {
			return User{}, err
		}
	}

	return c.GetUser(id)
}

func (c Client) GetUser(id uuid.UUID) (User, error) {

	var user User
	err := c.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user.Id, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil

}

func (c Client) GetUserWithUsername(username string) (User, error) {

	var user User
	err := c.db.QueryRow("SELECT * FROM users WHERE username = ?", username).Scan(&user.Id, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// #region Refresh Tokens

func (c Client) AddRefreshToken(userId uuid.UUID, token string, expiresAt time.Time) (RefreshTokenInfo, error) {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return RefreshTokenInfo{}, err
		}
		defer c.Rollback()
	}

	_, err := c.tx.Exec(`
	INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
	SELECT ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, users.id, ?, NULL
	FROM users WHERE users.id = ?
	`, token, expiresAt, userId)

	if err != nil {
		return RefreshTokenInfo{}, err
	}

	if indyTx {
		err = c.Commit()
		if err != nil {
			return RefreshTokenInfo{}, err
		}
	}

	return c.GetRefreshToken(token)
}

func (c Client) GetRefreshToken(token string) (RefreshTokenInfo, error) {

	var info RefreshTokenInfo
	err := c.db.QueryRow("SELECT * FROM refresh_tokens WHERE token = ?", token).Scan(
		&info.Token,
		&info.CreatedAt,
		&info.UpdatedAt,
		&info.UserId,
		&info.ExpiresAt,
		&info.RevokedAt,
	)
	if err != nil {
		return RefreshTokenInfo{}, err
	}
	return info, nil
}
