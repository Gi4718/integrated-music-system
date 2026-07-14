package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

type SystemUser struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	LastLoginAt  time.Time
	FailedAttempts int
	LockedUntil  time.Time
}

// CreateSystemUser 创建系统用户
func CreateSystemUser(username, password string) error {
	// 检查是否已有用户
	var count int
	err := dbConn.QueryRow("SELECT COUNT(*) FROM system_users").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("系统已存在管理员账号")
	}

	// 密码加密
	hash := hashPassword(password)

	_, err = dbConn.Exec(
		"INSERT INTO system_users (username, password_hash, created_at, failed_attempts) VALUES (?, ?, ?, 0)",
		username, hash, time.Now(),
	)
	return err
}

// AuthenticateSystemUser 验证系统用户
func AuthenticateSystemUser(username, password string) (*SystemUser, error) {
	var user SystemUser
	var lastLoginAt sql.NullTime
	var lockedUntil sql.NullTime

	err := dbConn.QueryRow(
		"SELECT id, username, password_hash, created_at, last_login_at, failed_attempts, locked_until FROM system_users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &lastLoginAt, &user.FailedAttempts, &lockedUntil)
	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = lockedUntil.Time
	}

	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查是否被锁定
	if lockedUntil.Valid && time.Now().Before(lockedUntil.Time) {
		return nil, errors.New("账号已被锁定，请稍后再试")
	}

	// 验证密码
	hash := hashPassword(password)
	if hash != user.PasswordHash {
		// 增加失败次数
		user.FailedAttempts++
		var lockedUntil time.Time
		if user.FailedAttempts >= 5 {
			lockedUntil = time.Now().Add(15 * time.Minute)
		}

		lockedUntilStr := ""
		if !lockedUntil.IsZero() {
			lockedUntilStr = lockedUntil.Format(time.RFC3339)
		}

		dbConn.Exec(
			"UPDATE system_users SET failed_attempts = ?, locked_until = ? WHERE id = ?",
			user.FailedAttempts, lockedUntilStr, user.ID,
		)

		if user.FailedAttempts >= 5 {
			return nil, errors.New("密码错误次数过多，账号已被锁定15分钟")
		}
		return nil, errors.New("用户名或密码错误")
	}

	// 登录成功，重置失败次数
	dbConn.Exec(
		"UPDATE system_users SET failed_attempts = 0, locked_until = NULL, last_login_at = ? WHERE id = ?",
		time.Now(), user.ID,
	)

	return &user, nil
}

// HasSystemUser 检查是否已有系统用户
func HasSystemUser() (bool, error) {
	var count int
	err := dbConn.QueryRow("SELECT COUNT(*) FROM system_users").Scan(&count)
	return count > 0, err
}

// hashPassword 使用SHA256加密密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
