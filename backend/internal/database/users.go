package database

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
}

type UserParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
func (c *Client) CountUsers() (int, error) {

	var count int
	err := c.handler.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (c *Client) AddUser(params UserParams) (User, string, error) {

	_, err := c.handler.Exec(
		"INSERT INTO users (id, username, password_hash, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		uuid.New(),
		params.Username,
		params.Password,
	)
	if err != nil {
		return User{}, "", err
	}

	return c.GetUserWithUsername(params.Username)
}

func (c *Client) UpdatePassword(id uuid.UUID, password string) (User, string, error) {

	_, err := c.handler.Exec("UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", password, id)
	if err != nil {
		return User{}, "", err
	}

	return c.GetUser(id)
}

func (c *Client) GetUser(id uuid.UUID) (User, string, error) {

	var user User
	var password string
	err := c.handler.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user.Id, &user.Username, &password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, "", err
	}
	return user, password, nil

}

func (c *Client) GetUserWithUsername(username string) (User, string, error) {

	var user User
	var password string
	err := c.handler.QueryRow("SELECT * FROM users WHERE username = ?", username).Scan(&user.Id, &user.Username, &password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, "", err
	}
	return user, password, nil
}

func (c *Client) DeleteUser(id uuid.UUID) (int, error) {

	_, err := c.handler.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return -1, err
	}

	return c.CountUsers()
}

func (c *Client) RemoveAllUsers() error {

	_, err := c.handler.Exec("DELETE FROM users")
	if err != nil {
		return err
	}

	_, err = c.handler.Exec("DELETE FROM refresh_tokens")
	if err != nil {
		return err
	}

	return nil
}

// #region Refresh Tokens

func (c *Client) AddRefreshToken(userId uuid.UUID, token string, expiresAt time.Time) (RefreshTokenInfo, error) {

	_, err := c.handler.Exec(`
	INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
	SELECT ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, users.id, ?, NULL
	FROM users WHERE users.id = ?
	`, token, expiresAt, userId)

	if err != nil {
		return RefreshTokenInfo{}, err
	}

	return c.GetRefreshToken(token)
}

func (c *Client) GetRefreshToken(token string) (RefreshTokenInfo, error) {

	var info RefreshTokenInfo
	err := c.handler.QueryRow("SELECT * FROM refresh_tokens WHERE token = ?", token).Scan(
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

func (c *Client) RevokeRefreshToken(token string) error {

	_, err := c.handler.Exec("UPDATE refresh_tokens SET revoked_at = ? WHERE token = ?", time.Now().UTC(), token)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteRefreshToken(token string) error {

	_, err := c.handler.Exec("DELETE FROM refresh_tokens WHERE token = ?", token)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CullRefreshTokens(cutoff time.Time) error {

	rows, err := c.handler.Query("SELECT token, expires_at, revoked_at FROM refresh_tokens")
	if err != nil {
		return err
	}

	var id string
	var expiry time.Time
	var revokation time.Time
	toRemove := []string{}
	for rows.Next() {
		err = rows.Scan(&id, &expiry, &revokation)

		if expiry.Before(cutoff) || revokation.Before(cutoff) {
			toRemove = append(toRemove, "'"+id+"'")
		}
	}

	if len(toRemove) > 0 {
		_, err = c.handler.Exec("DELETE FROM refresh_tokens WHERE token IN (" + strings.Join(toRemove, ",") + ")")
		if err != nil {
			return err
		}
	}

	return nil
}
