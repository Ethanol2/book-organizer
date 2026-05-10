package database

import (
	"errors"
	"strings"
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

	_, err := c.tx.Exec(
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

func (c Client) DeleteUser(id uuid.UUID) (int, error) {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return -1, err
		}
		defer c.Rollback()
	}

	_, err := c.tx.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return -1, err
	}

	if indyTx {
		err = c.Commit()
		if err != nil {
			return -1, err
		}
	}

	return c.CountUsers()
}

func (c Client) RemoveAllUsers() error {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return err
		}
		defer c.Rollback()
	}

	_, err := c.tx.Exec("DELETE FROM users")
	if err != nil {
		return err
	}

	_, err = c.tx.Exec("DELETE FROM refresh_tokens")
	if err != nil {
		return err
	}

	if indyTx {
		err = c.Commit()
		if err != nil {
			return err
		}
	}
	return nil
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

func (c Client) DeleteRefreshToken(token string) error {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return err
		}
		defer c.Rollback()
	}

	_, err := c.tx.Exec("DELETE FROM request_tokens WHERE token = ?", token)
	if err != nil {
		return err
	}

	if indyTx {
		err = c.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Client) CullRefreshTokens(cutoff time.Time) error {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return err
		}
		defer c.Rollback()
	}

	rows, err := c.tx.Query("SELECT token, expires_at FROM refresh_tokens")
	if err != nil {
		return err
	}

	var id string
	var expiry time.Time
	toRemove := []string{}
	for rows.Next() {
		err = rows.Scan(&id, &expiry)

		if expiry.Before(cutoff) {
			toRemove = append(toRemove, "'"+id+"'")
		}
	}

	if len(toRemove) > 0 {
		_, err = c.tx.Exec("DELETE FROM refresh_tokens WHERE token IN (" + strings.Join(toRemove, ",") + ")")
		if err != nil {
			return err
		}
	}

	if indyTx {
		err = c.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}
