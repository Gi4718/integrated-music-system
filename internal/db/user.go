package db

import (
	"database/sql"
	"endfield-music/internal/model"
	"time"
)

func GetCurrentUser() (*model.User, error) {
	query := `SELECT id, user_id, nickname, avatar_url, cookie, cookie_expires, created_at, updated_at, COALESCE(system_user_id, 0)
			  FROM users ORDER BY updated_at DESC LIMIT 1`

	row := dbConn.QueryRow(query)

	var user model.User
	var systemUserID int
	err := row.Scan(
		&user.ID,
		&user.UserID,
		&user.Nickname,
		&user.AvatarURL,
		&user.Cookie,
		&user.CookieExpires,
		&user.CreatedAt,
		&user.UpdatedAt,
		&systemUserID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func SaveUser(user *model.User) error {
	query := `INSERT INTO users (user_id, nickname, avatar_url, cookie, cookie_expires, system_user_id, updated_at)
			  VALUES (?, ?, ?, ?, ?, 0, ?)
			  ON CONFLICT(user_id) DO UPDATE SET
			  nickname = excluded.nickname,
			  avatar_url = excluded.avatar_url,
			  cookie = excluded.cookie,
			  cookie_expires = excluded.cookie_expires,
			  updated_at = excluded.updated_at`

	_, err := dbConn.Exec(
		query,
		user.UserID,
		user.Nickname,
		user.AvatarURL,
		user.Cookie,
		user.CookieExpires,
		time.Now(),
	)

	return err
}

func ClearCurrentUser() error {
	_, err := dbConn.Exec(`DELETE FROM users`)
	return err
}

func IsLoggedIn() bool {
	user, err := GetCurrentUser()
	if err != nil || user == nil {
		return false
	}
	return !time.Now().After(user.CookieExpires)
}

func GetCookie() (string, error) {
	user, err := GetCurrentUser()
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", nil
	}
	return user.Cookie, nil
}
