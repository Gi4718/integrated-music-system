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

// GetSystemUserByID 根据 ID 获取系统用户
func GetSystemUserByID(id int) (*SystemUser, error) {
	var user SystemUser
	var lastLoginAt sql.NullTime
	var lockedUntil sql.NullTime

	err := dbConn.QueryRow(
		"SELECT id, username, password_hash, role, created_at, last_login_at, failed_attempts, locked_until FROM system_users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &lastLoginAt, &user.FailedAttempts, &lockedUntil)

	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = lockedUntil.Time
	}

	return &user, nil
}

// GetAllSystemUsers 获取所有系统用户
func GetAllSystemUsers() ([]SystemUser, error) {
	rows, err := dbConn.Query("SELECT id, username, password_hash, role, created_at, last_login_at, failed_attempts, locked_until FROM system_users ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []SystemUser
	for rows.Next() {
		var user SystemUser
		var lastLoginAt sql.NullTime
		var lockedUntil sql.NullTime

		err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &lastLoginAt, &user.FailedAttempts, &lockedUntil)
		if err != nil {
			return nil, err
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = lastLoginAt.Time
		}
		if lockedUntil.Valid {
			user.LockedUntil = lockedUntil.Time
		}

		users = append(users, user)
	}

	return users, nil
}

// UpdateUserRole 更新用户角色
func UpdateUserRole(userID int, role string) error {
	_, err := dbConn.Exec("UPDATE system_users SET role = ? WHERE id = ?", role, userID)
	return err
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(userID int, password string) error {
	hash := hashPassword(password)
	_, err := dbConn.Exec("UPDATE system_users SET password_hash = ? WHERE id = ?", hash, userID)
	return err
}

// DeleteSystemUser 删除系统用户
func DeleteSystemUser(userID int) error {
	// 删除用户关联的网易云账号
	dbConn.Exec("DELETE FROM users WHERE system_user_id = ?", userID)
	// 删除用户关联的设置
	dbConn.Exec("DELETE FROM settings WHERE user_id = ?", userID)
	// 删除用户
	_, err := dbConn.Exec("DELETE FROM system_users WHERE id = ?", userID)
	return err
}

// UsernameExists 检查用户名是否已存在
func UsernameExists(username string) bool {
	var count int
	err := dbConn.QueryRow("SELECT COUNT(*) FROM system_users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// CreateSystemUserWithRole 创建指定角色的系统用户
func CreateSystemUserWithRole(username, password, role string) error {
	hash := hashPassword(password)
	_, err := dbConn.Exec(
		"INSERT INTO system_users (username, password_hash, role, created_at, failed_attempts) VALUES (?, ?, ?, ?, 0)",
		username, hash, role, time.Now(),
	)
	return err
}

// hashPassword 使用SHA256加密密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
